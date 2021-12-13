# Log-Generator
_API managed testing tool for logging-operator_


##  Deploy log-generator with Helm
Add the chart repository of the Log-generator using the following commands:

```sh
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm repo update
```

### Install the Log-generator.

```sh
helm install --wait --generate-name banzaicloud-stable/log-generator
```


## Usage

You can start the daemon serving the API on e.g. port 11000 by running the following command:
```sh
$ go run main.go
```

Now you can connect to http://localhost:11000 from your browser or using your favorite HTTP client.

## Available API Calls 

### Manage Memory Load Function
#### [GET] /state/memory 

call:
```sh
curl --location --request GET 'localhost:11000/state/memory' 
```

Response:
```sh
{
    "megabyte": 0,
    "active": "0001-01-01T00:00:00Z",
    "duration": 0,
    "last_modified": "0001-01-01T00:00:00Z"
}
```

#### [PATCH] /state/memory 

Call:
```sh
curl --location --request PATCH 'localhost:11000/state/memory' \
--header 'Content-Type: application/json' \
--data-raw '{
    "megabyte": 100,
    "duration": 15
}'
```

Response:
```sh
{
    "megabyte": 100,
    "active": "2021-09-09T17:41:47.813508+02:00",
    "duration": 15,
    "last_modified": "2021-09-09T17:41:32.813508+02:00"
}
```


### Manage CPU Load Function

#### [GET] /state/cpu 

Call:
```sh
curl --location --request GET 'localhost:11000/state/cpu'
```
Response:
```sh
{
    "load": 0,
    "duration": 0,
    "active": "0001-01-01T00:00:00Z",
    "core": 0,
    "last_modified": "0001-01-01T00:00:00Z"
}

```

#### [PATCH] /state/cpu 
Call:
```sh
curl --location --request PATCH 'localhost:11000/state/cpu' \
--header 'Content-Type: application/json' \
--data-raw '{
    "load": 5.7,
    "duration": 10,
    "core": 2
}'
```
Response:
```sh
{
    "load": 5.7,
    "duration": 10,
    "active": "2021-09-10T14:50:00.525809+02:00",
    "core": 2,
    "last_modified": "2021-09-10T14:49:50.525808+02:00"
}
```

### Manage Log Level Configuration
#### [GET] /state/log_level 
Call:
```sh
curl --location --request GET 'localhost:11000/state/log_level' 
```
Response:
```sh
{
    "level": "debug",
    "last_modified": "0001-01-01T00:00:00Z"
}

```

#### [PATCH] /state/log_level 
Call:
```sh
curl --location --request PATCH 'localhost:11000/state/log_level' \
--header 'Content-Type: application/json' \
--data-raw '{
    "level": "info"
}'
```

Response:
```sh
{
    "level": "info",
    "last_modified": "2021-09-10T14:51:56.639658+02:00"
}
```


### Status

#### [GET] /state 
Call:
```sh
curl --location --request GET 'localhost:11000/state’
```
Response:
```sh
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

#### [PATCH] /state/
Call:
```sh
curl --location --request PATCH 'localhost:11000/state' \
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
```sh
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
Install newman(https://github.com/postmanlabs/newman) with homebrew

`$ brew install newman`

Run the collection test

```sh
newman run Log-Generator.postman_collection.json --env-var "baseUrl=localhost:11000"
```

Expected Output:
```sh
Log-Generator

❏ Test / Memory
↳ memory
  GET localhost:11000/state/memory [200 OK, 247B, 43ms]
  ✓  Status test

↳ memory
  PATCH localhost:11000/state/memory [200 OK, 249B, 5ms]
  ✓  Status test
  ✓  Content test

❏ Test / CPU
↳ cpu
  GET localhost:11000/state/cpu [200 OK, 252B, 4ms]
  ✓  Status test

↳ cpu
  PATCH localhost:11000/state/cpu [200 OK, 254B, 5ms]
  ✓  Status test
  ✓  Content test

❏ Test / LogLevel
↳ log_level
  GET localhost:11000/state/log_level [200 OK, 191B, 3ms]
  ✓  Status test

↳ log_level
  PATCH localhost:11000/state/log_level [200 OK, 191B, 2ms]
  ✓  Status test
  ✓  Content test

❏ Test / State
↳ state
  GET localhost:11000/state [200 OK, 478B, 4ms]
  ✓  Status test

↳ state
  PATCH localhost:11000/state [200 OK, 474B, 4ms]
  ✓  Status test
  ✓  Content test

┌─────────────────────────┬──────────────────┬─────────────────┐
│                         │         executed │          failed │
├─────────────────────────┼──────────────────┼─────────────────┤
│              iterations │                1 │               0 │
├─────────────────────────┼──────────────────┼─────────────────┤
│                requests │                8 │               0 │
├─────────────────────────┼──────────────────┼─────────────────┤
│            test-scripts │               16 │               0 │
├─────────────────────────┼──────────────────┼─────────────────┤
│      prerequest-scripts │                8 │               0 │
├─────────────────────────┼──────────────────┼─────────────────┤
│              assertions │               12 │               0 │
├─────────────────────────┴──────────────────┴─────────────────┤
│ total run duration: 323ms                                    │
├──────────────────────────────────────────────────────────────┤
│ total data received: 1.31KB (approx)                         │
├──────────────────────────────────────────────────────────────┤
│ average response time: 8ms [min: 2ms, max: 43ms, s.d.: 12ms] │
└──────────────────────────────────────────────────────────────┘
```

