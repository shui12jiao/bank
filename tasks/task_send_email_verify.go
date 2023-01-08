package tasks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	SendVerifyEmail = "task:send_verify_email"
)

type SendVerifyEmailPayload struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) SendVerifyEmail(ctx context.Context, payload *SendVerifyEmailPayload, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(SendVerifyEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().
		Str("id", info.ID).
		Str("type", info.Type).
		Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).
		Bytes("payload", info.Payload).
		Msgf("%s enqueued", SendVerifyEmail)
	return nil
}

func (processor *RedisTaskProcessor) SendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload SendVerifyEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	//TODO: send email to user

	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msgf("%s processed", SendVerifyEmail)
	return nil
}
