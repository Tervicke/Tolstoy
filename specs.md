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
1. #### Acknowledge 
This is a broker only packet send by the broker to indidicate the accepetance of the previous packet , the topic and message data remains same
Type - 1
Topic - whatever topic was sent in the packet
Payload - whatever message was sent in the packet
2. #### Error
This is a broker only packte send by the broker to indidicate an error in the previous sent packet 
Type - 2 (1 byte)
Error - This is a string indicating an error ( 2048 bytes)

3. #### server packet
This is a server packet sent to all the subscribers with the topic name and the message 
Type - 2 
Topic - The topic in which u are subscribed and the message is sent in
Payload - The message

4. #### publish request
the payload will be sent to every subscriber subscribed to the topic provided in the packet
Type - 4
Topic - Topic to publish the message 
Message - Message to be published
