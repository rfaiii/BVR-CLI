package model

import "testing"

func TestLandingButtonTopAccountsForCWD(t *testing.T) {
	// Buttons appear after the CWD line + 1 blank separator.
	// mainY=10, cwdHeight=1 → 10 + 1 + 1 = 12
	if got := landingButtonTop(10, 1); got != 12 {
		t.Fatalf("landing button top = %d, want 12", got)
	}
}
