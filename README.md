# Ready Set Go
A minimal Go starter project for building web APIs

## Features
- simple `http/net` based router with route groups and middleware support
- session authentication

## Tools
- [`goose`](https://github.com/pressly/goose) for db migrations
- [`sqlc`](https://sqlc.dev/) for generating typesafe Go code from sql queries

## Getting started
- Install [`goose`](https://github.com/pressly/goose)
- Copy this repository by clicking the `Use This Template` button.
- Start a new postgres DB
- Create an env variable called `DB_CONN` which should be follow this [format](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS)
- Run the migrate command using `goose`

  ```
  $ goose postgres $DB_CONN --dir db/migrations/ up;
  ```

- Run `go run .`
