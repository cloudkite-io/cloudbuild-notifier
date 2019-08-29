FROM golang:1.12-alpine as build
WORKDIR /src

# ca-certs and git required for go modules
RUN apk add ca-certificates git

# Build deps
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Compile
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app

FROM alpine
WORKDIR /app
COPY --from=build /bin/app ./
CMD ./app
