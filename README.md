# web-auth

A basic auth package, allowing people to signup, reset passwords, generating API tokens, and view an internal page.

## Usage

You can get the latest tagged version with `go get github.com/ystv/web-auth`, otherwise follow the [build instructions](#building).

Copy the .env file as .env.local and set the variables as required

After all that is set you should be able to visit it at `:8080`.

## Building

Both methods require cloning the repo

`git clone https://github.com/ystv/web-api`

### Docker

Execute `docker build -t webauth .` in the root directory and you'll have a container `webauth:latest`.

### Static binary

`go build -o web-auth`

Then use that produced binary along with the usage instructions.
