# obbba-breaks

[![Go Reference](https://pkg.go.dev/badge/github.com/theluckystrike/obbba-breaks.svg)](https://pkg.go.dev/github.com/theluckystrike/obbba-breaks)

The family of tax breaks governed by the **One Big Beautiful Bill Act (OBBBA,
2025)** and related permanent provisions, modeled independently so caps and
phase-outs never double-count:

- **Child Tax Credit (§24)** — $2,200/child, **PERMANENT** (does not sunset 2028).
- **SALT cap (§164(b)(6))** — $40,400 through 2029, reverts to **$10,000 in 2030**.
- **No-tax-on-tips deduction (§224)** — $25,000 flat per return (marriage penalty), 2025-2028.
- **Senior deduction (§512)** — +$6,000 single / +$12,000 MFJ (both 65+), 2025-2028.
- **Car-loan interest deduction (§163(h)(4))** — $10,000, NEW US-assembled vehicle, 2025-2028.

Pure, dependency-free Go math — the same engine behind the
[taxbreakcalc.com](https://taxbreakcalc.com/) tax-break calculator suite.

## Install

```sh
go get github.com/theluckystrike/obbba-breaks
```

## Quick example

```go
package main

import (
	"fmt"

	obbbabreaks "github.com/theluckystrike/obbba-breaks"
)

func main() {
	// 2 kids, single, MAGI $80k
	ctc := obbbabreaks.ChildTaxCredit(2, obbbabreaks.Single, 80_000.0)
	fmt.Println(ctc.Savings) // 4400 (dollar-for-dollar credit)

	// SALT: $50k paid in 2026 -> capped at $40,400
	salt := obbbabreaks.SALTCap(50_000.0, 2026)
	fmt.Println(salt.Deductible) // 40400

	// Tips: $30k tips, single, 22% bracket -> capped $25k
	tips := obbbabreaks.TipsDeduction(30_000.0, obbbabreaks.Single, 0.0, 0.22)
	fmt.Println(tips.Savings) // 5500 (25000 * 0.22)
}
```

## API

| Function | Description |
| --- | --- |
| `ChildTaxCredit(children int, fs FilingStatus, magi float64) Result` | §24 credit, $2,200/child, phase-out $200k/$400k. |
| `SALTCap(stateLocalTaxPaid float64, year int) Result` | §164(b)(6) itemized cap ($40,400 → $10,000 in 2030). |
| `TipsDeduction(qualifiedTips float64, fs FilingStatus, magi, marginalRate float64) Result` | §224 above-the-line deduction, $25k flat cap. |
| `SeniorDeduction(spouses65 int, fs FilingStatus, magi, marginalRate float64) Result` | §512 additional senior deduction. |
| `CarLoanInterestDeduction(interestPaid float64, isNewVehicleUSAAssembly bool, fs FilingStatus, magi, marginalRate float64) Result` | §163(h)(4) car-loan interest; gated on new US-assembled vehicle. |

Each `Result` reports `GrossAmount`, `Capped`, `Deductible` (after phase-out),
`IsCredit`, and `Savings` (federal tax value), so callers can sum savings
across provisions without double-counting caps.

## Verification

All figures verified against IRS / Joint Committee on Taxation / Tax
Foundation primary sources. Note: **energy/EV credits (§25D/§30D/§25E) were
TERMINATED by OBBBA for 2026** and are deliberately NOT modeled here.

## License

MIT.

## Links

- Tax-break calculator suite: <https://taxbreakcalc.com/>
- Site: <https://taxbreakcalc.com/>
- Package docs: <https://pkg.go.dev/github.com/theluckystrike/obbba-breaks>
