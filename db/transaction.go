package db

import (
	goctx "context"
	"database/sql"
	"strconv"
	"strings"
	"time"
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"
)

const (
	// TYPEDefaultTransaction is a type id of rowset class
	TYPEDefaultTransaction = "default"
)

var (
	buildTransactionHandlers = map[string]func(*sql.Tx) (Transaction, error){}
)

func init() {
	RegisterTransaction(TYPEDefaultTransaction, NewDefaultTransaction)
}

// Transaction represents transaction interface
type Transaction interface {
	SetAdapter(adpt Adapter) error
	SetContext(ctx context.Context) error
	Commit() error
	Rollback() error
	Update(table string, data map[string]interface{}, cond map[string]interface{}) (bool, error)
	Insert(table string, data map[string]interface{}) (int, error)
	Delete(table string, cond map[string]interface{}) (bool, error)
	Query(dbs Select) ([]map[string]interface{}, error)
}

// NewTransaction creates a new rowset
func NewTransaction(transactionType string, tx *sql.Tx) (Transaction, error) {
	if f, ok := buildTransactionHandlers[transactionType]; ok {
		return f(tx)
	}

	return nil, errors.Errorf("Unrecognized database transaction type \"%v\"", transactionType)
}

// RegisterTransaction registers a handler for database rowset creation
func RegisterTransaction(transactionType string, handler func(*sql.Tx) (Transaction, error)) {
	buildTransactionHandlers[transactionType] = handler
}

// DefaultTransaction represents database transaction
type DefaultTransaction struct {
	Tx  *sql.Tx
	Adp Adapter
	Ctx context.Context
}

// SetAdapter sets the transaction sql adapter
func (t *DefaultTransaction) SetAdapter(adpt Adapter) error {
	t.Adp = adpt
	return nil
}

// SetContext sets the transaction context
func (t *DefaultTransaction) SetContext(ctx context.Context) error {
	t.Ctx = ctx
	return nil
}

// Commit commits a transaction
func (t *DefaultTransaction) Commit() error {
	err := t.Tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Transaction commit error")
	}

	return nil
}

// Rollback roll back a transaction
func (t *DefaultTransaction) Rollback() error {
	err := t.Tx.Rollback()
	if err != nil {
		return errors.Wrap(err, "Transaction roll back error")
	}

	return nil
}

// Insert inserts new row into table
func (t *DefaultTransaction) Insert(table string, data map[string]interface{}) (int, error) {
	cols := []string{}
	vals := []string{}
	binds := []interface{}{}
	i := 0
	for col, val := range data {
		cols = append(cols, t.Adp.QuoteIdentifier(col, true))

		switch val.(type) {
		case *SQLExpr:
			vals = append(vals, val.(*SQLExpr).ToString())

		default:
			if t.Adp.SupportsParameters("positional") {
				vals = append(vals, "?")
				binds = append(binds, val)
			} else if t.Adp.SupportsParameters("named") {
				vals = append(vals, ":col"+strconv.Itoa(i))
				binds = append(binds, sql.Named("col"+strconv.Itoa(i), val))
				i++
			} else {
				return 0, errors.New("Adapter doesn't support positional or named binding")
			}
		}
	}

	query := "INSERT INTO " + t.Adp.QuoteIdentifier(table, true) + " (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(vals, ", ") + ")"

	pctx, cancel := goctx.WithTimeout(t.Ctx, time.Duration(t.Adp.GetOptions().PingTimeout)*time.Second)
	defer cancel()

	stmt, err := t.Tx.PrepareContext(pctx, query)
	if err != nil {
		return 0, errors.Wrap(err, "Database insert Error")
	}

	qctx, cancel := goctx.WithTimeout(t.Ctx, time.Duration(t.Adp.GetOptions().QueryTimeout)*time.Second)
	defer cancel()

	result, err := stmt.ExecContext(qctx, binds...)
	if err != nil {
		stmt.Close()
		return 0, errors.Wrap(err, "Database insert Error")
	}
	stmt.Close()

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "Database insert Error")
	}

	return int(lastInsertID), nil
}

// Update updates rows into table be condition
func (t *DefaultTransaction) Update(table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
	set := []string{}
	binds := []interface{}{}
	i := 1
	for col, val := range data {
		var value string

		switch val.(type) {
		case *SQLExpr:
			value = val.(*SQLExpr).ToString()

		default:
			if t.Adp.SupportsParameters("positional") {
				value = "?"
				binds = append(binds, val)
				i++
			} else if t.Adp.SupportsParameters("named") {
				value = ":col" + strconv.Itoa(i)
				binds = append(binds, sql.Named("col"+strconv.Itoa(i), val))
				i++
			} else {
				return false, errors.New("Adapter doesn't support positional or named binding")
			}
		}

		set = append(set, t.Adp.QuoteIdentifier(col, true)+" = "+value)
	}

	where := t.Adp.WhereExpresion(cond)

	query := "UPDATE " + t.Adp.QuoteIdentifier(table, true) + " SET " + strings.Join(set, ", ") + ""
	if where != "" {
		query = query + " WHERE " + where
	}

	pctx, cancel := goctx.WithTimeout(t.Ctx, time.Duration(t.Adp.GetOptions().PingTimeout)*time.Second)
	defer cancel()

	stmt, err := t.Tx.PrepareContext(pctx, query)
	if err != nil {
		return false, errors.Wrap(err, "Database insert Error")
	}

	qctx, cancel := goctx.WithTimeout(t.Ctx, time.Duration(t.Adp.GetOptions().QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(qctx, binds...)
	if err != nil {
		stmt.Close()
		return false, err
	}
	stmt.Close()
	defer rows.Close()

	for rows.Next() {
		var updatedID int
		if err := rows.Scan(&updatedID); err != nil {
			return true, err
		}
	}

	return true, nil
}

// Delete removes rows from table
func (t *DefaultTransaction) Delete(table string, cond map[string]interface{}) (bool, error) {
	where := t.Adp.WhereExpresion(cond)

	sql := "DELETE FROM " + t.Adp.QuoteIdentifier(table, true)
	if where != "" {
		sql = sql + " WHERE " + where
	}

	pctx, cancel := goctx.WithTimeout(t.Ctx, time.Duration(t.Adp.GetOptions().PingTimeout)*time.Second)
	defer cancel()

	stmt, err := t.Tx.PrepareContext(pctx, sql)
	if err != nil {
		return false, err
	}

	qctx, cancel := goctx.WithTimeout(t.Ctx, time.Duration(t.Adp.GetOptions().QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(qctx)
	if err != nil {
		stmt.Close()
		return false, err
	}
	stmt.Close()
	rows.Close()

	return true, nil
}

// Query runs a query
func (t *DefaultTransaction) Query(dbs Select) ([]map[string]interface{}, error) {
	if t.Adp == nil {
		return nil, errors.New("Database adapter is not set")
	}

	if err := dbs.Err(); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	qctx, cancel := goctx.WithTimeout(t.Ctx, time.Duration(t.Adp.GetOptions().QueryTimeout)*time.Second)
	defer cancel()

	rows, err := t.Tx.QueryContext(qctx, dbs.Assemble(), dbs.Binds()...)
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}
	defer rows.Close()

	return t.Adp.PrepareRowset(rows)
}

// NewDefaultTransaction creates default transaction
func NewDefaultTransaction(tx *sql.Tx) (Transaction, error) {
	return &DefaultTransaction{
		Tx:  tx,
		Ctx: nil,
	}, nil
}
