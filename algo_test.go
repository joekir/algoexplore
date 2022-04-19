package algoexplore

import (
	"testing"
)

type Fake struct {
	foobar int
}

func (fake *Fake) Name() string                        { return "fake" }
func (fake *Fake) Init(inputLen int)                   { return }
func (fake *Fake) Step(d byte)                         { return }
func (fake *Fake) SerializeState() string              { return "serialized" }
func (fake *Fake) DeserializeState(state string) error { return nil }

func TestRegister_withValidFactory_addsToRegistry(t *testing.T) {
	Register(func() AlgoPlugin { return &Fake{} })

	_, err := GetAlgo("random")
	if err == nil {
		t.Fatal("found some unregistered algorithm!")
	}

	_, err = GetAlgo("fake")
	if err != nil {
		t.Fatal("valid algorithm not found")
	}
}

func TestRegister_twiceWithSamePluginName_Panics(t *testing.T) {
	defer func() { recover() }()

	Register(func() AlgoPlugin { return &Fake{} })
	Register(func() AlgoPlugin { return &Fake{} })

	t.Errorf("did not panic on duplicate registration")
}
