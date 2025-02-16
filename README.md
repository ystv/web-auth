# web-auth

A basic auth package, allowing people to signup, reset passwords, generating API tokens, and view an internal page.

## Usage

You can get the latest tagged version with `go get github.com/ystv/web-auth`, otherwise follow the [build instructions](#building).

Copy the .env file as .env.local and set the variables as required

After all that is set you should be able to visit it at `:8080`.

### Setting up development database

You will need a PostgreSQL server running locally - either install it on your computer or use Docker.

Once you've done that, create a database to work with:

```shell
$ createdb ystv2020
# or in Docker
$ docker run -d --name ystv2020 -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -e POSTGRES_DB=ystv2020 -p 5432:5432 postgres
```

Then run the migrations to initialise the tables:

```shell
$ go run ./cmd/migrate-db
```

### Dev side note

If you are trying to connect to a database from your dev machine then I can recommend you use the following command to make your life easier.

`ssh -L [local port]:127.0.0.1:[db port on remote server] [remote server user and ip]`

This will prevent the full deploy being your dev environment and is much quicker.

## Building

Both methods require cloning the repo

`git clone https://github.com/ystv/web-auth`

### Docker

Execute `docker build -t webauth .` in the root directory, and you'll have a container `webauth:latest`.

### Static binary

`go build -o web-auth`

Then use that produced binary along with the usage instructions.
