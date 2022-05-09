package states

var Backends = make(map[string]func() StateStorage)

func AddToBackends(name string, storage func() StateStorage) {
	Backends[name] = storage
}
