# Log-Generator

API managed testing tool for logging-operator

## Deploy log-generator with Helm

```sh
helm install --wait --generate-name oci://ghcr.io/kube-logging/helm-charts/log-generator
```

## Usage

You can start the daemon serving the API on e.g. port 11000 by running the following command:

```sh
go run main.go
```

Now you can connect to <http://localhost:11000> from your browser or using your favorite HTTP client.

## Available API Calls

### Log generator

#### [GET] /loggen

Call:

```sh
curl --location --request GET 'localhost:11000/loggen'
```

Response:

```json
{
  "event_per_sec": 100,
  "byte_per_sec": 200,
  "randomise": true,
  "active_requests": [],
  "golang_log": {
    "error_weight": 0,
    "warning_weight": 0,
    "info_weight": 1,
    "debug_weight": 0
  }
}
```

#### [GET] /loggen/formats

Call:

```sh
curl --location --request GET 'localhost:11000/loggen/formats'
```

Response:

```json
{
  "web": [
    "apache",
    "nginx"
  ]
}
```

#### [POST] /loggen

Call:

```sh
curl --location --request POST 'localhost:11000/loggen' \
--header 'Content-Type: application/json' \
--data-raw '{
    "type": "web",
    "format": "nginx",
    "count": 1000,
    "framing": false
}'
```

Response:

```json
{
  "type": "web",
  "format": "nginx",
  "count": 1000,
  "framing": false
}
```

### Manage Memory Load Function

#### [GET] /memory

Call:

```sh
curl --location --request GET 'localhost:11000/memory'
```

Response:

```json
{
    "megabyte": 0,
    "active": "0001-01-01T00:00:00Z",
    "duration": 0,
    "last_modified": "0001-01-01T00:00:00Z"
}
```

#### [PATCH] /memory

Call:

```sh
curl --location --request PATCH 'localhost:11000/memory' \
--header 'Content-Type: application/json' \
--data-raw '{
    "megabyte": 100,
    "duration": 15
}'
```

Response:

```json
{
    "megabyte": 100,
    "active": "2021-09-09T17:41:47.813508+02:00",
    "duration": 15,
    "last_modified": "2021-09-09T17:41:32.813508+02:00"
}
```

### Manage CPU Load Function

#### [GET] /cpu

Call:

```sh
curl --location --request GET 'localhost:11000/cpu'
```

Response:

```json
{
    "load": 0,
    "duration": 0,
    "active": "0001-01-01T00:00:00Z",
    "core": 0,
    "last_modified": "0001-01-01T00:00:00Z"
}

```

#### [PATCH] /cpu

Call:

```sh
curl --location --request PATCH 'localhost:11000/cpu' \
--header 'Content-Type: application/json' \
--data-raw '{
    "load": 5.7,
    "duration": 10,
    "core": 2
}'
```

Response:

```json
{
    "load": 5.7,
    "duration": 10,
    "active": "2021-09-10T14:50:00.525809+02:00",
    "core": 2,
    "last_modified": "2021-09-10T14:49:50.525808+02:00"
}
```

### Manage Log Level Configuration

#### [GET] /log_level

Call:

```sh
curl --location --request GET 'localhost:11000/log_level'
```

Response:

```json
{
    "level": "debug",
    "last_modified": "0001-01-01T00:00:00Z"
}

```

#### [PATCH] /log_level

Call:

```sh
curl --location --request PATCH 'localhost:11000/log_level' \
--header 'Content-Type: application/json' \
--data-raw '{
    "level": "info"
}'
```

Response:

```json
{
    "level": "info",
    "last_modified": "2021-09-10T14:51:56.639658+02:00"
}
```

### Status

#### [GET] /

Call:

```sh
curl --location --request GET 'localhost:11000/'
```

Response:

```json
{
    "memory": {
        "megabyte": 0,
        "active": "0001-01-01T00:00:00Z",
        "duration": 0,
        "last_modified": "0001-01-01T00:00:00Z"
    },
    "cpu": {
        "load": 5.7,
        "duration": 10,
        "active": "2021-09-10T14:50:00.525809+02:00",
        "core": 2,
        "last_modified": "2021-09-10T14:49:50.525808+02:00"
    },
    "log_level": {
        "level": "info",
        "last_modified": "2021-09-10T14:51:56.639658+02:00"
    }
}
```

#### [PATCH] /

Call:

```sh
curl --location --request PATCH 'localhost:11000/' \
--header 'Content-Type: application/json' \
--data-raw '{
    "memory": {
        "megabyte": 70,
        "duration": 2
    },
    "cpu": {
        "load": 5.3,
        "duration": 11,
        "core": 1
    },
    "log_level": {
        "level": "debug"
    }
}'
```

Response:

```json
{
    "memory": {
        "megabyte": 70,
        "active": "2021-09-10T14:53:42.425137+02:00",
        "duration": 2,
        "last_modified": "2021-09-10T14:53:40.425137+02:00"
    },
    "cpu": {
        "load": 5.3,
        "duration": 11,
        "active": "2021-09-10T14:53:51.42514+02:00",
        "core": 1,
        "last_modified": "2021-09-10T14:53:40.42514+02:00"
    },
    "log_level": {
        "level": "debug",
        "last_modified": "2021-09-10T14:53:40.425195+02:00"
    }
}
```

## Testing with newman

Install [newman](https://github.com/postmanlabs/newman) with homebrew

```sh
brew install newman
```

Run the collection test

```sh
newman run Log-Generator.postman_collection.json --env-var "baseUrl=localhost:11000"
```

Expected Output:

```sh
Log-Generator

❏ Test / Loggen
↳ loggen
  GET localhost:11000/loggen [200 OK, 284B, 31ms]
  ✓  Status test

↳ loggen
  POST localhost:11000/loggen [200 OK, 171B, 7ms]
  ✓  Status test
  ✓  Content test

❏ Test / Memory
↳ memory
  GET localhost:11000/memory [200 OK, 221B, 5ms]
  ✓  Status test

↳ memory
  PATCH localhost:11000/memory [200 OK, 255B, 3ms]
  ✓  Status test
  ✓  Content test

❏ Test / CPU
↳ cpu
  GET localhost:11000/cpu [200 OK, 227B, 4ms]
  ✓  Status test

↳ cpu
  PATCH localhost:11000/cpu [200 OK, 260B, 3ms]
  ✓  Status test
  ✓  Content test

❏ Test / LogLevel
↳ log_level
  GET localhost:11000/log_level [200 OK, 179B, 3ms]
  ✓  Status test

↳ log_level
  PATCH localhost:11000/log_level [200 OK, 194B, 3ms]
  ✓  Status test
  ✓  Content test

❏ Test / State
↳ state
  GET localhost:11000// [200 OK, 711B, 5ms]
  ✓  Status test

↳ state
  PATCH localhost:11000// [200 OK, 708B, 4ms]
  ✓  Status test
  ✓  Content test

┌─────────────────────────┬─────────────────┬─────────────────┐
│                         │        executed │          failed │
├─────────────────────────┼─────────────────┼─────────────────┤
│              iterations │               1 │               0 │
├─────────────────────────┼─────────────────┼─────────────────┤
│                requests │              10 │               0 │
├─────────────────────────┼─────────────────┼─────────────────┤
│            test-scripts │              20 │               0 │
├─────────────────────────┼─────────────────┼─────────────────┤
│      prerequest-scripts │              10 │               0 │
├─────────────────────────┼─────────────────┼─────────────────┤
│              assertions │              15 │               0 │
├─────────────────────────┴─────────────────┴─────────────────┤
│ total run duration: 280ms                                   │
├─────────────────────────────────────────────────────────────┤
│ total data received: 1.97kB (approx)                        │
├─────────────────────────────────────────────────────────────┤
│ average response time: 6ms [min: 3ms, max: 31ms, s.d.: 8ms] │
└─────────────────────────────────────────────────────────────┘
```
