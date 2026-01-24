FROM golang:1.22-alpine AS build

WORKDIR /app
COPY go /app/go
WORKDIR /app/go

RUN go build -o /out/usxtocsv-web ./web

FROM node:20-alpine AS webui
WORKDIR /app/web-ui
COPY web-ui/package.json web-ui/package-lock.json* web-ui/
RUN npm install --silent
COPY web-ui /app/web-ui
RUN npm run build

FROM alpine:3.20
RUN adduser -D app
USER app

WORKDIR /app
COPY --from=build /out/usxtocsv-web /app/usxtocsv-web
COPY --from=webui /app/web-ui/dist /app/web-ui

ENV PORT=8080
ENV WEB_UI_DIR=/app/web-ui
EXPOSE 8080
CMD ["/app/usxtocsv-web"]
