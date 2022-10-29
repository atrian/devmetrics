package storage

type Observer interface {
	RunOnClose()
	RunOnStart()
}
