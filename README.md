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
  "type": "request",
  "id": "request1"
}
```

#### server response

```json
{
  "op": "get",
  "type": "response",
  "id": "request1",
  "peers": [
    "192.168.0.100:7070",
    "192.168.0.101:7070"
  ],
  "timeout": "20s",
  "config": {
    "addr": ":7070",
    "tls": {
      "addr": ":443",
      "cert": "./certs/latest/server.pem",
      "key": "./certs/latest/server.key"
    }
  }
}
```

### Connect

```json
{
  "op": "connect",
  "type": "request",
  "id": "c1",
  "addr": "192.168.0.100:7070"
}
```

#### server response (same as `get` operation)

```json
{
  "op": "connect",
  "type": "response",
  "id": "c1",
  "peers": [
    "192.168.0.100:7070",
    "192.168.0.101:7070"
  ],
  "timeout": "20s",
  "config": {
    "addr": ":7070",
    "tls": {
      "addr": ":443",
      "cert": "./certs/latest/server.pem",
      "key": "./certs/latest/server.key"
    }
  }
}
```


### Ping

```json
{
  "op": "ping",
  "type": "request",
  "id": "ping1"
}
```

#### server response

```json
{
  "op": "pong",
  "type": "response",
  "id": "ping1"
}
```

## SERVER SENT EVENTS

```json
{
  "op": "peers_changed",
  "type": "push",
  "id": "1",
  "peers": [
    "192.168.0.100:7070",
    "192.168.0.101:7070",
    "192.168.0.102:7070"
  ]
}
```

```json
{
  "op": "config_changed",
  "type": "push",
  "id": "2",
  "config": {
    "addr": ":7070",
    "feature1": "enabled",
    "tls": {
      "addr": ":443",
      "cert": "./certs/latest/server.pem",
      "key": "./certs/latest/server.key"
    }
  }
}
```
