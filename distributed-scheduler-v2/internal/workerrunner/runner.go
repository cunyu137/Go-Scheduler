package workerrunner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"distributed-scheduler-v2/internal/executor"
	"distributed-scheduler-v2/internal/model"

	"github.com/sirupsen/logrus"
)

type Job struct {
	InstanceID     int64
	HandlerName    string
	Payload        string
	TimeoutSeconds int
	RetryCount     int
	MaxRetry       int
	IdempotentKey  string
}

type Runner struct {
	workerID     string
	address      string
	adminBaseURL string
	exec         *executor.Executor
	logger       *logrus.Logger
	client       *http.Client
}

func New(workerID, address, adminBaseURL string, exec *executor.Executor, logger *logrus.Logger) *Runner {
	return &Runner{workerID: workerID, address: address, adminBaseURL: strings.TrimRight(adminBaseURL, "/"), exec: exec, logger: logger, client: &http.Client{Timeout: 5 * time.Second}}
}

func (r *Runner) WorkerID() string { return r.workerID }

func (r *Runner) StartHeartbeat(intervalSeconds int) {
	if intervalSeconds <= 0 {
		intervalSeconds = 5
	}
	_ = r.register()
	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			_ = r.register()
		}
	}()
}

func (r *Runner) register() error {
	body, _ := json.Marshal(map[string]string{"id": r.workerID, "address": r.address})
	_, err := r.client.Post(r.adminBaseURL+"/internal/workers/register", "application/json", bytes.NewReader(body))
	if err != nil {
		r.logger.WithError(err).Warn("worker register/heartbeat failed")
	}
	return err
}

func (r *Runner) ExecuteAsync(job Job) {
	go r.execute(job)
}

func (r *Runner) execute(job Job) {
	_ = r.post(fmt.Sprintf("/internal/instances/%d/start", job.InstanceID), map[string]any{"worker_id": r.workerID})
	inst := &model.TaskInstance{ID: job.InstanceID, HandlerName: job.HandlerName, Payload: job.Payload, TimeoutSeconds: job.TimeoutSeconds, RetryCount: job.RetryCount, MaxRetry: job.MaxRetry, IdempotentKey: job.IdempotentKey}
	success, timedOut, message := r.exec.Execute(r.workerID, inst)
	_ = r.post(fmt.Sprintf("/internal/instances/%d/finish", job.InstanceID), map[string]any{"worker_id": r.workerID, "success": success, "timed_out": timedOut, "message": message, "retry_count": job.RetryCount, "max_retry": job.MaxRetry})
}

func (r *Runner) post(path string, payload any) error {
	body, _ := json.Marshal(payload)
	resp, err := r.client.Post(r.adminBaseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		r.logger.WithError(err).WithField("path", path).Warn("admin callback failed")
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		err := fmt.Errorf("callback status=%d", resp.StatusCode)
		r.logger.WithError(err).WithField("path", path).Warn("admin callback bad status")
		return err
	}
	return nil
}
