FROM golang:1.22-alpine

ARG GitUser=deploy
ARG GitSecret=secret
ARG AppName=mfg

ENV CodePath=app/$AppName/*.go
ENV GOFLAGS=-mod=vendor

WORKDIR /app
COPY . .
RUN apk add --update --no-cache ca-certificates git
RUN go env -w GOPRIVATE=git.kanosolution.net
RUN git config --global url."https://${GitUser}:${GitSecret}@git.kanosolution.net".insteadOf "https://git.kanosolution.net"
RUN go mod tidy
RUN go build -o mfg.exe ${CodePath}

RUN mkdir -p /app/data/template
RUN mkdir /app/config
COPY data/template /app/data/template

CMD ["sh","-c","/app/mfg.exe --config=/app/config/app.yml"]
