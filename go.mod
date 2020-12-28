module github.com/hbtc-chain/bhchain

go 1.13

require (
	github.com/aristanetworks/goarista v0.0.0-20200310212843-2da4c1f5881b
	github.com/bartekn/go-bip39 v0.0.0-20171116152956-a05967ea095d
	github.com/bgentry/speakeasy v0.1.0
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/cosmos/ledger-cosmos-go v0.11.1
	github.com/emirpasic/gods v1.12.0
	github.com/ethereum/go-ethereum v1.9.15
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.4.0
	github.com/gorilla/mux v1.7.3
	github.com/hbtc-chain/chainnode v0.9.4
	github.com/mattn/go-isatty v0.0.12
	github.com/otiai10/copy v1.0.2
	github.com/pelletier/go-toml v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.6
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.1
	github.com/stumble/gorocksdb v0.0.3 // indirect
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/crypto v0.0.0-20191022145703-50d29ede1e15
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/iavl v0.13.2
	github.com/tendermint/tendermint v0.32.8
	github.com/tendermint/tm-db v0.2.0
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/tendermint/tendermint => github.com/hbtc-chain/tendermint v0.9.0

replace github.com/tendermint/iavl => github.com/hbtc-chain/iavl v0.9.0
