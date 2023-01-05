package main

import (
	"database/sql"
	_ "embed"
	"log"
	"net"
	"net/http"

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
		log.Fatal("failed to load config:", err)
	}

	//connect to db
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("failed to connect to db:", err)
	}

	//run migrations
	runDatabaseMigrations(config.MigrationURL, config.DBSource)

	//create store
	store := db.NewStore(conn)

	//run http/grpc/grpc-gateway server
	// runHTTPServer(config, store)
	go runGatewayServer(config, store)
	runGRPCServer(config, store)
}

func runDatabaseMigrations(migrationURL, databaseSource string) {
	m, err := migrate.New(migrationURL, databaseSource)
	if err != nil {
		log.Fatal("failed to create new migrate instance:", err)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrations:", err)
	}
	log.Println("migrations run successfully")
}

func runHTTPServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("failed to create http server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("can not start http server:", err)
	}
}

func runGRPCServer(config util.Config, store db.Store) {
	server, err := apig.NewServer(config, store)
	if err != nil {
		log.Fatal("failed to create grpc server")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("failed to start grpc server:", err)
	}

	log.Println("starting grpc server at", config.GRPCServerAddress)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("failed to start grpc server:", err)
	}
}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := apig.NewServer(config, store)
	if err != nil {
		log.Fatal("failed to create gateway server")
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
		log.Fatal("failed to register grpc gateway:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(doc.SwaggerFS()))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("failed to start gateway server:", err)
	}

	log.Println("starting gateway server at", config.HTTPServerAddress)
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("failed to start gateway server:", err)
	}
}
