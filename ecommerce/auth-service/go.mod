module github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service

go 1.23.0

toolchain go1.23.11

require (
	github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto v0.0.0-20250803130941-72ac65b63311
	google.golang.org/grpc v1.74.2
)

require (
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto => ../proto
