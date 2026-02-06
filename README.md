# sqls

[![Go Reference](https://pkg.go.dev/badge/github.com/skamenetskiy/sqls)](https://pkg.go.dev/github.com/skamenetskiy/sqls)
[![go test](https://github.com/skamenetskiy/sqls/actions/workflows/test.yml/badge.svg)](https://github.com/skamenetskiy/sqls/actions/workflows/test.yml)
[![go report](https://goreportcard.com/badge/github.com/skamenetskiy/sqls)](https://goreportcard.com/report/github.com/skamenetskiy/sqls)

A generic interface to simplify sql selects from database to struct.

## Installation

```shell
go get -u github.com/skamenetskiy/sqls
```

## Example

```golang
package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/skamenetskiy/sqls"
)

func main() {
	db, err := sql.Open("postgres", "...")
	if err != nil {
		panic(err)
	}

	q := sqls.Must[Model](db)

	one, err := QueryModel(context.Background(), q, 1)
	if err != nil {
		if sqls.IsNotFound(err) {
			log.Fatal("not found")
			return
		}
		log.Fatalf("failed to query one: %s\n", err)
	}
	log.Println("found one", one)

	many, err := QueryModels(context.Background(), q)
	if err != nil {
		log.Fatalf("failed to query one: %s\n", err)
	}

	if len(many) == 0 {
		log.Println("no models found")
		return
	}

	log.Println(many)
}

type Model struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func QueryModel(ctx context.Context, s *sqls.Selector[Model], id int) (Model, error) {
	const query = "select id, name from table where id = $1"

	result, err := s.SelectOne(ctx, query, id)
	if err != nil {
		return Model{}, err
	}

	return result, nil
}

func QueryModels(ctx context.Context, s *sqls.Selector[Model]) ([]Model, error) {
	const query = "select id, name from table"

	result, err := s.Select(ctx, query)
	if err != nil {
		return nil, err
	}

	return result, nil
}

```