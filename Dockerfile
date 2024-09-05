ARG GO_VERSION=1
ARG PORT
ARG DATABASE_URL
ARG PREVIEW_ENV
ARG IMAGE_DIR
FROM golang:${GO_VERSION}-bookworm as builder

ENV PORT=${PORT}
ENV DATABASE_URL=${DATABASE_URL}
ENV PREVIEW_ENV=${PREVIEW_ENV}
ENV IMAGE_DIR=${IMAGE_DIR}

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .


FROM debian:bookworm

COPY --from=builder /run-app /usr/local/bin/
CMD ["run-app"]
