FROM golang:1.22-alpine AS build
ARG GOOS=linux
ARG CGO_ENABLED=0
WORKDIR /build
RUN apk add --no-cache --update tzdata ca-certificates
COPY . .
RUN go build -o homepage ./app

FROM scratch
WORKDIR /srv
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/homepage /srv/homepage
EXPOSE 3000
ENTRYPOINT ["/srv/homepage"]
