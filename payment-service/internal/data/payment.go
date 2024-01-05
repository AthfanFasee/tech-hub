package data

import (
	"context"
	"database/sql"
	"time"
)

type PaymentModel struct {
	DB *sql.DB
}

// TransactionStatus is the type for transaction statuses
type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transaction is the type for transactions
type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusID int       `json:"transaction_status_id"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

// Customer is the type for customers
type Customer struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// Order is the type for orders
type Order struct {
	ID            int  `json:"id"`
	TransactionID int  `json:"transaction_id"`
	CustomerID    int  `json:"customer_id"`
	StatusID      int  `json:"status_id"`
	IsRecurring   bool `json:"is_recurring"`
	Amount        int  `json:"amount"`
}

// Inserts a transaction
func (p PaymentModel) InsertTransaction(transaction Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO transactions (amount, currency, last_four, bank_return_code, 
		expiry_month, expiry_year, payment_intent, payment_method, transaction_status_id)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	args := []interface{}{transaction.Amount, transaction.Currency, transaction.LastFour, transaction.BankReturnCode, transaction.ExpiryMonth, transaction.ExpiryYear, transaction.PaymentIntent, transaction.PaymentMethod, transaction.TransactionStatusID}

	result, err := p.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Inserts a new order, and returns its id
func (p PaymentModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO orders (transaction_id, status_id, customer_id, is_recurring, amount)
		VALUES (?, ?, ?, ?, ?)`

	result, err := p.DB.ExecContext(ctx, stmt,
		order.TransactionID,
		order.StatusID,
		order.CustomerID,
		order.IsRecurring,
		order.Amount,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Inserts a new order, and returns its id
func (p PaymentModel) InsertCustomer(customer Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO customers (first_name, last_name, email, created_at, updated_at)
		values (?, ?, ?)`

	result, err := p.DB.ExecContext(ctx, stmt,
		customer.FirstName,
		customer.LastName,
		customer.Email,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Updates order status
func (p PaymentModel) UpdateOrderStatus(id, statusID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE orders set status_id = ? WHERE id = ?`

	_, err := p.DB.ExecContext(ctx, stmt, statusID, id)
	if err != nil {
		return err
	}

	return nil
}
