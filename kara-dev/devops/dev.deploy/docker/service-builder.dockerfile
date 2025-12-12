ARG AppName
ARG Compiler

FROM registry.kanosolution.net/sebar/${AppName}/compiler:${Compiler} as compiler
ARG AppName

ENV CodePath=app/$AppName/*.go
ENV GOFLAGS=-mod=vendor

WORKDIR /app
COPY . .
RUN go build -o $AppName.exe ${CodePath}

FROM alpine:latest as runner
ARG AppName
ENV AppName=$AppName

RUN apk --no-cache add ca-certificates tzdata

RUN mkdir -p /app/data/template
RUN mkdir /app/config
COPY data/template /app/data/template

WORKDIR /app

COPY --from=compiler /app/$AppName.exe /app/$AppName.exe
RUN rm -vfR /app/*.go

CMD ["sh","-c","/app/${AppName}.exe --config=/app/config/app.yml"]