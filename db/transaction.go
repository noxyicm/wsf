package db

import (
	"database/sql"
	"wsf/context"
	"wsf/errors"
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
	SetContext(ctx context.Context) error
	Commit() error
	Rollback() error
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
	Ctx context.Context
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

// NewDefaultTransaction creates default transaction
func NewDefaultTransaction(tx *sql.Tx) (Transaction, error) {
	return &DefaultTransaction{
		Tx:  tx,
		Ctx: nil,
	}, nil
}
