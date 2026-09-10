package model

import "testing"

func TestLandingButtonTopAccountsForHero(t *testing.T) {
	if got := landingButtonTop(10, 1, 9); got != 23 {
		t.Fatalf("landing button top = %d, want 23", got)
	}
}
