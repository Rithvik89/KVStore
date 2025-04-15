package store

type IStoreManager interface {
	Get(key string) (string, error)
	Put(key string, value string) error
	Delete(key string)
}

type StoreManager struct {
	Store IStoreManager `json:"store"`
}

func NewStoreManager() *StoreManager {
	return &StoreManager{
		Store: NewInMemStore(),
	}
}
