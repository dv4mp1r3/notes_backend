FROM golang:1.22.5-alpine3.20 as stage-build
WORKDIR /usr/go
COPY ./*.go go.mod go.sum ./
RUN apk add --no-cache gcc=15.1.1_git20250726-r0 musl-dev=1.2.5-r17 && \
    go mod download && \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build .


FROM alpine:3.20 as prod
WORKDIR /usr/go
COPY --from=stage-build /usr/go/notes ./notes
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN addgroup -S app \
    && adduser \
    --disabled-password \
    --gecos "" \
    --home /home/app \
    --ingroup app \
    --uid "1000" \
    app \
    && chown -R app:app /usr/go
USER app
CMD ["./notes"]