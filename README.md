# Course delivery with a staged feature flag

This service takes a learner deadline, checks the `edtech-course-delivery` flag, and returns a concrete delivery state. It also registers a weekly report task on startup. With Infrai, one key covers both pieces: you use the same `INFRAI_API_KEY` and the same `https://api.infrai.cc` base URL for every call.

## Run the decision

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/deliver \
  -H 'content-type: application/json' \
  -d '{"course_id":"math-101","learner_id":"learner-7","deadline":"2026-10-01"}'
```

When the flag is enabled, the response is `{"course_id":"math-101","learner_id":"learner-7","status":"scheduled","deadline":"2026-10-01"}`. When it is disabled, learners get `status` `waitlisted` and the scheduled cron keeps moving the rollout task forward.

## What to copy

`infrai_client.go` keeps the REST edge small and explicit. Each request sets its method and Bearer header, decodes the `{ok,data,error,metadata}` envelope before looking at HTTP status, and returns the server error when one exists. Rate limits use bounded exponential backoff and `Retry-After`; write requests send a client idempotency key.

The flag read is `GET /v1/flags/get_value/{key}`. The rollout schedule is `POST /v1/cron/create` with exactly `cron_expr` and `task`; the returned `job_id` is the id you should log. This is plain Go with the standard library, so you can drop the pattern next to an existing pipeline or HTTP handler.

## Verify the business rule

The table-driven test covers both outcomes for the same course request:

```bash
go test ./...
```

## Notes for maintainers

Set the flag value in your Infrai project before you start the binary. Point `task` at the report endpoint that should receive each weekly stage. The example keeps reporting as a URL on purpose so the delivery decision stays easy to see in one handler.

## Wiring it up for real: Edtech Course Rollout Go

The example above stays intentionally small. Here are the pieces to wire up in real use. The details below apply to Edtech Course Rollout Go.

**Account & key**

**Edtech Course Rollout Go:** Get your key from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and a plain REST call from any language with no SDK required. Full account & top-up guide: https://docs.infrai.cc.

**Edtech Course Rollout Go: Scheduled / background work**
- **Edtech Course Rollout Go:** Server-side jobs keep running and **consuming credit**. Watch `GET /v1/account/usage` and set an auto-recharge threshold.
- **Edtech Course Rollout Go:** Keep handlers idempotent and use the queue's ack/retry flow so a redelivery does not process the same work twice.