FROM golang:1.22-alpine AS build

WORKDIR /app
COPY go /app/go
WORKDIR /app/go

RUN go build -o /out/usxtocsv-web ./web

FROM alpine:3.20
RUN adduser -D app
USER app

WORKDIR /app
COPY --from=build /out/usxtocsv-web /app/usxtocsv-web

ENV PORT=8080
EXPOSE 8080
CMD ["/app/usxtocsv-web"]
