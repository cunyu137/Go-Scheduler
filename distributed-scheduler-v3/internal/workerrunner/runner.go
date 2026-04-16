package workerrunner

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"distributed-scheduler-v3/internal/executor"
	"distributed-scheduler-v3/internal/model"
)

type Runner struct {
	exec           *executor.Executor
	adminBaseURL   string
	callbackClient *http.Client
}

func New(exec *executor.Executor, adminBaseURL string, callbackTimeoutSec int) *Runner {
	if callbackTimeoutSec <= 0 {
		callbackTimeoutSec = 5
	}
	return &Runner{
		exec:         exec,
		adminBaseURL: strings.TrimRight(adminBaseURL, "/"),
		callbackClient: &http.Client{Timeout: time.Duration(callbackTimeoutSec) * time.Second},
	}
}

func (r *Runner) Run(instanceID int64, handlerName, payload string, timeoutSeconds int, idempotentKey string) {
	r.exec.Execute(instanceID, handlerName, payload, timeoutSeconds, idempotentKey, func(status, errMsg string) {
		body, _ := json.Marshal(map[string]any{
			"instance_id": instanceID,
			"status":      status,
			"error_msg":   errMsg,
		})
		_, _ = r.callbackClient.Post(r.adminBaseURL+"/internal/tasks/callback", "application/json", bytes.NewReader(body))
	})
	_ = model.InstanceStatusFailed
}
