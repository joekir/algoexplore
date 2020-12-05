package algoexplore

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"sync"
)

var (
	algos      = make(map[string]AlgoInfo)
	algosMutex sync.RWMutex
)

type Algo interface {
	Algo() AlgoInfo
	Init(obj interface{}, inputLen int)
	Step(ser string, d byte)
	SerializeState() string
	DeserializeState(state string)
}

type AlgoInfo struct {
	Name string
	New  func() Algo
}

// RegisterAlgo <TODO>
func RegisterAlgo(instance Algo) {
	algo := instance.Algo()

	if val := algo.New(); val == nil {
		panic("AlgoInfo.New must return a non-nil module instance")
	}

	algosMutex.Lock()
	defer algosMutex.Unlock()

	if _, ok := algos[string(algo.Name)]; ok {
		panic(fmt.Sprintf("algo already registered: %s", algo.Name))
	}
	algos[string(algo.Name)] = algo
}

// GetAlgo <TODO>
func GetAlgo(name string) (AlgoInfo, error) {
	algosMutex.RLock()
	defer algosMutex.RUnlock()
	m, ok := algos[name]
	if !ok {
		return AlgoInfo{}, fmt.Errorf("algo not registered: %s", name)
	}
	return m, nil
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

func StrictUnmarshalJSON(data *io.ReadCloser, v interface{}) error {
	dec := json.NewDecoder(*data)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
