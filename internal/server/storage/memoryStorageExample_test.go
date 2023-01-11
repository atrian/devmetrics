package storage_test

import (
	"fmt"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/storage"
	"github.com/atrian/devmetrics/pkg/logger"
)

func ExampleMemoryStorage() {
	var (
		l          logger.ILogger
		c          *serverconfig.Config
		memStorage *storage.MemoryStorage
	)

	l = logger.NewZapLogger()
	c = serverconfig.NewServerConfigWithoutFlags(l)
	memStorage = storage.NewMemoryStorage(c, l)

	err := memStorage.StoreCounter("YourCounter", int64(100500))
	if err != nil {
		l.Error("Store counter error", err)
	}

	value, exist := memStorage.GetCounter("YourCounter")
	fmt.Println(value, exist)

	value2, exist2 := memStorage.GetCounter("RandomCounter")
	fmt.Println(value2, exist2)

	// Output:
	// 100500 true
	// 0 false
}
