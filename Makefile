DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable

network:
	docker network create bank-network

postgres:
	docker run --name postgres15 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14-alpine
	
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

sqlc:
	sqlc generate

mock:
	mockgen -package mockdb  -destination db/mock/store.go github.com/shui12jiao/my_simplebank/db/sqlc Store

test:
	go test -v -cover ./...

server:
	go run main.go

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgresql -o doc/schema.sql doc/db.dbml

.PHONY: network postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server mock db_docs db_schema
