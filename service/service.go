package service

type Service interface {
	Shorten(string) (string, error)
	Resolve(string) (string, error)
}
