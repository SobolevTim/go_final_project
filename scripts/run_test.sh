#!/bin/bash
export TODO_PORT=8080 # port to run the server
export TODO_DBFILE="../data/scheduler.db" # path to db file

# To start the test, select the necessary tests.
#go test -run ^TestApp$ ./tests # run first tesr
#go test -run ^TestDB$ ./tests # run second test
#go test -run ^TestNextDate$ ./tests # run third test
#go test -run ^TestAddTask$ ./tests # run fourth test
#go test -run ^TestTasks$ ./tests # run fifth test
#go test -run ^TestTask$ ./tests # run sixth test
#go test -run ^TestEditTask$ ./tests # run seventh test
#go test -run ^TestDone$ ./tests # run eighth test
#go test -run ^TestDelTask$ ./tests # run ninth test

# Run all tests
go test ./tests
# Clear test cashe
go clean -testcache 