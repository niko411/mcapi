MCApi
=====

An API for fetching the status of and querying Minecraft servers.

It is running at [mcapi.us](https://mcapi.us).

## Configuration

```bash
go build
./mcapi -gencfg
```

Options in config.json:
* HTTPAppHost &mdash; host and port to listen on
* RedisHost &mdash; host of redis server
* StaticFiles &mdash; path to static files
* TemplateFile &mdash; path to index file
* SentryDSN &mdash; optional sentry dsn to report errors to
* AdminKey &mdash; secret token used to get list of servers or clear the list

Rate limiting with Cloudflare requires setting the following environment variables:
* CLOUDFLARE_EMAIL &mdash; your Cloudflare account email address
* CLOUDFLARE_AUTH &mdash; your Cloudflare authentication token

Disable ratelimiting generally or with Cloudflare by using built-time variables
to update `rateLimitEnabled` and `cloudflareEnabled` or modifying the code
in `ratelimit.go#16-17`.
