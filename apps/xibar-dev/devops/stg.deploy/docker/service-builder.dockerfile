ARG AppName
ARG Compiler

FROM registry.kanosolution.net/sebar/apps/xibar/compiler:${Compiler} as compiler
ARG AppName

ENV CodePath=app/$AppName/*.go
ENV GOFLAGS=-mod=vendor

WORKDIR /app
COPY . .
RUN go build -o $AppName.exe ${CodePath}

FROM alpine:latest as runner
ARG AppName
ENV AppName=$AppName

RUN apk --no-cache add ca-certificates

RUN mkdir -p /app/data/
RUN mkdir -p /app/data/iam
RUN mkdir -p /app/data/asset
RUN mkdir -p /app/data/output
RUN mkdir /config
COPY data/template /app/data/template
# COPY devops/dev/app.yml /app/app.yml

WORKDIR /app

COPY --from=compiler /app/$AppName.exe /app/$AppName.exe
RUN rm -vfR /app/*.go

CMD ["sh","-c","/app/${AppName}.exe --config=/config/app.yml"]
