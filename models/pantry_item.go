package models

type PantryItem struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Quantity     float64      `json:"quantity"`
	Unit         QuantityUnit `json:"unit"`
	Price        float64      `json:"price"`
	ExpirationMs string       `json:"expirationMs"`
	CategoryID   string       `json:"categoryId"`
	UserID       string       `json:"userId"`
}
