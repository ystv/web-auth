FROM golang:1.21.0-alpine3.18 AS build
LABEL site="auth"
LABEL stage="builder"

WORKDIR /src/

ARG WAUTH_VERSION_ARG

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
    echo -n "'" >> ./ldflags

# Build the executable
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(cat ./ldflags)" -o /bin/auth

# Run the executable
FROM scratch
LABEL site="auth"
ENTRYPOINT ["/bin/auth"]