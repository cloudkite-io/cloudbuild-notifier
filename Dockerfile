FROM golang:1.16-alpine as build
WORKDIR /src
# ca-certs and git required for go modules
RUN apk add ca-certificates git
# Build deps
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Compile
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd

########################################################################################################################

FROM alpine
COPY --from=build /app /
RUN apk add ca-certificates
CMD ./app
