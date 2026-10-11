# Testing

- Prefer repeatable, verifiable end-to-end CLI workflows for complex behavior, and define their expected artifacts before implementation.
- Create relevant regression or isolated tests before writing new behavior. Use focused isolated tests when E2E cannot exercise a failure, after enumerating the failure paths. Follow the [E2E-first testing decision](../adr/0005-e2e-first-testing.md).
- Do not add tests that merely mirror constants, test getters, or match arbitrary implementation strings.
- Treat coverage as evidence, not as a numerical target to game.
- `mise run check` must pass before push; it runs formatting, build, test, lint and conventions.
