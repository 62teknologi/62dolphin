dev:
	nodemon --exec go run main.go --signal SIGTERM

migrationcreate:
	migrate create -ext sql -dir database/migrations -seq init_schema

migrateup:
	migrate -path database/migrations -database "$(DB_SOURCE)" -verbose up

migratedown:
	migrate -path database/migrations -database "$(DB_SOURCE)" -verbose down


.PHONY: dev migrationcreate migrateup migratedown

