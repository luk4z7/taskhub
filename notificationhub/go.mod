module github.com/luk4z7/notificationhub

go 1.23.0

require (
	github.com/ThreeDotsLabs/go-event-driven v0.0.12
	github.com/ThreeDotsLabs/watermill v1.3.5
	github.com/ThreeDotsLabs/watermill-redisstream v1.2.2
	github.com/luk4z7/messages v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.4.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/Rican7/retry v0.3.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	golang.org/x/sys v0.28.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/luk4z7/messages => ../messages/
