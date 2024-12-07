# 参考 https://note.com/jtamas_engineer/n/n28cc15e61a1f

FROM golang:1.23.0 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o main /app/main.go

FROM gcr.io/distroless/base-debian10

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD [ "/app/main" ]


