module github.com/orlangure/gnomockd

go 1.14

require (
	github.com/aws/aws-sdk-go v1.29.34
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/orlangure/gnomock v0.3.0
	github.com/orlangure/gnomock-localstack v0.0.0-00010101000000-000000000000
	github.com/orlangure/gnomock-mongo v0.0.0-00010101000000-000000000000
	github.com/orlangure/gnomock-mssql v0.0.0-00010101000000-000000000000
	github.com/orlangure/gnomock-mysql v0.0.0-00010101000000-000000000000
	github.com/orlangure/gnomock-postgres v0.0.0-00010101000000-000000000000
	github.com/orlangure/gnomock-redis v0.0.0-00010101000000-000000000000
	github.com/orlangure/gnomock-splunk v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.4.0
	go.mongodb.org/mongo-driver v1.3.2
)

replace github.com/orlangure/gnomock => ../gnomock

replace github.com/orlangure/gnomock-localstack => ../gnomock-localstack

replace github.com/orlangure/gnomock-mongo => ../gnomock-mongo

replace github.com/orlangure/gnomock-mssql => ../gnomock-mssql

replace github.com/orlangure/gnomock-mysql => ../gnomock-mysql

replace github.com/orlangure/gnomock-postgres => ../gnomock-postgres

replace github.com/orlangure/gnomock-redis => ../gnomock-redis

replace github.com/orlangure/gnomock-splunk => ../gnomock-splunk
