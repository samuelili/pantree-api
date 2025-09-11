package models

type Category string

const (
	CategoryUnset   Category = "unset"
	CategoryPantry  Category = "pantry"
	CategoryFridge  Category = "fridge"
	CategoryFreezer Category = "freezer"
)
