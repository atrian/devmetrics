package crypto

type Hasher interface {
	Hash(metric string, key string) string
}
