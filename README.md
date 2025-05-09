# A lightweight pub sub system written in golang 
read the [specs](specs.md) to get a understanding of how the pub sub protocol.

![Tests](https://github.com/tervicke/Tolstoy/actions/workflows/test.yml/badge.svg)

##  Running the example
TO run the example follow these steps 
1. Start the main broker
```
make runbroker
```
2. Run the publisher , in a different terminal instance
```
make publisher
```
> **send a initial message so that the server has topic created**

3. Run subscriber 
```
make subscriber
```
![example screenshot](examples/examplescreenshot.png)


## Running the broker
```
./broker --config path_to_config.yaml
```

## Sample config
```yaml
# The port number on which the broker server will listen for incoming connections
Port: 8080

# The host address the server will bind to (usually "localhost" for local development)
Host: "localhost"

# A list of initial topics that the broker will recognize; clients can publish/subscribe to these
Topics:
  - "mytopic"
  - "anothertopic"

# Configuration related to message persistence
Persistence:
  # Whether to enable saving messages to disk (false means messages are kept in memory only)
  Enabled: false
  
  # Directory path where topic data will be stored if persistence is enabled
  Directory: "broker/data/topics/"
```

## ToDo
- [x] Make a simple demo publisher in Go
- [x] Make a simple demo subscriber in go
- [ ] Add baisc authentication in the topic access
- [x] Persistent storage of messages 
- [ ] Write a small benchmark tool
- [x] Add logging
- [x] better readme and build & run instructions