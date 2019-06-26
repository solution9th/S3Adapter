package db

// DB DB function
type DB interface {
	LinkDB(config map[string]interface{}) error
	AddTable() (err error)
	CountInfo(ak, sk, engine string) (int, error)
	GetInfo(ak string) (m interface{}, err error)
	SaveInfo(data map[string]interface{}) (id int, err error)
	DeleteInfo(oak, osk string) error
}
