# Phase 2: 核心功能开发 - Validation Criteria

**Created:** 2026-05-08
**Phase:** 2-核心功能开发
**Status:** Ready for execution

## Overview

This document defines testable acceptance criteria for Phase 2 core features. Each criterion includes:
- Given/When/Then test scenarios
- Performance benchmarks
- Security validation requirements
- Edge case coverage

## Feature 1: Bifrost Integration (D-01, D-02, D-03)

### AC-1.1: Bifrost Library Initialization
**Given:** A running Gateway service
**When:** Initializing the Bifrost service
**Then:**
- Bifrost instance is created without errors
- Custom Account interface is registered
- Provider queues are initialized
- Object pools are configured (ChannelMessage, Response, etc.)

**Test Command:**
```bash
go test -v -run TestBifrostInit ./internal/services/bifrostservice/...
```

**Performance:** Initialization < 500ms

**Edge Cases:**
- Empty provider list
- Invalid provider configuration
- Missing API keys

### AC-1.2: Chat Completion Non-Streaming
**Given:** A valid API key and chat request
```
POST /v1/chat/completions
{
  "model": "gpt-3.5-turbo",
  "messages": [{"role": "user", "content": "Hello"}],
  "stream": false
}
```
**When:** Sending request to gateway
**Then:**
- Response is OpenAI-compatible format
- Response includes `id`, `object`, `created`, `model`, `choices`, `usage`
- `usage.prompt_tokens` > 0
- `usage.completion_tokens` > 0
- Response time < 20ms (gateway overhead only, excludes provider time)

**Test Command:**
```bash
go test -v -run TestChatCompletionNonStream ./pkg/handlers/...
```

### AC-1.3: Chat Completion Streaming
**Given:** A valid API key and streaming chat request
```
POST /v1/chat/completions
{
  "model": "gpt-3.5-turbo",
  "messages": [{"role": "user", "content": "Hello"}],
  "stream": true
}
```
**When:** Sending request with `stream: true`
**Then:**
- Response has `Content-Type: text/event-stream`
- Multiple `data:` chunks received
- Final message is `data: [DONE]`
- Each chunk has valid JSON structure
- Chunk ordering is preserved

**Test Command:**
```bash
go test -v -run TestChatCompletionStream ./pkg/handlers/...
```

**Performance:** First chunk < 100ms, subsequent chunks < 50ms

### AC-1.4: All OpenAI Endpoints Implemented
**Given:** Gateway is running
**When:** Calling each OpenAI-compatible endpoint
**Then:**

| Endpoint | Method | Status |
|----------|--------|--------|
| `/v1/chat/completions` | POST | ✅ Implemented |
| `/v1/completions` | POST | ✅ Implemented |
| `/v1/images/generations` | POST | ✅ Implemented |
| `/v1/audio/speech` | POST | ✅ Implemented |
| `/v1/audio/transcriptions` | POST | ✅ Implemented |
| `/v1/embeddings` | POST | ✅ Implemented |
| `/v1/models` | GET | ✅ Implemented |

**Test Command:**
```bash
go test -v -run TestAllEndpoints ./pkg/handlers/...
```

### AC-1.5: Provider Failover
**Given:** Model configured with 2+ suppliers
**When:** Primary supplier fails (returns 5xx)
**Then:**
- Request automatically routes to backup supplier
- Response is successful
- Failover happens within 1 second
- Event is logged

**Test Command:**
```bash
go test -v -run TestProviderFailover ./internal/services/...
```

## Feature 2: Model Routing & Pricing (D-02, D-07, D-09, D-10)

### AC-2.1: Profit-Based Routing
**Given:** External model "gpt-4-turbo" with 3 suppliers:
- Supplier A: cost $10/M, priority 1
- Supplier B: cost $12/M, priority 2
- Supplier C: cost $15/M, priority 3 (backup)
**When:** User requests "gpt-4-turbo" with selling price $18/M
**Then:**
- Supplier A is selected (highest profit: $8/M)
- Usage record shows correct cost and profit
- Profit margin = ($18 - $10) / $18 = 55.56%

**Test Command:**
```bash
go test -v -run TestProfitRouting ./internal/services/modelrouter/...
```

### AC-2.2: Profit Protection
**Given:** Supplier cost $20/M, user selling price $18/M
**When:** Router evaluates this supplier
**Then:**
- Supplier is rejected from candidate list
- Error "No available supplier with positive margin" returned
- No request sent to this supplier

**Test Command:**
```bash
go test -v -run TestProfitProtection ./internal/services/modelrouter/...
```

### AC-2.3: Tiered Pricing Calculation
**Given:** Model with tiered pricing:
- Base: $0.50/1k tokens (input), $1.50/1k tokens (output)
- Above 128k: $0.60/1k tokens (input), $1.80/1k tokens (output)
- Above 200k: $0.70/1k tokens (input), $2.00/1k tokens (output)
**When:** Request uses 150k input tokens, 20k output tokens
**Then:**
- Input cost = (128k × $0.50 + 22k × $0.60) / 1000 = $76.40
- Output cost = 20k × $1.50 / 1000 = $30.00
- Total cost = $106.40

**Test Command:**
```bash
go test -v -run TestTieredPricing ./internal/pricing/...
```

### AC-2.4: Cache Pricing
**Given:** Model with semantic cache enabled:
- Cache creation: $0.025/1k tokens
- Cache read: $0.005/1k tokens
**When:** Request with 10k cached tokens, 5k new tokens
**Then:**
- Input cost = (5k × $0.50 + 10k × $0.005) / 1000 = $2.55
- Usage record shows `cached_read_tokens: 10000`

**Test Command:**
```bash
go test -v -run TestCachePricing ./internal/pricing/...
```

## Feature 3: API Key Management (D-11, D-12, D-13)

### AC-3.1: API Key Format
**Given:** Creating a new API key
**When:** Key generation completes
**Then:**
- Key format is `sk-<UUIDv4>` (36 chars total with prefix)
- Example: `sk-550e8400-e29b-41d4-a716-446655440000`
- UUID is valid v4 (version bit set)
- Key is unique in database

**Test Command:**
```bash
go test -v -run TestAPIKeyFormat ./internal/services/keymanager/...
```

### AC-3.2: API Key Creation
**Given:** Authenticated user with < 10 keys
**When:** POST /api/keys with `{"name": "My Key"}`
**Then:**
- Key created successfully
- Response includes full key (only time visible)
- Key is hashed in database
- Default limits applied (concurrency: 10, quota: unlimited)

**Test Command:**
```bash
go test -v -run TestCreateAPIKey ./internal/services/keymanager/...
```

### AC-3.3: API Key Validation
**Given:** Request with `Authorization: Bearer sk-<key>`
**When:** Key validation middleware runs
**Then:**
- Key format validated
- Database lookup performed
- User ID extracted and stored in context
- Invalid keys return 401

**Test Command:**
```bash
go test -v -run TestValidateAPIKey ./internal/middleware/...
```

### AC-3.4: Concurrency Limit
**Given:** API key with `concurrency_limit: 5`
**When:** 6 concurrent requests sent
**Then:**
- 5 requests proceed normally
- 6th request gets 429 (concurrency limit exceeded)
- After one request completes, new requests can proceed

**Test Command:**
```bash
go test -v -run TestConcurrencyLimit ./internal/middleware/...
```

**Performance:** Concurrency check < 5ms via Redis INCR

### AC-3.5: Model-Specific Concurrency
**Given:** API key with:
- Default concurrency: 10
- Model-specific: `{"gpt-4": 2, "gpt-3.5-turbo": 10}`
**When:** Sending 3 concurrent requests to "gpt-4"
**Then:**
- 2 requests proceed
- 3rd request gets 429 for model-specific limit
- Requests to "gpt-3.5-turbo" unaffected

**Test Command:**
```bash
go test -v -run TestModelConcurrency ./internal/middleware/...
```

### AC-3.6: Daily Quota Limit
**Given:** API key with `daily_quota: 1000000` tokens
**When:** Daily usage reaches 1,000,001 tokens
**Then:**
- Request rejected with 429
- Error message includes quota exceeded info
- Reset at midnight UTC

**Test Command:**
```bash
go test -v -run TestDailyQuota ./internal/services/keymanager/...
```

## Feature 4: Membership & Discounts (D-14, D-15, D-16)

### AC-4.1: Membership Tier Application
**Given:** User with "VIP" membership
**When:** User makes request
**Then:**
- VIP discount applied to pricing
- Discount rate per model from MembershipDiscount table
- Final price = base_price × (1 - discount_rate)

**Test Command:**
```bash
go test -v -run TestMembershipDiscount ./internal/services/membership/...
```

### AC-4.2: Model-Specific Discounts
**Given:** VIP membership with discounts:
- gpt-4: 20%
- gpt-3.5-turbo: 10%
**When:** User requests gpt-4 at $20/M
**Then:**
- Selling price = $20 × 0.8 = $16/M
- gpt-3.5-turbo requests get 10% discount

**Test Command:**
```bash
go test -v -run TestModelSpecificDiscount ./internal/services/membership/...
```

### AC-4.3: Immediate Membership Changes
**Given:** User upgraded from "Basic" to "VIP"
**When:** Next request is made
**Then:**
- VIP pricing applied immediately
- No restart required
- Database change propagates within 100ms

**Test Command:**
```bash
go test -v -run TestImmediateMembership ./internal/services/membership/...
```

## Feature 5: Billing & Usage Recording (D-08, D-17, D-18)

### AC-5.1: Dual-Record Billing
**Given:** Successful model response with usage data
**When:** Billing recorder processes usage
**Then:**
- Record written to database (async)
- Record pushed to Redis Stream (sync)
- RequestID is unique and indexed
- Both records contain identical data

**Test Command:**
```bash
go test -v -run TestDualRecord ./internal/services/billing/...
```

**Performance:** Recording overhead < 5ms (async DB)

### AC-5.2: Idempotent Recording
**Given:** Usage record with RequestID "req-123"
**When:** Same RequestID recorded twice
**Then:**
- First insert succeeds
- Second insert ignored (unique index violation)
- No duplicate records in database

**Test Command:**
```bash
go test -v -run TestIdempotentRecording ./internal/services/billing/...
```

### AC-5.3: Redis Stream Consumer
**Given:** Redis Stream with pending messages
**When:** Consumer processes messages
**Then:**
- Each message written to database
- Message ACKed after successful write
- Failed messages remain pending
- Consumer can claim stale messages

**Test Command:**
```bash
go test -v -run TestStreamConsumer ./internal/services/billing/...
```

### AC-5.4: Monthly Bill Generation
**Given:** User with usage in May 2026
**When:** Cron job runs on June 1, 2026
**Then:**
- Bill generated for period 2026-05-01 to 2026-05-31
- Bill includes all usage records
- Total cost, revenue, profit calculated
- Bill saved to database

**Test Command:**
```bash
go test -v -run TestBillGeneration ./internal/services/billing/...
```

### AC-5.5: PDF Bill Export
**Given:** Generated bill with ID "bill-123"
**When:** User downloads PDF
**Then:**
- PDF includes bill summary
- Model usage table included
- Total amounts displayed
- UTF-8 characters (Chinese) render correctly
- File is valid PDF format

**Test Command:**
```bash
go test -v -run TestPDFExport ./internal/utils/...
```

### AC-5.6: CSV Usage Export
**Given:** Bill with multiple usage records
**When:** User exports CSV
**Then:**
- CSV includes all usage records
- Headers: timestamp, model, supplier, tokens, cost, revenue, profit
- UTF-8 encoded
- Valid CSV format

**Test Command:**
```bash
go test -v -run TestCSVExport ./internal/utils/...
```

## Feature 6: Image & Audio Endpoints

### AC-6.1: Image Generation
**Given:** Valid image generation request
```
POST /v1/images/generations
{
  "model": "dall-e-3",
  "prompt": "A cat",
  "n": 1,
  "size": "1024x1024"
}
```
**When:** Request processed
**Then:**
- Response includes `data` array with image URLs
- `data[0].url` is valid URL
- Usage recorded (per-image pricing)
- Response time < 5s (provider dependent)

**Test Command:**
```bash
go test -v -run TestImageGeneration ./pkg/handlers/...
```

### AC-6.2: TTS (Text-to-Speech)
**Given:** Valid TTS request
```
POST /v1/audio/speech
{
  "model": "tts-1",
  "input": "Hello world",
  "voice": "alloy"
}
```
**When:** Request processed
**Then:**
- Response has `Content-Type: audio/mpeg`
- Audio data returned
- Usage recorded (character or second-based pricing)

**Test Command:**
```bash
go test -v -run TestTTS ./pkg/handlers/...
```

### AC-6.3: STT (Speech-to-Text)
**Given:** Valid STT request with audio file
**When:** Request processed
**Then:**
- Transcription text returned
- Usage recorded (per-second pricing)
- Language detection works

**Test Command:**
```bash
go test -v -run TestSTT ./pkg/handlers/...
```

## Feature 7: Embeddings & Reranking

### AC-7.1: Embeddings
**Given:** Valid embeddings request
```
POST /v1/embeddings
{
  "model": "text-embedding-3-small",
  "input": "Hello world"
}
```
**When:** Request processed
**Then:**
- Response includes `data` array with embeddings
- `data[0].embedding` is array of floats (dimension 1536)
- Usage recorded (input token pricing)

**Test Command:**
```bash
go test -v -run TestEmbeddings ./pkg/handlers/...
```

### AC-7.2: Reranking
**Given:** Valid rerank request
```
POST /v1/rerank
{
  "model": "rerank-v2",
  "query": "What is AI?",
  "documents": ["doc1", "doc2", "doc3"],
  "top_n": 2
}
```
**When:** Request processed
**Then:**
- Response includes reranked documents
- Relevance scores included
- Top N documents returned

**Test Command:**
```bash
go test -v -run TestRerank ./pkg/handlers/...
```

## Performance Benchmarks

### PB-1: Gateway Overhead
**Metric:** Gateway processing time (excluding provider)
**Target:** < 20ms p95
**Test:** Load test with 1000 concurrent requests

```bash
go test -v -run BenchmarkGatewayOverhead ./benchmarks/...
```

### PB-2: Concurrency Check Performance
**Metric:** Redis INCR + check + DECR
**Target:** < 5ms p95
**Test:** 10000 concurrent concurrency checks

```bash
go test -v -run BenchmarkConcurrency ./benchmarks/...
```

### PB-3: Streaming Latency
**Metric:** Time to first chunk
**Target:** < 100ms p95
**Test:** 100 streaming requests

```bash
go test -v -run BenchmarkStreaming ./benchmarks/...
```

### PB-4: QPS Capacity
**Metric:** Maximum sustainable QPS
**Target:** > 10000 QPS
**Test:** Sustained load for 5 minutes

```bash
go test -v -run BenchmarkMaxQPS ./benchmarks/...
```

## Security Validation

### SV-1: API Key Hashing
**Test:** API keys never stored in plain text
**Check:**
```bash
# Query database for plain text keys (should return 0)
psql -c "SELECT COUNT(*) FROM user_api_keys WHERE key_value NOT LIKE '\$%'"
```

### SV-2: Input Validation
**Test:** All user inputs validated
**Checks:**
- SQL injection attempts blocked
- XSS attempts in prompts blocked
- Overly large requests rejected (> 10MB)
- Malformed JSON returns 400

### SV-3: Rate Limiting
**Test:** Rate limits enforced
**Checks:**
- Per-IP rate limiting works
- Per-key rate limiting works
- 429 response includes Retry-After header

### SV-4: Authentication Bypass
**Test:** Cannot access without valid API key
**Checks:**
- Missing API key returns 401
- Invalid API key returns 401
- Expired API key returns 401
- Disabled API key returns 401

## Edge Cases

### EC-1: Empty Messages Array
**Given:** Request with `"messages": []`
**When:** Validation runs
**Then:** Returns 400 with "messages cannot be empty"

### EC-2: Invalid Model Name
**Given:** Request with non-existent model
**When:** Router attempts to find supplier
**Then:** Returns 404 with "model not found"

### EC-3: No Available Suppliers
**Given:** All suppliers for model are disabled/down
**When:** Request sent
**Then:** Returns 503 with "no available suppliers"

### EC-4: Concurrent Limit Edge
**Given:** API key at concurrency limit N
**When:** N+1 request arrives, then 1 request completes
**Then:** N+1th request proceeds after completion

### EC-5: Streaming Disconnect
**Given:** Active streaming connection
**When:** Client disconnects
**Then:** Goroutine exits cleanly, no resource leak

### EC-6: Redis Down
**Given:** Redis unavailable
**When:** Concurrency check attempted
**Then:** Falls back to in-memory (degraded mode), logs error

### EC-7: Bill with No Usage
**Given:** User with zero usage in month
**When:** Bill generation runs
**Then:** No bill created (or empty bill created with warning)

## Integration Tests

### IT-1: Full Request Flow
**Scenario:** User → Gateway → Bifrost → Provider → Gateway → User
**Steps:**
1. Create API key
2. Make chat request
3. Verify routing to correct supplier
4. Verify response is OpenAI-compatible
5. Verify usage recorded
6. Verify billing calculation

**Test Command:**
```bash
go test -v -run TestFullFlow ./internal/integration/...
```

### IT-2: Failover Flow
**Scenario:** Primary supplier fails, backup succeeds
**Steps:**
1. Configure 2 suppliers
2. Make primary supplier return 500
3. Verify request routed to backup
4. Verify response successful
5. Verify failover logged

**Test Command:**
```bash
go test -v -run TestFailoverFlow ./internal/integration/...
```

### IT-3: Billing End-to-End
**Scenario:** Request → Usage → Bill → Export
**Steps:**
1. Make multiple requests
2. Verify usage records created
3. Generate monthly bill
4. Export PDF
5. Export CSV
6. Verify totals match

**Test Command:**
```bash
go test -v -run TestBillingFlow ./internal/integration/...
```

## Test Execution Order

### Wave 0: Foundation (Week 1-2)
- AC-1.1: Bifrost initialization
- AC-3.1: API key format
- AC-2.3: Tiered pricing calculation

### Wave 1: Core API (Week 3-4)
- AC-1.2: Chat completion non-streaming
- AC-1.3: Chat completion streaming
- AC-3.2: API key creation
- AC-3.3: API key validation

### Wave 2: Advanced Features (Week 5-6)
- AC-1.4: All OpenAI endpoints
- AC-2.1: Profit-based routing
- AC-2.2: Profit protection
- AC-3.4: Concurrency limit
- AC-4.1: Membership discounts

### Wave 3: Billing & Export (Week 7-8)
- AC-5.1: Dual-record billing
- AC-5.2: Idempotent recording
- AC-5.4: Monthly bill generation
- AC-5.5: PDF export
- AC-5.6: CSV export

### Wave 4: Media & Integration (Week 9-10)
- AC-6.1: Image generation
- AC-6.2: TTS
- AC-6.3: STT
- AC-7.1: Embeddings
- AC-7.2: Reranking
- IT-1 to IT-3: Integration tests

## Validation Checklist

Before `/gsd-verify-work`:

- [ ] All AC-1.x tests pass (Bifrost integration)
- [ ] All AC-2.x tests pass (Routing & pricing)
- [ ] All AC-3.x tests pass (API keys)
- [ ] All AC-4.x tests pass (Membership)
- [ ] All AC-5.x tests pass (Billing)
- [ ] All AC-6.x tests pass (Media)
- [ ] All AC-7.x tests pass (Embeddings)
- [ ] All performance benchmarks met (PB-1 to PB-4)
- [ ] All security validations pass (SV-1 to SV-4)
- [ ] All edge cases handled (EC-1 to EC-7)
- [ ] All integration tests pass (IT-1 to IT-3)
- [ ] Code coverage > 80% for new code
- [ ] No critical vulnerabilities in security scan
- [ ] Documentation updated

---

**Phase:** 2-核心功能开发
**Validation criteria created:** 2026-05-08
**Ready for execution:** Yes
