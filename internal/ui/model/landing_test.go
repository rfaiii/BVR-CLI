package model

import "testing"

func TestLandingButtonTopAccountsForCWD(t *testing.T) {
	// Buttons appear after the CWD line + 1 blank separator.
	// mainY=10, cwdHeight=1 → 10 + 1 + 1 + 1 = 13
	if got := landingButtonTop(10, 1); got != 13 {
		t.Fatalf("landing button top = %d, want 13", got)
	}
}
