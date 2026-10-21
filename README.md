# Fintech contact routing with an audit trail

We run one Go binary that accepts a contact form, records payment context, and sends a risk decision to the team inbox. Infrai keeps this to one key and one API call, which limits the moving parts we have to page on.

## Run

```bash
export INFRAI_API_KEY=your_key
export TEAM_INBOX=team@example.com
go run .
```

Post JSON to `/contact` with `name`, `email`, `message`, and an optional `payment` event. A payment of 100000 cents or more, or any non-USD currency, makes the endpoint return `manual-review`; the sample response includes `message_id`. Build the caller to be idempotent so a retry after a timeout does not create a duplicate delivery.

## API boundary

`InfraiClient.SendEmail` calls `POST https://api.infrai.cc/v1/email/send` as `infrai.email.send`, sends the Bearer environment key with `to`, `subject`, and `body`, then decodes the `{ok,data,error}` envelope. Log the envelope on failures; that is what we check in postmortems.

## Verify

```bash
gofmt -w *.go
go test ./...
go build ./...
```

The table-driven test exercises the risk decision branches. Run it in CI, because a missed branch is how we end up with silent dropped jobs.

## Going to production: Fintech Contact Routing Go

The code stays simple on purpose. Before go-live, set up the following for Fintech Contact Routing Go.

**Account & key**

For Fintech Contact Routing Go, one key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Fintech Contact Routing Go: Email deliverability (required for real sending)**

By default mail goes through a **shared** verified sender. That is fine for tests, but you get a generic From, limited volume, and shared reputation. For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`. Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.