package adminclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"distributed-scheduler-v3/internal/model"
)

type WorkerDispatcher struct{ client *http.Client }

func NewWorkerDispatcher(timeoutSeconds int) *WorkerDispatcher {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 5
	}
	return &WorkerDispatcher{client: &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}}
}

type DispatchRequest struct {
	InstanceID     int64  `json:"instance_id"`
	HandlerName    string `json:"handler_name"`
	Payload        string `json:"payload"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	RetryCount     int    `json:"retry_count"`
	MaxRetry       int    `json:"max_retry"`
	IdempotentKey  string `json:"idempotent_key"`
}

type DispatchAck struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message"`
}

func (d *WorkerDispatcher) Dispatch(worker model.Worker, inst model.TaskInstance) error {
	reqBody := DispatchRequest{
		InstanceID:     inst.ID,
		HandlerName:    inst.HandlerName,
		Payload:        inst.Payload,
		TimeoutSeconds: inst.TimeoutSeconds,
		RetryCount:     inst.RetryCount,
		MaxRetry:       inst.MaxRetry,
		IdempotentKey:  inst.IdempotentKey,
	}
	body, _ := json.Marshal(reqBody)
	url := strings.TrimRight(worker.Address, "/") + "/internal/execute"
	resp, err := d.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dispatch failed with status %d", resp.StatusCode)
	}
	var ack DispatchAck
	if err := json.NewDecoder(resp.Body).Decode(&ack); err != nil {
		return err
	}
	if !ack.Accepted {
		return fmt.Errorf("worker rejected dispatch: %s", ack.Message)
	}
	return nil
}
