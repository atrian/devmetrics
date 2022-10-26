package storage

type Dumper interface {
	DumpToFile(filename string) error
	RestoreFromFile(filename string) error
}
