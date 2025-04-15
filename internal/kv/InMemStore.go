package store

type InMemStore struct {
	store map[string]string
}

func NewInMemStore() *InMemStore {
	return &InMemStore{
		store: make(map[string]string),
	}
}

func (s *InMemStore) Get(key string) (string, error) {
	value, exists := s.store[key]
	if !exists {
		return "", nil
	}
	return value, nil
}

func (s *InMemStore) Put(key string, value string) error {
	s.store[key] = value
	return nil
}

func (s *InMemStore) Delete(key string) {
	delete(s.store, key)
}
