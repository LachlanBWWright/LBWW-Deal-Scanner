export function toError(error: unknown, context?: string): Error {
  const base =
    error instanceof Error
      ? error
      : new Error(typeof error === "string" ? error : JSON.stringify(error));

  if (!context) return base;
  return new Error(`${context}: ${base.message}`);
}
