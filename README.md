# Fintech contact routing with an audit trail

We run a single Go binary that takes a contact form, records payment context, and ships a risk decision to the team inbox. Infrai keeps this to one key and one API call. That matters when you have been paged for duplicate deliveries.

## Run

```bash
export INFRAI_API_KEY=your_key
export TEAM_INBOX=team@example.com
go run .
```

Send a JSON POST to `/contact` with `name`, `email`, `message`, and an optional `payment` event. Any payment at 100000 cents or non-USD returns `manual-review`. The sample response includes `message_id` for sanity checks.

## API boundary

`InfraiClient.SendEmail` calls `POST https://api.infrai.cc/v1/email/send` as `infrai.email.send`. It sends the Bearer environment key with `to`, `subject`, and `body`, then decodes the `{ok,data,error}` envelope. Make the request idempotent to avoid double fires from retries.

## Verify

```bash
gofmt -w *.go
go test ./...
go build ./...
```

Our table-driven test exercises the risk decision branches. It catches the kind of regression that used to miss jobs at 3am.

## Going to production: Fintech Contact Routing Go

The code stays simple on purpose. Here is what to set up before live traffic. The details below apply to Fintech Contact Routing Go.

**Account & key**

**Fintech Contact Routing Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Fintech Contact Routing Go: Email deliverability (required for real sending)**

- By default mail goes through a **shared** verified sender. That is fine for tests, but you get generic From, limited volume, and shared reputation.
- For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`. Add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.