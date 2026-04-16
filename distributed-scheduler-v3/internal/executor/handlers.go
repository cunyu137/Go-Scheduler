package executor

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var failOnceState = map[string]bool{}

func RegisterBuiltinHandlers(exec *Executor) {
	exec.Register("demo.print", func(ctx context.Context, payload json.RawMessage) error {
		var req struct {
			Msg string `json:"msg"`
		}
		_ = json.Unmarshal(payload, &req)
		select {
		case <-time.After(2 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	exec.Register("demo.fail_once", func(ctx context.Context, payload json.RawMessage) error {
		var req struct {
			Key string `json:"key"`
		}
		_ = json.Unmarshal(payload, &req)
		if req.Key == "" {
			req.Key = "default"
		}
		if !failOnceState[req.Key] {
			failOnceState[req.Key] = true
			return errors.New("first run failed intentionally")
		}
		select {
		case <-time.After(1 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	exec.Register("demo.timeout", func(ctx context.Context, payload json.RawMessage) error {
		select {
		case <-time.After(20 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}
