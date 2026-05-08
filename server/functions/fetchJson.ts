import { errAsync, okAsync, ResultAsync } from "neverthrow";
import { z } from "zod";
import type { Dispatcher } from "undici";

import { toError } from "./neverthrowUtils.js";

type FetchInit = RequestInit & { dispatcher?: Dispatcher };

function fetchJson(
  url: string,
  init: FetchInit | undefined,
  context: string,
): ResultAsync<unknown, Error> {
  return ResultAsync.fromPromise(fetch(url, init), (error) =>
    toError(error, context),
  ).andThen((response) => {
    if (!response.ok) {
      return errAsync(
        new Error(`${context}: ${response.status} ${response.statusText}`),
      );
    }

    return ResultAsync.fromPromise(response.json(), (error) =>
      toError(error, `${context}: invalid JSON response`),
    );
  });
}

export function fetchAndValidateJson<S extends z.ZodTypeAny>(
  url: string,
  schema: S,
  init: FetchInit | undefined,
  context: string,
): ResultAsync<z.output<S>, Error> {
  return fetchJson(url, init, context).andThen((data) => {
    const validated = schema.safeParse(data);

    if (!validated.success) {
      return errAsync(new Error(`${context}: ${validated.error.message}`));
    }

    return okAsync(validated.data);
  });
}