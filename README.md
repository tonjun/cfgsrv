cfgsrv
=========

Go lang apps config server

```bash
./cfgsrv -c cofig.json -p 8080 -timeout 20s
```

## API

### Get Config

```json
{
  "op": "get",
}
```

#### server response

```json
{
  "servers": [
    "192.168.0.100:7070",
    "192.168.0.101:7070",
  ],
  "timeout": "20s",
  "config": {
    "addr": ":7070",
    "tls": {
      "addr": ":443",
      "cert": "./certs/latest/server.pem",
      "key": "./certs/latest/server.key",
    },
  }
}
```

### Connect

```json
{
  "op": "connect",
  "addr": "192.168.0.100:7070",
}
```

#### server response (same as `get` operation)

```json
{
  "servers": [
    "192.168.0.100:7070",
    "192.168.0.101:7070",
  ],
  "timeout": "20s",
  "config": {
    "addr": ":7070",
    "tls": {
      "addr": ":443",
      "cert": "./certs/latest/server.pem",
      "key": "./certs/latest/server.key",
    },
  }
}
```


### Ping

```json
{
  "op": "ping"
}
```

#### server response

```json
{
  "op": "pong"
}
```

## SERVER SENT EVENTS

```json
{
  "op": "servers_changed",
  "servers": [
    "192.168.0.100:7070",
    "192.168.0.101:7070",
    "192.168.0.102:7070",
  ],
}
```

```json
{
  "op": "config_changed",
  "config": {
    "addr": ":7070",
    "feature1": "enabled",
    "tls": {
      "addr": ":443",
      "cert": "./certs/latest/server.pem",
      "key": "./certs/latest/server.key",
    },
  }
}
```
