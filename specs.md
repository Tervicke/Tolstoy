## **Pub-Sub System Specifications**

Each packet in the system is structured as follows:

- **Type**: An integer representing the packet type.
- **Topic**: A 1024-byte ASCII string containing the topic name.
- **Payload**: A 1024-byte ASCII string containing the message payload.

| **Field**   | **Size**       | **Description**                                                                                                 |
|-------------|----------------|-----------------------------------------------------------------------------------------------------------------|
| **Type**    | 1 byte         | Integer value:<br> 1 = Connection request, 2 = Disconnection request, etc.                                      |
| **Topic**   | 1024 bytes     | ASCII string representing the topic (padded if necessary).                                                     |
| **Payload** | 1024 bytes     | ASCII string representing the payload/message (padded if necessary).                                           |

---

## **Examples**

### 1. **Acknowledge Packet**
- **Description**: Sent by the broker to indicate the acceptance of a previously received packet. The topic and payload remain unchanged.
  
  **Fields**:
  - **Type**: 1
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
