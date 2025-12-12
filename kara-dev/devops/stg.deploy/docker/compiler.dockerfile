FROM golang:1.21.6-alpine3.19

ARG GitUser=deploy
ARG GitSecret=secret

WORKDIR /app
COPY . .

RUN apk add --update --no-cache ca-certificates git tzdata
RUN go env -w GOPRIVATE=git.kanosolution.net
RUN git config --global url."https://${GitUser}:${GitSecret}@git.kanosolution.net".insteadOf "https://git.kanosolution.net"
RUN go mod tidy