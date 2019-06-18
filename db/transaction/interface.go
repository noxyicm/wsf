package transaction

import (
	"context"
	"database/sql"
	"wsf/errors"
)

const (
	// TYPEDefault is a type id of rowset class
	TYPEDefault = "default"
)

var (
	buildHandlers = map[string]func(*sql.Tx) (Interface, error){}
)

func init() {
	Register(TYPEDefault, NewDefaultTransaction)
}

// Interface represents transaction interface
type Interface interface {
	SetContext(ctx context.Context) error
	Commit() error
	Rollback() error
}

// NewTransaction creates a new rowset
func NewTransaction(transactionType string, tx *sql.Tx) (Interface, error) {
	if f, ok := buildHandlers[transactionType]; ok {
		return f(tx)
	}

	return nil, errors.Errorf("Unrecognized database transaction type \"%v\"", transactionType)
}

// Register registers a handler for database rowset creation
func Register(transactionType string, handler func(*sql.Tx) (Interface, error)) {
	buildHandlers[transactionType] = handler
}

// Transaction represents database transaction
type Transaction struct {
	tx  *sql.Tx
	ctx context.Context
}

// SetContext sets the transaction context
func (t *Transaction) SetContext(ctx context.Context) error {
	t.ctx = ctx
	return nil
}

// Commit commits a transaction
func (t *Transaction) Commit() error {
	err := t.tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Transaction commit error")
	}

	return nil
}

// Rollback roll back a transaction
func (t *Transaction) Rollback() error {
	err := t.tx.Rollback()
	if err != nil {
		return errors.Wrap(err, "Transaction roll back error")
	}

	return nil
}

// NewDefaultTransaction creates default transaction
func NewDefaultTransaction(tx *sql.Tx) (Interface, error) {
	return &Transaction{
		tx:  tx,
		ctx: context.Background(),
	}, nil
}
