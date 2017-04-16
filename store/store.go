package store

type Storer interface {
	GetHost(name string) *Host
	SaveHost(host *Host)
	HostExist(name string) bool
	HostToken(*Host) string
	Close() error
}
