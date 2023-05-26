FROM golang:1.20.4-alpine3.18 AS build
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

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="$(cat ./ldflags)" -o /bin/auth

FROM scratch
LABEL site="auth"
# Copy binary
COPY --from=build /bin/auth .
ENTRYPOINT ["./auth"]