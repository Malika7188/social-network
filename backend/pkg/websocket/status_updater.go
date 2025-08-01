package websocket

// StatusUpdater defines the interface for updating user status
type StatusUpdater interface {
	SetUserOnline(userID string) error
	SetUserOffline(userID string) error
}
