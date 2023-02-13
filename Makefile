serverdev:
	nodemon --exec go run main.go --signal SIGTERM
.PHONY: serverdev
