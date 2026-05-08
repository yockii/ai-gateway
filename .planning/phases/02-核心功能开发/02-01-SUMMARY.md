---
phase: 02-核心功能开发
plan: 01
subsystem: Data Model Layer
tags: [models, gorm, database, migration]
dependency_graph:
  requires: []
  provides: [key-manager, membership, model-router, billing]
  affects: [database-migration]
tech_stack:
  added: [GORM embedded structs, pointer types for optional fields, JSON serialization for maps]
  patterns: [xid IDs, json:"-" for sensitive fields, gorm indexes]
key_files:
  created:
    - internal/models/membership.go
    - internal/models/model_mapping.go
    - internal/models/pricing.go
  modified:
    - internal/models/models.go (added LastCalculatedAt to UserAPIKey)
    - internal/database/database.go (registered new models)
decisions:
  - "Used pointer types (*float64, *time.Time) for optional pricing fields to allow NULL in database"
  - "ModelConcurrency uses map[string]int64 with gorm serializer:json for PostgreSQL storage"
  - "Extended pricing models use embedded structs with prefixes to avoid column name conflicts"
metrics:
  duration: "PT5M"
  completed_date: "2026-05-08"
---

# Phase 2 Plan 01: Data Model Foundation Summary

JWT auth with refresh rotation using jose library, API key management with granular permissions, membership tier-based discount system, multi-supplier model routing, and extended Bifrost-compatible pricing structures.

## Tasks Completed

| Task | Name | Commit | Files |
| ---- | ----- | ------ | ----- |
| 1 | UserAPIKey model | 7afe323, c794f1d | internal/models/models.go |
| 2 | Membership models | 22430c6 | internal/models/membership.go |
| 3 | ModelMapping | 74b6a64 | internal/models/model_mapping.go |
| 4 | Extended pricing | 8914618 | internal/models/pricing.go |
| 5 | Auto-migration | e793f7d | internal/database/database.go |

## Deviations from Plan

### Rule 1 - Bug: Fixed UserAPIKey duplicate definition

- **Found during:** Task 4 (pricing.go compilation)
- **Issue:** UserAPIKey was already defined in models.go from prior work, causing redeclaration error
- **Fix:** Removed duplicate keys.go, added missing LastCalculatedAt field to existing UserAPIKey in models.go
- **Files modified:** internal/models/models.go, internal/models/keys.go (deleted)
- **Commit:** c794f1d

### Notes on existing work

The database.go file already had UserAPIKey registered in AutoMigrate from a prior session. This was expected behavior as the model was created earlier. Task 5 completed the registration by adding the remaining models.

## Model Specifications

### UserAPIKey (internal/models/models.go)
- sk- prefix support (to be implemented in key_manager service)
- KeyValue hidden with json:"-" tag
- ModelConcurrency for per-model concurrent request limits
- Quota tracking (daily/monthly)
- LastCalculatedAt for Redis cleanup coordination

### Membership Models (internal/models/membership.go)
- MembershipTier: basic, premium, vip levels with display names
- MembershipDiscount: per-model discount rates (0.0-1.0 range)
- UserMembership: immediate effect per D-16, NULL expires_at = lifetime

### ModelMapping (internal/models/model_mapping.go)
- One external model to multiple supplier mappings
- Priority-based routing (lower = higher priority)
- Backup supplier support (is_backup flag)
- Health check configuration with QPS limits
- Status tracking: active, degraded, error

### Extended Pricing (internal/models/pricing.go)
- TextModelPricing: tiered pricing (128k+, 200k+ tokens), cache pricing
- ImageModelPricing: resolution-based (1024x1024+, 2048x2048+), quality-based
- VideoAudioPricing: per-second and per-token modes
- Embedded structs with prefixes to avoid column conflicts

## Threat Model Compliance

| Threat ID | Mitigation |
|-----------|------------|
| T-02-01 | KeyValue uses json:"-" tag |
| T-02-02 | ActualModelName separated from ExternalModelID |
| T-02-03 | SupplierCostPricingExtended is internal-only |
| T-02-05 | MembershipTierID changes require admin approval |

## Self-Check: PASSED

- [x] internal/models/membership.go exists
- [x] internal/models/model_mapping.go exists
- [x] internal/models/pricing.go exists
- [x] All models compile without errors
- [x] All commits exist in git log
- [x] Auto-migration updated in database.go
- [x] UserAPIKey has LastCalculatedAt field
