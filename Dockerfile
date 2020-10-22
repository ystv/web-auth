FROM golang:1.15-alpine AS build
LABEL site="auth"
LABEL stage="builder"

WORKDIR /src/

# Stores our dependencies
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Copy source
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /bin/auth

FROM scratch
LABEL site="auth"
# Copy binary
COPY --from=build /bin/auth .
# Copy static content and templates
COPY --from=build src/public ./public
ENTRYPOINT ["./auth"]