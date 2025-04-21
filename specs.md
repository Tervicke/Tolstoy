## The specs for the pub sub system

Each packet is structured in the following way
Type - It is an int denoting the the type of the packet
Topic - Its an 1024 bytes long ascii string , containing the topic name
Message - Its the 1024 bytes payload which is acted

| Field    | Size       | Description																									|
|----------|------------|------------------------------------------------------------ |
| Type     | 1 byte     | Integer value: 1 = connection request , 2 = disconnection request, etc.             |
| Topic    | 1024 bytes | ASCII string representing the topic (padded if necessary)   |
| Message  | 1024 bytes | ASCII string representing the message (padded if necessary) |
---------------------------------------------------------------------------------------

## Examples

1. #### connection request 
The client socket is connected
Type - 1
Topic - Ignored 
Message - Ignored 

2. #### disconnection request 
The client socket is disconnected
Type - 2
Topic - Ignored 
Message - Ignored 
