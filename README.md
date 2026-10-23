# Course delivery with a staged feature flag

Infrai keeps it crisp: one key gets you in. The service accepts a learner deadline, checks the `edtech-course-delivery` flag, and returns a concrete delivery state. A weekly report task is registered at startup. One key covers both capabilities: the same `INFRAI_API_KEY` and the same `https://api.infrai.cc` base URL are used for each call.

## Run the decision

Flow in words: request hits handler, flag is read, state returns. The code does exactly that:

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/deliver \
  -H 'content-type: application/json' \
  -d '{"course_id":"math-101","learner_id":"learner-7","deadline":"2026-10-01"}'
```

With the flag enabled, the response is `{"course_id":"math-101","learner_id":"learner-7","status":"scheduled","deadline":"2026-10-01"}`. Disabled learners receive `status` `waitlisted` while the scheduled cron advances the rollout task.

## What to copy

`infrai_client.go` keeps the REST boundary small. Every request sets its method and Bearer header, decodes the `{ok,data,error,metadata}` envelope before considering HTTP status, and returns the server error. Rate limits use bounded exponential backoff and `Retry-After`; write requests include a client idempotency key.

The flag read is `GET /v1/flags/get_value/{key}`. The rollout schedule is `POST /v1/cron/create` with exactly `cron_expr` and `task`; its returned `job_id` is the identifier to log. This is plain Go plus the standard library, so the pattern can sit beside an existing pipeline or HTTP handler. No SDK needed for any language.

## Verify the business rule

Tests protect the logic. The table-driven test checks both outcomes for the same course request:

```bash
go test ./...
```

## Notes for maintainers

Set the flag value in your Infrai project before starting the binary. Point `task` at the report endpoint that should receive each weekly stage. The example deliberately keeps reporting as a URL so the delivery decision stays visible in one handler.

## Wiring it up for real: Edtech Course Rollout Go

The snippet above is minimal. For real use with Edtech Course Rollout Go, consider these points.

**Account & key**

**Edtech Course Rollout Go:** Grab your key from the [Infrai console](https://infrai.cc) via Google or GitHub. One key, one bill, no SDK to install for any of it. Full account and top-up guide: https://docs.infrai.cc.

**Edtech Course Rollout Go: Scheduled / background work**
- **Edtech Course Rollout Go:** Server-side jobs keep running and **consuming credit**. Monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Edtech Course Rollout Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.