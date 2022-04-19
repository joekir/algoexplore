package algoexplore

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"sync"
)

var (
	algos      = map[string]AlgoFactory{}
	algosMutex sync.RWMutex
)

type AlgoPlugin interface {
	Name() string
	Init(inputLen int)
	Step(d byte)
	SerializeState() string
	DeserializeState(state string) error
}

type AlgoFactory func() AlgoPlugin

// RegisterAlgo <TODO>
func Register(algoFactory AlgoFactory) {
	algosMutex.Lock()
	defer algosMutex.Unlock()

	if _, ok := algos[algoFactory().Name()]; ok {
		panic(fmt.Sprintf("algo already registered: %s", algoFactory().Name()))
	}
	algos[algoFactory().Name()] = algoFactory
}

// GetAlgo <TODO>
func GetAlgo(name string) (AlgoPlugin, error) {
	algosMutex.RLock()
	defer algosMutex.RUnlock()
	m, ok := algos[name]
	if !ok {
		return nil, fmt.Errorf("algo not registered: %s", name)
	}
	return m(), nil
}

// Algos returns the names of all registered algorithms
func Algos() []string {
	algosMutex.RLock()
	defer algosMutex.RUnlock()

	names := make([]string, 0, len(algos))
	for name := range algos {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

func StrictUnmarshalJSON(data *io.Reader, v interface{}) error {
	dec := json.NewDecoder(*data)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
