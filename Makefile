db_driver := mysql
db_username := root
db_password := password

dev:
	nodemon --exec go run main.go --signal SIGTERM

migrationcreate:
	migrate create -ext sql -dir database/migrations -seq init_schema

migrateup:
	migrate -path database/migrations -database "$(db_driver)://$(db_username):$(db_password)@tcp(127.0.0.1:3306)/tourid_dev?charset=utf8mb4&parseTime=True&loc=Local" -verbose up

migratedown:
	migrate -path database/migrations -database "$(db_driver)://$(db_username):$(db_password)@tcp(127.0.0.1:3306)/tourid_dev?charset=utf8mb4&parseTime=True&loc=Local" -verbose down


.PHONY: dev migrationcreate migrateup migratedown

