#!/bin/bash
export TODO_PORT=8080
export TODO_DBFILE="./data/scheduler.db"
export TODO_PASSWORD="a-simple-password-1234"
go run ./cmd/server/main.go
