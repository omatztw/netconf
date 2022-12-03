package netconf

type NotificationListener interface {
	Do(msg NotificationMsg)
}
