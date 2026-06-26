// Package obbbabreaks implements the family of tax breaks governed by the One
// Big Beautiful Bill Act (OBBBA, 2025) and related permanent provisions:
// the child tax credit (§24, permanent), the SALT cap (§164(b)(6),
// $40,400 through 2029 then $10,000 from 2030), the no-tax-on-tips deduction
// (§224, 2025-2028), the senior deduction (§512, 2025-2028), and the
// car-loan interest deduction (§163(h)(4), 2025-2028). It is the same math
// behind the taxbreakcalc.com multi-provision tax-break calculator suite.
//
// Each break is modeled independently with its own cap, phase-out, and
// (where applicable) credit-vs-deduction character, so a caller can sum
// Savings across provisions without double-counting caps.
package obbbabreaks

// Filing status selects per-provision caps and phase-out thresholds.
type FilingStatus int

const (
	Single FilingStatus = iota
	Joint
	HeadOfHousehold
)

// Break identifies a provision.
type Break int

const (
	BreakChildTaxCredit Break = iota
	BreakSALTCap
	BreakTipsDeduction
	BreakSeniorDeduction
	BreakCarLoanInterest
)

// Result is the per-break outcome.
type Result struct {
	Break       Break
	Eligible    bool    // gross amount before cap and phase-out
	GrossAmount float64 // the provision's full statutory amount
	Capped      float64 // after the per-return/per-child cap
	Deductible  float64 // after MAGI phase-out (the actual claimable amount)
	IsCredit    bool    // true for credits (dollar-for-dollar); false for deductions
	MarginalRate float64 // rate applied to deductions
	Savings     float64 // federal tax value (credit: dollar-for-dollar; deduction: x marginal rate)
}

// --- Child Tax Credit (§24, PERMANENT under OBBBA) ---
//
// $2,200 per qualifying child (indexed), non-refundable portion up to $1,800
// with phase-in; phase-out $200,000 single / $400,000 MFJ (-$50 per $1,000
// over). PERMANENT — does NOT sunset 2028.

const (
	ctcPerChild      = 2_200.0
	ctcPhaseoutS     = 200_000.0
	ctcPhaseoutJ     = 400_000.0
	ctcPhaseoutSlope = 50.0 // -$50 per $1,000 over
)

// ChildTaxCredit returns the §24 child tax credit for `children` qualifying
// children at a given MAGI. The full $2,200/child is modeled (the refundable
// split is upstream bookkeeping).
func ChildTaxCredit(children int, fs FilingStatus, magi float64) Result {
	gross := float64(children) * ctcPerChild
	start := ctcPhaseoutS
	if fs == Joint {
		start = ctcPhaseoutJ
	}
	capped := gross
	deductible := gross
	eligible := children > 0
	if magi > start {
		reduction := ctcPhaseoutSlope * float64(int((magi-start)/1_000.0))
		deductible = gross - reduction
		if deductible < 0 {
			deductible = 0
		}
	}
	return Result{
		Break:        BreakChildTaxCredit,
		Eligible:     eligible,
		GrossAmount:  gross,
		Capped:       capped,
		Deductible:   deductible,
		IsCredit:     true,
		MarginalRate: 0,
		Savings:      deductible, // credit is dollar-for-dollar
	}
}

// --- SALT cap (§164(b)(6)) ---
//
// $40,400 (2025-2029, indexed), reverts to $10,000 in 2030. No MAGI phase-out.
// This models the itemized-deduction CAP, not the underlying property/income
// tax paid.

const (
	saltCap2025to2029 = 40_400.0
	saltCap2030onward = 10_000.0
)

// SALTCap returns the capped SALT deduction. Pass year to select the cap
// (2025-2029 -> $40,400; 2030+ -> $10,000).
func SALTCap(stateLocalTaxPaid float64, year int) Result {
	cap := saltCap2025to2029
	if year >= 2030 {
		cap = saltCap2030onward
	}
	amount := stateLocalTaxPaid
	if amount > cap {
		amount = cap
	}
	if amount < 0 {
		amount = 0
	}
	return Result{
		Break:       BreakSALTCap,
		Eligible:    stateLocalTaxPaid > 0,
		GrossAmount: stateLocalTaxPaid,
		Capped:      amount,
		Deductible:  amount,
	}
}

// --- No-tax-on-tips deduction (§224, 2025-2028) ---
//
// $25,000 FLAT per return (marriage penalty — NOT $50k MFJ). Same MAGI
// phase-out as overtime (§225): -$100 per $1,000 over $150k/$300k, fully $0
// at $275k/$550k.

const (
	tipsCapSingle = 25_000.0
	tipsCapJoint  = 25_000.0 // flat — marriage penalty
	tipsStartS    = 150_000.0
	tipsStartJ    = 300_000.0
	tipsEndS      = 275_000.0
	tipsEndJ      = 550_000.0
)

// TipsDeduction returns the §224 above-the-line deduction for qualified tips.
// $25,000 flat per return; MAGI phase-out -100/$1k over $150k/$300k.
func TipsDeduction(qualifiedTips float64, fs FilingStatus, magi, marginalRate float64) Result {
	cap := tipsCapSingle
	if fs == Joint {
		cap = tipsCapJoint
	}
	capped := qualifiedTips
	if capped > cap {
		capped = cap
	}
	start := tipsStartS
	end := tipsEndS
	if fs == Joint {
		start = tipsStartJ
		end = tipsEndJ
	}
	deductible := capped
	if magi >= end {
		deductible = 0
	} else if magi > start {
		reduction := 100.0 * float64(int((magi-start)/1_000.0))
		deductible = capped - reduction
		if deductible < 0 {
			deductible = 0
		}
	}
	return Result{
		Break:        BreakTipsDeduction,
		Eligible:     qualifiedTips > 0,
		GrossAmount:  qualifiedTips,
		Capped:       capped,
		Deductible:   deductible,
		IsCredit:     false,
		MarginalRate: marginalRate,
		Savings:      deductible * marginalRate,
	}
}

// --- Senior deduction (§512, 2025-2028) ---
//
// +$6,000 single / +$12,000 MFJ (both spouses 65+). Stacks on the existing
// additional standard deduction for the elderly. Phase-out ~6% over
// $75,000 single / $150,000 MFJ -> -$360 single / -$720 MFJ per $1,000 over,
// fully gone at ~$91,667/$175,000. This models the OBBBA additional amount.

const (
	seniorSingle = 6_000.0
	seniorJoint  = 12_000.0
	seniorStartS = 75_000.0
	seniorStartJ = 150_000.0
)

// SeniorDeduction returns the §512 additional senior deduction. spouses65
// is the count of spouses aged 65+ on the return (1 or 2; for single filers
// pass 1).
func SeniorDeduction(spouses65 int, fs FilingStatus, magi, marginalRate float64) Result {
	gross := seniorSingle
	start := seniorStartS
	if fs == Joint && spouses65 >= 2 {
		gross = seniorJoint
		start = seniorStartJ
	} else if fs == Joint && spouses65 == 1 {
		gross = seniorSingle
	}
	eligible := spouses65 > 0
	deductible := gross
	if magi > start {
		// ~6% of the gross amount reduced per $1,000 over the threshold.
		per := 0.06 * gross * float64(int((magi-start)/1_000.0))
		deductible = gross - per
		if deductible < 0 {
			deductible = 0
		}
	}
	return Result{
		Break:        BreakSeniorDeduction,
		Eligible:     eligible,
		GrossAmount:  gross,
		Capped:       gross,
		Deductible:   deductible,
		IsCredit:     false,
		MarginalRate: marginalRate,
		Savings:      deductible * marginalRate,
	}
}

// --- Car-loan interest deduction (§163(h)(4), 2025-2028) ---
//
// Up to $10,000 of INTEREST on a loan for a NEW vehicle, US final assembly,
// under 14,000 lbs. No MSRP cap. Phase-out -$200 per $1,000 MAGI over
// $100,000 single / $200,000 MFJ -> fully gone at $150,000/$250,000.

const (
	carLoanCap      = 10_000.0
	carLoanStartS   = 100_000.0
	carLoanStartJ   = 200_000.0
	carLoanEndS     = 150_000.0
	carLoanEndJ     = 250_000.0
	carLoanSlope    = 200.0 // -$200 per $1,000 over
)

// CarLoanInterestDeduction returns the §163(h)(4) car-loan interest
// deduction. Requires isNewVehicleUSAAssembly=true to be eligible (the statute
// gates the deduction on a new, US-assembled vehicle under 14,000 lbs).
func CarLoanInterestDeduction(interestPaid float64, isNewVehicleUSAAssembly bool, fs FilingStatus, magi, marginalRate float64) Result {
	if !isNewVehicleUSAAssembly || interestPaid <= 0 {
		return Result{Break: BreakCarLoanInterest, Eligible: false}
	}
	capped := interestPaid
	if capped > carLoanCap {
		capped = carLoanCap
	}
	start := carLoanStartS
	end := carLoanEndS
	if fs == Joint {
		start = carLoanStartJ
		end = carLoanEndJ
	}
	deductible := capped
	if magi >= end {
		deductible = 0
	} else if magi > start {
		reduction := carLoanSlope * float64(int((magi-start)/1_000.0))
		deductible = capped - reduction
		if deductible < 0 {
			deductible = 0
		}
	}
	return Result{
		Break:        BreakCarLoanInterest,
		Eligible:     true,
		GrossAmount:  interestPaid,
		Capped:       capped,
		Deductible:   deductible,
		IsCredit:     false,
		MarginalRate: marginalRate,
		Savings:      deductible * marginalRate,
	}
}
