module github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service

go 1.23.0

toolchain go1.23.11

require (
	github.com/google/uuid v1.6.0
	github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger v0.0.0-00010101000000-000000000000
	github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto v0.0.0-20250803130941-72ac65b63311
	golang.org/x/crypto v0.38.0
	google.golang.org/grpc v1.74.2
	gorm.io/driver/mysql v1.6.0
	gorm.io/gorm v1.30.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto => ../proto

replace github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger => ../logger
