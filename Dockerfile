FROM golang:alpine as builder

# No need for go.sum since the code is stdlib based
COPY go.mod /app/exiftool2json/
RUN apk add exiftool build-base
# Un-comment in case you add thirdy-party libs
# RUN go mod download
COPY . /app/exiftool2json
WORKDIR /app/exiftool2json
RUN go test ./...
RUN CGO_ENABLED=1 go build -o exiftool2json

FROM alpine
RUN apk add exiftool
COPY --from=builder /app/exiftool2json/exiftool2json /usr/local/bin/exiftool2json
EXPOSE 8080
CMD [ "exiftool2json", "-port", "8080", "-addr", "0.0.0.0" ]
