DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable

network:
	docker network create bank-network

postgres:
	docker run --name postgres15 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15-alpine
	
createdb:
	docker exec -it postgres15 createdb --username=root --owner=root simple_bank
	
dropdb:
	docker exec -it postgres15 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

redis:
	docker run --name redis --network bank-network -p 6379:6379 -d redis:7-alpine

sqlc:
	sqlc generate

mock:
	mockgen -package mockdb  -destination db/mock/store.go github.com/shui12jiao/my_simplebank/db/sqlc Store

test:
	go test -v -cover -short ./...

server:
	go run main.go

proto:
	rm -f pb/*.pb.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt logtostderr=true,allow_merge=true,merge_file_name=simple_bank \
	proto/*.proto

evans:
	evans --host localhost --port 8081 -r repl

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgresql -o doc/schema.sql doc/db.dbml

.PHONY: network postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 new_migration sqlc redis test server mock proto evans db_docs db_schema 
