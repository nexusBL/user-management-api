# File Guide

This document explains why each file exists, what its key functions do, how its imports are used, and how it interacts with the rest of the architecture.

## Architecture Overview

Required flow:

`Request -> Route -> Handler -> Service -> Repository -> SQLC -> PostgreSQL`

### Route Layer

- Responsibility: map HTTP method and URL to the correct handler.
- Why it exists: it keeps endpoint registration out of business logic.
- Why logic belongs here: only transport routing concerns should live here.
- If removed: `main.go` becomes cluttered and endpoint management becomes harder to explain.

### Handler Layer

- Responsibility: parse requests, validate input, call services, shape HTTP responses.
- Why it exists: it isolates HTTP-specific concerns from business rules.
- Why logic belongs here: body parsing, query parsing, status codes, and JSON formatting are HTTP concerns.
- If removed: service code becomes tightly coupled to Fiber and harder to test or reuse.

### Service Layer

- Responsibility: business logic, age calculation orchestration, pagination rules, DOB rules.
- Why it exists: it is the place for application decisions that are not specific to HTTP or SQL.
- Why logic belongs here: age calculation and pagination rules are business behavior, not database behavior.
- If removed: handlers would become fat and mix validation, business rules, and transport concerns.

### Repository Layer

- Responsibility: abstract data access behind a small interface and translate SQLC output into app models.
- Why it exists: it isolates SQL and SQLC details from the service layer.
- Why logic belongs here: not-found translation and SQLC parameter mapping are persistence concerns.
- If removed: service code would depend directly on generated SQLC code and `database/sql`.

### SQLC Layer

- Responsibility: generate type-safe Go methods from SQL queries.
- Why it exists: it avoids handwritten row scanning and reduces runtime SQL mistakes.
- Why logic belongs here: SQL shape and query result types are persistence mechanics.
- If removed: you would write manual SQL execution, manual scanning, and duplicate query boilerplate.

## File-by-File Notes

### `go.mod`

- Why it exists: defines the Go module and enables dependency management.
- Important content: the module path is used by every internal import.
- Imports: none, because it is a module metadata file.
- Interaction: every `.go` file depends on the module path declared here.

### `go.sum`

- Why it exists: records cryptographic checksums for module versions downloaded by Go.
- Important content: exact dependency verification hashes.
- Imports: none, because it is a dependency lock file.
- Interaction: Go uses it during dependency resolution to ensure reproducible and tamper-checked downloads.

### `.gitignore`

- Why it exists: prevents local artifacts such as `.env`, compiled binaries, and coverage reports from entering version control.
- Interaction: protects repo hygiene but does not affect runtime behavior.

### `.env.example`

- Why it exists: documents the environment variables the application expects.
- Interaction: mirrors the config keys parsed in `config/config.go`.

### `cmd/server/main.go`

- Why it exists: application entry point.
- Important functions:
  - `main`: loads config, creates logger, opens DB, wires dependencies, registers middleware/routes, and handles graceful shutdown.
  - `openDB`: creates the PostgreSQL connection pool and pings the database before serving traffic.
- Imports:
  - `context`, `os/signal`, `syscall`, `time`: graceful startup and shutdown.
  - `database/sql`: shared DB handle for SQLC.
  - `log`: bootstrap logging before Zap is ready.
  - `github.com/gofiber/fiber/v2`: HTTP app.
  - `github.com/gofiber/fiber/v2/middleware/recover`: panic recovery middleware.
  - `github.com/lib/pq`: PostgreSQL driver registered through side effects.
  - `go.uber.org/zap`: structured startup/shutdown logs.
  - internal packages: dependency wiring.
- Interaction:
  - calls `config.Load`
  - creates logger via `internal/logger`
  - creates repository/service/handler chain
  - registers middleware and routes
  - starts the Fiber server

### `config/config.go`

- Why it exists: central place for environment-based configuration.
- Important functions:
  - `Load`: reads environment variables, validates pagination values, and returns a typed config struct.
  - `getEnv`, `getEnvInt32`, `getEnvDuration`: helper parsers.
- Imports:
  - `os`: reads environment variables.
  - `strconv`: parses integer env values.
  - `time`: parses `SHUTDOWN_TIMEOUT`.
  - `fmt`: returns useful configuration errors.
- Interaction: used only by `main.go`, but its values influence DB, server port, and pagination behavior.

### `db/migrations/000001_create_users.up.sql`

- Why it exists: creates the `users` table.
- SQL explanation:
  - `id SERIAL PRIMARY KEY`: auto-increment integer identifier.
  - `name TEXT NOT NULL`: required user name.
  - `dob DATE NOT NULL`: date of birth stored as a real date, not a string.
- Interaction: PostgreSQL schema source for both runtime storage and SQLC type generation.

### `db/migrations/000001_create_users.down.sql`

- Why it exists: rollback script for the table creation migration.
- SQL explanation:
  - `DROP TABLE IF EXISTS users`: safely removes the table during rollback.
- Interaction: used by migration tools for reverse changes.

### `db/sqlc/query.sql`

- Why it exists: contains the raw SQL that SQLC reads to generate Go methods.
- Queries:
  - `CreateUser`: inserts a row and returns the inserted columns.
  - `GetUserByID`: reads one user by primary key.
  - `UpdateUser`: updates the full resource for `PUT` and returns the updated row.
  - `DeleteUser`: deletes by ID and returns affected row count.
  - `ListUsers`: stable list query ordered by `id` with `LIMIT/OFFSET`.
  - `CountUsers`: supports pagination metadata.
- Interaction: SQLC uses the names like `CreateUser` and `ListUsers` to generate Go methods with matching names.

### `db/sqlc/sqlc.yaml`

- Why it exists: tells SQLC where the schema and queries live and where generated Go code should be written.
- Important fields:
  - `version: "2"`: current SQLC config format.
  - `engine: "postgresql"`: PostgreSQL dialect.
  - `schema`: points to migration files so SQLC knows table shapes.
  - `queries`: points to `query.sql`.
  - `package`: generated Go package name.
  - `out`: generated code destination.
  - `sql_package: "database/sql"`: SQLC generates code around Go's standard DB abstraction.
  - `emit_interface: true`: generates a `Querier` interface for easier testing/mocking.
  - `emit_json_tags: true`: adds JSON tags to generated structs.
- Interaction: running `sqlc generate` reads this file and writes code into `internal/repository/sqlc`.

### `internal/handler/error_handler.go`

- Why it exists: centralized Fiber error handler that always returns `{"error":"message"}`.
- Important functions:
  - `NewErrorHandler`: converts returned errors into proper HTTP status codes and JSON responses.
  - `requestIDFromCtx`: enriches logs with request tracing data.
- Imports:
  - `errors`: unwraps `*fiber.Error`.
  - `github.com/gofiber/fiber/v2`: Fiber error types and status codes.
  - `go.uber.org/zap`: structured error logging.
  - `internal/models`: error response payload struct.
- Interaction: plugged into `fiber.Config` by `main.go`; every handler error flows through this function.

### `internal/handler/user_handler.go`

- Why it exists: HTTP layer for the user resource.
- Important functions:
  - `NewValidator`: creates the validator instance and registers the custom `notblank` tag.
  - `NewUserHandler`: constructor for dependency injection.
  - `CreateUser`, `GetUserByID`, `UpdateUser`, `DeleteUser`, `ListUsers`: HTTP endpoint handlers.
  - `parseUserID`, `parseOptionalInt32`: convert path/query strings into typed values.
  - `validationErrorMessage`: converts validator output into user-friendly messages.
  - `mapServiceError`: maps domain/service errors to HTTP errors.
  - `toUserResponse`, `toListUsersResponse`: response shaping helpers.
- Imports:
  - `strconv`, `strings`: parse IDs and format validation messages.
  - `errors`: classify validation/service errors.
  - `validator/v10`: request validation engine.
  - `fiber/v2`: HTTP request/response handling.
  - internal `models`, `repository`, `service`: payloads and business dependencies.
- Interaction:
  - receives the request from `routes.Register`
  - validates input before calling the service
  - maps service output into JSON responses

### `internal/logger/logger.go`

- Why it exists: central logger construction.
- Important function:
  - `New`: returns a development logger locally and a production JSON logger in production.
- Imports:
  - `strings`: case-insensitive environment check.
  - `zap`: logger creation.
- Interaction: `main.go` creates one logger and injects it into middleware and the error handler.

### `internal/middleware/request_id.go`

- Why it exists: attaches a request ID to each request for traceability.
- Important function:
  - `RequestID`: reads `X-Request-ID` if supplied, otherwise generates a UUID, stores it in Fiber locals, response headers, and Go context.
- Imports:
  - `context`: attaches request-scoped values to Go context.
  - `fiber/v2`: middleware contract.
  - `google/uuid`: unique ID generation.
- Interaction: runs before handlers so logs and downstream code can access the request ID.

### `internal/middleware/request_duration.go`

- Why it exists: logs request latency and metadata after the downstream chain finishes.
- Important function:
  - `RequestDuration`: measures duration around `c.Next()` and writes a structured log.
- Imports:
  - `time`: duration measurement.
  - `fiber/v2`: middleware contract.
  - `zap`: structured request logs.
- Interaction: wraps the request lifecycle and logs data produced by handlers and the error handler.

### `internal/models/user.go`

- Why it exists: keeps shared request/response/domain data structures in one place.
- Important structs:
  - `CreateUserRequest`, `UpdateUserRequest`: request DTOs with validator tags.
  - `User`: internal domain model used between repository, service, and handler.
  - `UserList`: service output for pagination.
  - `UserResponse`, `ListUsersResponse`, `ErrorResponse`: HTTP response DTOs.
- Imports:
  - `time`: `DOB` is a real `time.Time` inside the application.
- Interaction: shared by handler, service, and repository.

### `internal/repository/user_repository.go`

- Why it exists: persistence adapter that hides SQLC from the service layer.
- Important functions:
  - `NewUserRepository`: wraps the generated SQLC `Queries`.
  - CRUD methods: convert method arguments into SQLC params and translate SQLC results into `models.User`.
  - `toModel`: maps generated SQLC struct to app model.
- Imports:
  - `context`: DB calls are context-aware.
  - `database/sql`: detects `sql.ErrNoRows`.
  - `errors`: maps not-found cases.
  - `time`: DOB persistence argument type.
  - internal `models`: domain return type.
  - generated `sqlc`: low-level query methods and SQL param/result structs.
- Interaction:
  - called by service
  - calls generated SQLC code
  - maps DB-specific errors into repository-friendly errors

### `internal/routes/routes.go`

- Why it exists: cleanly registers the REST endpoints in one place.
- Important function:
  - `Register`: binds URL patterns to handler methods.
- Imports:
  - `fiber/v2`: router type.
  - `internal/handler`: handler dependency.
- Interaction: bridge between the Fiber app and the handler layer.

### `internal/service/user_service.go`

- Why it exists: contains business logic.
- Important functions:
  - `NewUserService`: constructor with dependency injection and an injectable clock.
  - `CreateUser`, `GetUserByID`, `UpdateUser`, `DeleteUser`, `ListUsers`: service use cases.
  - `normalizePagination`: enforces pagination rules and caps the limit.
  - `withAge`, `withAges`: enrich users with dynamic age.
  - `parseDOB`: parses `YYYY-MM-DD` and rejects future birth dates.
- Imports:
  - `context`: passes request context to repository.
  - `errors`: domain error values.
  - `strings`: trims names before persistence.
  - `time`: DOB parsing and injected clock.
  - internal `models`, `repository`, `utils`: domain models, persistence abstraction, and age calculation.
- Interaction:
  - called by handler
  - calls repository
  - uses `utils.CalculateAge`

### `internal/utils/age.go`

- Why it exists: dedicated reusable age calculation utility required by the assignment.
- Important functions:
  - `CalculateAge`: computes age from DOB and current date.
  - `normalizeDate`: strips time-of-day and forces UTC so comparisons are date-based.
- Imports:
  - `time`: date math.
- Interaction: only the service calls it, which keeps business logic centralized.

### `internal/utils/age_test.go`

- Why it exists: verifies the age algorithm, especially edge cases.
- Important test cases:
  - birthday today
  - birthday tomorrow
  - leap-year birthday before March 1 in a non-leap year
  - leap-year birthday on March 1 in a non-leap year
  - leap-day birthday on an actual leap year
  - future DOB defensive behavior
- Imports:
  - `testing`: Go test framework.
  - `time`: deterministic test dates.
- Interaction: validates the utility independently from HTTP or DB code.

### `Dockerfile`

- Why it exists: containerizes the API for repeatable builds and deployment.
- Key lines:
  - first stage uses `golang` image to download modules, run tests, and compile the binary.
  - second stage uses a small Alpine runtime image.
  - `ca-certificates` enables outbound TLS if needed.
  - non-root user improves container security.
- Interaction: used by `docker compose` and standalone `docker build`.

### `docker-compose.yml`

- Why it exists: runs app, database, and migrations together.
- Services:
  - `postgres`: PostgreSQL database.
  - `migrate`: applies SQL files before the app starts.
  - `app`: builds and runs the Fiber API.
- Interaction: makes onboarding and demo execution simpler.

### `README.md`

- Why it exists: main onboarding document for setup, commands, endpoints, and sample requests.
- Interaction: first document a reviewer or interviewer can read to understand how to run the project.

### `FILE_GUIDE.md`

- Why it exists: detailed implementation explanation for every important source/config/generated file.
- Interaction: your companion guide for line-by-line interview review.

### `INTERVIEW_PREPARATION.md`

- Why it exists: focused question-and-answer preparation tailored to this project.
- Interaction: helps you rehearse the architecture, design choices, and edge cases before the interview.

### `internal/repository/sqlc/db.go`

- Why it exists: SQLC-generated foundation for executing queries.
- Important content:
  - `DBTX`: interface for any object that can execute SQL calls, such as `*sql.DB` and `*sql.Tx`.
  - `New(db DBTX)`: constructor that creates a `Queries` instance.
  - `Queries`: generated query holder struct.
  - `WithTx(tx *sql.Tx)`: creates a `Queries` instance bound to a transaction.
- Imports:
  - `context`: required by DB method signatures.
  - `database/sql`: transaction and result types.
- Interaction: repository injects a `*sql.DB` into `sqlcgen.New(db)` to get a query executor.

### `internal/repository/sqlc/models.go`

- Why it exists: SQLC-generated structs that map directly to database rows.
- Important content:
  - `User`: row model for the `users` table with `ID`, `Name`, and `Dob`.
- Imports:
  - `time`: PostgreSQL `DATE` becomes Go `time.Time`.
- Interaction: repository converts this generated row model into the app-level `models.User`.

### `internal/repository/sqlc/querier.go`

- Why it exists: SQLC-generated interface for all query methods.
- Important content:
  - `Querier`: includes `CountUsers`, `CreateUser`, `DeleteUser`, `GetUserByID`, `ListUsers`, and `UpdateUser`.
  - `var _ Querier = (*Queries)(nil)`: compile-time assertion that `Queries` implements `Querier`.
- Imports:
  - `context`: every generated query method is context-aware.
- Interaction: helpful if you ever want to mock generated queries or swap implementations in tests.

### `internal/repository/sqlc/query.sql.go`

- Why it exists: SQLC-generated Go implementation of the SQL in `db/sqlc/query.sql`.
- Important content:
  - `CountUsers(ctx) (int64, error)`: runs `COUNT(*)`.
  - `CreateUserParams` and `CreateUser(ctx, arg)`: inserts a user and scans the returned row.
  - `DeleteUser(ctx, id) (int64, error)`: executes delete and returns affected rows because the query used `:execrows`.
  - `GetUserByID(ctx, id) (User, error)`: fetches one row.
  - `ListUsersParams` and `ListUsers(ctx, arg) ([]User, error)`: paginated list query.
  - `UpdateUserParams` and `UpdateUser(ctx, arg)`: full update query with returned row.
- Imports:
  - `context`: query cancellation and deadlines.
  - `time`: parameter and result type for `dob`.
- Interaction: repository calls these generated methods instead of writing raw `QueryRowContext` or `Scan` logic by hand.

## SQLC Generated Code

### How SQLC Works

1. It reads the schema from the migration files.
2. It reads named queries from `query.sql`.
3. It infers parameter types and result row types.
4. It generates type-safe Go code so you call methods instead of manually scanning rows.

### Why SQLC Is Useful

- compile-time query signatures
- no reflection
- no ORM magic
- less boilerplate than `database/sql`
- easier to review in interviews because the SQL stays explicit

## Age Calculation Notes

### Why Age Is Not Stored In The Database

- Age changes every year without the DOB changing.
- Stored age becomes stale unless you run background updates.
- DOB is the source of truth; age is a derived value.

### Line-by-Line Algorithm

1. Convert both `dob` and `today` to UTC and remove the time-of-day.
2. Start with `today.Year() - dob.Year()`.
3. Build `birthdayThisYear` using the current year and the birth month/day.
4. If today is before this year's birthday, subtract one.
5. If the result is negative, return zero defensively.

### Leap-Year Handling

- For a `Feb 29` birth date in a non-leap year, `time.Date(currentYear, February, 29, ...)` normalizes to `March 1`.
- That means this implementation treats the birthday as `March 1` in non-leap years.
- This is deterministic and easy to defend in an interview.

## Validation Notes

- `required`: field must be present and non-zero.
- `notblank`: custom rule so `"   "` is rejected after trimming.
- `min=2`: avoids single-character names.
- `max=100`: protects the API from unbounded string sizes.
- `datetime=2006-01-02`: forces `YYYY-MM-DD`.

Validation is done before business logic because:

- invalid data should be rejected early
- service code should receive already well-formed input
- error messages become clearer and more consistent

## Logging Notes

- Zap uses structured fields like `request_id`, `status`, and `duration_ms`.
- Structured logging is better than `fmt.Println` because logs are searchable, machine-readable, and consistent.
- Zap is preferred because it is fast, production-proven, and has a clean API for contextual fields.

## Middleware Notes

- Middleware exists to run cross-cutting logic around many handlers without duplicating code.
- Fiber executes middleware in registration order.
- `c.Next()` passes control to the next middleware or final handler.
- When the downstream chain finishes, execution resumes in the current middleware.

For this project:

- `RequestID` runs early so every later log can include the ID.
- `RequestDuration` wraps the whole request and logs the final status and elapsed time.

## Error Handling Notes

- `400 Bad Request`: malformed JSON, invalid IDs, invalid query params, invalid DOB format, future DOB, validation failures.
- `404 Not Found`: user ID does not exist.
- `500 Internal Server Error`: unexpected DB or server error.
- `201 Created`: successful `POST /users`.
- `204 No Content`: successful delete with no response body.
