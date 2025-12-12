ARG AppName
ARG Compiler

FROM registry.kanosolution.net/sebar/tenantcore/compiler:${Compiler} as compiler
ARG AppName

ENV CodePath=app/$AppName/*.go
ENV GOFLAGS=-mod=vendor

WORKDIR /app
COPY . .
RUN go build -o $AppName.exe ${CodePath}

FROM surnet/alpine-wkhtmltopdf:3.19.0-0.12.6-full as wkhtmltopdf

FROM alpine:latest as runner
ARG AppName
ENV AppName=$AppName

RUN apk add --no-cache \
    ca-certificates \
    libstdc++ \
    libx11 \
    libxrender \
    libxext \
    libssl3 \
    fontconfig \
    freetype \
    ttf-dejavu \
    ttf-droid \
    ttf-freefont \
    ttf-liberation \
    # more fonts
  && apk add --no-cache --virtual .build-deps \
    msttcorefonts-installer \
  # Install microsoft fonts
  && update-ms-fonts \
  && fc-cache -f \
  # Clean up when done
  && rm -rf /tmp/* \
  && apk del .build-deps

COPY --from=wkhtmltopdf /bin/wkhtmltopdf /usr/bin/wkhtmltopdf
COPY --from=wkhtmltopdf /bin/wkhtmltoimage /usr/bin/wkhtmltoimage
COPY --from=wkhtmltopdf /bin/libwkhtmltox* /usr/bin/

RUN mkdir -p /app/data/template
RUN mkdir /app/config
COPY data/template /app/data/template
# COPY devops/dev/app.yml /app/app.yml

WORKDIR /app

COPY --from=compiler /app/$AppName.exe /app/$AppName.exe
RUN rm -vfR /app/*.go

CMD ["sh","-c","/app/${AppName}.exe --config=/app/config/app.yml"]
