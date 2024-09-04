FROM golang:1.20.7 as builder
RUN apt-get update && apt-get install -y
WORKDIR /go/drt-chain-go
COPY . .
WORKDIR /go/drt-chain-go/cmd/termui
RUN go build -v
RUN cp /go/pkg/mod/github.com/DharitriOne/$(cat /go/drt-chain-go/go.mod | grep drt-chain-vm-v | sort -n | tail -n -1| awk -F '/' '{print$3}'| sed 's/ /@/g')/wasmer/libwasmer_linux_amd64.so /lib_amd64/
RUN cp /go/pkg/mod/github.com/DharitriOne/$(cat /go/drt-chain-go/go.mod | grep drt-chain-vm-v | sort -n | tail -n -1 | awk -F '/' '{print$3}' | sed 's/ /@/g')/wasmer/libwasmer_linux_arm64_shim.so /lib_arm64/


# ===== SECOND STAGE ======
FROM ubuntu:22.04
COPY --from=builder /go/drt-chain-go/cmd/termui /go/drt-chain-go/cmd/termui

# Copy architecture-specific files
COPY --from=builder "/lib_${TARGETARCH}/*" "/lib/"

WORKDIR /go/drt-chain-go/cmd/termui/
ENTRYPOINT ["./termui"]
