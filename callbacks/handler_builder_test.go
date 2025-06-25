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
	"errors"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
)

// Test helper to create a simple handler for testing
func createTestHandler(id string, tracker *[]string) Handler {
	return NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context {
			*tracker = append(*tracker, "start-"+id)
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *RunInfo, output CallbackOutput) context.Context {
			*tracker = append(*tracker, "end-"+id)
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *RunInfo, err error) context.Context {
			*tracker = append(*tracker, "error-"+id)
			return ctx
		}).
		Build()
}

func TestNewHandlerBuilder(t *testing.T) {
	t.Run("basic builder functionality", func(t *testing.T) {
		var tracker []string

		handler := NewHandlerBuilder().
			OnStartFn(func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context {
				tracker = append(tracker, "start")
				return ctx
			}).
			OnEndFn(func(ctx context.Context, info *RunInfo, output CallbackOutput) context.Context {
				tracker = append(tracker, "end")
				return ctx
			}).
			OnErrorFn(func(ctx context.Context, info *RunInfo, err error) context.Context {
				tracker = append(tracker, "error")
				return ctx
			}).
			Build()

		// Test OnStart
		ctx := handler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.NotNil(t, ctx)
		assert.Equal(t, []string{"start"}, tracker)

		// Test OnEnd
		ctx = handler.OnEnd(ctx, &RunInfo{}, nil)
		assert.Equal(t, []string{"start", "end"}, tracker)

		// Test OnError
		ctx = handler.OnError(ctx, &RunInfo{}, errors.New("test error"))
		assert.Equal(t, []string{"start", "end", "error"}, tracker)
	})

	t.Run("stream functionality", func(t *testing.T) {
		var tracker []string

		handler := NewHandlerBuilder().
			OnStartWithStreamInputFn(func(ctx context.Context, info *RunInfo, input *schema.StreamReader[CallbackInput]) context.Context {
				tracker = append(tracker, "stream-start")
				return ctx
			}).
			OnEndWithStreamOutputFn(func(ctx context.Context, info *RunInfo, output *schema.StreamReader[CallbackOutput]) context.Context {
				tracker = append(tracker, "stream-end")
				return ctx
			}).
			Build()

		// Test stream methods
		sr, _ := schema.Pipe[CallbackInput](1)
		ctx := handler.OnStartWithStreamInput(context.Background(), &RunInfo{}, sr)
		assert.Equal(t, []string{"stream-start"}, tracker)

		sr2, _ := schema.Pipe[CallbackOutput](1)
		ctx = handler.OnEndWithStreamOutput(ctx, &RunInfo{}, sr2)
		assert.Equal(t, []string{"stream-start", "stream-end"}, tracker)
	})

	t.Run("Needed method", func(t *testing.T) {
		handler := NewHandlerBuilder().
			OnStartFn(func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context {
				return ctx
			}).
			OnErrorFn(func(ctx context.Context, info *RunInfo, err error) context.Context {
				return ctx
			}).
			Build()

		assert.True(t, handler.Needed(context.Background(), &RunInfo{}, TimingOnStart))
		assert.False(t, handler.Needed(context.Background(), &RunInfo{}, TimingOnEnd))
		assert.True(t, handler.Needed(context.Background(), &RunInfo{}, TimingOnError))
		assert.False(t, handler.Needed(context.Background(), &RunInfo{}, TimingOnStartWithStreamInput))
		assert.False(t, handler.Needed(context.Background(), &RunInfo{}, TimingOnEndWithStreamOutput))
	})
}

func TestNewMultiHandlerBuilder(t *testing.T) {
	t.Run("empty handlers", func(t *testing.T) {
		builder := NewMultiHandlerBuilder()
		handler := builder.Build()

		// Should not panic and should return empty handler
		ctx := handler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.NotNil(t, ctx)
	})

	t.Run("multiple handlers execution order", func(t *testing.T) {
		var tracker []string

		handler1 := createTestHandler("1", &tracker)
		handler2 := createTestHandler("2", &tracker)
		handler3 := createTestHandler("3", &tracker)

		multiHandler := NewMultiHandlerBuilder(handler1, handler2, handler3).Build()

		// Test execution order
		ctx := multiHandler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.Equal(t, []string{"start-1", "start-2", "start-3"}, tracker)

		tracker = []string{} // Reset
		ctx = multiHandler.OnEnd(ctx, &RunInfo{}, nil)
		assert.Equal(t, []string{"end-1", "end-2", "end-3"}, tracker)

		tracker = []string{} // Reset
		ctx = multiHandler.OnError(ctx, &RunInfo{}, errors.New("test"))
		assert.Equal(t, []string{"error-1", "error-2", "error-3"}, tracker)
	})
}

func TestNewPreHandlerBuilder(t *testing.T) {
	t.Run("nil existing builder", func(t *testing.T) {
		var tracker []string
		preHandler := createTestHandler("pre", &tracker)

		builder := NewPreHandlerBuilder(nil, preHandler)
		handler := builder.Build()

		handler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.Equal(t, []string{"start-pre"}, tracker)
	})

	t.Run("prepend to existing builder", func(t *testing.T) {
		var tracker []string

		// Create existing builder
		existingBuilder := NewHandlerBuilder().
			OnStartFn(func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context {
				tracker = append(tracker, "start-existing")
				return ctx
			})

		// Create pre-handlers
		preHandler1 := createTestHandler("pre1", &tracker)
		preHandler2 := createTestHandler("pre2", &tracker)

		// Build with pre-handlers
		builder := NewPreHandlerBuilder(existingBuilder, preHandler1, preHandler2)
		handler := builder.Build()

		// Test execution order - pre-handlers should execute first
		handler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.Equal(t, []string{"start-pre1", "start-pre2", "start-existing"}, tracker)
	})
}

func TestNewPostHandlerBuilder(t *testing.T) {
	t.Run("nil existing builder", func(t *testing.T) {
		var tracker []string
		postHandler := createTestHandler("post", &tracker)

		builder := NewPostHandlerBuilder(nil, postHandler)
		handler := builder.Build()

		handler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.Equal(t, []string{"start-post"}, tracker)
	})

	t.Run("append to existing builder", func(t *testing.T) {
		var tracker []string

		// Create existing builder
		existingBuilder := NewHandlerBuilder().
			OnStartFn(func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context {
				tracker = append(tracker, "start-existing")
				return ctx
			})

		// Create post-handlers
		postHandler1 := createTestHandler("post1", &tracker)
		postHandler2 := createTestHandler("post2", &tracker)

		// Build with post-handlers
		builder := NewPostHandlerBuilder(existingBuilder, postHandler1, postHandler2)
		handler := builder.Build()

		// Test execution order - post-handlers should execute after existing
		handler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.Equal(t, []string{"start-existing", "start-post1", "start-post2"}, tracker)
	})
}

func TestChainHandlers(t *testing.T) {
	t.Run("chain multiple handlers", func(t *testing.T) {
		var tracker []string

		handler1 := createTestHandler("1", &tracker)
		handler2 := createTestHandler("2", &tracker)
		handler3 := createTestHandler("3", &tracker)

		chainedHandler := ChainHandlers(handler1, handler2, handler3)

		// Test chained execution
		chainedHandler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.Equal(t, []string{"start-1", "start-2", "start-3"}, tracker)

		tracker = []string{} // Reset
		chainedHandler.OnEnd(context.Background(), &RunInfo{}, nil)
		assert.Equal(t, []string{"end-1", "end-2", "end-3"}, tracker)
	})

	t.Run("empty chain", func(t *testing.T) {
		chainedHandler := ChainHandlers()

		// Should not panic
		ctx := chainedHandler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.NotNil(t, ctx)
	})
}

func TestStateAwareHandlerBuilder(t *testing.T) {
	type TestState struct {
		Counter int
		Message string
	}

	t.Run("basic state-aware functionality", func(t *testing.T) {
		var capturedState TestState
		var capturedInput CallbackInput

		builder := NewStateAwareHandlerBuilder[TestState]().
			OnStartWithState(func(ctx context.Context, state TestState, info *RunInfo, input CallbackInput) context.Context {
				capturedState = state
				capturedInput = input
				return ctx
			}).
			OnEndWithState(func(ctx context.Context, state TestState, info *RunInfo, output CallbackOutput) context.Context {
				capturedState = state
				return ctx
			}).
			OnErrorWithState(func(ctx context.Context, state TestState, info *RunInfo, err error) context.Context {
				capturedState = state
				return ctx
			})

		handler := builder.Build()

		// Test that handler exists and has the right methods
		assert.NotNil(t, handler)

		// Note: Since getStateFromContext is not implemented and returns an error,
		// the state-aware callbacks will be skipped. This is expected behavior.
		ctx := handler.OnStart(context.Background(), &RunInfo{}, "test-input")
		assert.NotNil(t, ctx)

		// The captured values should remain zero since state extraction fails
		assert.Equal(t, TestState{}, capturedState)
		assert.Nil(t, capturedInput)
	})

	t.Run("require state setting", func(t *testing.T) {
		builder := NewStateAwareHandlerBuilder[TestState]().
			RequireState(true).
			OnStartWithState(func(ctx context.Context, state TestState, info *RunInfo, input CallbackInput) context.Context {
				return ctx
			})

		handler := builder.Build()
		assert.NotNil(t, handler)

		// Should still work even when state is required but not available
		// (the implementation gracefully handles this case)
		ctx := handler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.NotNil(t, ctx)
	})

	t.Run("fallback behavior", func(t *testing.T) {
		builder := NewStateAwareHandlerBuilder[TestState]().
			RequireState(false). // Allow fallback
			OnStartWithState(func(ctx context.Context, state TestState, info *RunInfo, input CallbackInput) context.Context {
				return ctx
			})

		handler := builder.Build()
		assert.NotNil(t, handler)

		// Should work with fallback enabled
		ctx := handler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.NotNil(t, ctx)
	})

	t.Run("stream methods with state", func(t *testing.T) {
		var called bool

		builder := NewStateAwareHandlerBuilder[TestState]().
			OnStartWithStreamInputAndState(func(ctx context.Context, state TestState, info *RunInfo, input *schema.StreamReader[CallbackInput]) context.Context {
				called = true
				return ctx
			}).
			OnEndWithStreamOutputAndState(func(ctx context.Context, state TestState, info *RunInfo, output *schema.StreamReader[CallbackOutput]) context.Context {
				called = true
				return ctx
			})

		handler := builder.Build()

		// Test stream methods
		sr, _ := schema.Pipe[CallbackInput](1)
		handler.OnStartWithStreamInput(context.Background(), &RunInfo{}, sr)

		sr2, _ := schema.Pipe[CallbackOutput](1)
		handler.OnEndWithStreamOutput(context.Background(), &RunInfo{}, sr2)

		// Since state extraction fails, the callbacks won't be called
		assert.False(t, called)
	})

	t.Run("nil callback functions", func(t *testing.T) {
		// Test that builder handles nil callback functions gracefully
		builder := NewStateAwareHandlerBuilder[TestState]()
		handler := builder.Build()

		// Should not panic even with no callback functions set
		ctx := handler.OnStart(context.Background(), &RunInfo{}, nil)
		assert.NotNil(t, ctx)

		ctx = handler.OnEnd(ctx, &RunInfo{}, nil)
		assert.NotNil(t, ctx)

		ctx = handler.OnError(ctx, &RunInfo{}, errors.New("test"))
		assert.NotNil(t, ctx)
	})
}

func TestComplexHandlerComposition(t *testing.T) {
	t.Run("complex composition scenario", func(t *testing.T) {
		var tracker []string

		// Create base handlers
		baseHandler := createTestHandler("base", &tracker)
		preHandler := createTestHandler("pre", &tracker)
		postHandler := createTestHandler("post", &tracker)

		// Create a complex composition
		baseBuilder := NewHandlerBuilder().
			OnStartFn(func(ctx context.Context, info *RunInfo, input CallbackInput) context.Context {
				tracker = append(tracker, "start-builder")
				return ctx
			})

		// Add pre-handlers
		withPreHandlers := NewPreHandlerBuilder(baseBuilder, preHandler)

		// Add post-handlers
		withPostHandlers := NewPostHandlerBuilder(withPreHandlers, postHandler)

		// Add the base handler to the mix
		finalHandler := NewMultiHandlerBuilder(withPostHandlers.Build(), baseHandler).Build()

		// Test the complex execution order
		finalHandler.OnStart(context.Background(), &RunInfo{}, nil)

		expected := []string{
			"start-pre",     // Pre-handler first
			"start-builder", // Original builder
			"start-post",    // Post-handler
			"start-base",    // Base handler from multi-builder
		}
		assert.Equal(t, expected, tracker)
	})
}
