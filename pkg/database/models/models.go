// Package models holds the shared domain types that the gRPC adapters under
// pkg/adapter return to their consumers. Keeping them in one place lets a
// service depend on a single domain shape regardless of which remote service
// produced the data.
package models

import "time"

// Merchant is the domain view of a merchant produced by the merchant adapter.
type Merchant struct {
	MerchantID   int32
	UserID       int32
	Name         string
	Description  *string
	Address      *string
	ContactEmail *string
	ContactPhone *string
	Status       string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

// Cashier is the domain view of a cashier produced by the cashier adapter.
type Cashier struct {
	CashierID  int32
	MerchantID int32
	Name       string
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

// Order is the domain view of an order produced by the order adapter.
type Order struct {
	OrderID    int32
	MerchantID int32
	CashierID  int32
	TotalPrice int64
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

// OrderItem is the domain view of an order item produced by the order item
// adapter.
type OrderItem struct {
	OrderItemID int32
	OrderID     int32
	ProductID   int32
	Quantity    int32
	Price       int64
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// Product is the domain view of a product produced by the product adapter.
type Product struct {
	ProductID    int32
	MerchantID   int32
	CategoryID   int32
	Name         string
	Description  *string
	Price        int32
	CountInStock int32
	Brand        *string
	Weight       *int32
	SlugProduct  *string
	ImageProduct *string
	Barcode      *string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

// Transaction is the domain view of a transaction produced by the transaction
// adapter. PaymentStatus is nullable because the payment may still be pending.
type Transaction struct {
	TransactionID int32
	OrderID       int32
	MerchantID    int32
	PaymentMethod string
	Amount        int32
	PaymentStatus *string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

// Category is the domain view of a category produced by the category adapter.
type Category struct {
	CategoryID   int32
	Name         string
	Description  *string
	SlugCategory *string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

// User is the domain view of a user produced by the user adapter. Password is
// only populated by the auth-oriented lookups.
type User struct {
	UserID    int32
	Firstname string
	Lastname  string
	Email     string
	Password  string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// Role is the domain view of a role produced by the role adapter.
type Role struct {
	RoleID    int32
	RoleName  string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// UserRole is the domain view of a user-role assignment produced by the role
// adapter.
type UserRole struct {
	UserRoleID int32
	UserID     int32
	RoleID     int32
}
