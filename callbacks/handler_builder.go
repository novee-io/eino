/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package callbacks

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/schema"
)

type HandlerBuilder struct {
	onStartFns                []func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context
	onEndFns                  []func(ctx context.Context, info *RunInfo, output CallbackOutput) context.Context
	onErrorFns                []func(ctx context.Context, info *RunInfo, err error) context.Context
	onStartWithStreamInputFns []func(ctx context.Context, info *RunInfo, input *schema.StreamReader[CallbackInput]) context.Context
	onEndWithStreamOutputFns  []func(ctx context.Context, info *RunInfo, output *schema.StreamReader[CallbackOutput]) context.Context
}

type handlerImpl struct {
	HandlerBuilder
}

func (hb *handlerImpl) OnStart(ctx context.Context, info *RunInfo, input CallbackInput) context.Context {
	for _, fn := range hb.onStartFns {
		if fn != nil {
			ctx = fn(ctx, info, input)
		}
	}
	return ctx
}

func (hb *handlerImpl) OnEnd(ctx context.Context, info *RunInfo, output CallbackOutput) context.Context {
	for _, fn := range hb.onEndFns {
		if fn != nil {
			ctx = fn(ctx, info, output)
		}
	}
	return ctx
}

func (hb *handlerImpl) OnError(ctx context.Context, info *RunInfo, err error) context.Context {
	for _, fn := range hb.onErrorFns {
		if fn != nil {
			ctx = fn(ctx, info, err)
		}
	}
	return ctx
}

func (hb *handlerImpl) OnStartWithStreamInput(ctx context.Context, info *RunInfo,
	input *schema.StreamReader[CallbackInput]) context.Context {

	for _, fn := range hb.onStartWithStreamInputFns {
		if fn != nil {
			ctx = fn(ctx, info, input)
		}
	}
	return ctx
}

func (hb *handlerImpl) OnEndWithStreamOutput(ctx context.Context, info *RunInfo,
	output *schema.StreamReader[CallbackOutput]) context.Context {

	for _, fn := range hb.onEndWithStreamOutputFns {
		if fn != nil {
			ctx = fn(ctx, info, output)
		}
	}
	return ctx
}

func (hb *handlerImpl) Needed(_ context.Context, _ *RunInfo, timing CallbackTiming) bool {
	switch timing {
	case TimingOnStart:
		return len(hb.onStartFns) > 0
	case TimingOnEnd:
		return len(hb.onEndFns) > 0
	case TimingOnError:
		return len(hb.onErrorFns) > 0
	case TimingOnStartWithStreamInput:
		return len(hb.onStartWithStreamInputFns) > 0
	case TimingOnEndWithStreamOutput:
		return len(hb.onEndWithStreamOutputFns) > 0
	default:
		return false
	}
}

// NewHandlerBuilder creates and returns a new HandlerBuilder instance.
// HandlerBuilder is used to construct a Handler with custom callback functions
func NewHandlerBuilder() *HandlerBuilder {
	return &HandlerBuilder{}
}

func (hb *HandlerBuilder) OnStartFn(
	fn func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context) *HandlerBuilder {

	hb.onStartFns = append(hb.onStartFns, fn)
	return hb
}

func (hb *HandlerBuilder) OnEndFn(
	fn func(ctx context.Context, info *RunInfo, output CallbackOutput) context.Context) *HandlerBuilder {

	hb.onEndFns = append(hb.onEndFns, fn)
	return hb
}

func (hb *HandlerBuilder) OnErrorFn(
	fn func(ctx context.Context, info *RunInfo, err error) context.Context) *HandlerBuilder {

	hb.onErrorFns = append(hb.onErrorFns, fn)
	return hb
}

// OnStartWithStreamInputFn sets the callback function to be called.
func (hb *HandlerBuilder) OnStartWithStreamInputFn(
	fn func(ctx context.Context, info *RunInfo, input *schema.StreamReader[CallbackInput]) context.Context) *HandlerBuilder {

	hb.onStartWithStreamInputFns = append(hb.onStartWithStreamInputFns, fn)
	return hb
}

// OnEndWithStreamOutputFn sets the callback function to be called.
func (hb *HandlerBuilder) OnEndWithStreamOutputFn(
	fn func(ctx context.Context, info *RunInfo, output *schema.StreamReader[CallbackOutput]) context.Context) *HandlerBuilder {

	hb.onEndWithStreamOutputFns = append(hb.onEndWithStreamOutputFns, fn)
	return hb
}

// Build returns a Handler with the functions set in the builder.
func (hb *HandlerBuilder) Build() Handler {
	return &handlerImpl{*hb}
}

// addHandlersToBuilder is a helper function that adds multiple handlers to a builder
// by converting them to callback functions and appending to the builder's slices
func addHandlersToBuilder(builder *HandlerBuilder, handlers []Handler) {
	for _, handler := range handlers {
		builder.onStartFns = append(builder.onStartFns, func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context {
			return handler.OnStart(ctx, info, input)
		})
		builder.onEndFns = append(builder.onEndFns, func(ctx context.Context, info *RunInfo, output CallbackOutput) context.Context {
			return handler.OnEnd(ctx, info, output)
		})
		builder.onErrorFns = append(builder.onErrorFns, func(ctx context.Context, info *RunInfo, err error) context.Context {
			return handler.OnError(ctx, info, err)
		})
		builder.onStartWithStreamInputFns = append(builder.onStartWithStreamInputFns, func(ctx context.Context, info *RunInfo, input *schema.StreamReader[CallbackInput]) context.Context {
			return handler.OnStartWithStreamInput(ctx, info, input)
		})
		builder.onEndWithStreamOutputFns = append(builder.onEndWithStreamOutputFns, func(ctx context.Context, info *RunInfo, output *schema.StreamReader[CallbackOutput]) context.Context {
			return handler.OnEndWithStreamOutput(ctx, info, output)
		})
	}
}

// appendBuilderCallbacks is a helper function that appends all callback functions
// from the source builder to the destination builder
func appendBuilderCallbacks(dest *HandlerBuilder, src *HandlerBuilder) {
	dest.onStartFns = append(dest.onStartFns, src.onStartFns...)
	dest.onEndFns = append(dest.onEndFns, src.onEndFns...)
	dest.onErrorFns = append(dest.onErrorFns, src.onErrorFns...)
	dest.onStartWithStreamInputFns = append(dest.onStartWithStreamInputFns, src.onStartWithStreamInputFns...)
	dest.onEndWithStreamOutputFns = append(dest.onEndWithStreamOutputFns, src.onEndWithStreamOutputFns...)
}

// NewMultiHandlerBuilder creates a HandlerBuilder that can chain multiple handlers together.
// This allows for composable middleware-like behavior where multiple handlers can process
// the same callback event in sequence.
func NewMultiHandlerBuilder(handlers ...Handler) *HandlerBuilder {
	if len(handlers) == 0 {
		return NewHandlerBuilder()
	}

	builder := NewHandlerBuilder()
	addHandlersToBuilder(builder, handlers)
	return builder
}

// NewPreHandlerBuilder creates a HandlerBuilder that prepends handlers to an existing HandlerBuilder.
// The provided handlers will be executed before any existing handlers in the builder.
func NewPreHandlerBuilder(existingBuilder *HandlerBuilder, preHandlers ...Handler) *HandlerBuilder {
	if existingBuilder == nil {
		return NewMultiHandlerBuilder(preHandlers...)
	}

	builder := NewHandlerBuilder()

	// Add pre-handlers first
	addHandlersToBuilder(builder, preHandlers)

	// Add existing handlers after pre-handlers
	appendBuilderCallbacks(builder, existingBuilder)

	return builder
}

// NewPostHandlerBuilder creates a HandlerBuilder that appends handlers to an existing HandlerBuilder.
// The provided handlers will be executed after any existing handlers in the builder.
func NewPostHandlerBuilder(existingBuilder *HandlerBuilder, postHandlers ...Handler) *HandlerBuilder {
	if existingBuilder == nil {
		return NewMultiHandlerBuilder(postHandlers...)
	}

	builder := NewHandlerBuilder()

	// Add existing handlers first
	appendBuilderCallbacks(builder, existingBuilder)

	// Add post-handlers after existing handlers
	addHandlersToBuilder(builder, postHandlers)

	return builder
}

// ChainHandlers creates a single handler that executes multiple handlers in sequence.
// This is useful for creating middleware-like behavior where each handler can modify
// the context and pass it to the next handler.
func ChainHandlers(handlers ...Handler) Handler {
	return NewMultiHandlerBuilder(handlers...).Build()
}

// StateAwareHandlerBuilder creates a HandlerBuilder that provides easy access to state.
// This ensures handlers always have access to the graph's state when available.
type StateAwareHandlerBuilder[S any] struct {
	onStart                func(ctx context.Context, state S, info *RunInfo, input CallbackInput) context.Context
	onEnd                  func(ctx context.Context, state S, info *RunInfo, output CallbackOutput) context.Context
	onError                func(ctx context.Context, state S, info *RunInfo, err error) context.Context
	onStartWithStreamInput func(ctx context.Context, state S, info *RunInfo, input *schema.StreamReader[CallbackInput]) context.Context
	onEndWithStreamOutput  func(ctx context.Context, state S, info *RunInfo, output *schema.StreamReader[CallbackOutput]) context.Context
	fallbackToNoState      bool
}

// NewStateAwareHandlerBuilder creates a new StateAwareHandlerBuilder for the given state type.
func NewStateAwareHandlerBuilder[S any]() *StateAwareHandlerBuilder[S] {
	return &StateAwareHandlerBuilder[S]{
		fallbackToNoState: true, // By default, allow execution even without state
	}
}

// OnStartWithState sets the callback function to be called on start with state access.
func (b *StateAwareHandlerBuilder[S]) OnStartWithState(fn func(ctx context.Context, state S, info *RunInfo, input CallbackInput) context.Context) *StateAwareHandlerBuilder[S] {
	b.onStart = fn
	return b
}

// OnEndWithState sets the callback function to be called on end with state access.
func (b *StateAwareHandlerBuilder[S]) OnEndWithState(fn func(ctx context.Context, state S, info *RunInfo, output CallbackOutput) context.Context) *StateAwareHandlerBuilder[S] {
	b.onEnd = fn
	return b
}

// OnErrorWithState sets the callback function to be called on error with state access.
func (b *StateAwareHandlerBuilder[S]) OnErrorWithState(fn func(ctx context.Context, state S, info *RunInfo, err error) context.Context) *StateAwareHandlerBuilder[S] {
	b.onError = fn
	return b
}

// OnStartWithStreamInputAndState sets the callback function for stream input with state access.
func (b *StateAwareHandlerBuilder[S]) OnStartWithStreamInputAndState(fn func(ctx context.Context, state S, info *RunInfo, input *schema.StreamReader[CallbackInput]) context.Context) *StateAwareHandlerBuilder[S] {
	b.onStartWithStreamInput = fn
	return b
}

// OnEndWithStreamOutputAndState sets the callback function for stream output with state access.
func (b *StateAwareHandlerBuilder[S]) OnEndWithStreamOutputAndState(fn func(ctx context.Context, state S, info *RunInfo, output *schema.StreamReader[CallbackOutput]) context.Context) *StateAwareHandlerBuilder[S] {
	b.onEndWithStreamOutput = fn
	return b
}

// RequireState sets whether the handler should fail if state is not available.
// By default, handlers will skip execution if state is not available.
func (b *StateAwareHandlerBuilder[S]) RequireState(require bool) *StateAwareHandlerBuilder[S] {
	b.fallbackToNoState = !require
	return b
}

// Build creates a Handler from the StateAwareHandlerBuilder.
func (b *StateAwareHandlerBuilder[S]) Build() Handler {
	return NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context {
			if b.onStart == nil {
				return ctx
			}

			state, err := getStateFromContext[S](ctx)
			if err != nil && !b.fallbackToNoState {
				// State required but not available, skip
				return ctx
			}
			if err != nil {
				// State not available but fallback allowed, skip this specific handler
				return ctx
			}

			return b.onStart(ctx, state, info, input)
		}).
		OnEndFn(func(ctx context.Context, info *RunInfo, output CallbackOutput) context.Context {
			if b.onEnd == nil {
				return ctx
			}

			state, err := getStateFromContext[S](ctx)
			if err != nil && !b.fallbackToNoState {
				return ctx
			}
			if err != nil {
				return ctx
			}

			return b.onEnd(ctx, state, info, output)
		}).
		OnErrorFn(func(ctx context.Context, info *RunInfo, err error) context.Context {
			if b.onError == nil {
				return ctx
			}

			state, stateErr := getStateFromContext[S](ctx)
			if stateErr != nil && !b.fallbackToNoState {
				return ctx
			}
			if stateErr != nil {
				return ctx
			}

			return b.onError(ctx, state, info, err)
		}).
		OnStartWithStreamInputFn(func(ctx context.Context, info *RunInfo, input *schema.StreamReader[CallbackInput]) context.Context {
			if b.onStartWithStreamInput == nil {
				return ctx
			}

			state, err := getStateFromContext[S](ctx)
			if err != nil && !b.fallbackToNoState {
				return ctx
			}
			if err != nil {
				return ctx
			}

			return b.onStartWithStreamInput(ctx, state, info, input)
		}).
		OnEndWithStreamOutputFn(func(ctx context.Context, info *RunInfo, output *schema.StreamReader[CallbackOutput]) context.Context {
			if b.onEndWithStreamOutput == nil {
				return ctx
			}

			state, err := getStateFromContext[S](ctx)
			if err != nil && !b.fallbackToNoState {
				return ctx
			}
			if err != nil {
				return ctx
			}

			return b.onEndWithStreamOutput(ctx, state, info, output)
		}).
		Build()
}

// getStateFromContext is a helper function to extract state from context.
// This uses the same state management system as the compose package.
func getStateFromContext[S any](ctx context.Context) (S, error) {
	// Use the same stateKey and state extraction logic as compose.GetState
	type stateKey struct{}

	state := ctx.Value(stateKey{})
	if state == nil {
		var zero S
		return zero, fmt.Errorf("have not set state")
	}

	// internalState matches the structure in compose package
	type internalState struct {
		state any
		mu    sync.Mutex
	}

	iState, ok := state.(*internalState)
	if !ok {
		var zero S
		return zero, fmt.Errorf("invalid state structure")
	}

	iState.mu.Lock()
	defer iState.mu.Unlock()

	cState, ok := iState.state.(S)
	if !ok {
		var zero S
		return zero, fmt.Errorf("unexpected state type. expected: %T, got: %T",
			zero, iState.state)
	}

	return cState, nil
}
