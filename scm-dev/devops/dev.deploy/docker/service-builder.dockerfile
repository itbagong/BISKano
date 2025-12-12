ARG AppName
ARG GitUser
ARG GitSecret
ARG Compiler=latest

FROM registry.kanosolution.net/sebar/scm/compiler:${Compiler} as compiler
ARG AppName

ENV CodePath=app/$AppName/*.go
ENV GOFLAGS=-mod=vendor

WORKDIR /app
COPY . .
RUN apk add --update --no-cache ca-certificates git
RUN go env -w GOPRIVATE=git.kanosolution.net
RUN git config --global url."https://${GitUser}:${GitSecret}@git.kanosolution.net".insteadOf "https://git.kanosolution.net"
RUN go mod tidy
RUN go build -o $AppName.exe ${CodePath}

FROM alpine:latest as runner
ARG AppName
ENV AppName=$AppName

RUN apk --no-cache add ca-certificates

RUN mkdir -p /app/data/template
RUN mkdir /app/config
COPY data/template /app/data/template
# COPY devops/dev/app.yml /app/app.yml

WORKDIR /app

COPY --from=compiler /app/$AppName.exe /app/$AppName.exe
RUN rm -vfR /app/*.go

CMD ["sh","-c","/app/${AppName}.exe --config=/app/config/app.yml"]
