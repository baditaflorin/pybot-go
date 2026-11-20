package pybot

import "testing"

func TestArithmeticEvaluation(t *testing.T) {
	cases := []struct {
		question string
		want     string
	}{
		{"What is 12 plus 50?", "62"},
		{"calculate 8 times 7", "56"},
		{"compute (3 + 4) * 2", "14"},
		{"what is 24 divided by 8", "3"},
		{"please add 7 and 5", "12"},
		{"what is the difference between 10 and 3?", "7"},
		{"compute (-5 + 3) * 2", "-4"},
		{"-5 + 3", "-2"},
	}
	for _, test := range cases {
		t.Run(test.question, func(t *testing.T) {
			response := New(0).Reply(test.question)
			if response.Module != "reasoning/arithmetic" || response.Text != test.want {
				t.Fatalf("response = %#v, want %q", response, test.want)
			}
		})
	}
}

func TestArithmeticDivisionByZeroClarifies(t *testing.T) {
	response := New(0).Reply("what is 9 / 0")
	if !response.NeedsClarification || response.Text != "Division by zero is undefined." {
		t.Fatalf("response = %#v", response)
	}
}

func TestSquareRootUsesPreciseCPUResult(t *testing.T) {
	response := New(0).Reply("What is the square root of 2?")
	if response.Module != "reasoning/arithmetic" || response.Text != "≈ 1.414213562373" {
		t.Fatalf("response = %#v, want a precise square-root approximation", response)
	}
}

func TestSquareRootOfPerfectSquareStaysExact(t *testing.T) {
	response := New(0).Reply("Find the square root of 81")
	if response.Module != "reasoning/arithmetic" || response.Text != "9" {
		t.Fatalf("response = %#v, want exact result 9", response)
	}
}

func TestSquareRootCanAppearInsideExpression(t *testing.T) {
	response := New(0).Reply("sqrt(2) * sqrt(2)")
	if response.Module != "reasoning/arithmetic" || response.Text != "≈ 2" {
		t.Fatalf("response = %#v, want approximate result 2", response)
	}
}

func TestOrdinaryAndDoesNotBecomeAddition(t *testing.T) {
	response := New(0).Reply("I like bread and butter")
	if response.Module == "reasoning/arithmetic" {
		t.Fatalf("ordinary sentence routed to arithmetic: %#v", response)
	}
}
