---
phase: 02-核心功能开发
plan: 04
subsystem: membership, billing, routing
tags: [membership, billing, model-router, profit-optimization]
dependency_graph:
  requires: [02-01]
  provides: [membership-tier-management, monthly-billing, profit-based-routing]
  affects: [api-handlers, cost-optimization]
tech_stack:
  added: []
  patterns: [service-layer, gorm-model, profit-calculation, bill-aggregation]
key_files:
  created:
    - internal/services/membership.go
    - internal/services/billing.go
    - internal/services/model_router.go
  modified:
    - internal/models/models.go
    - internal/database/database.go
decisions: []
metrics:
  duration: "15 minutes"
  completed_date: "2026-05-08"
---

# Phase 02 Plan 04: Membership, Billing & Routing Services Summary

**One-liner:** Membership tier management with per-model discounts, monthly bill generation with PDF/CSV export stubs, and profit-based supplier selection router.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed UserMembership preload issue**
- **Found during:** Task 1
- **Issue:** UserMembership model has no MembershipTier field for Preload
- **Fix:** Query MembershipTier separately instead of using Preload
- **Files modified:** internal/services/membership.go
- **Commit:** 49556a3

**2. [Rule 3 - Fix] Added Bill and BillModelDetail models before billing service**
- **Found during:** Task 2 compilation
- **Issue:** BillingService references undefined models.Bill
- **Fix:** Created Bill and BillModelDetail models in models.go before compiling billing service
- **Files modified:** internal/models/models.go, internal/database/database.go
- **Commit:** a1c78b2

**3. [Rule 3 - Fix] Fixed BillModelDetail pointer/value append issue**
- **Found during:** Task 2 compilation
- **Issue:** append expects value not pointer for BillModelDetail
- **Fix:** Dereference stat pointer when appending to bill.Items
- **Files modified:** internal/services/billing.go
- **Commit:** a1c78b2

**4. [Rule 3 - Fix] Fixed unused bill variable in ExportBillAsPDF**
- **Found during:** Task 2 compilation
- **Issue:** Variable declared but not used (TODO implementation)
- **Fix:** Added bill fields to log statement to use the variable
- **Files modified:** internal/services/billing.go
- **Commit:** a1c78b2

## Authentication Gates

None encountered during this plan.

## Known Stubs

**1. PDF Export (billing.go)**
- File: internal/services/billing.go
- Lines: 145-177
- Reason: PDF generation library not yet integrated
- TODO: Use github.com/jung-kurt/gofpdf or github.com/signintech/gopdf

**2. CSV Export (billing.go)**
- File: internal/services/billing.go
- Lines: 179-217
- Reason: CSV generation not yet implemented
- TODO: Use encoding/csv to generate detailed usage records

**3. Supplier Health Check (model_router.go)**
- File: internal/services/model_router.go
- Lines: 186-193
- Reason: Health monitoring infrastructure not yet built
- TODO: Check error rate, response time, QPS limits

**4. QPS Overload Check (model_router.go)**
- File: internal/services/model_router.go
- Lines: 195-203
- Reason: Real-time QPS tracking not yet implemented
- TODO: Track current QPS per route

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: tampering | internal/services/membership.go | UpdateUserMembership affects pricing - requires audit logging |
| threat_flag: information_disclosure | internal/services/model_router.go | Cost prices must not be exposed in API responses |
| threat_flag: information_disclosure | internal/services/billing.go | Export functions return sensitive usage data |

## Files Created/Modified

### Created Files
1. **internal/services/membership.go** (223 lines)
   - MembershipService with tier management
   - Per-model discount calculation per D-15
   - Immediate effect membership updates per D-16

2. **internal/services/billing.go** (267 lines)
   - Monthly bill generation per D-17
   - PDF/CSV export stubs per D-18
   - Auto-generate on 1st of month

3. **internal/services/model_router.go** (199 lines)
   - Profit-based supplier selection per D-09
   - EvaluateSupplier with profit margin calculation
   - Health check and QPS overload stubs

### Modified Files
1. **internal/models/models.go**
   - Added RequestID field to UsageRecord for idempotency per D-08
   - Added Bill model with period-based aggregation
   - Added BillItem model for detailed line items
   - Added BillModelDetail for JSON responses

2. **internal/database/database.go**
   - Registered Bill and BillItem models in AutoMigrate

## Commits

1. **49556a3** feat(02-04): create Membership service with discount calculation
2. **a1c78b2** feat(02-04): create Billing service with PDF/CSV export and Bill model
3. **2db1d6b** feat(02-04): create Model Router with profit-based selection

## Self-Check: PASSED

All files created, all commits verified, all code compiles successfully.
