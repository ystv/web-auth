# web-auth

A basic auth package, allowing people to signup, reset passwords, generating API tokens, and view an internal page.

## Usage

You can get the latest tagged version with `go get github.com/ystv/web-auth`, otherwise follow the [build instructions](#building).

You'll need to set a couple of environment variables first, usually by doing a good `export key=value`

- `SIGNING_KEY` used to sign JWTs

- `DATABASE_URL` Database connection URL.

- `SMTP_HOST` SMTP host, used to send forgotten password emails

- `SMTP_USERNAME` SMTP username

- `SMTP_PASSWORD` SMTP password

- `DOMAIN_NAME` Domain name of where it's hosted so we can restrict callbacks to a certain domain

Optional variables. Used to keep cookies secure, if left blank it will generate random keys instead

- `AUTHENTICATION_KEY` 64 bytes of hex, used for cookies

- `ENCRYPTION_KEY` 32 bytes of hex, used for cookies

- `LOGOUT_ENDPOINT` Where web-auth redirects after successful logout

After all that is set you should be able to visit it at `:8080`.

If it succeeds on startup it won't print anything in the console, otherwise it will print the errors.

## Building

Both methods require cloning the repo

`git clone https://github.com/ystv/web-api`

### Docker

Execute `docker build -t webauth .` in the root directory and you'll have a container `webauth:latest`.

### Static binary

`go build -o web-auth`

Then use that produced binary along with the usage instructions.
