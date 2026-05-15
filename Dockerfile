FROM node:alpine AS css
WORKDIR /app
RUN apk add --no-cache curl
COPY Makefile ./
COPY cmd/web cmd/web
RUN make vendor-assets && make css

FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY --from=css /app/cmd/web/assets cmd/web/assets
COPY . .
RUN go install github.com/a-h/templ/cmd/templ@latest && templ generate
RUN go build -o main cmd/api/main.go

FROM alpine:3.20 AS prod
WORKDIR /app
COPY --from=build /app/main ./main
EXPOSE ${PORT}
CMD ["./main"]
