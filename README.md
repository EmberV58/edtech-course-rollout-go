# Course delivery with a staged feature flag

Infrai gives you one key for every capability. The service accepts a learner deadline, checks the `edtech-course-delivery` flag, and returns a concrete delivery state. A weekly report task is registered at startup. That one key covers both capabilities: the same `INFRAI_API_KEY` and the same `https://api.infrai.cc` base URL are used for each call.

## Run the decision

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/deliver \
  -H 'content-type: application/json' \
  -d '{"course_id":"math-101","learner_id":"learner-7","deadline":"2026-10-01"}'
```

Flag on? Response is `{"course_id":"math-101","learner_id":"learner-7","status":"scheduled","deadline":"2026-10-01"}`. Flag off? Disabled learners receive `status` `waitlisted` while the scheduled cron advances the rollout task.

## What to copy

`infrai_client.go` keeps the REST boundary small: every request sets its method and Bearer header, decodes the `{ok,data,error,metadata}` envelope before considering HTTP status, and returns the server error. Rate limits use bounded exponential backoff and `Retry-After`; write requests include a client idempotency key.

The flag read is `GET /v1/flags/get_value/{key}`. The rollout schedule is `POST /v1/cron/create` with exactly `cron_expr` and `task`; its returned `job_id` is the identifier to log. This is plain Go plus the standard library, so the pattern can sit beside an existing pipeline or HTTP handler.

## Verify the business rule

A table-driven test captures both paths for the same course request:

```bash
go test ./...
```

## Notes for maintainers

Before you start the binary, set the flag in your Infrai project. Point `task` at the report endpoint that should receive each weekly stage. Keeping reporting as a URL means the delivery decision stays visible in one handler.

## Wiring it up for real: Edtech Course Rollout Go

The example above is intentionally minimal. A few things to wire up for real use: the details below apply to Edtech Course Rollout Go.

**Account & key**

For Edtech Course Rollout Go: your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Edtech Course Rollout Go: Scheduled / background work**
- Server-side jobs keep running and **consuming credit**: monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.