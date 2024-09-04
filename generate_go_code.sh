source ~/.profile

# vm/systemSmartContracts
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/vm/systemSmartContracts" \
    --gogoslick_out="$PWD/vm/systemSmartContracts" \
    $PWD/vm/systemSmartContracts/*.proto

# trie
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/trie" \
    --gogoslick_out="$PWD/trie" \
    $PWD/trie/*.proto

# state
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/state" \
    --gogoslick_out="$PWD/state" \
    $PWD/state/*.proto

# state/accounts
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/state/accounts" \
    --gogoslick_out="$PWD/state/accounts" \
    $PWD/state/accounts/*.proto

# state/dataTrieValue
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/state/dataTrieValue" \
    --gogoslick_out="$PWD/state/dataTrieValue" \
    $PWD/state/dataTrieValue/*.proto

# consensus
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/consensus" \
    --gogoslick_out="$PWD/consensus" \
    $PWD/consensus/*.proto

# heartbeat
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/heartbeat" \
    --gogoslick_out="$PWD/heartbeat" \
    $PWD/heartbeat/proto/*.proto

mv $PWD/heartbeat/proto/heartbeat.pb.go $PWD/heartbeat/

# dblookupext
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/dblookupext" \
    --gogoslick_out="$PWD/dblookupext" \
    $PWD/dblookupext/*.proto

# dblookupext/dcdtSupply
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/dblookupext/dcdtSupply" \
    --gogoslick_out="$PWD/dblookupext/dcdtSupply" \
    $PWD/dblookupext/dcdtSupply/proto/*.proto

mv $PWD/dblookupext/dcdtSupply/proto/*.pb.go $PWD/dblookupext/dcdtSupply/

# dataRetriever
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/dataRetriever" \
    --gogoslick_out="$PWD/dataRetriever" \
    $PWD/dataRetriever/*.proto

# testscommon/state
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/testscommon/state" \
    --gogoslick_out="$PWD/testscommon/state" \
    $PWD/testscommon/state/*.proto

# process/block/bootstrapStorage
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/process/block/bootstrapStorage" \
    --gogoslick_out="$PWD/process/block/bootstrapStorage" \
    $PWD/process/block/bootstrapStorage/*.proto

# sharding/nodesCoordinator
protoc \
    -I="$HOME/go/src" \
    -I="$HOME/go/src/github.com/DharitriOne/protobuf" \
    -I="$PWD/sharding/nodesCoordinator" \
    --gogoslick_out="$PWD/sharding/nodesCoordinator" \
    $PWD/sharding/nodesCoordinator/*.proto