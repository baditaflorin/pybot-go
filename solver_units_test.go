package pybot

import "testing"

func TestUnitConversions(t *testing.T) {
	cases := []struct {
		question string
		want     string
	}{
		{"convert 100 centimeters to meters", "100 cm = 1 m."},
		{"how many inches are in 2 meters", "2 m = 78.740157 in."},
		{"convert 20 celsius to fahrenheit", "20 °C = 68 °F."},
		{"convert 1 degree Celsius to degrees Fahrenheit", "1 °C = 33.8 °F."},
		{"convert 1000 grams to kilograms", "1000 g = 1 kg."},
	}
	for _, test := range cases {
		t.Run(test.question, func(t *testing.T) {
			response := New(0).Reply(test.question)
			if response.Module != "reasoning/unit-conversion" || response.Text != test.want {
				t.Fatalf("response = %#v, want %q", response, test.want)
			}
		})
	}
}

func TestUnitConversionRejectsDifferentDimensions(t *testing.T) {
	response := New(0).Reply("convert 5 kilograms to meters")
	if !response.NeedsClarification {
		t.Fatalf("expected a clarification, got %#v", response)
	}
}

func TestUnknownUnitAsksForClarification(t *testing.T) {
	response := New(0).Reply("convert 5 furlongs to meters")
	if !response.NeedsClarification {
		t.Fatalf("expected unsupported-unit clarification, got %#v", response)
	}
}
