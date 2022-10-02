FROM golang:1.19-alpine as builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o main
RUN ls

FROM alpine:3.16
WORKDIR /app

COPY --from=builder /app/main .
COPY prod.env ./

RUN ls

RUN mkdir -p /app/inputMedia
RUN mkdir -p /app/inputMedia/photos
RUN mkdir -p /app/inputMedia/files
RUN mkdir -p /app/inputMedia/videos
RUN mkdir -p /app/inputMedia/music
RUN ls

CMD [ "/app/main" ]

