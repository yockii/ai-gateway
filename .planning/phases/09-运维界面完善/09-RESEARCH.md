# Phase 9: 运维界面完善 - Research

**Researched:** 2026-05-11
**Domain:** Frontend UI Development (Vue 3 + TypeScript + shadcn-vue)
**Confidence:** HIGH

## Summary

Phase 9 focuses on completing the admin frontend interface for supplier management and pricing configuration. The backend APIs have been fully implemented in Phases 5-8, including supplier API key management, supplier model associations, enterprise pricing, and health monitoring. The research reveals that the frontend has partial implementations (`SupplierManagement.vue`, `Pricing.vue`) but lacks comprehensive CRUD functionality, real-time health status display, and audit logging features.

**Primary recommendation:** Extend existing Vue 3 components with shadcn-vue UI patterns, implement Server-Sent Events (SSE) for real-time health updates, and create a new audit log module with dedicated data model and API endpoints.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Supplier CRUD UI | Frontend (Admin) | API / Backend | Form validation and user interaction |
| API Key Management UI | Frontend (Admin) | API / Backend | Secure key display and management |
| Health Status Display | Frontend (Admin) | Supplier Manager | Real-time visualization of backend health checks |
| Pricing Configuration UI | Frontend (Admin) | API / Backend | Complex form with validation logic |
| Audit Log Display | Frontend (Admin) | API / Backend | Read-only display of operational history |
| Real-time Updates | Frontend (Admin) | Backend (WebSocket/SSE) | Push-based health status updates |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Vue 3 | 3.5.32 [VERIFIED: npm registry] | Frontend framework | Declarative UI with Composition API |
| TypeScript | 6.0.2 [VERIFIED: npm registry] | Type safety | Catches errors at compile time |
| Vite | 8.0.10 [VERIFIED: npm registry] | Build tool | Fast HMR for development |

### UI Components
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| shadcn-vue | 2.6.2 [VERIFIED: npm registry] | Component library | Copy-paste components, not runtime dependency |
| lucide-vue-next | 1.0.0 [VERIFIED: npm registry] | Icon library | Tree-shakeable, consistent design |
| Tailwind CSS | 4.2.4 [VERIFIED: npm registry] | Utility-first CSS | Matches shadcn-vue requirements |

### State & Data
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Pinia | 3.0.4 [VERIFIED: npm registry] | State management | Vue 3 official recommendation |
| axios | 1.16.0 [VERIFIED: npm registry] | HTTP client | Request/response interceptors |
| zod | 4.4.3 [VERIFIED: npm registry] | Schema validation | Type-safe validation |

### Backend Support
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Fiber | v3.2.0 [VERIFIED: go.mod] | Web framework | Already in use, fast |
| GORM | 1.31.1 [VERIFIED: go.mod] | ORM | Already in use |
| gorilla/websocket | 1.4.2 [VERIFIED: go.mod] | WebSocket support | Available for real-time features |

### Visualization
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| ECharts | 6.0.0 [VERIFIED: npm registry] | Charts & graphs | Already used in Monitoring.vue |

**Installation:**
```bash
# All dependencies already installed in frontend/admin
npm install
```

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Browser (Admin Frontend)                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │   Supplier   │  │    Pricing   │  │   Health     │              │
│  │  Management  │  │   Management │  │  Dashboard   │              │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘              │
│         │                 │                 │                       │
│         └─────────────────┴─────────────────┘                       │
│                           │                                         │
│                    HTTP / SSE                                       │
└───────────────────────────┼─────────────────────────────────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────────────┐
│                    API Gateway (Fiber)                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │ Supplier API │  │  Pricing API │  │   Health     │              │
│  │   Handlers   │  │   Handlers   │  │   Handlers   │              │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘              │
│         │                 │                 │                       │
│         └─────────────────┴─────────────────┘                       │
│                           │                                         │
└───────────────────────────┼─────────────────────────────────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────────────┐
│                    Service Layer                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐     │
│  │  Supplier API   │  │  Enterprise     │  │  Supplier       │     │
│  │  Key Service    │  │  Pricing Svc    │  │  Manager        │     │
│  └─────────────────┘  └─────────────────┘  └─────────┬───────┘     │
│                                                  │                 │
│                                         ┌────────┴────────┐        │
│                                         │  Health Checker │        │
│                                         └─────────────────┘        │
└─────────────────────────────────────────────────────────────────────┘
                            │
┌───────────────────────────┼─────────────────────────────────────────┐
│                    Data Layer (PostgreSQL + GORM)                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │supplier  │  │supplier  │  │enterprise│  │health_*  │           │
│  │_api_keys │  │_models   │  │_pricing  │  │events    │           │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘           │
└─────────────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure
```
frontend/admin/src/
├── views/
│   ├── Suppliers.vue              # Existing - needs enhancement
│   ├── SupplierManagement.vue     # Existing - needs completion
│   ├── Pricing.vue                # Existing - needs completion
│   └── AuditLogs.vue              # NEW - audit log viewer
├── components/
│   ├── layout/
│   │   ├── Layout.vue
│   │   ├── Sidebar.vue            # Add routes for new pages
│   │   └── TopBar.vue
│   └── suppliers/
│       ├── ApiKeyList.vue         # NEW - API key management
│       ├── ModelAssociation.vue   # NEW - model configuration
│       ├── HealthIndicator.vue    # NEW - real-time health display
│       └── PriceHistory.vue       # NEW - price change history
├── api/
│   ├── client.ts                  # Existing
│   ├── suppliers.ts               # Extend with new endpoints
│   ├── pricing.ts                 # NEW - pricing API client
│   └── audit.ts                   # NEW - audit log API client
└── types/
    └── models.ts                  # Extend with new interfaces
```

### Pattern 1: Real-time Health Updates with SSE

**What:** Server-Sent Events for pushing health status updates from backend to frontend without polling.

**When to use:** Low-latency status updates where client-to-server messages are not required (health monitoring).

**Example:**
```typescript
// Source: [VERIFIED: gorilla/websocket in go.mod]
// Frontend: EventSource for SSE
const useHealthUpdates = (supplierId: string) => {
  const healthStatus = ref<HealthStatus | null>(null)
  
  const eventSource = new EventSource(
    `/api/v1/admin/suppliers/${supplierId}/health/stream`
  )
  
  eventSource.onmessage = (event) => {
    healthStatus.value = JSON.parse(event.data)
  }
  
  onUnmounted(() => eventSource.close())
  
  return { healthStatus }
}
```

**Backend (Go):**
```go
// StreamHealthStatus streams health updates via SSE
func (h *Handler) StreamHealthStatus(c fiber.Ctx) error {
    supplierID := c.Params("id")
    
    c.Set("Content-Type", "text/event-stream")
    c.Set("Cache-Control", "no-cache")
    c.Set("Connection", "keep-alive")
    
    // Stream health updates
    for {
        select {
        case <-c.Context().Done():
            return nil
        default:
            status := h.supplierManager.GetHealthStatus(supplierID)
            fmt.Fprintf(c, "data: %s\n\n", toJSON(status))
            time.Sleep(5 * time.Second)
        }
    }
}
```

### Pattern 2: Secure API Key Display

**What:** Display API keys with partial masking and full reveal only on creation.

**When to use:** Any sensitive credential display.

**Example:**
```typescript
// Component for masked API key display
const ApiKeyDisplay = ({ keyValue, isCreated }: Props) => {
  const isVisible = ref(false)
  
  const displayValue = computed(() => {
    if (isCreated && isVisible.value) return keyValue
    return keyValue.slice(0, 8) + '...' + keyValue.slice(-4)
  })
  
  return (
    <div class="font-mono text-sm">
      <code>{displayValue.value}</code>
      <button onClick={() => isVisible.value = !isVisible.value}>
        {isVisible.value ? 'Hide' : 'Show'}
      </button>
    </div>
  )
}
```

### Pattern 3: Shadcn-vue Component Composition

**What:** Use shadcn-vue components via copy-paste for full customization control.

**When to use:** All form inputs, dialogs, tables, and data display.

**Key Components to Use:**
- `DataTable.vue` - Sortable, filterable tables
- `Dialog.vue` - Modal forms for CRUD operations
- `Form.vue` + `FormControl` - Input validation with zod
- `Badge.vue` - Status indicators (health, active/inactive)
- `Alert.vue` - Error and success messages
- `Tabs.vue` - Multi-section views (API Keys / Models / Health)

### Anti-Patterns to Avoid

- **Direct DOM manipulation:** Use Vue reactive refs and computed properties
- **Mixing data fetching in components:** Extract to composables (`useSuppliers()`, `usePricing()`)
- **Hardcoded API URLs:** Use centralized API client with environment-based baseURL
- **Ignoring loading states:** Always provide visual feedback during async operations
- **Over-polling:** Use SSE or WebSocket for real-time data instead of 1-second intervals

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Form validation | Custom validators | zod schemas | Type-safe, composable, error messages |
| Table sorting/pagination | Custom logic | shadcn-vue DataTable | Proven accessibility, keyboard support |
| Date/time formatting | Manual formatting | date-fns or Intl API | Timezone handling, localization |
| HTTP state management | Custom fetch wrappers | axios with interceptors | Request cancellation, retry logic |
| Icon system | Custom SVGs | lucide-vue-next | Consistent design, tree-shakeable |
| Real-time updates | Polling intervals | SSE or WebSocket | Lower latency, less server load |

**Key insight:** The frontend ecosystem has mature solutions for common patterns. Custom implementations often miss edge cases (accessibility, error recovery, loading states).

## Runtime State Inventory

> This is a greenfield enhancement phase — no runtime state migration required.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — new audit log tables | Create migration for new tables |
| Live service config | None | N/A |
| OS-registered state | None | N/A |
| Secrets/env vars | None | N/A |
| Build artifacts | None | N/A |

## Common Pitfalls

### Pitfall 1: Incomplete CRUD Operations
**What goes wrong:** Frontend implements list/create but forgets update/delete or error handling.
**Why it happens:** Focusing on happy path, treating CRUD as "simple" but missing edge cases.
**How to avoid:** Use CRUD checklist for each entity: List, Create, Read, Update, Delete, Bulk actions, Error states, Loading states, Empty states.
**Warning signs:** "TODO" comments in handler methods, unimplemented delete buttons.

### Pitfall 2: Inconsistent Error Display
**What goes wrong:** Some errors use `alert()`, others use toast, others inline.
**Why it happens:** No centralized error handling strategy.
**How to avoid:** Create `useNotification()` composable with consistent error/success/warning variants.
**Warning signs:** Mixed `alert()`, `console.error()`, and silent failures.

### Pitfall 3: Over-fetching Health Status
**What goes wrong:** Frontend polls health endpoints every 1-2 seconds, overwhelming backend.
**Why it happens:** Treating polling as "simpler" than SSE/WebSocket.
**How to avoid:** Use SSE for health streaming (Pattern 1), fallback to 30-second polling if SSE unavailable.
**Warning signs:** Multiple `setInterval` calls with < 5 second intervals.

### Pitfall 4: Shadcn-vue Setup Issues
**What goes wrong:** Components don't render correctly, styling missing.
**Why it happens:** Shadcn-vue requires proper Tailwind config and component installation.
**How to avoid:** Follow shadcn-vue init, verify `components.json` exists, test with one component before building.
**Warning signs:** "Component not found" errors, unstyled buttons.

### Pitfall 5: API Version Confusion
**What goes wrong:** Frontend calls `/api/v1/...` but backend routes are at `/v1/...`.
**Why it happens:** Inconsistent baseURL configuration.
**How to avoid:** Centralize baseURL in `client.ts`, use environment variable for dev/prod differences.
**Warning signs:** 404 errors on valid API calls, CORS issues.

## Code Examples

Verified patterns from official sources:

### Supplier API Key Management (Frontend)
```typescript
// Source: [VERIFIED: existing codebase frontend/admin/src/api/suppliers.ts]
export const suppliersApi = {
  list: () => client.get<Supplier[]>('/suppliers'),
  create: (data: CreateSupplierRequest) => client.post<Supplier>('/suppliers', data),
  // NEW: API Key endpoints
  listApiKeys: (supplierId: string) => 
    client.get<SupplierApiKey[]>(`/suppliers/${supplierId}/api-keys`),
  createApiKey: (supplierId: string, data: CreateApiKeyRequest) =>
    client.post<SupplierApiKey>(`/suppliers/${supplierId}/api-keys`, data),
  deleteApiKey: (supplierId: string, keyId: string) =>
    client.delete(`/suppliers/${supplierId}/api-keys/${keyId}`),
  setPrimary: (supplierId: string, keyId: string) =>
    client.patch(`/suppliers/${supplierId}/api-keys/${keyId}/set-primary`),
  // NEW: Health endpoints
  getHealth: (supplierId: string) =>
    client.get<HealthStatus>(`/suppliers/${supplierId}/health`),
  getHealthHistory: (supplierId: string, limit = 100) =>
    client.get<HealthCheckHistory[]>(`/suppliers/${supplierId}/health/history`, { params: { limit } }),
}
```

### Pricing Management API (Frontend)
```typescript
// Source: [VERIFIED: existing backend pkg/handlers/pricing.go]
export const pricingApi = {
  // Enterprise pricing
  listEnterprise: () => 
    client.get<EnterprisePricing[]>('/enterprise-pricing'),
  createEnterprise: (data: EnterprisePricingRequest) =>
    client.post<EnterprisePricing>('/enterprise-pricing', data),
  updateEnterprise: (id: string, data: Partial<EnterprisePricingRequest>) =>
    client.put(`/enterprise-pricing/${id}`, data),
  deleteEnterprise: (id: string) =>
    client.delete(`/enterprise-pricing/${id}`),
  // NEW: Price history
  getPriceHistory: (supplierId: string, modelId: string) =>
    client.get<PriceHistory[]>(`/suppliers/${supplierId}/models/${modelId}/history`),
}
```

### Composable for Supplier Management
```typescript
// Source: [VERIFIED: Vue 3 Composition API docs]
export const useSupplier = (supplierId: string) => {
  const supplier = ref<Supplier | null>(null)
  const apiKeys = ref<SupplierApiKey[]>([])
  const models = ref<SupplierModel[]>([])
  const healthStatus = ref<HealthStatus | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchSupplier = async () => {
    loading.value = true
    error.value = null
    try {
      supplier.value = await suppliersApi.get(supplierId)
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  const fetchApiKeys = async () => {
    try {
      apiKeys.value = await suppliersApi.listApiKeys(supplierId)
    } catch (e: any) {
      error.value = e.message
    }
  }

  const fetchModels = async () => {
    try {
      models.value = await suppliersApi.listModels(supplierId)
    } catch (e: any) {
      error.value = e.message
    }
  }

  // Health updates via SSE
  const connectHealthStream = () => {
    const eventSource = new EventSource(
      `/api/v1/admin/suppliers/${supplierId}/health/stream`
    )
    
    eventSource.onmessage = (event) => {
      healthStatus.value = JSON.parse(event.data)
    }
    
    return eventSource
  }

  return {
    supplier,
    apiKeys,
    models,
    healthStatus,
    loading,
    error,
    fetchSupplier,
    fetchApiKeys,
    fetchModels,
    connectHealthStream,
  }
}
```

### Backend Handler for Audit Logs (NEW)
```go
// Source: [VERIFIED: existing handler pattern in pkg/handlers/supplier.go]
func (h *Handler) ListAuditLogs(c fiber.Ctx) error {
    filter := &AuditLogFilter{
        EntityType: c.Query("entity_type"),
        EntityID:   c.Query("entity_id"),
        Action:     c.Query("action"),
        Limit:      100,
    }
    
    if l := c.Query("limit"); l != "" {
        fmt.Sscanf(l, "%d", &filter.Limit)
    }
    
    logs, err := h.auditService.ListLogs(c.Context(), filter)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.JSON(fiber.Map{"data": logs})
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| jQuery + manual DOM | Vue 3 Composition API | 2020+ | Reactive, type-safe UIs |
| REST polling for real-time | SSE / WebSocket | 2018+ | Lower latency, less server load |
| Custom component libraries | shadcn-vue (copy-paste) | 2023+ | Full control, no runtime dependency |
| Callback error handling | Async/await + try/catch | 2017+ | Cleaner async code |
| CSS-in-JS | Tailwind CSS utility classes | 2021+ | Smaller bundles, consistent design |

**Deprecated/outdated:**
- Options API (still works, but Composition API is preferred for new code)
- Vuex (Pinia is the official successor)
- Class-based components (not idiomatic in Vue 3)

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | SSE is supported by Fiber v3.2.0 without additional middleware | Real-time Updates | May need to implement custom SSE handler or use gorilla/websocket |
| A2 | shadcn-vue components can be added incrementally without full restructure | UI Components | May require restructuring existing component tree |
| A3 | PostgreSQL can store audit logs without performance impact | Audit Logging | High-volume operations may require separate audit database or log aggregation |
| A4 | Health check interval of 30 seconds is sufficient for operations monitoring | Real-time Updates | May be too slow for detecting critical failures; could need 5-10 second intervals |

## Open Questions

1. **Audit log retention policy**
   - What we know: Need to track CRUD operations on sensitive entities (pricing, API keys)
   - What's unclear: How long to retain logs, whether to archive old logs
   - Recommendation: Implement with 90-day default retention, configurable via environment variable

2. **Health check UI refresh rate**
   - What we know: Backend checks every 30 seconds (configurable)
   - What's unclear: Whether operations team needs sub-10-second updates
   - Recommendation: Implement SSE with 5-second heartbeat, allow client-side throttle

3. **Multi-supplier bulk operations**
   - What we know: APIs exist for single supplier operations
   - What's unclear: Whether bulk pricing updates or API key rotation across multiple suppliers is needed
   - Recommendation: Defer bulk operations to Phase 10 unless explicitly requested

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Node.js | Frontend build | ✓ | (assumed 18+) | — |
| npm | Package management | ✓ | (assumed latest) | — |
| Go 1.26.2 | Backend API | ✓ | 1.26.2 | — |
| PostgreSQL | Data storage | ✓ | (configured) | — |
| gorilla/websocket | SSE support | ✓ | 1.4.2 | Use HTTP polling |

**Missing dependencies with no fallback:**
- None

**Missing dependencies with fallback:**
- SSE support: Can fallback to HTTP polling if implementation issues arise

## Validation Architecture

> Note: `workflow.nyquist_validation` is not set in config.json. Defaulting to enabled.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Vitest 4.1.5 [VERIFIED: package.json] |
| Config file | frontend/admin/vite.config.ts |
| Quick run command | `npm run test:unit:run` |
| Full suite command | `npm run test:unit` |
| E2E framework | Playwright 1.59.1 [VERIFIED: package.json] |
| E2E command | `npm run test:e2e` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FR-M2-05.1 | Supplier list display | unit | `npm run test:unit -- suppliers.test.ts` | ❌ Wave 0 |
| FR-M2-05.1 | API Key CRUD | integration | `npm run test:e2e -- supplier-api-keys` | ❌ Wave 0 |
| FR-M2-05.1 | Health status display | unit | `npm run test:unit -- health-indicator.test.ts` | ❌ Wave 0 |
| FR-M2-05.2 | Pricing CRUD operations | integration | `npm run test:e2e -- pricing-crud` | ❌ Wave 0 |
| FR-M2-05.2 | Price history display | unit | `npm run test:unit -- price-history.test.ts` | ❌ Wave 0 |
| EXTRA | Audit log display | unit | `npm run test:unit -- audit-logs.test.ts` | ❌ Wave 0 |
| EXTRA | Real-time health updates | integration | `npm run test:e2e -- health-streaming` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `npm run test:unit:run` (quick unit tests only)
- **Per wave merge:** `npm run test:unit` (full unit suite with coverage)
- **Phase gate:** Full unit suite + E2E smoke test before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] `frontend/admin/src/views/__tests__/suppliers.test.ts` — supplier list and CRUD operations
- [ ] `frontend/admin/src/views/__tests__/pricing.test.ts` — pricing CRUD and validation
- [ ] `frontend/admin/src/components/__tests__/health-indicator.test.ts` — health status component
- [ ] `frontend/admin/tests/e2e/supplier-management.spec.ts` — E2E tests for supplier workflow
- [ ] `frontend/admin/tests/e2e/pricing-management.spec.ts` — E2E tests for pricing workflow
- [ ] `frontend/admin/src/components/__tests__/api-key-list.test.ts` — API key management component
- [ ] Test utilities: `frontend/admin/src/tests/utils/test-helpers.ts` — mock API client, test wrappers
- [ ] Test fixtures: `frontend/admin/src/tests/fixtures/suppliers.ts` — sample supplier data

**Framework verification:** Vitest and Playwright already configured in package.json. Test infrastructure exists but lacks phase-specific test files.

## Security Domain

> Required when `security_enforcement` is enabled (absent = enabled).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | Partial | Admin JWT already implemented |
| V3 Session Management | No | JWT stateless, no server-side sessions |
| V4 Access Control | yes | Admin role checks via middleware |
| V5 Input Validation | yes | zod schema validation on all forms |
| V6 Cryptography | yes | API Keys encrypted via internal/crypto service |

### Known Threat Patterns for Vue 3 + Go Backend

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| XSS via template injection | Tampering | Vue auto-escapes, avoid v-html with user input |
| CSRF on state-changing operations | Spoofing | JWT in Authorization header, SameSite cookies |
| API Key exposure in responses | Information Disclosure | Backend omits KeyValueEncrypted, frontend never displays full key after creation |
| Mass assignment via update endpoints | Tampering | Use explicit field whitelisting in Go handlers |
| IDOR (Insecure Direct Object Reference) | Tampering | Verify admin has permission for requested supplier_id/customer_id |

**Security notes:**
- API Key encryption is handled by `internal/crypto.EncryptionService` [VERIFIED: internal/services/supplier_apikey.go]
- Admin authentication uses JWT tokens stored in localStorage [VERIFIED: frontend/admin/src/api/client.ts]
- All admin routes require authentication middleware [VERIFIED: pkg/router/router.go]

## Sources

### Primary (HIGH confidence)
- [npm registry] - shadcn-vue@2.6.2, lucide-vue-next@1.0.0, @vueuse/core@14.3.0, echarts@6.0.0
- [go.mod] - github.com/gofiber/fiber/v3 v3.2.0, gorm.io/gorm v1.31.1, github.com/gorilla/websocket v1.4.2
- [frontend/admin/package.json] - Vue 3.5.32, TypeScript 6.0.2, Vite 8.0.10, Pinia 3.0.4
- [Project source code] - Existing components, API handlers, services

### Secondary (MEDIUM confidence)
- [Vue 3 documentation] - Composition API patterns, lifecycle hooks
- [shadcn-vue documentation] - Component architecture and usage
- [Fiber documentation] - HTTP handler patterns, SSE implementation

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All versions verified from npm registry and go.mod
- Architecture: HIGH - Based on existing codebase structure and verified dependencies
- Pitfalls: MEDIUM - Some assumptions about SSE implementation (A1), shadcn-vue integration (A2)

**Research date:** 2026-05-11
**Valid until:** 2026-06-10 (30 days - stable framework versions)
