package models

type QuantityUnit string

const (
	// universal
	QuantityUnset QuantityUnit = "unset"
	QuantityCount QuantityUnit = "count"

	// metric
	QuantityGrams       QuantityUnit = "grams"
	QuantityMilliliters QuantityUnit = "milliliters"
	QuantityLiters      QuantityUnit = "liters"

	// imperial
	QuantityTeaspoon    QuantityUnit = "teaspoon"
	QuantityTablespoon  QuantityUnit = "tablespoon"
	QuantityFluidOunces QuantityUnit = "fluid_ounces"
	QuantityCup         QuantityUnit = "cup"
	QuantityPint        QuantityUnit = "pint"
	QuantityQuart       QuantityUnit = "quart"

	QuantityOunces QuantityUnit = "ounces"
	QuantityPounds QuantityUnit = "pounds"
)
