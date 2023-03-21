package main

import (
	"database/sql"
	_ "embed"
	"net"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/shui12jiao/my_simplebank/api"
	"github.com/shui12jiao/my_simplebank/apig"
	db "github.com/shui12jiao/my_simplebank/db/sqlc"
	"github.com/shui12jiao/my_simplebank/doc"
	"github.com/shui12jiao/my_simplebank/pb"
	"github.com/shui12jiao/my_simplebank/tasks"
	"github.com/shui12jiao/my_simplebank/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	//load config
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	//set up logger
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	//connect to db
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to db")
	}

	//run migrations
	runDatabaseMigrations(config.MigrationURL, config.DBSource)

	//create store
	store := db.NewStore(conn)

	//create task distributor and task processor
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := tasks.NewRedisTaskDistributor(&redisOpt)
	go runRedisTaskProcessor(config, redisOpt, store)

	//run http/grpc/grpc-gateway server
	/* runHTTPServer(config, store) */
	go runGatewayServer(config, store, taskDistributor)
	runGRPCServer(config, store, taskDistributor)
}

func runDatabaseMigrations(migrationURL, databaseSource string) {
	m, err := migrate.New(migrationURL, databaseSource)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create new migrate instance")
	}

	err = m.Up()
	switch err {
	case nil:
		log.Info().Msg("migrated up successfully")
	case migrate.ErrNoChange:
		log.Info().Msg("no migrations to run")
	default:
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
}

func runRedisTaskProcessor(config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	log.Info().Msgf("starting redis task processor at %s", redisOpt.Addr)
	processor := tasks.NewRedisTaskProcessor(redisOpt, store, util.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword))
	err := processor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start redis task processor")
	}
}

func runHTTPServer(config util.Config, store db.Store, taskDistributor tasks.TaskDistributor) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("failed to create http server")
	}

	log.Info().Msgf("starting http server at %s", config.HTTPServerAddress)
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("can not start http server")
	}
}

func runGRPCServer(config util.Config, store db.Store, taskDistributor tasks.TaskDistributor) {
	server, err := apig.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("failed to create grpc server")
	}

	grpcLogger := grpc.UnaryInterceptor(apig.GRPCLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start grpc server")
	}

	log.Info().Msgf("starting grpc server at %s", config.GRPCServerAddress)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start grpc server")
	}
}

func runGatewayServer(config util.Config, store db.Store, taskDistributor tasks.TaskDistributor) {
	server, err := apig.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("failed to create gateway server")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	grpcMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}))
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register grpc gateway")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(doc.SwaggerFS()))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start gateway server")
	}

	log.Info().Msgf("starting gateway server at %s", config.HTTPServerAddress)
	err = http.Serve(listener, apig.HTTPLogger(mux))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start gateway server")
	}
}
