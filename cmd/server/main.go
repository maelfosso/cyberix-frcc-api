package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cyberix.fr/frcc/jobs"
	"cyberix.fr/frcc/messaging"
	"cyberix.fr/frcc/server"
	"cyberix.fr/frcc/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/smithy-go/logging"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"maragu.dev/env"
)

var release string

func init() {
	godotenv.Load()
}

func main() {
	os.Exit(start())
}

func start() int {
	logEnv := env.GetStringOrDefault("LOG_ENV", "development")
	log, err := createLogger(logEnv)
	if err != nil {
		fmt.Println("Error setting up the logger: ", err)
		return 1
	}

	log = log.With(zap.String("release", release))

	defer func() {
		_ = log.Sync()
	}()

	host := env.GetStringOrDefault("HOST", "0.0.0.0")
	port := env.GetIntOrDefault("PORT", 8080)

	awsConfig, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithLogger(createAWSLogAdapter(log)),
		config.WithEndpointResolver(createAWSEndpointResolver()),
	)
	if err != nil {
		log.Info("Error creating AWS config", zap.Error(err))
		return 1
	}

	queue := createQueue(log, awsConfig)

	s := server.New(server.Options{
		Database: createDatabase(log),
		Host:     host,
		Port:     port,
		Log:      log,
		Queue:    queue,
	})

	runner := jobs.NewRunner(jobs.NewRunnerOptions{
		Emailer: createEmailer(log, host, port),
		Log:     log,
		Queue:   queue,
	})

	var eg errgroup.Group
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	eg.Go(func() error {
		if err := s.Start(); err != nil {
			log.Info("Error starting server", zap.Error(err))
			return err
		}
		return nil
	})

	eg.Go(func() error {
		runner.Start(ctx)
		return nil
	})

	<-ctx.Done()

	eg.Go(func() error {
		if err := s.Stop(); err != nil {
			log.Info("Error stopping server", zap.Error(err))
			return err
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return 1
	}

	return 0
}

func createLogger(env string) (*zap.Logger, error) {
	switch env {
	case "production":
		return zap.NewProduction()
	case "development":
		return zap.NewDevelopment()
	default:
		return zap.NewNop(), nil
	}
}

func createDatabase(log *zap.Logger) *storage.Database {
	return storage.NewDatabase(storage.NewDatabaseOptions{
		Host:                  env.GetStringOrDefault("DB_HOST", "localhost"),
		Port:                  env.GetIntOrDefault("DB_PORT", 5432),
		User:                  env.GetStringOrDefault("DB_USER", "frcc"),
		Password:              env.GetStringOrDefault("DB_PASSWORD", "123"),
		Name:                  env.GetStringOrDefault("DB_NAME", "frcc"),
		MaxOpenConnections:    10,
		MaxIdleConnections:    10,
		ConnectionMaxLifetime: time.Hour,
		Log:                   log,
	})
}

func createAWSLogAdapter(log *zap.Logger) logging.LoggerFunc {
	return func(classification logging.Classification, format string, v ...interface{}) {
		switch classification {
		case logging.Debug:
			log.Sugar().Debugf(format, v...)
		case logging.Warn:
			log.Sugar().Warnf(format, v...)
		}
	}
}

func createAWSEndpointResolver() aws.EndpointResolverFunc {
	sqsEndpointURL := env.GetStringOrDefault("SQS_ENDPOINT_URL", "")

	return func(service, region string) (aws.Endpoint, error) {
		if sqsEndpointURL != "" && service == sqs.ServiceID {
			return aws.Endpoint{
				URL: sqsEndpointURL,
			}, nil
		}

		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	}
}

func createQueue(log *zap.Logger, awsConfig aws.Config) *messaging.Queue {
	return messaging.NewQueue(messaging.NewQueueOptions{
		Config:   awsConfig,
		Log:      log,
		Name:     env.GetStringOrDefault("QUEUE_NAME", "jobs"),
		WaitTime: env.GetDurationOrDefault("QUEUE_WAIT_TIME", 20*time.Second),
	})
}

func createEmailer(log *zap.Logger, host string, port int) *messaging.Emailer {
	return messaging.NewEmailer(messaging.NewEmailerOptions{
		BaseURL:                   env.GetStringOrDefault("BASE_URL", fmt.Sprintf("http://%v:%v", host, port)),
		Log:                       log,
		MarketingEmailName:        env.GetStringOrDefault("MARKETING_EMAIL_NAME", "Forum Regional de Cybersecurité de la CEMAC"),
		MarketingEmailAddress:     env.GetStringOrDefault("MARKETING_EMAIL_ADDRESS", "bot@frcc.example.com"),
		Token:                     env.GetStringOrDefault("POSTMARK_TOKEN", ""),
		TransactionalEmailName:    env.GetStringOrDefault("TRANSACTIONAL_EMAIL_NAME", "Forum Regional de Cybersecurité de la CEMAC"),
		TransactionalEmailAddress: env.GetStringOrDefault("TRANSACTIONAL_EMAIL_ADDRESS", "bot@frcc.example.com"),
	})
}
