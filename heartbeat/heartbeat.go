//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/DharitriOne/protobuf/protobuf --gogoslick_out=. heartbeat.proto

package heartbeat
