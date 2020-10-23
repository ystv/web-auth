# web-auth

A basic auth package, allowing people to signup, reset passwords, generating API tokens, and view an internal page.

## Usage

You can get the latest tagged version with `go get github.com/ystv/web-auth`, otherwise follow the [build instructions](#building).

You'll need to set a couple of environment variables first, usually by doing a good `export key=value`

- `WAUTH_SIGNING_KEY` used to sign JWTs

- `WAUTH_DATABASE_URL` Database connection URL.

Optional variables

will print reset codes instead of mailing

- `WAUTH_SMTP_HOST` SMTP host, used to send forgotten password emails

- `WAUTH_SMTP_USERNAME` SMTP username

- `WAUTH_SMTP_PASSWORD` SMTP password

- `WAUTH_DOMAIN_NAME` Domain name of where it's hosted so we can restrict callbacks to a certain domain

Used to keep cookies secure, if left blank it will generate random keys instead

- `WAUTH_AUTHENTICATION_KEY` 64 bytes of hex, used for cookies

- `WAUTH_ENCRYPTION_KEY` 32 bytes of hex, used for cookies

- `WAUTH_LOGOUT_ENDPOINT` Where web-auth redirects after successful logout

After all that is set you should be able to visit it at `:8080`.

## Building

Both methods require cloning the repo

`git clone https://github.com/ystv/web-api`

### Docker

Execute `docker build -t webauth .` in the root directory and you'll have a container `webauth:latest`.

### Static binary

`go build -o web-auth`

Then use that produced binary along with the usage instructions.
