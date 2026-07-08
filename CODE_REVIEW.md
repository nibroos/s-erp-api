# S-ERP-API — Code Review & Recommendations

> **Scope:** Full-project review of the Go/Fiber ERP backend under `service/`, plus the
> `docker/`, `gateway/`, and `scripts/` supporting infrastructure.
> **Nature:** This is an advisory document only. **No code has been changed.** Each item
> lists the problem, why it matters, the trade-off, and a concrete recommendation.
> **Reviewed at:** commit `8ec977e` (branch `main`), 2026-07-02.

---

## How to read this document

Findings are grouped by area and tagged with a severity:

| Tag | Meaning |
|-----|---------|
| 🔴 **Critical** | Exploitable security hole or data-loss/crash risk. Fix before any production exposure. |
| 🟠 **High** | Serious correctness, security, or maintainability problem. Schedule soon. |
| 🟡 **Medium** | Real weakness; worth fixing but not urgent. |
| 🔵 **Low / Nit** | Polish, hygiene, or DX improvement. |

Each finding cites a representative file/line. Many issues repeat across the ~41
controller/repository pairs, so treat single citations as *examples of a pattern*.

---

## 1. Executive Summary

The project is a feature-rich ERP backend (users, RBAC data model, inventory, sales/purchase
orders, invoicing, quotations, scheduling, messaging, observability). The engineering
ambition is high: Prometheus + Grafana, Jaeger tracing, RabbitMQ, Asynq workers, k6/vegeta
load tests, and a Docker/Nginx deployment stack are all wired up.

However, several **critical security controls are disabled or missing**, there is a
**systemic SQL-injection pattern**, and the codebase carries a large amount of
**copy-paste duplication with zero automated tests**. The most urgent items:

1. 🔴 **Authorization is turned off** — every `PermissionMiddleware` call is commented out, so any authenticated user can call every endpoint (see §2.1).
2. 🔴 **SQL injection** via unvalidated `ORDER BY` and `IN (...)` interpolation in ~36 repositories (see §2.2).
3. 🔴 **Secrets committed to the repo** and **baked into the Docker image** (see §2.3).
4. 🔴 **Unauthenticated destructive seeder endpoint** and public debug/panic routes (see §2.4).
5. 🟠 **Zero tests** despite a `mocks/` folder and `testify` in `go.mod` (see §5).
6. 🟠 **Data race** on a shared error variable in the concurrent query pattern (see §4.2).

Fixing 1–4 is a prerequisite to calling this service production-ready. Items 5–6 and the
architecture notes in §3 are what will keep it maintainable as it grows.

---

## 2. Security

### 2.1 🔴 Authorization (RBAC) is completely disabled

`PermissionMiddleware` exists and works, but **every single usage in the route files is
commented out** (22 commented occurrences, 0 active):

```go
// internal/routes/product_routes.go
// products.Post("/index-product", middleware.PermissionMiddleware("read_masters"), productController.GetProducts)
products.Post("/index-product", productController.GetProducts)   // <-- no permission check
```

The JWT carries `roles`, `permissions`, and `bid` (branch id), and the DB has a full
groups/roles/permissions model — but none of it is enforced at the HTTP layer.

**Impact:** Any user who can log in (including a self-registered `RoleStudent` from
`/auth/register`) can read and mutate **all** ERP data across **all** branches — create
users, delete invoices, adjust inventory, etc. This is a textbook *Broken Access Control*
(OWASP A01) and *broken tenant isolation*: the `bid` claim is issued but never used to scope
queries.

**Recommendation:**
- Re-enable `PermissionMiddleware` on every route, or (better) enforce it centrally in each
  `Setup*Routes` group so a new endpoint can't accidentally ship unprotected.
- Add **branch scoping**: filter list/detail queries by the authenticated user's `bid` unless
  they hold an explicit cross-branch permission.
- Consider a default-deny posture: a small wrapper that requires an explicit permission string
  per handler and fails closed if none is declared.

**Trade-off:** Turning RBAC back on will break clients that currently rely on the open
behavior; roll out behind a flag and seed a correct permission matrix first.

---

### 2.2 🔴 SQL injection via `ORDER BY` and `IN (...)` interpolation

The list repositories build SQL by string-concatenating user-controlled filter values.

**(a) `ORDER BY` column & direction are not validated** in ~36 repositories:

```go
// internal/repository/user_repository.go:105
orderColumn := utils.GetStringOrDefault(filters["order_column"], "id")
orderDirection := utils.GetStringOrDefault(filters["order_direction"], "asc")
query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)
```

`order_column`/`order_direction` come straight from the request body (via
`ConvertRequestToFilters`). Column names and `ASC/DESC` **cannot** be passed as bound
parameters, so this interpolation is injectable
(e.g. `order_column = "id; <subquery>"` or boolean-based extraction through ordering).

A safe helper already exists — `utils.GetStringOrDefaultFromArray(value, allowedValues, ...)`
— but only **3** repositories use it while **36** use the unchecked `GetStringOrDefault`.

**(b) `IN (...)` lists are interpolated raw:**

```go
// internal/repository/user_repository.go:88-95
query += fmt.Sprintf(" AND %s IN (%s)", valueID, value)   // role_ids
query += fmt.Sprintf(" AND id IN (%s)", filters["ids"])    // ids
```

`filters["ids"]` and `filters["role_ids"]` are user-supplied strings dropped directly into
the query. `... IN (1); DROP ...`-style payloads and blind-injection are possible.

**Recommendation:**
- **`ORDER BY`:** whitelist column names against an allowed set per model (use the existing
  `GetStringOrDefaultFromArray` everywhere), and map direction to a strict `"asc"`/`"desc"`.
- **`IN (...)`:** split the CSV, validate each element is an integer, and bind them as
  `$n` placeholders (`pq.Array` or expanded placeholders), never interpolate.
- Consider adopting a query builder (`squirrel`, or Bun/ent) so parameterization is the
  default and hand-rolled string SQL disappears.

**Trade-off:** A whitelist means adding new sortable columns requires a one-line change —
a small, worthwhile cost for closing the hole.

---

### 2.3 🔴 Secrets in the repository and in the image

- `service/.env.example` and `docker/.env.example` contain **real-looking credentials**, not
  placeholders: `JWT_SECRET=nibrossecret`, `RABBITMQ_PASSWORD=yubi9911`, and
  live-looking SMTP creds (`SMTP_EMAIL_USER=developer@yubiteck.com`,
  `SMTP_HOST=mail.yubiteck.com`). If any of these match production, they are compromised.
- The `Dockerfile` bakes the env file into the image:
  ```dockerfile
  COPY .env /apps/.env
  ```
  Anyone who can pull the image gets the secrets, and they persist in image layers even if
  later removed.
- The Postman collection (`S-ERP-API.postman_collection.json`) and README publish working
  login credentials (`yubi@email.com` / `erppostman1!`).

**Recommendation:**
- Rotate every secret that has ever been committed (JWT secret, DB, RabbitMQ, SMTP).
- Make `.env.example` contain **only dummy placeholders**; keep real `.env` out of git
  (it already is via `.gitignore`) *and* out of the image — inject at runtime via Docker
  secrets, compose `env_file`, or an orchestrator secret store.
- A leaked JWT secret is the worst of these: with `nibrossecret` anyone can forge admin
  tokens. Treat rotation as urgent.

**Trade-off:** Runtime secret injection adds a deployment step but is the only way to keep
secrets out of build artifacts.

---

### 2.4 🔴 Unauthenticated destructive & debug endpoints

These routes are registered **before** `app.Use(middleware.JWTMiddleware())` in
`internal/routes/routes.go`, so they require **no authentication**:

```go
version.Post("/seeders/run", rest.NewSeederController(sqlDB.DB).RunSeeders)
app.Get("/api/v1/test-panic", ...)          // deliberately panics
app.Get("/api/v1/test-panic-message", ...)  // deliberately panics
```

- **`/seeders/run`** wipes/reseeds core tables and is gated *only* by comparing a body field
  to `POSTGRES_PASSWORD` using a plain `!=` (`seeder_controller.go:26`). That's an
  unauthenticated, internet-reachable data-reset button, protected by a reused infra password
  and a **non-constant-time** comparison.
- **`/test-panic*`** are DoS/`recover` probes that should never exist in a production build.

**Recommendation:**
- Remove the panic routes (or compile them only under a dev build tag).
- Move seeding out of the HTTP surface entirely (a CLI command / migration job). If it must
  stay, put it behind JWT + an admin permission + a dedicated secret, and use
  `crypto/subtle.ConstantTimeCompare`.

---

### 2.5 🟠 Permissive CORS with credentials

```go
// main.go
app.Use(cors.New(cors.Config{ AllowOrigins: "*" }))
```

`AllowOrigins: "*"` allows any site to call the API from a browser. Since auth is a
bearer token in a header (not a cookie) this isn't as bad as `*` + credentials, but it still
invites token-replay from malicious pages and gives no origin control.

**Recommendation:** Restrict to known front-end origins (configurable via env). Avoid `*` in
production.

---

### 2.6 🟠 JWT weaknesses

`internal/middleware/jwt.go`:
- **Long lifetime:** `JWT_EXPIRES_MINUTE_AT=10800` = **7.5 days**, with **no refresh token and
  no revocation/blacklist**. A stolen token is valid for a week and cannot be killed (logout,
  password change, and user deletion don't invalidate issued tokens).
- **No `iss`/`aud`/`nbf` claims** and no leeway config; only `exp` is set.
- Algorithm confusion is handled (HMAC-only check) — good — but the weak shared secret (§2.3)
  undermines it.

**Recommendation:** Short-lived access tokens (minutes) + refresh tokens; a revocation list or
token-version claim checked against the DB; set and validate `iss`/`aud`. Consider `jti` for
per-token revocation.

---

### 2.7 🟠 Login has no brute-force protection

`Login` (`user_controller.go:251`) does a bcrypt compare with **no rate limiting, no lockout,
and no CAPTCHA/backoff**. bcrypt slows guessing but doesn't stop distributed credential
stuffing.

**Recommendation:** Add per-IP/per-account rate limiting (Fiber `limiter` middleware or a
Redis counter — Redis is already available) and exponential lockout after N failures.

---

### 2.8 🟡 Internal error details leaked to clients

Handlers routinely return `err.Error()` in the response body, and the global `ErrorHandler`
does too:

```go
return ctx.Status(...).JSON(fiber.Map{ "errors": err.Error(), ... })
```

This exposes SQL text, driver errors, and internal messages to callers — useful
reconnaissance for an attacker.

**Recommendation:** Return generic messages + a correlation id to clients; log the detailed
error server-side (you already have tracing/logging). Reserve verbose errors for non-prod via
an env flag.

---

### 2.9 🟡 Container runs as root

The `Dockerfile` creates `appuser` but the `USER appuser` line is **commented out**, so the
process runs as root. Combined with `app.Static("/public", "./public")` serving a
world-writable-ish upload dir, this widens the blast radius of any RCE.

**Recommendation:** Run as the non-root user; ensure the `public/` upload paths are owned by
it and validate/limit uploaded file types and sizes.

---

## 3. Architecture & Design

### 3.1 🟠 "Microservices" that are actually a modular monolith

The README and badges describe a microservice/gRPC system, but in practice there is **one Go
binary** (`service/`) that switches behavior via `SERVICE_TYPE` (`rest`/`consumer`/`scheduler`)
in `main.go`. gRPC is fully scaffolded (`proto/`, `interceptor/`) but **commented out**.

This isn't wrong — a modular monolith is often the *right* call — but the naming and dead
scaffolding create confusion and maintenance drag.

**Recommendation:** Pick a story and commit to it. Either (a) embrace the monolith and remove
the gRPC/microservice scaffolding and misleading docs, or (b) actually split services. Don't
carry both half-built.

---

### 3.2 🟠 Two DB access layers stacked on each other (GORM + sqlx + raw SQL)

`main.go` opens GORM, then wraps GORM's underlying `*sql.DB` in sqlx:

```go
gormDB, _ := gorm.Open(postgres.Open(dbURL), ...)
sqlDBGorm, _ := gormDB.DB()
sqlDB := sqlx.NewDb(sqlDBGorm, "postgres")
```

Both `gormDB` and `sqlDB` are then threaded through every repository, and connection-pool
settings are configured twice on the same underlying pool. Reads use hand-written sqlx SQL;
writes/transactions often use GORM. Two mental models, two escaping stories (see §2.2), double
the surface area.

**Recommendation:** Standardize on **one** primary data layer. If you keep raw SQL for complex
reads, wrap it behind a query builder for safe parameterization; use GORM consistently for CRUD.
Configure the pool once.

---

### 3.3 🟠 Massive duplication across ~41 controller/repository pairs

There are 41 controllers and 41 repositories, and the list/filter/search/order/paginate logic
is **copy-pasted** in each (the `for key, value := range filters` + `ORDER BY` + `LIMIT/OFFSET`
+ concurrent count/select block appears near-identically in `user_repository.go`,
`task_repository.go`, `io_type_repository.go`, `purchase_type_repository.go`, …). Some files are
enormous: `inventory_repository.go` is **4,700 lines**, `sales_order_repository.go` **4,053**,
`dtos.go` **2,912**.

**Impact:** A fix (like the SQL-injection whitelist) must be applied 36+ times; drift is
guaranteed; onboarding is slow; god-files are hard to test and review.

**Recommendation:**
- Extract a generic, reusable "list query" builder that takes an allowed-columns set,
  filter spec, and pagination — implement the safe pattern once.
- Split god-files by aggregate/use-case. Move business logic out of repositories (they should
  be thin data access) into the service layer.
- Consider Go generics for the repeated CRUD skeleton.

---

### 3.4 🟡 Everything is `POST` with a stringly-typed filter bag

428 of 431 routes are `POST` (only 3 `GET`), including pure reads like
`/index-product`, `/show-product`. The `ConvertRequestToFilters` middleware flattens the JSON
body into `map[string]string` for *every* request, so all handlers lose type information and
re-parse strings.

**Impact:** Non-idiomatic/non-RESTful (no HTTP caching, confusing for API consumers), and the
stringly-typed bag is exactly what enables the injection issues in §2.2. Numbers/bools become
strings and are re-converted ad hoc.

**Recommendation:** Use `GET` with query params for reads and typed request structs (you
already have `dtos` and a validator) instead of a global map. Keep DTO validation at the edge.

---

### 3.5 🟡 Duplicated helpers and dead/commented code

- `GetAuthUser` is defined **twice**, identically, in `internal/auth/auth.go` and
  `internal/middleware/jwt.go`.
- Large commented-out blocks litter `main.go` (gRPC server, Redis init, scheduler wiring) and
  route files. Multiple stale Dockerfiles exist (`Dockerfile.prod.backup`,
  `Dockerfile.prod.pass`).
- Root contains committed artifacts that don't belong in VCS
  (`HTTP Request Count-1739486699698.json`).

**Recommendation:** Delete dead code and duplicate helpers; git history is the backup. Keep one
canonical Dockerfile per environment. Move scratch/export artifacts out of the repo.

---

## 4. Concurrency & Correctness

### 4.1 🟠 `nil` RabbitMQ dereferenced in a `defer` → startup panic

```go
// main.go
rabbitmq, err := config.NewRabbitMQ()
if err != nil {
    log.Printf("Warning: Failed to initialize RabbitMQ: %v", err)
    // Don't fatal here, allow service to run without notifications
}
defer rabbitmq.Close()
```

`NewRabbitMQ` returns `nil, err` on failure (`config.go:51/56`). The code deliberately
continues, then `defer rabbitmq.Close()` runs `r.Channel` on a **nil receiver** → panic. So the
"allow service to run without RabbitMQ" intent is defeated: if the broker is down, the process
crashes at shutdown (and the deferred call is set up regardless).

**Recommendation:** Guard the defer (`if rabbitmq != nil { defer rabbitmq.Close() }`) and make
`Close` safe on a nil receiver. Better: represent "no broker" with a no-op publisher so callers
don't branch on nil.

---

### 4.2 🟠 Data race on the shared `selectErr` in the concurrent query pattern

The list repositories run count and select queries in two goroutines that both write the same
variable without synchronization:

```go
// user_repository.go (pattern repeated across repos)
var selectErr error
go func(){ ...; selectErr = err }()   // goroutine 1
go func(){ ...; selectErr = err }()   // goroutine 2
wg.Wait()
```

Concurrent unsynchronized writes to `selectErr` are a data race (would fail `go test -race`),
and one error can silently overwrite the other. There's also no context cancellation, so if one
query fails the other still runs to completion.

**Recommendation:** Use `errgroup.Group` with a shared `context` — it collects the first error,
cancels siblings, and is race-free. Or give each goroutine its own error and combine after
`Wait()`.

---

### 4.3 🟡 Server "restart loop" hides fatal errors

```go
for {
    if err := app.Listen(":4001"); err != nil {
        log.Printf("Server error: %v", err); time.Sleep(5*time.Second); continue
    }
    break
}
```

If the port is permanently unavailable or config is bad, this spins forever logging the same
error instead of failing fast. There's also **no graceful shutdown** (no signal handling to
drain in-flight requests / close the DB and broker cleanly).

**Recommendation:** Fail fast on unrecoverable listen errors; add `signal.NotifyContext` +
`app.ShutdownWithContext` for graceful termination. Let the orchestrator (Docker/systemd — you
already have an auto-restart unit) handle restarts.

---

### 4.4 🟡 Hardcoded ports and an odd `FetchCachedData(&fiber.Ctx{}, ...)` call

- Ports `:4001`, `:9090`, `:4010` are hardcoded in `main.go`; everything else is env-driven.
- `config.FetchCachedData(&fiber.Ctx{}, sqlDB)` passes a throwaway empty `*fiber.Ctx` just to
  satisfy a signature — a sign the function's dependency on the request context is artificial.

**Recommendation:** Drive ports from env/config; refactor `FetchCachedData` to take only what it
needs (a `context.Context` and the DB), not a fake request.

---

## 5. Testing & Quality Gates

### 5.1 🟠 No automated tests at all

`find . -name '*_test.go'` returns **0 files**, even though `internal/mocks/` exists and
`testify` is a dependency. For a financial/ERP system (invoices, inventory, tax/PPh23/VAT
calculations), the absence of tests around money math and state transitions is a significant
risk.

**Recommendation:** Prioritize tests for:
- Auth (JWT issue/verify, permission checks once re-enabled).
- The generic list/query builder (once extracted) — especially the `ORDER BY`/`IN` whitelist.
- Money and tax calculations in the invoice/sales/purchase services.
Run with `-race` in CI to catch §4.2. `Jenkinsfile` exists — wire a real `go test ./... -race`
stage and fail the build on regressions.

### 5.2 🔵 Observability metric cardinality

`PromDurationMiddleware` labels metrics with `c.Path()`. If any paths embed IDs this explodes
cardinality; current routes are static POSTs so it's fine today, but keep it in mind if you add
`/:id` style routes.

---

## 6. Ops, Docker & Docs

- 🟡 **DB/broker ports exposed to host** in `docker-compose.yml` (`5672`, `15672`, `9090`,
  `9200`, Postgres, etc.). Fine for local dev; ensure production compose does **not** publish
  datastore/management ports to the public interface.
- 🟡 **Manual, error-prone setup steps** in the README (hand-editing `schema_migrations` to fix
  migrations; manually creating the Grafana role/DB; manual systemd unit editing). These should
  be scripted/idempotent.
- 🔵 **Multiple Dockerfiles** (`Dockerfile`, `.dev`, `.test`, `.prod.backup`, `.prod.pass`)
  create ambiguity about what actually ships. Consolidate with build args/targets.
- 🔵 **`update-ms-fonts` / msttcorefonts** in the image pulls fonts over the network at build
  time (reproducibility/licensing concern for wkhtmltopdf PDF rendering). Vendor the fonts you
  need.

---

## 7. What's already good (keep doing this)

- **Algorithm-confusion protection** on JWT parsing (HMAC-only check) in both `auth.go` and
  `jwt.go`.
- **bcrypt** with `DefaultCost` for password hashing — correct choice.
- **Parameterized placeholders** *are* used for the `ILIKE`/search filters (`$n` + args) — the
  injection gaps are specifically `ORDER BY` and `IN`, not the whole layer.
- **Real observability**: Prometheus, Jaeger tracing spans threaded through the layers, health
  endpoints, and load-test tooling (k6/vegeta) show operational maturity.
- **Transactions** are used around multi-step writes (e.g. user create/delete) with rollback.
- **Panic recovery** is layered (custom `SafetyMiddleware` + Fiber `recover` with stack traces).
- **Connection pooling** is explicitly configured.
- A **safe ordering helper already exists** (`GetStringOrDefaultFromArray`) — the fix for §2.2
  is mostly adopting it everywhere rather than writing something new.

---

## 8. Suggested remediation order

| # | Item | Section | Severity | Effort |
|---|------|---------|----------|--------|
| 1 | Rotate all committed secrets; stop baking `.env` into image | §2.3 | 🔴 | Low |
| 2 | Re-enable RBAC + branch scoping | §2.1 | 🔴 | Med |
| 3 | Whitelist `ORDER BY`; bind `IN (...)` lists | §2.2 | 🔴 | Med |
| 4 | Remove panic routes; lock down/relocate seeder | §2.4 | 🔴 | Low |
| 5 | Guard nil RabbitMQ; fix `selectErr` race (errgroup) | §4.1, §4.2 | 🟠 | Low |
| 6 | Restrict CORS; shorten JWT + add revocation; login rate-limit | §2.5–2.7 | 🟠 | Med |
| 7 | Stop leaking `err.Error()`; run container as non-root | §2.8, §2.9 | 🟡 | Low |
| 8 | Extract generic list/query builder; split god-files | §3.3 | 🟠 | High |
| 9 | Add tests (auth, money math, query builder) + CI `-race` | §5.1 | 🟠 | High |
| 10 | Consolidate DB layer; docs/Dockerfile cleanup | §3.2, §6 | 🟡 | Med |

---

*Generated as a static review — no runtime exploitation was performed and no source files were
modified. Validate each finding against your current threat model and deployment topology
before acting.*
