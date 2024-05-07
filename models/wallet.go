package models

type Wallet struct {
	UserID int `json:"user_id"`
	User   Patient
	Amount float64 `json:"amount"`
}
