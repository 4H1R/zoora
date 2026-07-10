// Package chat implements IN-MEETING chat: LiveRoom-scoped messages and
// reactions delivered in realtime over LiveKit data channels.
//
// Not to be confused with the standalone messaging stack:
//   - internal/conversations — direct/group/channel conversations (REST)
//   - internal/chathub       — its WebSocket hub + Redis fan-out
package chat
