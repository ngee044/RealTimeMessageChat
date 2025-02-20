module RestAPI

go 1.24.0

require (
	Common v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/redis/go-redis/v9 v9.7.0 // indirect
	github.com/streadway/amqp v1.1.0 // indirect
)

replace Common => ../Common
