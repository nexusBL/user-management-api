# Interview Preparation

Below are likely interview questions and concise but strong answers based specifically on this project.

## Go Language

1. **Why did you use `time.Time` for DOB inside the application?**  
Because `time.Time` is the natural Go representation for dates and allows safe comparison, parsing, and formatting. I only convert it to a string at the HTTP boundary.

2. **Why is age not stored in the database?**  
Age is derived from DOB and changes over time. Storing it creates stale data and forces scheduled updates. DOB is the source of truth; age should be computed.

3. **Why inject `time.Now` into the service?**  
It makes the service deterministic and testable. In tests I could replace the clock with a fixed function if needed.

4. **Why use pointers for the handler and service structs?**  
They hold dependencies and avoid copying. Pointer receivers are the standard choice for stateful structs.

5. **Why are your error variables package-level vars?**  
Sentinel errors make classification easy with `errors.Is`, which keeps transport mapping clean.

6. **Why parse DOB with `time.ParseInLocation`?**  
DOB is date-only data. Parsing with a fixed location like UTC avoids timezone ambiguity.

7. **Why trim the name in the service layer?**  
The service owns business-level data normalization before persistence. The handler validates shape; the service normalizes domain values.

8. **Why return zero for negative ages in the utility?**  
It is a defensive guard. The service should already reject future DOB values, but the utility stays safe if reused elsewhere.

9. **What is the purpose of `context.Context` here?**  
It carries request-scoped cancellation, deadlines, and values like request ID through repository calls.

10. **Why use interfaces for repository and service dependencies?**  
Interfaces reduce coupling and make testing easier, especially if I later add mocks or alternate implementations.

## Fiber

11. **Why choose Fiber for this assignment?**  
It is fast, lightweight, and gives an Express-like routing experience while still letting me keep a layered architecture.

12. **Why use a custom Fiber error handler?**  
It centralizes error formatting, hides internal errors from clients, and guarantees the required JSON shape.

13. **How does Fiber execute middleware?**  
Fiber runs middleware in the order registered. Each middleware calls `c.Next()` to continue, and control returns back up the chain afterward.

14. **Why add `recover` middleware?**  
Without it, panics can crash the process. Recover converts panics into errors so the app can log them and return a safe 500 response.

15. **Why return errors instead of writing every error response manually in handlers?**  
Returning errors keeps handlers smaller and pushes cross-cutting response logic into one place.

## REST API

16. **Why use `PUT` for update?**  
`PUT` fits full-resource replacement semantics. The request requires both `name` and `dob`, which matches a full update.

17. **Why return `201 Created` on create?**  
The resource was successfully created, and `201` is the correct REST status for that outcome.

18. **Why return `204 No Content` on delete?**  
Deletion succeeded and there is no representation to return. `204` communicates success without a body.

19. **Why return `404` for missing users?**  
The client asked for a specific resource identifier that does not exist, which is exactly what `404 Not Found` means.

20. **Why not return `200` for every operation?**  
Meaningful status codes improve API clarity, client behavior, debugging, and standards compliance.

## Validation

21. **Why validate in the handler layer?**  
Validation of incoming HTTP payloads is a transport concern. It prevents invalid data from reaching business logic.

22. **What does `required` do?**  
It ensures the field is present and not the zero value for its type.

23. **Why add a custom `notblank` validator?**  
`required` does not reject strings made only of spaces. `notblank` closes that gap.

24. **Why use `datetime=2006-01-02`?**  
It enforces the exact `YYYY-MM-DD` format expected by the API and matches how DOB is stored as a date.

25. **Could validation also happen in the service?**  
Yes for defensive checks. In this project, formatting validation happens in the handler, while business validation like future DOB is still enforced in the service.

## Service Layer

26. **What logic belongs in the service layer here?**  
DOB parsing, future-date rejection, pagination rules, trimming names, and dynamic age enrichment.

27. **Why not calculate age in the handler?**  
Age is business behavior, not HTTP formatting. Keeping it in the service keeps handlers thin.

28. **Why cap the page limit in the service?**  
Pagination rules are application policy. The service is the right place to enforce that policy consistently.

29. **Why does the service still validate ID positivity if the handler already parses it?**  
The service defends its own contract. That makes it safer if another handler or caller uses it in the future.

30. **What would happen if the service layer was removed?**  
Handlers would become large and mix HTTP parsing, business logic, and persistence coordination, which hurts maintainability and testability.

## Repository and SQLC

31. **Why use SQLC instead of an ORM?**  
I wanted explicit SQL with compile-time-generated Go methods. SQLC keeps SQL visible and type-safe without ORM abstraction overhead.

32. **What does SQLC generate from `query.sql`?**  
Method signatures, parameter structs, row structs, and query execution code based on the named SQL statements.

33. **Why keep a repository layer if SQLC already generates code?**  
The repository isolates generated code from the service and translates persistence-specific errors like `sql.ErrNoRows`.

34. **How does the delete flow detect `404`?**  
The SQLC method generated from `:execrows` returns affected row count. If it is zero, the repository returns `ErrUserNotFound`.

35. **Why is `CountUsers` a separate query?**  
Pagination metadata like `total` should come from a count query, not from the page slice alone.

36. **What is the benefit of `emit_interface: true` in SQLC?**  
It generates a `Querier` interface for the SQLC package, which helps testing and abstraction if needed later.

37. **What would happen if the repository layer was removed?**  
The service would directly import generated SQLC types and become more tightly coupled to persistence details.

## PostgreSQL and SQL

38. **Why store DOB as `DATE` instead of `TEXT` or `TIMESTAMP`?**  
DOB is a calendar date, not free-form text and not a moment-in-time. `DATE` is the correct database type.

39. **Why use `SERIAL` for `id`?**  
It provides a simple auto-increment integer primary key that fits the assignment.

40. **Why order `ListUsers` by `id`?**  
Pagination needs deterministic ordering. Without `ORDER BY`, page results can be unstable.

41. **Why use `RETURNING` in `INSERT` and `UPDATE`?**  
It lets PostgreSQL send the final row back immediately, which avoids an extra query.

42. **What does `COUNT(*)` return in PostgreSQL?**  
It returns the number of rows as a bigint, which maps well to Go `int64`.

## Logging and Middleware

43. **Why use Zap instead of `fmt.Println`?**  
Zap produces structured, leveled, high-performance logs that are much better for production debugging and aggregation.

44. **What makes a log 'structured'?**  
Instead of plain text only, it stores named fields like `request_id`, `status`, and `duration_ms`, which log tools can filter and query.

45. **Why log request duration in middleware instead of inside each handler?**  
Latency logging is cross-cutting. Middleware applies it consistently to every endpoint.

46. **Why place request ID middleware before request duration logging?**  
So the duration logger can include the request ID in every log line.

47. **How is the request ID propagated?**  
It is stored in Fiber locals, returned in the response header, and added to Go context for downstream access.

## Docker and Deployment

48. **Why use a multi-stage Docker build?**  
The builder image has Go tooling, while the runtime image is smaller and cleaner. This reduces attack surface and image size.

49. **Why create a non-root user in the runtime image?**  
Running as non-root is a standard container security practice.

50. **Why add a separate migration service in Docker Compose?**  
It keeps schema setup reproducible and makes the app start on a ready schema.

51. **Why expose port `3000` in the container?**  
That matches the application default and keeps host-to-container networking simple.

## Error Handling

52. **Why centralize error responses?**  
It guarantees consistency, reduces duplication, and prevents accidental leakage of raw internal errors.

53. **When do you return `500`?**  
For unexpected failures such as database connectivity issues or unhandled internal errors.

54. **Why hide internal error details from the client?**  
Internal messages may expose implementation details or sensitive information. Clients only need a safe message.

## System Design and Architecture

55. **What is the main architectural benefit of this layered design?**  
Separation of concerns. Each layer changes for a different reason and can be tested or replaced more independently.

56. **If the app needed caching later, where would it go?**  
Usually in the service layer orchestrating behavior, with the cache client injected as another dependency.

57. **If you added authentication later, where would it live?**  
Authentication would usually enter through middleware, with authorization checks in handlers or services depending on the rule.

58. **How would you make this production stronger?**  
Add integration tests, metrics, health checks, DB migrations in CI/CD, configuration validation on startup, and observability tooling.

59. **How would you support partial updates?**  
Add a `PATCH /users/:id` endpoint with optional fields and distinct DTO validation rules.

60. **How would you scale the app horizontally?**  
Keep the app stateless, place it behind a load balancer, rely on PostgreSQL for shared persistence, and use request IDs for tracing across instances.

## Age-Specific Questions

61. **How does the age algorithm work?**  
It subtracts birth year from current year, then subtracts one more if the birthday has not happened yet this year.

62. **How do you handle leap-year birthdays?**  
This implementation lets Go normalize Feb 29 to March 1 in non-leap years, so age increases on March 1 in those years.

63. **Why normalize both dates to midnight UTC before comparing?**  
Age depends on dates, not times. Normalization prevents timezone and time-of-day differences from creating off-by-one errors.

64. **What edge cases did you test?**  
Birthday today, birthday tomorrow, leap-day birthday in leap and non-leap years, and a defensive future-date case.

65. **Why are unit tests especially important for age calculation?**  
Date logic often looks simple but fails on boundaries. Small mistakes lead to incorrect ages that are hard to notice without tests.
