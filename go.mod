module github.com/orlangure/gnomock

go 1.14

require (
	github.com/aws/aws-sdk-go v1.30.19
	github.com/denisenkom/go-mssqldb v0.0.0-20200428022330-06a60b6afbbc // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/lib/pq v1.5.1 // indirect
	github.com/orlangure/gnomock-localstack v0.3.1
	github.com/orlangure/gnomock-mongo v0.1.1
	github.com/orlangure/gnomock-mssql v0.1.2
	github.com/orlangure/gnomock-mysql v0.3.2
	github.com/orlangure/gnomock-postgres v0.3.2
	github.com/orlangure/gnomock-redis v0.3.1
	github.com/orlangure/gnomock-splunk v0.4.2
	github.com/stretchr/testify v1.5.1
	go.mongodb.org/mongo-driver v1.3.2
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79 // indirect
	golang.org/x/net v0.0.0-20200501053045-e0ff5e5a1de5 // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/sys v0.0.0-20200501145240-bc7a7d42d5c3 // indirect
)

replace github.com/orlangure/gnomock-localstack => ../gnomock-localstack

replace github.com/orlangure/gnomock-mongo => ../gnomock-mongo

replace github.com/orlangure/gnomock-mssql => ../gnomock-mssql

replace github.com/orlangure/gnomock-mysql => ../gnomock-mysql

replace github.com/orlangure/gnomock-postgres => ../gnomock-postgres

replace github.com/orlangure/gnomock-redis => ../gnomock-redis

replace github.com/orlangure/gnomock-splunk => ../gnomock-splunk
