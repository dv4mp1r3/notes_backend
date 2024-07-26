FROM golang:1.22.5-alpine3.20 as stage-build
WORKDIR /usr/go
COPY ./*.go go.mod go.sum ./
RUN go mod tidy && go build .


FROM alpine:3.20 as prod
WORKDIR /usr/go
COPY --from=stage-build /usr/go/notes ./notes
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
CMD ["./notes"]