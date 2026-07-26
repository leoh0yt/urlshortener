package storage

type Storage interface {
	SaveId(string, string) error
	GetOriginalId(string) (string, error)
	GetShortenId(string) (string, error)
	GetNextId() (uint64, error)
	Close() error
}
