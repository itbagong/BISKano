FROM golang:1.20-alpine3.18

ARG GitUser=deploy
ARG GitSecret=secret

WORKDIR /app
COPY . .

RUN apk add --update --no-cache ca-certificates git
RUN go env -w GOPRIVATE=git.kanosolution.net
RUN git config --global url."https://${GitUser}:${GitSecret}@git.kanosolution.net".insteadOf "https://git.kanosolution.net"
RUN go mod tidy