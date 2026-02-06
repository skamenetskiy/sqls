package sqls

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSelector_FindOne(t *testing.T) {
	type target struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	db, mock, _ := sqlmock.New()
	q, err := New[target](db)
	if err != nil {
		t.Fatalf("failed to init Selector: %s", err)
	}

	const query = "SELECT id, name FROM table"

	mock.ExpectQuery(query).
		RowsWillBeClosed().
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test"))

	result, err := q.SelectOne(t.Context(), query)
	if err != nil {
		t.Fatalf("err should be nil, but got %v", err)
	}

	expected1 := target{1, "test"}
	if !reflect.DeepEqual(result, expected1) {
		t.Fatalf("result should be %v, but got %v", expected1, result)
	}

	mock.ExpectQuery(query).
		WillReturnError(sql.ErrNoRows)

	_, err = q.SelectOne(t.Context(), query)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err should be sql.ErrNoRows, but got %v", err)
	}

	mock.ExpectQuery(query).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test1").AddRow(2, "test2"))

	result, err = q.SelectOne(t.Context(), query)
	if err != nil {
		t.Fatalf("err should be nil, but got %v", err)
	}

	expected2 := target{1, "test1"}
	if !reflect.DeepEqual(result, expected2) {
		t.Fatalf("result should be %v, but got %v", expected2, result)
	}
}

func TestSelector_Select(t *testing.T) {
	type target struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	db, mock, _ := sqlmock.New()
	q, err := New[target](db)
	if err != nil {
		t.Fatalf("failed to init Selector: %s", err)
	}

	const query = "SELECT id, name FROM table"

	mock.ExpectQuery(query).
		RowsWillBeClosed().
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test"))

	result, err := q.Select(t.Context(), query)
	if err != nil {
		t.Fatalf("err should be nil, but got %v", err)
	}

	expected1 := []target{
		{1, "test"},
	}
	if !reflect.DeepEqual(result, expected1) {
		t.Fatalf("result should be %v, but got %v", expected1, result)
	}

	mock.ExpectQuery(query).
		WillReturnError(sql.ErrNoRows)

	_, err = q.Select(t.Context(), query)
	if err != nil {
		t.Fatalf("err should be nil, but got %v", err)
	}

	mock.ExpectQuery(query).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test1").AddRow(2, "test2"))

	result, err = q.Select(t.Context(), query)
	if err != nil {
		t.Fatalf("err should be nil, but got %v", err)
	}

	expected2 := []target{
		{1, "test1"},
		{2, "test2"},
	}
	if !reflect.DeepEqual(result, expected2) {
		t.Fatalf("result should be %v, but got %v", expected2, result)
	}
}
