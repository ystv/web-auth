# web-auth

A very basic auth package, allowing people to signup, generating API tokens, and view an internal page.

## Usage

You'll need to set a couple of environment variables first, usually by doing a good `export key=value`

- `SIGNING_KEY` used to sign JWTs

- `DATABASE_URL` Database connection URL.

- `SMTP_HOST` SMTP host, used to send forgotten password emails

- `SMTP_USERNAME` SMTP username

- `SMTP_PASSWORD` SMTP password

After all that is set you should be able to visit it at `:8080`
