package port

type Library interface {
	Install(args ...any) error
	Uninstall() error
}

type Connector interface {
	Library
	Connect() error
	Disconnect() error
}
