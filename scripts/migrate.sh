#! /bin/sh

# go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate -database ${MYSQL_DSN} -path ./db/migrations $1 $2
