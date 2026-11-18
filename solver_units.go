package pybot

import (
	"math/big"
	"regexp"
	"strings"
)

type unitDefinition struct {
	dimension string
	symbol    string
	factor    string // factor to the base unit; temperature units use conversion helpers
}

var (
	unitLexeme          = `(?:degrees?\s+(?:[cf]|celsius|fahrenheit)|°\s?[cfk]|°?[a-z]+)`
	unitPattern         = regexp.MustCompile(`(?i)\b(-?\d+(?:\.\d+)?)\s*(` + unitLexeme + `)\s+(?:to|into|in)\s+(` + unitLexeme + `)\b`)
	unitQuestionPattern = regexp.MustCompile(`(?i)\bhow many\s+(` + unitLexeme + `)\s+(?:are|is)?\s+in\s+(-?\d+(?:\.\d+)?)\s*(` + unitLexeme + `)\b`)
)

var unitAliases = map[string]unitDefinition{
	"mm": {dimension: "length", symbol: "mm", factor: "0.001"}, "millimeter": {dimension: "length", symbol: "mm", factor: "0.001"}, "millimeters": {dimension: "length", symbol: "mm", factor: "0.001"}, "millimetre": {dimension: "length", symbol: "mm", factor: "0.001"}, "millimetres": {dimension: "length", symbol: "mm", factor: "0.001"},
	"cm": {dimension: "length", symbol: "cm", factor: "0.01"}, "centimeter": {dimension: "length", symbol: "cm", factor: "0.01"}, "centimeters": {dimension: "length", symbol: "cm", factor: "0.01"}, "centimetre": {dimension: "length", symbol: "cm", factor: "0.01"}, "centimetres": {dimension: "length", symbol: "cm", factor: "0.01"},
	"m": {dimension: "length", symbol: "m", factor: "1"}, "meter": {dimension: "length", symbol: "m", factor: "1"}, "meters": {dimension: "length", symbol: "m", factor: "1"}, "metre": {dimension: "length", symbol: "m", factor: "1"}, "metres": {dimension: "length", symbol: "m", factor: "1"},
	"km": {dimension: "length", symbol: "km", factor: "1000"}, "kilometer": {dimension: "length", symbol: "km", factor: "1000"}, "kilometers": {dimension: "length", symbol: "km", factor: "1000"}, "kilometre": {dimension: "length", symbol: "km", factor: "1000"}, "kilometres": {dimension: "length", symbol: "km", factor: "1000"},
	"in": {dimension: "length", symbol: "in", factor: "0.0254"}, "inch": {dimension: "length", symbol: "in", factor: "0.0254"}, "inches": {dimension: "length", symbol: "in", factor: "0.0254"},
	"ft": {dimension: "length", symbol: "ft", factor: "0.3048"}, "foot": {dimension: "length", symbol: "ft", factor: "0.3048"}, "feet": {dimension: "length", symbol: "ft", factor: "0.3048"},
	"yd": {dimension: "length", symbol: "yd", factor: "0.9144"}, "yard": {dimension: "length", symbol: "yd", factor: "0.9144"}, "yards": {dimension: "length", symbol: "yd", factor: "0.9144"},
	"mi": {dimension: "length", symbol: "mi", factor: "1609.344"}, "mile": {dimension: "length", symbol: "mi", factor: "1609.344"}, "miles": {dimension: "length", symbol: "mi", factor: "1609.344"},
	"mg": {dimension: "mass", symbol: "mg", factor: "0.000001"}, "milligram": {dimension: "mass", symbol: "mg", factor: "0.000001"}, "milligrams": {dimension: "mass", symbol: "mg", factor: "0.000001"},
	"g": {dimension: "mass", symbol: "g", factor: "0.001"}, "gram": {dimension: "mass", symbol: "g", factor: "0.001"}, "grams": {dimension: "mass", symbol: "g", factor: "0.001"},
	"kg": {dimension: "mass", symbol: "kg", factor: "1"}, "kilogram": {dimension: "mass", symbol: "kg", factor: "1"}, "kilograms": {dimension: "mass", symbol: "kg", factor: "1"},
	"oz": {dimension: "mass", symbol: "oz", factor: "0.028349523125"}, "ounce": {dimension: "mass", symbol: "oz", factor: "0.028349523125"}, "ounces": {dimension: "mass", symbol: "oz", factor: "0.028349523125"},
	"lb": {dimension: "mass", symbol: "lb", factor: "0.45359237"}, "lbs": {dimension: "mass", symbol: "lb", factor: "0.45359237"}, "pound": {dimension: "mass", symbol: "lb", factor: "0.45359237"}, "pounds": {dimension: "mass", symbol: "lb", factor: "0.45359237"},
	"ml": {dimension: "volume", symbol: "mL", factor: "0.001"}, "milliliter": {dimension: "volume", symbol: "mL", factor: "0.001"}, "milliliters": {dimension: "volume", symbol: "mL", factor: "0.001"}, "millilitre": {dimension: "volume", symbol: "mL", factor: "0.001"}, "millilitres": {dimension: "volume", symbol: "mL", factor: "0.001"},
	"l": {dimension: "volume", symbol: "L", factor: "1"}, "liter": {dimension: "volume", symbol: "L", factor: "1"}, "liters": {dimension: "volume", symbol: "L", factor: "1"}, "litre": {dimension: "volume", symbol: "L", factor: "1"}, "litres": {dimension: "volume", symbol: "L", factor: "1"},
	"gallon": {dimension: "volume", symbol: "US gal", factor: "3.785411784"}, "gallons": {dimension: "volume", symbol: "US gal", factor: "3.785411784"},
	"c": {dimension: "temperature", symbol: "°C"}, "°c": {dimension: "temperature", symbol: "°C"}, "celsius": {dimension: "temperature", symbol: "°C"}, "degree c": {dimension: "temperature", symbol: "°C"}, "degrees c": {dimension: "temperature", symbol: "°C"}, "degree celsius": {dimension: "temperature", symbol: "°C"}, "degrees celsius": {dimension: "temperature", symbol: "°C"},
	"f": {dimension: "temperature", symbol: "°F"}, "°f": {dimension: "temperature", symbol: "°F"}, "fahrenheit": {dimension: "temperature", symbol: "°F"}, "degree f": {dimension: "temperature", symbol: "°F"}, "degrees f": {dimension: "temperature", symbol: "°F"}, "degree fahrenheit": {dimension: "temperature", symbol: "°F"}, "degrees fahrenheit": {dimension: "temperature", symbol: "°F"},
	"k": {dimension: "temperature", symbol: "K"}, "°k": {dimension: "temperature", symbol: "K"}, "kelvin": {dimension: "temperature", symbol: "K"},
}

func solveUnitConversion(text string) (Response, bool) {
	var amountText, fromText, toText string
	if match := unitQuestionPattern.FindStringSubmatch(text); len(match) == 4 {
		toText, amountText, fromText = match[1], match[2], match[3]
	} else if match := unitPattern.FindStringSubmatch(text); len(match) == 4 {
		amountText, fromText, toText = match[1], match[2], match[3]
	} else {
		if strings.Contains(text, "convert") || strings.Contains(text, "unit conversion") {
			return Response{Text: "Give me a number, a supported source unit, and a target unit. I support common length, mass, volume, and temperature units.", Module: "reasoning/unit-conversion", NeedsClarification: true}, true
		}
		return Response{}, false
	}

	from, fromOK := unitAliases[normalizeUnit(fromText)]
	to, toOK := unitAliases[normalizeUnit(toText)]
	if !fromOK || !toOK {
		return Response{Text: "I don't recognize one of those units. I support common length, mass, volume, and temperature units.", Module: "reasoning/unit-conversion", NeedsClarification: true}, true
	}
	if from.dimension != to.dimension {
		return Response{Text: "Those units measure different things, so I can't convert between them.", Module: "reasoning/unit-conversion", NeedsClarification: true}, true
	}
	value := new(big.Rat)
	if _, ok := value.SetString(amountText); !ok {
		return Response{Text: "I couldn't read that quantity as a number.", Module: "reasoning/unit-conversion", NeedsClarification: true}, true
	}
	converted, ok := convertUnit(value, from, to)
	if !ok {
		return Response{Text: "I couldn't convert that quantity.", Module: "reasoning/unit-conversion", NeedsClarification: true}, true
	}
	return Response{Text: formatDecimal(value) + " " + from.symbol + " = " + formatDecimal(converted) + " " + to.symbol + ".", Module: "reasoning/unit-conversion"}, true
}

func normalizeUnit(unit string) string {
	unit = strings.ToLower(strings.TrimSpace(unit))
	unit = strings.ReplaceAll(unit, "° ", "°")
	return unit
}

func convertUnit(value *big.Rat, from, to unitDefinition) (*big.Rat, bool) {
	if from.dimension == "temperature" {
		celsius := temperatureToCelsius(value, from.symbol)
		if celsius == nil {
			return nil, false
		}
		return celsiusToTemperature(celsius, to.symbol), true
	}
	fromFactor := new(big.Rat)
	toFactor := new(big.Rat)
	fromFactor.SetString(from.factor)
	toFactor.SetString(to.factor)
	base := new(big.Rat).Mul(value, fromFactor)
	return base.Quo(base, toFactor), true
}

func temperatureToCelsius(value *big.Rat, symbol string) *big.Rat {
	switch symbol {
	case "°C":
		return new(big.Rat).Set(value)
	case "°F":
		result := new(big.Rat).Sub(value, big.NewRat(32, 1))
		return result.Mul(result, big.NewRat(5, 9))
	case "K":
		return new(big.Rat).Sub(value, big.NewRat(27315, 100))
	default:
		return nil
	}
}

func celsiusToTemperature(value *big.Rat, symbol string) *big.Rat {
	switch symbol {
	case "°C":
		return new(big.Rat).Set(value)
	case "°F":
		result := new(big.Rat).Mul(value, big.NewRat(9, 5))
		return result.Add(result, big.NewRat(32, 1))
	case "K":
		return new(big.Rat).Add(value, big.NewRat(27315, 100))
	default:
		return nil
	}
}

func formatDecimal(value *big.Rat) string {
	text := value.FloatString(6)
	return strings.TrimRight(strings.TrimRight(text, "0"), ".")
}
