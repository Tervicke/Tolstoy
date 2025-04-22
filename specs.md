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
1.#### Acknowledge 
This is a broker only packet send by the broker to indidicate the accepetance of the previous packet , the topic and message data remains same
Type - 1
Topic - whatever topic was sent in the packet
Payload - whatever message was sent in the packet

2. #### connection request 
The client socket is connected , The broker will start handling the upcoming packets
Type - 1
Topic - Ignored 
Payload - Ignored 

3. #### disconnection request 
The client socket is disconnected and the broker wont handle any future packets 
Type - 3 
Topic - Ignored 
Message - Ignored 
