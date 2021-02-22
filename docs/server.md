# Using Gnomock over HTTP

If you use Go, please refer to [Using Gnomock in Go
applications](../README.md#using-gnomock-in-go-applications) section.
Otherwise, you'll need to setup a helper container, and communicate with it
over HTTP.

## Running the server

### Using Github Actions

For convenience, there is a [Github
Action](https://github.com/marketplace/actions/gnomock) that starts a Gnomock
server on port 23042 in a single step:

```
steps:
  - name: Gnomock
    uses: gnomock/github-action@master

  - name: Test
    run: |
      echo "running tests..."
      # run tests that use Gnomock server on port 23042
```

### Run directly

To start a Gnomock server without using Github Action, run the following on any
Unix-based system:

```bash
docker run --rm \
    -p 23042:23042 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $PWD:$PWD \
    --privileged \ # this flag may be required on some systems
    orlangure/gnomock
```

`-p 23042:23042` exposes a port on the host to communicate with `gnomock`. You
can use any port you like, just make sure to configure the client properly.

`-v /var/run/docker.sock:/var/run/docker.sock` allows `gnomock` to communicate
with the docker engine running on host. Without it `gnomock` can't access
docker.

If you use any file-related `gnomock` options, like `WithQueriesFile`, you have
to make the path you use available inside the container:

```
# this makes the current folder appear inside the container under the same
# path and name:
-v `pwd`:`pwd`
```

Any program in any language can communicate with `gnomock` server using OpenAPI
3.0 [specification](https://app.swaggerhub.com/apis/orlangure/gnomock/).

Below is an example of setting up a **MySQL** container using a `POST` request:

```
$ cat mysql-preset.json
{
  "preset": {
    "db": "mydb",
    "user": "gnomock",
    "password": "p@s$w0rD",
    "queries": [
      "create table foo(bar int)",
      "insert into foo(bar) values(1)"
    ],
    "queries_file": "/home/gnomock/project/testdata/mysql/queries.sql"
  },
  "options": {}
}

$ curl --data @mysql-preset.json http://127.0.0.1:23042/start/mysql
{
  "id": "f5d08dc84421",
  "host": "string",
  "ports": {
    "default": {
      "protocol": "tcp",
      "port": 35973
    }
  }
}
```

There are auto-generated wrappers for the available API:

| Client | Sample code |
|--------|-------------|
| [Python SDK](https://github.com/orlangure/gnomock-python-sdk) | [Code](https://github.com/orlangure/gnomock/blob/master/sdktest/python/test/test_sdk.py) |
| JavaScript SDK | |
| Ruby SDK | |
| PHP SDK | |
| Java SDK | |
| [Other](https://openapi-generator.tech/docs/generators) languages | |

**For more details and a full specification, see
[documentation](https://app.swaggerhub.com/apis/orlangure/gnomock/).**

