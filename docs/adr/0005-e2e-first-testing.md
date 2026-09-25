# 0005: E2E-First Testing

Status: Accepted

Context: Help-text, flag-default, envelope-shape, and fake-helper self-tests asserted constants via string matching. Coverage patch target drove test creation.

Decision: E2E is default. FakeZenodo integration plus file roundtrips are the artifact tests. Isolated tests remain only for file permissions, secret indirection, XDG paths, retry parsing, fuzz parsers, auth and read-only gates, and error mapping with exact codes. Coverage informational: patch 80 to 50, threshold 2 to 5, output package excluded.

Consequences: Deleted context, completion, version, api-seal, example, renderer, errors-mirror, and fake self-tests. Exit-code table retained.
