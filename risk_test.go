package main

import "testing"

func TestRiskDecision(t *testing.T) {
	for _, x := range []struct {
		p *PaymentEvent
		w string
	}{{nil, "standard-review"}, {&PaymentEvent{Amount: 100, Currency: "USD"}, "standard-review"}, {&PaymentEvent{Amount: 100000, Currency: "USD"}, "manual-review"}, {&PaymentEvent{Amount: 1, Currency: "EUR"}, "manual-review"}} {
		if g := riskDecision(x.p); g != x.w {
			t.Errorf("got %s want %s", g, x.w)
		}
	}
}
