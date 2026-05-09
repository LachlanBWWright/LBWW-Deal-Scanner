# Agent Instructions

## Error Handling

- Use `neverthrow` `Result` and `ResultAsync` as the default error-handling model.
- Never throw exceptions from application code.
- Never rely on `try`/`catch` for normal control flow.
- Return explicit `ok(...)` and `err(...)` values instead of throwing.
- Model expected failures with typed error values so callers are forced to handle them.

## Wrapping Throwable Code

- Treat external libraries, platform APIs, generated clients, and legacy modules as potentially throwable unless their contract proves otherwise.
- Wrap throwable functions at the boundary with `neverthrow` primitives such as `fromThrowable`, `Result.fromThrowable`, or `ResultAsync.fromThrowable`.
- Export wrapped functions that return `Result` or `ResultAsync`; do not export raw throwable library calls.
- Convert unknown external errors into typed domain errors at the boundary.
- Keep throwing behavior contained inside the smallest possible wrapper and never let it leak into application code.

Example pattern:

```ts
import { Result } from "neverthrow";

type ParseJsonError = {
  readonly type: "ParseJsonError";
  readonly message: string;
};

const parseJsonThrowable = Result.fromThrowable(
  JSON.parse,
  (error): ParseJsonError => ({
    type: "ParseJsonError",
    message: error instanceof Error ? error.message : "Failed to parse JSON",
  }),
);

export const parseJson = (input: string): Result<unknown, ParseJsonError> =>
  parseJsonThrowable(input);
```

## Type Safety

- Preserve strict type safety at all times.
- Never use `any`.
- Never use type assertions such as `as SomeType`, non-null assertions, or double assertions.
- Prefer `unknown` plus narrowing when handling untrusted values.
- Use discriminated unions for typed errors and state machines.
- Make invalid states unrepresentable where practical.
- Prefer explicit return types for exported functions.

## Linting And TypeScript

- Never disable ESLint rules.
- Never disable TypeScript checks.
- Do not add `eslint-disable`, `ts-ignore`, `ts-expect-error`, or equivalent suppressions.
- Fix the underlying type or lint issue instead of silencing it.
- Keep code compatible with the repository's strictest configured lint and TypeScript settings.

## Implementation Expectations

- Keep error boundaries close to external systems.
- Do not mix exception-based and `Result`-based APIs in the same module unless the exception-based API is private and immediately wrapped.
- Prefer small, composable functions that make success and failure paths explicit.
- Ensure tests cover both `ok` and `err` paths for `Result`-returning code.
