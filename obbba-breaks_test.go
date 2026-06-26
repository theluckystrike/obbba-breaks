package obbbabreaks

import (
	"math"
	"testing"
)

func approx(a, b, tol float64) bool { return math.Abs(a-b) < tol }

// Child tax credit: 2 children, single, MAGI $80k (under $200k phase-out).
// Gross = 2 * 2200 = 4400, full credit, dollar-for-dollar savings = 4400.
func TestChildTaxCreditFull(t *testing.T) {
	r := ChildTaxCredit(2, Single, 80_000.0)
	if !approx(r.GrossAmount, 4_400.0, 1e-9) {
		t.Fatalf("gross = %v, want 4400", r.GrossAmount)
	}
	if !approx(r.Deductible, 4_400.0, 1e-9) {
		t.Fatalf("deductible = %v, want 4400 (under phase-out)", r.Deductible)
	}
	if !r.IsCredit {
		t.Fatal("CTC must be a credit (dollar-for-dollar)")
	}
	if !approx(r.Savings, 4_400.0, 1e-9) {
		t.Fatalf("savings = %v, want 4400 (credit = dollar-for-dollar)", r.Savings)
	}
}

// CTC phase-out: 2 children, single, MAGI $240k.
// reduction = 50 * int((240000-200000)/1000) = 50*40 = 2000 -> deductible 2400.
func TestChildTaxCreditPhaseout(t *testing.T) {
	r := ChildTaxCredit(2, Single, 240_000.0)
	if !approx(r.Deductible, 2_400.0, 1e-9) {
		t.Fatalf("deductible = %v, want 2400 after -$2000 phase-out", r.Deductible)
	}
	if !approx(r.Savings, 2_400.0, 1e-9) {
		t.Fatalf("savings = %v, want 2400", r.Savings)
	}
}

// CTC fully phased out by $300k single (2 kids): reduction 5000 > 4400 -> 0.
func TestChildTaxCreditFullyPhasedOut(t *testing.T) {
	r := ChildTaxCredit(2, Single, 300_000.0)
	if r.Deductible != 0.0 {
		t.Fatalf("deductible = %v, want 0 (fully phased out)", r.Deductible)
	}
}

// SALT cap: $50k paid -> capped at $40,400 (2026).
func TestSALTCap2026(t *testing.T) {
	r := SALTCap(50_000.0, 2026)
	if !approx(r.Deductible, 40_400.0, 1e-9) {
		t.Fatalf("SALT = %v, want 40400 (2025-2029 cap)", r.Deductible)
	}
}

// SALT reverts to $10k in 2030.
func TestSALTCap2030Revert(t *testing.T) {
	r := SALTCap(50_000.0, 2030)
	if !approx(r.Deductible, 10_000.0, 1e-9) {
		t.Fatalf("SALT = %v, want 10000 (2030 revert)", r.Deductible)
	}
}

// SALT below cap: $8k paid -> $8k (no cap binding).
func TestSALTCapUnderCap(t *testing.T) {
	r := SALTCap(8_000.0, 2026)
	if !approx(r.Deductible, 8_000.0, 1e-9) {
		t.Fatalf("SALT = %v, want 8000 (under cap)", r.Deductible)
	}
}

// Tips: $30k tips, single, low MAGI -> capped to $25k flat. Savings at 22%.
func TestTipsDeductionCap(t *testing.T) {
	r := TipsDeduction(30_000.0, Single, 0.0, 0.22)
	if !approx(r.Capped, 25_000.0, 1e-9) {
		t.Fatalf("tips capped = %v, want 25000 (flat per-return cap)", r.Capped)
	}
	if !approx(r.Savings, 25_000.0*0.22, 1e-9) {
		t.Fatalf("tips savings = %v, want %v", r.Savings, 25_000.0*0.22)
	}
}

// Tips marriage penalty: MFJ cap is ALSO $25k (not $50k).
func TestTipsDeductionMarriagePenalty(t *testing.T) {
	r := TipsDeduction(30_000.0, Joint, 0.0, 0.22)
	if !approx(r.Capped, 25_000.0, 1e-9) {
		t.Fatalf("MFJ tips capped = %v, want 25000 (marriage penalty)", r.Capped)
	}
}

// Senior deduction single 65+, low MAGI -> $6,000.
func TestSeniorDeductionSingle(t *testing.T) {
	r := SeniorDeduction(1, Single, 40_000.0, 0.22)
	if !approx(r.GrossAmount, 6_000.0, 1e-9) {
		t.Fatalf("senior gross = %v, want 6000", r.GrossAmount)
	}
	if !approx(r.Savings, 6_000.0*0.22, 1e-9) {
		t.Fatalf("senior savings = %v, want %v", r.Savings, 6_000.0*0.22)
	}
}

// Senior MFJ both 65+ -> $12,000.
func TestSeniorDeductionMFJBoth(t *testing.T) {
	r := SeniorDeduction(2, Joint, 100_000.0, 0.24)
	if !approx(r.GrossAmount, 12_000.0, 1e-9) {
		t.Fatalf("senior MFJ gross = %v, want 12000", r.GrossAmount)
	}
}

// Car-loan: $8k interest, new US vehicle, single, MAGI $120k, 22% bracket.
// reduction = 200 * int((120000-100000)/1000) = 200*20 = 4000 -> deductible 4000.
func TestCarLoanInterestDeduction(t *testing.T) {
	r := CarLoanInterestDeduction(8_000.0, true, Single, 120_000.0, 0.22)
	if !approx(r.Capped, 8_000.0, 1e-9) {
		t.Fatalf("car-loan capped = %v, want 8000 (under 10k cap)", r.Capped)
	}
	if !approx(r.Deductible, 4_000.0, 1e-9) {
		t.Fatalf("car-loan deductible = %v, want 4000 after phase-out", r.Deductible)
	}
	if !approx(r.Savings, 4_000.0*0.22, 1e-9) {
		t.Fatalf("car-loan savings = %v, want %v", r.Savings, 4_000.0*0.22)
	}
}

// Car-loan eligibility gate: NOT a new US-assembled vehicle -> ineligible.
func TestCarLoanInterestDeductionIneligible(t *testing.T) {
	r := CarLoanInterestDeduction(8_000.0, false, Single, 50_000.0, 0.22)
	if r.Eligible {
		t.Fatal("car-loan should be INELIGIBLE for non-new-US vehicle")
	}
	if r.Deductible != 0 {
		t.Fatalf("deductible = %v, want 0 (ineligible)", r.Deductible)
	}
}

// Car-loan fully phased out at $150k single.
func TestCarLoanInterestDeductionFullPhaseout(t *testing.T) {
	r := CarLoanInterestDeduction(8_000.0, true, Single, 150_000.0, 0.22)
	if r.Deductible != 0.0 {
		t.Fatalf("deductible = %v, want 0 (fully phased out at $150k)", r.Deductible)
	}
}

// Car-loan $15k interest -> capped at $10k, no phase-out (low MAGI).
func TestCarLoanInterestDeductionCap(t *testing.T) {
	r := CarLoanInterestDeduction(15_000.0, true, Single, 50_000.0, 0.22)
	if !approx(r.Capped, 10_000.0, 1e-9) {
		t.Fatalf("car-loan capped = %v, want 10000", r.Capped)
	}
}
