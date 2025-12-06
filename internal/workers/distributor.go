package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const (
	TaskProcessImage = "task:process_image"
)

type TaskDistributor interface {
	DistributeProcessImageTask(ctx context.Context, payload *ProcessImagePayload, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

type ProcessImagePayload struct {
	JobID        string `json:"job_id"`
	ImageID      string `json:"image_id"`
	OriginalPath string `json:"original_path"`
}

func NewTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}

func (distributor *RedisTaskDistributor) DistributeProcessImageTask(
	ctx context.Context,
	payload *ProcessImagePayload,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TaskProcessImage, jsonPayload, opts...)
	info, err := distributor.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	_ = info
	return nil
}
