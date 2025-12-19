package repository

type Storage interface {
	SaveURL(url string) (string, error)
	GetURL(id string) (string, bool)
}
