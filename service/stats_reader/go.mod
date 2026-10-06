module github.com/MamangRust/microservice-point-of-sale-stats-reader

go 1.26.0

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.48.0
	github.com/MamangRust/microservice-point-of-sale-pkg v1.0.7
	github.com/MamangRust/microservice-point-of-sale-shared v1.0.8
	github.com/redis/go-redis/v9 v9.21.0
	github.com/spf13/viper v1.20.1
	go.uber.org/zap v1.28.0
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12 // indirect
)

require (
	github.com/ClickHouse/ch-go v0.74.0 // indirect
	github.com/andybalholm/brotli v1.2.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.8.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/paulmach/orb v0.13.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.29 // indirect
	github.com/sagikazarmark/locafero v0.9.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.9.2 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.46.0 // indirect
	go.opentelemetry.io/otel/log v0.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.19.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260831171406-18b4a7587f8a // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/MamangRust/microservice-point-of-sale-pkg => ../../pkg

replace github.com/MamangRust/microservice-point-of-sale-shared => ../../shared

require github.com/MamangRust/microservice-point-of-sale-pb v1.0.8

replace github.com/MamangRust/microservice-point-of-sale-pb => ../../pb
