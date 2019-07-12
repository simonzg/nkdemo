# Load Test for Nakama server

## Prerequisite

* DB has been migrated with `nakama migrate up`
* DB has been inserted with essential data within `sql\init.sql`

## Usage

```
# with binary
./loadtest-linux [params]

  -batchsize int
      numbers of sync requests that needs to be sent out every second (default 10)
  -duration int
      test duration in seconds (default 30)
  -inc int
      increment the batchsize every second by this number
  -ip string
      server ip address (default "localhost")
  -n int
      number of clients running at the same time (default 10)
  -port int
      server port number (default 7350)
```

## Build

```
# cross platform
GOOS=linux GOARCH=amd64 go build . -o build/loadtest-linux
```

