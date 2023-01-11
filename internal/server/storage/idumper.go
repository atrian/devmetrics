package storage

// IDumper интерфейс сброса метрик в файл и восстановления метрик из файла
type IDumper interface {
	DumpToFile(filename string) error      // DumpToFile созраняет все накопленные метрики в файл
	RestoreFromFile(filename string) error // RestoreFromFile восстанавливает все метрики из файла в хранилище
}
