FROM golang:1.20.7 as builder

RUN apt-get update && apt-get install -y
WORKDIR /go/drt-chain-go
COPY . .
RUN go mod tidy
# Keygenerator node
WORKDIR /go/drt-chain-go/cmd/keygenerator
RUN go build

# ===== SECOND STAGE ======
FROM ubuntu:22.04
COPY --from=builder /go/drt-chain-go/cmd/keygenerator /go/drt-chain-go/cmd/keygenerator

ENTRYPOINT ["/go/drt-chain-go/cmd/keygenerator/keygenerator"]
