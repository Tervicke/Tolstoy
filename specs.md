## **Pub-Sub System Specifications**
Each packet in the system is structured as follows:

- **Type**: An integer representing the packet type.
- **Topic**: A 1024-byte ASCII string containing the topic name.
- **Payload**: A 1024-byte ASCII string containing the message payload.

| **Field**   | **Size**       | **Description**                                                                                                 |
|-------------|----------------|-----------------------------------------------------------------------------------------------------------------|
| **Type**    | 1 byte         | Integer value:<br> serverpacket , error packet etc.                                      |
| **Topic**   | 1024 bytes     | ASCII string representing the topic (padded if necessary).                                                     |
| **Payload** | 1024 bytes     | ASCII string representing the payload/message (padded if necessary).                                           |

---

## **Examples**

### 1. **Acknowledge Packet**
- **Description**: Sent by the broker to indicate the acceptance of a previously received packet. The topic and payload remain unchanged.
    
    Acknowledgment Types:

    Type Code | ACK Type         | Description
    ----------|------------------|--------------------------------------------------
    10        | PUBLISH_ACK      | Acknowledges receipt of a published message
    11        | SUBSCRIBE_ACK    | Acknowledges a successful subscribe
    12        | UNSUBSCRIBE_ACK  | Acknowledges unsubscribe completion


  **Fields**:
  - **Type**: One of the above
  - **Topic**: Same as received packet
  - **Payload**: Same as received packet

---

### 2. **Error Packet**
- **Description**: Sent by the broker to indicate an error in a previously received packet.

  **Fields**:
  - **Type**: 2
  - **Payload**: A string indicating the error (2048 bytes)

---

### 3. **Server Packet**
- **Description**: Sent by the broker to all subscribers of a topic with the relevant message content.

  **Fields**:
  - **Type**: 3
  - **Topic**: The topic being published to
  - **Payload**: The message payload

---

### 4. **Publish Packet**
- **Description**: Sent by a publisher to deliver a message to all subscribers of a given topic.A new topic is created if the topic doesnt exist. 

  **Fields**:
  - **Type**: 4
  - **Topic**: The topic to publish to
  - **Payload**: The message payload

---

### 5. **Subscribe Packet**
- **Description**: Sent by a client to subscribe to a topic. Will return an error packet if no such topic exists

  **Fields**:
  - **Type**: 5
  - **Topic**: The topic to subscribe to
  - **Payload**: Ignored
  
### 6. **Unsubscribe Packet**
- **Description**: Sent by a client to unsubscribe to a topic, Will return an ack even if the client is not subscribed to the topic or the topic does not exist

  **Fields**:
  - **Type**: 6
  - **Topic**: The topic to unsubscribe to
  - **Payload**: Ignored
