package models

type Recipe struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	Ingredients []string `json:"ingredients"`
}

type RecipeIngredients struct {
	IngredientID string  `json:"ingredientId"`
	Quantity     float64 `json:"quantity"`
}
