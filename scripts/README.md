# Repository command scripts

These scripts provide stable entry points for developers, CI, and code agents. Run
them from any working directory; each script resolves the repository root itself.

| Command | Suggested command name | Purpose |
| --- | --- | --- |
| `./scripts/dev.sh` | `dev` | Start the active Go API and Vue dev server together. |
| `./scripts/lint.sh` | `lint` | Run frontend typechecking/ESLint plus Go format, vet, and NilAway checks. |
| `./scripts/test.sh` | `test` | Run the active Go test suite. |
| `./scripts/build.sh` | `build` | Build the production Vue assets and Go binary. |
| `./scripts/prod-smoke.sh` | `prod-smoke` | Build and boot the production Docker image, then verify HTTP and SQLite startup. |
| `./scripts/verify.sh` | `verify` | Run lint, tests, and native production builds. |
| `./scripts/ci.sh` | `ci` | Install locked dependencies and reproduce CI verification plus the container smoke test. |

`prod-smoke` and `ci` require a running Docker daemon. Override the smoke-test
port with `SMOKE_PORT` and its local image tag with `IMAGE_NAME` when needed.
