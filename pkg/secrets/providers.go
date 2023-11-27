package secrets

import (
	"encoding/json"
	"fmt"
	"sync"

	"golang.org/x/exp/maps"

	"kusionstack.io/kusion/pkg/apis/secrets"
	"kusionstack.io/kusion/pkg/log"
)

type Providers struct {
	lock     sync.RWMutex
	registry map[string]SecretStoreFactory
}

func NewProviders() *Providers {
	return &Providers{}
}

// Register registers a provider with associated spec. This
// is expected to happen during app startup.
func (ps *Providers) Register(ssf SecretStoreFactory, spec *secrets.ProviderSpec) {
	providerName, err := getProviderName(spec)
	if err != nil {
		panic(fmt.Sprintf("provider registery failed to parse spec: %s", err.Error()))
	}

	ps.lock.Lock()
	defer ps.lock.Unlock()
	if ps.registry != nil {
		_, found := ps.registry[providerName]
		if found {
			log.Warnf("Provider %s was registered twice", providerName)
		}
	} else {
		ps.registry = map[string]SecretStoreFactory{}
	}

	log.Infof("Registered secret store provider %s", providerName)
	ps.registry[providerName] = ssf
}

// GetProviderByName returns registered provider by name.
func (ps *Providers) GetProviderByName(providerName string) (SecretStoreFactory, bool) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	provider, found := ps.registry[providerName]
	return provider, found
}

func getProviderName(spec *secrets.ProviderSpec) (string, error) {
	specBytes, err := json.Marshal(spec)
	if err != nil || specBytes == nil {
		return "", fmt.Errorf("failed to marshal secret store provider spec: %w", err)
	}

	specMap := make(map[string]interface{})
	err = json.Unmarshal(specBytes, &specMap)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal secret store provider spec: %w", err)
	}

	if len(specMap) != 1 {
		return "", fmt.Errorf("secret stores must only have exactly one provider specified, found %d", len(specMap))
	}

	return maps.Keys(specMap)[0], nil
}
