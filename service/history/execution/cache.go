// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package execution

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/pborman/uuid"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/shard"
)

type (
	// ReleaseFunc releases workflow execution context
	ReleaseFunc func(err error)

	// Cache caches workflow execution context
	Cache struct {
		cache.Cache
		shard            shard.Context
		executionManager persistence.ExecutionManager
		disabled         bool
		logger           log.Logger
		metricsClient    metrics.Client
		config           *config.Config
	}
)

var (
	// NoopReleaseFn is an no-op implementation for the ReleaseFunc type
	NoopReleaseFn ReleaseFunc = func(err error) {}
)

const (
	cacheNotReleased int32 = 0
	cacheReleased    int32 = 1
)

// NewCache creates a new workflow execution context cache
func NewCache(shard shard.Context) *Cache {
	opts := &cache.Options{}
	config := shard.GetConfig()
	opts.InitialCapacity = config.HistoryCacheInitialSize()
	opts.TTL = config.HistoryCacheTTL()
	opts.Pin = true
	opts.MaxCount = config.HistoryCacheMaxSize()

	return &Cache{
		Cache:            cache.New(opts),
		shard:            shard,
		executionManager: shard.GetExecutionManager(),
		logger:           shard.GetLogger().WithTags(tag.ComponentHistoryCache),
		metricsClient:    shard.GetMetricsClient(),
		config:           config,
	}
}

// GetOrCreateCurrentWorkflowExecution gets or creates workflow execution context for the current run
func (c *Cache) GetOrCreateCurrentWorkflowExecution(
	ctx context.Context,
	domainID string,
	workflowID string,
) (Context, ReleaseFunc, error) {

	scope := metrics.HistoryCacheGetOrCreateCurrentScope
	c.metricsClient.IncCounter(scope, metrics.CacheRequests)
	sw := c.metricsClient.StartTimer(scope, metrics.CacheLatency)
	defer sw.Stop()

	// using empty run ID as current workflow run ID
	runID := ""
	execution := types.WorkflowExecution{
		WorkflowID: workflowID,
		RunID:      runID,
	}

	return c.getOrCreateWorkflowExecutionInternal(
		ctx,
		domainID,
		execution,
		scope,
		true,
		scope,
	)
}

// GetAndCreateWorkflowExecution is for analyzing mutableState, it will try getting Context from cache
// and also load from database
func (c *Cache) GetAndCreateWorkflowExecution(
	ctx context.Context,
	domainID string,
	execution types.WorkflowExecution,
) (Context, Context, ReleaseFunc, bool, error) {

	scope := metrics.HistoryCacheGetAndCreateScope
	c.metricsClient.IncCounter(scope, metrics.CacheRequests)
	sw := c.metricsClient.StartTimer(scope, metrics.CacheLatency)
	defer sw.Stop()

	if err := c.validateWorkflowExecutionInfo(ctx, domainID, &execution); err != nil {
		c.metricsClient.IncCounter(scope, metrics.CacheFailures)
		return nil, nil, nil, false, err
	}

	key := definition.NewWorkflowIdentifier(domainID, execution.GetWorkflowID(), execution.GetRunID())
	contextFromCache, cacheHit := c.Get(key).(Context)
	// TODO This will create a closure on every request.
	//  Consider revisiting this if it causes too much GC activity
	releaseFunc := NoopReleaseFn
	// If cache hit, we need to lock the cache to prevent race condition
	if cacheHit {
		if err := contextFromCache.Lock(ctx); err != nil {
			// ctx is done before lock can be acquired
			c.Release(key)
			c.metricsClient.IncCounter(metrics.HistoryCacheGetAndCreateScope, metrics.CacheFailures)
			c.metricsClient.IncCounter(metrics.HistoryCacheGetAndCreateScope, metrics.AcquireLockFailedCounter)
			return nil, nil, nil, false, err
		}
		releaseFunc = c.makeReleaseFunc(key, contextFromCache, false, metrics.HistoryCacheGetAndCreateScope, "")
	} else {
		c.metricsClient.IncCounter(metrics.HistoryCacheGetAndCreateScope, metrics.CacheMissCounter)
	}

	// Note, the one loaded from DB is not put into cache and don't affect any behavior
	contextFromDB := NewContext(domainID, execution, c.shard, c.executionManager, c.logger)
	return contextFromCache, contextFromDB, releaseFunc, cacheHit, nil
}

// GetOrCreateWorkflowExecutionForBackground gets or creates workflow execution context with background context
// currently only used in tests
func (c *Cache) GetOrCreateWorkflowExecutionForBackground(
	domainID string,
	execution types.WorkflowExecution,
) (Context, ReleaseFunc, error) {

	return c.GetOrCreateWorkflowExecution(context.Background(), domainID, execution, metrics.HistoryCacheGetOrCreateScope)
}

// GetOrCreateWorkflowExecutionWithTimeout gets or creates workflow execution context with timeout
func (c *Cache) GetOrCreateWorkflowExecutionWithTimeout(
	domainID string,
	execution types.WorkflowExecution,
	timeout time.Duration,
	callerScope int,
) (Context, ReleaseFunc, error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.GetOrCreateWorkflowExecution(ctx, domainID, execution, callerScope)
}

// GetOrCreateWorkflowExecution gets or creates workflow execution context
func (c *Cache) GetOrCreateWorkflowExecution(
	ctx context.Context,
	domainID string,
	execution types.WorkflowExecution,
	callerScope int,
) (Context, ReleaseFunc, error) {

	scope := metrics.HistoryCacheGetOrCreateScope
	c.metricsClient.IncCounter(scope, metrics.CacheRequests)
	sw := c.metricsClient.StartTimer(scope, metrics.CacheLatency)
	defer sw.Stop()

	if err := c.validateWorkflowExecutionInfo(ctx, domainID, &execution); err != nil {
		c.metricsClient.IncCounter(scope, metrics.CacheFailures)
		return nil, nil, err
	}

	return c.getOrCreateWorkflowExecutionInternal(
		ctx,
		domainID,
		execution,
		scope,
		false,
		callerScope,
	)
}

func (c *Cache) getOrCreateWorkflowExecutionInternal(
	ctx context.Context,
	domainID string,
	execution types.WorkflowExecution,
	scope int,
	forceClearContext bool,
	callerScope int,
) (Context, ReleaseFunc, error) {

	// Test hook for disabling the cache
	if c.disabled {
		return NewContext(domainID, execution, c.shard, c.executionManager, c.logger), NoopReleaseFn, nil
	}

	key := definition.NewWorkflowIdentifier(domainID, execution.GetWorkflowID(), execution.GetRunID())
	workflowCtx, cacheHit := c.Get(key).(Context)
	if !cacheHit {
		c.metricsClient.IncCounter(scope, metrics.CacheMissCounter)
		// Let's create the workflow execution workflowCtx
		workflowCtx = NewContext(domainID, execution, c.shard, c.executionManager, c.logger)
		elem, err := c.PutIfNotExist(key, workflowCtx)
		if err != nil {
			c.metricsClient.IncCounter(scope, metrics.CacheFailures)
			return nil, nil, err
		}
		workflowCtx = elem.(Context)
	}

	domainName, err := c.shard.GetDomainCache().GetDomainName(domainID)
	if err != nil {
		domainName = ""
	}
	sw := c.metricsClient.Scope(callerScope, metrics.DomainTag(domainName)).StartTimer(metrics.LockWaitLatency)
	defer sw.Stop()
	if err := workflowCtx.Lock(ctx); err != nil {
		// ctx is done before lock can be acquired
		c.Release(key)
		c.metricsClient.IncCounter(scope, metrics.CacheFailures)
		c.metricsClient.IncCounter(scope, metrics.AcquireLockFailedCounter)
		return nil, nil, err
	}
	// TODO This will create a closure on every request.
	//  Consider revisiting this if it causes too much GC activity
	releaseFunc := c.makeReleaseFunc(key, workflowCtx, forceClearContext, callerScope, domainName)
	return workflowCtx, releaseFunc, nil
}

func (c *Cache) validateWorkflowExecutionInfo(
	ctx context.Context,
	domainID string,
	execution *types.WorkflowExecution,
) error {

	if execution.GetWorkflowID() == "" {
		return &types.BadRequestError{Message: "Can't load workflow execution.  WorkflowId not set."}
	}

	// RunID is not provided, lets try to retrieve the RunID for current active execution
	if execution.GetRunID() == "" {
		response, err := c.getCurrentExecutionWithRetry(ctx, &persistence.GetCurrentExecutionRequest{
			DomainID:   domainID,
			WorkflowID: execution.GetWorkflowID(),
		})

		if err != nil {
			return err
		}

		execution.RunID = response.RunID
	} else if uuid.Parse(execution.GetRunID()) == nil { // immediately return if invalid runID
		return &types.BadRequestError{Message: "RunID is not valid UUID."}
	}
	return nil
}

func (c *Cache) makeReleaseFunc(
	key definition.WorkflowIdentifier,
	context Context,
	forceClearContext bool,
	callerScope int,
	domainName string,
) func(error) {

	status := cacheNotReleased
	start := time.Now()
	return func(err error) {
		defer func() {
			if atomic.CompareAndSwapInt32(&status, cacheNotReleased, cacheReleased) {
				if rec := recover(); rec != nil {
					context.Clear()
					context.Unlock()
					c.metricsClient.Scope(callerScope, metrics.DomainTag(domainName)).RecordTimer(metrics.LockHoldLatency, time.Since(start))
					c.Release(key)
					panic(rec)
				} else {
					if err != nil || forceClearContext {
						// TODO see issue #668, there are certain type or errors which can bypass the clear
						context.Clear()
					}
					if domainName == "cadence-canary" {
						if callerScope == metrics.TimerActiveTaskDeleteHistoryEventScope {
							//	time.Sleep(67 * time.Millisecond)
							//} else if callerScope == metrics.TransferActiveTaskStartChildExecutionScope {
							//	time.Sleep(32 * time.Millisecond)
						} else if callerScope == metrics.HistoryCacheGetOrCreateCurrentScope {
							//time.Sleep(29 * time.Millisecond)
							//} else if callerScope == metrics.TransferActiveTaskSignalExecutionScope {
							//	time.Sleep(20 * time.Millisecond)
							//} else if callerScope == metrics.TransferActiveTaskCancelExecutionScope {
							//	time.Sleep(18 * time.Millisecond)
						} else if callerScope == metrics.HistoryRespondDecisionTaskCompletedScope {
							time.Sleep(18 * time.Millisecond)
							//} else if callerScope == metrics.HistoryResetWorkflowExecutionScope {
							//	time.Sleep(17 * time.Millisecond)
						} else if callerScope == metrics.PersistenceUpdateWorkflowExecutionScope {
							time.Sleep(14 * time.Millisecond)
						} else if callerScope == metrics.TimerActiveTaskActivityTimeoutScope {
							time.Sleep(5 * time.Millisecond)
						}
					}
					context.Unlock()
					c.metricsClient.Scope(callerScope, metrics.DomainTag(domainName)).RecordTimer(metrics.LockHoldLatency, time.Since(start))
					c.Release(key)
				}
			}
		}()
	}
}

func (c *Cache) getCurrentExecutionWithRetry(
	ctx context.Context,
	request *persistence.GetCurrentExecutionRequest,
) (*persistence.GetCurrentExecutionResponse, error) {

	c.metricsClient.IncCounter(metrics.HistoryCacheGetCurrentExecutionScope, metrics.CacheRequests)
	sw := c.metricsClient.StartTimer(metrics.HistoryCacheGetCurrentExecutionScope, metrics.CacheLatency)
	defer sw.Stop()

	var response *persistence.GetCurrentExecutionResponse
	op := func() error {
		var err error
		response, err = c.executionManager.GetCurrentExecution(ctx, request)

		return err
	}

	throttleRetry := backoff.NewThrottleRetry(
		backoff.WithRetryPolicy(common.CreatePersistenceRetryPolicy()),
		backoff.WithRetryableError(persistence.IsTransientError),
	)
	err := throttleRetry.Do(ctx, op)
	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryCacheGetCurrentExecutionScope, metrics.CacheFailures)
		return nil, err
	}

	return response, nil
}
