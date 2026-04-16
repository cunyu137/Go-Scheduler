package executor

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type printPayload struct {
	Msg string `json:"msg"`
}

type failOncePayload struct {
	Key string `json:"key"`
}

var failOnceState sync.Map

func RegisterBuiltinHandlers(exec *Executor) {
	exec.Register("demo.print", func(ctx context.Context, payload json.RawMessage) error {
		var p printPayload
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &p); err != nil {
				return err
			}
		}
		select {
		case <-time.After(2 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	exec.Register("demo.fail_once", func(ctx context.Context, payload json.RawMessage) error {
		var p failOncePayload
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &p); err != nil {
				return err
			}
		}
		if p.Key == "" {
			p.Key = "default"
		}
		_, loaded := failOnceState.LoadOrStore(p.Key, true)
		if !loaded {
			return errors.New("intentional failure for first run")
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
