package ctph

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRollingHash(t *testing.T) {
	rh := NewRollingHash()

	x := rh.hash(byte(3))
	if x != 27 {
		t.Fatalf("expected 27, got %d", x)
	}

	x = rh.hash(byte(7))
	if x != 180 {
		t.Fatalf("expected 180, got %d", x)
	}

	rh.hash(byte(12))
	rh.hash(byte(59))
	x = rh.hash(byte(128))
	if x != 3390964 {
		t.Fatalf("expected 3390964, got %d", x)
	}
}

func TestCalcInitBlockSize(t *testing.T) {
	for _, tc := range []struct {
		length, expected uint32
	}{
		{767, 12},
		{1535, 24},
		{3071, 48},
	} {
		if x := calcInitBlockSize(tc.length); x != tc.expected {
			t.Fatalf("expected %d, got %d", tc.expected, x)
		}
	}
}

func TestCtphHash_WithMobyDick_MatchesExistingTool(t *testing.T) {
	// Comparison value generated with ssdeep tool
	expectedSSDeep := "384:S8G2SPXyDhU4nAnaFBtFrSx7zD74Z/kFSD:SM80YaFBtQDcZ/MSD"

	data, err := ioutil.ReadFile("testdata/mobydick.txt")
	if err != nil {
		t.Fatalf("could not read test file: %v", err)
	}

	var ctph *Ctph
	ctph = new(Ctph)
	ctph.Init(len(data))

	for ctph.Retry {
		for _, b := range data {
			ctph.Step(b)
			if !ctph.Retry {
				break
			}
		}
	}
	h := ctph.PrintSSDeep()

	if !cmp.Equal(expectedSSDeep, h) {
		t.Fatalf("Unexpected hash: %s", cmp.Diff(expectedSSDeep, h))
	}
}

func TestCtphHash_WithCrowAndFox_MatchesExistingTool(t *testing.T) {
	// Comparison value generated with ssdeep tool
	expectedSSDeep := "24:O7XC9FZ2LBfaW3h+XdcDljuQJtNMMqF5DjQuwM0OHC:O7S9FZ2LwWEdcM6tNMjDEuwwHC"

	data, err := ioutil.ReadFile("testdata/crowandthefox.txt")
	if err != nil {
		t.Fatalf("could not read test file: %v", err)
	}

	ctph := new(Ctph)
	ctph.Init(len(data))

	for ctph.Retry {
		for _, b := range data {
			ctph.Step(b)
			if !ctph.Retry {
				break
			}
		}
	}
	h := ctph.PrintSSDeep()

	if !cmp.Equal(expectedSSDeep, h) {
		t.Fatalf("Unexpected hash: %s", cmp.Diff(expectedSSDeep, h))
	}
}

func TestCompare_WithCrowAndFoxSlightlyTweaked_IsSimilar(t *testing.T) {
	original := "24:O7XC9FZ2LBfaW3h+XdcDljuQJtNMMqF5DjQuwM0OHC:O7S9FZ2LwWEdcM6tNMjDEuwwHC"

	d, err := ioutil.ReadFile("testdata/crowandthefox.txt")
	if err != nil {
		t.Fatalf("could not read test file: %v", err)
	}

	modified := []byte(strings.Replace(string(d), "vous", "tous", 1))
	ctph := new(Ctph)
	ctph.Init(len(modified))
	for ctph.Retry {
		for _, b := range modified {
			ctph.Step(b)
			if !ctph.Retry {
				break
			}
		}
	}
	newHash := ctph.PrintSSDeep()

	result, err := compareCtphSignatures(original, newHash)
	if err != nil {
		t.Fatalf("Failed to compare hashes: %v", err)
	}

	if result > 5 {
		t.Fatalf("comparison score (%d) incorrect for hashes:\n%s\n%s\n", result, original, newHash)
	}
}

func TestCompare_WithInvalidSignatures_ThrowsError(t *testing.T) {
	_, err := compareCtphSignatures("24O7XC9FZ2LBfaW3hM6tNMjDEuwwHC", "24:O7XC9FZ2LBfaW3h+XdcDljuQJtNMMqF5DjQuwM0OHC:O7S9FZ2LwWEdcM6tNMjDEuwwHC")

	if err == nil || err.Error() != "invalid pattern in string 1" {
		t.Fatalf("Compare should have failed with invalid sig pattern")
	}

	_, err = compareCtphSignatures("24:O7XC9FZ2LBfaW3h+XdcDljuQJtNMMqF5DjQuwM0OHC:O7S9FZ2LwWEdcM6tNMjDEuwwHC", "24O7XC9FZ2LBfaW3hM6tNMjDEuwwHC")

	if err == nil || err.Error() != "invalid pattern in string 2" {
		t.Fatalf("Compare should have failed with invalid sig pattern")
	}
}

func TestCompare_WithIncompatibleSignatures_ThrowsError(t *testing.T) {
	_, err := compareCtphSignatures("24:0:0", "12:O:O")

	if err == nil || err.Error() != "blocksize mismatch" {
		t.Fatalf("Compare should throw blocksize mismatch")
	}
}

func TestDeserializeState_withValidJson_parsestoCtph(t *testing.T) {
	json := `{"block_size":384,"index":-1,"input_length":12319,"is_trigger1":false,"is_trigger2":false,"rolling_hash":{"x":0,"y":0,"z":0,"c":0,"size":7,"window":[116,0,0,0,0,0,0]},"sig1":"","sig2":""}`
	ctph := new(Ctph)
	ctph.DeserializeState(json)

	if ctph.InputLen != 12319 {
		t.Fatalf("expected 12319, got %d", ctph.InputLen)
	}
}

func TestSerializeState_withCtphStruct_serializesToCorrectJSON(t *testing.T) {
	ctph := new(Ctph)
	ctph.InputLen = 37331
	if !strings.Contains(ctph.SerializeState(), "37331") {
		t.Fatal("expected serialized data to contain 37331 but it did not")
	}
}
