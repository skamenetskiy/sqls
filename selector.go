package sqls

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

type Selector[T any] struct {
	db     querier
	fields map[string]int
}

func New[T any](db querier) (q *Selector[T], err error) {
	q = &Selector[T]{db: db}
	q.fields, err = analyzeStruct[T]()
	return
}

func Must[T any](db querier) *Selector[T] {
	q, err := New[T](db)
	if err != nil {
		panic(err)
	}
	return q
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func (q *Selector[T]) Select(ctx context.Context, query string, args ...any) ([]T, error) {
	result := make([]T, 0)
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	for rows.Next() {
		t := new(T)
		if err = q.scan(rows, cols, t); err != nil {
			return nil, err
		}
		result = append(result, *t)
	}
	return result, rows.Err()
}

func (q *Selector[T]) SelectOne(ctx context.Context, query string, args ...any) (T, error) {
	result := new(T)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return *result, err
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return *result, fmt.Errorf("failed to get columns: %w", err)
	}

	if !rows.Next() {
		return *result, sql.ErrNoRows
	}

	if err = q.scan(rows, cols, result); err != nil {
		return *result, fmt.Errorf("failed to scan: %w", err)
	}

	return *result, rows.Err()
}

func (q *Selector[T]) scan(rows *sql.Rows, cols []string, target *T) error {
	vo := reflect.ValueOf(target).Elem()
	var (
		ok      bool
		pos     int
		field   reflect.Value
		targets = make([]any, len(cols))
	)
	for i, col := range cols {
		pos, ok = q.fields[col]
		if !ok {
			return fmt.Errorf("no target field found for column %s", col)
		}
		field = vo.Field(pos)
		if !field.CanAddr() {
			return fmt.Errorf("target field %s not addressable", field.String())
		}
		targets[i] = field.Addr().Interface()
	}
	if err := rows.Scan(targets...); err != nil {
		return fmt.Errorf("failed to scan: %w", err)
	}
	return nil
}

func analyzeStruct[T any]() (map[string]int, error) {
	var t T

	// check that T is a struct
	to := reflect.TypeOf(t)
	if k := to.Kind(); k != reflect.Struct {
		return nil, fmt.Errorf("expected struct got %s", k.String())
	}

	// parse struct fields
	numFields := to.NumField()
	fields := make(map[string]int, numFields)
	for i := 0; i < numFields; i++ {
		tag := to.Field(i).Tag.Get("db")
		switch tag {
		case "-", "":
			continue
		default:
			fields[tag] = i
		}
	}

	// do not allow structs without marked fields
	if len(fields) == 0 {
		return nil, fmt.Errorf("no db tags found in struct")
	}

	return fields, nil
}
