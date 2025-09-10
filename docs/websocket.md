# WebSocket Events API

This document describes how to use the WebSocket API for real-time Ethereum events.

## Connecting to the WebSocket API

Connect to the WebSocket endpoint at `/api/v1/events/ws`. Once connected, you will receive real-time updates for Ethereum events.

## Event Types

The following event types are supported:

- `new_block`: Triggered when a new block is mined
- `new_transaction`: Triggered when a new transaction is confirmed (in a block)
- `contract_event`: Triggered when a contract event is emitted

## Message Format

### Event Messages

Event messages sent from the server will have the following format:

```json
{
  "type": "new_block|new_transaction|contract_event",
  "blockHash": "0x...",
  "blockNum": 12345678,
  "txHash": "0x...",
  "data": { ... }
}
```

### Subscription Requests

To subscribe to specific contract events, send a message with the following format:

```json
{
  "type": "subscribe",
  "contract": "0x...",
  "events": ["Transfer(address,address,uint256)", "Approval(address,address,uint256)"]
}
```

If the `events` array is empty, you will subscribe to all events from the contract.

### Filter Requests

To filter events, send a message with the following format:

```json
{
  "type": "filter",
  "eventTypes": ["new_block", "new_transaction", "contract_event"],
  "contracts": ["0x..."]
}
```

## Event Data

### New Block Event

```json
{
  "type": "new_block",
  "blockHash": "0x...",
  "blockNum": 12345678,
  "data": {
    "number": "12345678",
    "hash": "0x...",
    "parentHash": "0x...",
    "timestamp": 1628097894,
    "txCount": 123
  }
}
```

### New Transaction Event

```json
{
  "type": "new_transaction",
  "blockHash": "0x...",
  "blockNum": 12345678,
  "txHash": "0x...",
  "data": {
    "hash": "0x...",
    "to": "0x...",
    "value": "1000000000000000000",
    "gasPrice": "20000000000",
    "gas": 21000,
    "nonce": 42
  }
}
```

### Contract Event

```json
{
  "type": "contract_event",
  "blockHash": "0x...",
  "blockNum": 12345678,
  "txHash": "0x...",
  "data": {
    "address": "0x...",
    "topics": ["0x...", "0x...", "0x..."],
    "data": "0x..."
  }
}
```

## Example Usage

Here's an example of how to connect to the WebSocket endpoint and subscribe to events:

```javascript
// Create WebSocket connection
const socket = new WebSocket('ws://localhost:8080/api/v1/events/ws');

// Connection opened
socket.addEventListener('open', function() {
  console.log('Connected to server');
  
  // Subscribe to a specific contract
  socket.send(JSON.stringify({
    type: 'subscribe',
    contract: '0x1234567890123456789012345678901234567890',
    events: ['Transfer(address,address,uint256)']
  }));
});

// Listen for messages
socket.addEventListener('message', function(event) {
  const data = JSON.parse(event.data);
  console.log('Received event:', data);
});

// Connection closed
socket.addEventListener('close', function() {
  console.log('Disconnected from server');
});
```
