FROM golang:1.24.1-alpine3.21 AS build

LABEL site="auth"
LABEL stage="builder"

WORKDIR /src/

ARG WAUTH_VERSION_ARG
ARG WAUTH_COMMIT_ARG

RUN apk --no-cache add ca-certificates

# Stores our dependencies
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Copy source
COPY . .

# Set build variables
RUN echo -n "-X 'main.Version=$WAUTH_VERSION_ARG" > ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "' -X 'main.Commit=$WAUTH_COMMIT_ARG" >> ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "'" >> ./ldflags

# Build the executable
RUN GOOS=linux GOARCH=amd64 go build -ldflags="$(cat ./ldflags)" -o /bin/auth

# Run the executable
FROM scratch
LABEL site="auth"
# Copy binary
COPY --from=build /bin/auth /bin/auth
ENTRYPOINT ["/bin/auth"]