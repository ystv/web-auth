# web-auth

A basic auth package, allowing people to signup, reset passwords, generating API tokens, and view an internal page.

## Usage

You'll need to set a couple of environment variables first, usually by doing a good `export key=value`

- `SIGNING_KEY` used to sign JWTs

- `DATABASE_URL` Database connection URL.

- `SMTP_HOST` SMTP host, used to send forgotten password emails

- `SMTP_USERNAME` SMTP username

- `SMTP_PASSWORD` SMTP password

Optional variables. Used to keep cookies secure, if left blank it will generate random keys instead

- `AUTHENTICATION_KEY` 64 bytes of hex, used for cookies

- `ENCRYPTION_KEY` 32 bytes of hex, used for cookies

After all that is set you should be able to visit it at `:8080`
