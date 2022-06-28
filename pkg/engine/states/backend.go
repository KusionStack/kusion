package states

// TODO: We will refactor this file with StateStorage later

var Backends = make(map[string]func() StateStorage)

func AddToBackends(name string, storage func() StateStorage) {
	Backends[name] = storage
}
