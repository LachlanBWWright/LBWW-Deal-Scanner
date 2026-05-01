import { fromThrowable, Result, ResultAsync } from "neverthrow";

export function toError(error: unknown, context?: string): Error {
  const base =
    error instanceof Error
      ? error
      : new Error(typeof error === "string" ? error : JSON.stringify(error));

  if (!context) return base;
  return new Error(`${context}: ${base.message}`);
}

export function fromThrowableSync<T>(
  operation: () => T,
  context?: string,
): Result<T, Error> {
  return fromThrowable(operation, (error) => toError(error, context))();
}

export function fromThrowableAsync<T>(
  operation: () => Promise<T>,
  context?: string,
): ResultAsync<T, Error> {
  return ResultAsync.fromThrowable(operation, (error) =>
    toError(error, context),
  )();
}
