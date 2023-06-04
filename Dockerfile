FROM golang:1.20-alpine AS go-build

WORKDIR /app

COPY go.* .

COPY ./cmd ./cmd
# COPY ./scoretracker ./scoretracker
COPY ./pkg ./pkg
COPY ./gameserver ./gameserver
COPY ./move ./move
COPY ./web ./web

RUN go install ./cmd/serve
RUN go build -o /app/sillygame ./cmd/serve

### ------------ ###
FROM node:18-alpine AS node-build

WORKDIR /app

COPY --from=go-build /app/web/package* .
RUN npm i

COPY --from=go-build /app/web/* .
COPY --from=go-build /app/web/src ./src
COPY --from=go-build /app/web/static ./static

ENV VITE_WEBSOCKET_URL=sillygame.chimid.rocks/subscribe
ENV VITE_WEB_URL=sillygame.chimid.rocks
ENV VITE_LOG_LEVEL=error

RUN npm run build

### ------------ ###
FROM alpine:3.18

WORKDIR /app
COPY --from=go-build /app/sillygame /app/sillygame
COPY --from=node-build /app/build /app/static

CMD [ "./sillygame", "0.0.0.0:80" ]

