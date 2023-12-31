package data

import (
	"database/sql"
)

type Models struct {
	Payment interface {
		InsertTransaction(transaction Transaction) (int, error)
		InsertOrder(order Order) (int, error)
		InsertCustomer(customer Customer) (int, error)
		UpdateOrderStatus(id, statudID int) error
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Payment: PaymentModel{DB: db},
	}
}
