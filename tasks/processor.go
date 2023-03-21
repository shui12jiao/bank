package tasks

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	db "github.com/shui12jiao/my_simplebank/db/sqlc"
	"github.com/shui12jiao/my_simplebank/util"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	SendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	mailer util.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer util.EmailSender) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 6,
				QueueDefault:  3,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("type", task.Type()).
					Bytes("payload", task.Payload()).
					Msg("task processor error")
			}),
			Logger: NewLogger(),
		},
	)
	redis.SetLogger(NewLogging())

	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(SendVerifyEmail, processor.SendVerifyEmail)

	return processor.server.Start(mux)
}

// asynq.Logger set to "zerolog Logger" format
type Logger struct {
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) print(level zerolog.Level, args ...interface{}) {
	log.WithLevel(level).Msg(fmt.Sprint(args...))
}

func (l *Logger) Debug(args ...interface{}) {
	l.print(zerolog.DebugLevel, args...)
}
func (l *Logger) Info(args ...interface{}) {
	l.print(zerolog.InfoLevel, args...)
}
func (l *Logger) Warn(args ...interface{}) {
	l.print(zerolog.WarnLevel, args...)
}
func (l *Logger) Error(args ...interface{}) {
	l.print(zerolog.ErrorLevel, args...)
}
func (l *Logger) Fatal(args ...interface{}) {
	l.print(zerolog.FatalLevel, args...)
}

// redis.Logging set to "zerolog Logger" format
type Logging struct{}

func NewLogging() *Logging {
	return &Logging{}
}

func (l *Logging) Printf(ctx context.Context, format string, v ...interface{}) {
	log.Info().Msgf(format, v...)
}
