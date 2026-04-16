package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var failOnceStore sync.Map

func RegisterBuiltinHandlers(exec *Executor) {
	exec.Register("demo.print", func(ctx context.Context, payload json.RawMessage) error {
		var req struct {
			Msg string `json:"msg"`
		}
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &req)
		}
		select {
		case <-time.After(2 * time.Second):
			fmt.Println("demo.print:", req.Msg)
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	exec.Register("demo.fail_once", func(ctx context.Context, payload json.RawMessage) error {
		var req struct {
			Key string `json:"key"`
		}
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &req)
		}
		if req.Key == "" {
			req.Key = "default"
		}
		if _, loaded := failOnceStore.LoadOrStore(req.Key, 1); !loaded {
			return errors.New("intentional first failure")
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
