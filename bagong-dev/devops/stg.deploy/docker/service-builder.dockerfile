ARG AppName
ARG Compiler

FROM registry.kanosolution.net/sebar/bagong/compiler:${Compiler} as compiler
ARG AppName

ENV CodePath=app/$AppName/*.go

WORKDIR /app
COPY . .
RUN go build -o $AppName.exe ${CodePath}

FROM alpine:latest as runner
ARG AppName
ENV AppName=$AppName

RUN apk --no-cache add ca-certificates
RUN apk --no-cache add tzdata

RUN mkdir -p /app/data/template
RUN mkdir /app/config
COPY data/template /app/data/template
# COPY devops/dev/app.yml /app/app.yml

WORKDIR /app

COPY --from=compiler /app/$AppName.exe /app/$AppName.exe
RUN rm -vfR /app/*.go

CMD ["sh","-c","/app/${AppName}.exe --config=/app/config/app.yml"]
