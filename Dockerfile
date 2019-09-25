FROM golang as builder
RUN mkdir /go/src/openbikeshare
WORKDIR /go/src/openbikeshare
COPY . .

RUN go get
RUN CGO_ENABLED=0 go build -o /go/bin/openbikeshare

FROM alpine
COPY --from=builder /go/bin/openbikeshare /app/openbikeshare_api
WORKDIR /app
CMD ["/app/openbikeshare_api"]

