package adminclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"distributed-scheduler-v2/internal/model"
)

type WorkerDispatcher struct{ client *http.Client }

func NewWorkerDispatcher(timeoutSeconds int) *WorkerDispatcher {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 5
	}
	return &WorkerDispatcher{client: &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}}
}

type dispatchRequest struct {
	InstanceID     int64  `json:"instance_id"`
	HandlerName    string `json:"handler_name"`
	Payload        string `json:"payload"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	RetryCount     int    `json:"retry_count"`
	MaxRetry       int    `json:"max_retry"`
	IdempotentKey  string `json:"idempotent_key"`
}

func (d *WorkerDispatcher) Dispatch(worker model.Worker, inst model.TaskInstance) error {
	reqBody := dispatchRequest{InstanceID: inst.ID, HandlerName: inst.HandlerName, Payload: inst.Payload, TimeoutSeconds: inst.TimeoutSeconds, RetryCount: inst.RetryCount, MaxRetry: inst.MaxRetry, IdempotentKey: inst.IdempotentKey}
	body, _ := json.Marshal(reqBody)
	url := strings.TrimRight(worker.Address, "/") + "/internal/execute"
	resp, err := d.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dispatch failed with status %d", resp.StatusCode)
	}
	return nil
}
