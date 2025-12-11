# RabbitMQ

## A Game studio scenario: Using MQ for Broadcast
In online games, you often need to send a message to all connected players, but players are connected to different servers (Gateway / WebSocket servers).
- Server maintenance announcement
- World Boss has been defated
- A new event is starting
- Global chat message

## Overview

```
        ┌──────────────────────────────┐
        │ Admin API (/admin/broadcast) │
        │      Publishes message       │
        └───────────────┬──────────────┘
                        ▼
            [RabbitMQ Fanout Exchange]
       ┌───────────────┼──────────────────┐
       ▼               ▼                  ▼
  Queue for WS1   Queue for WS2     Queue for WS3
       ▼               ▼                  ▼
  WS Server 1     WS Server 2       WS Server 3
       ▼               ▼                  ▼
 Broadcast to     Broadcast to      Broadcast to
 connected users  connected users   connected users
```