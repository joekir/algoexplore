package ctph

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/joekir/algoexplore"

	"github.com/agnivade/levenshtein"
)

func init() {
	log.Println("registering ctph algorithm")
	algoexplore.RegisterAlgo(Ctph{})
}

type Ctph struct {
	Bs         uint32      `json:"block_size"`
	Hash1      Sum32       `json:"-"`
	Hash2      Sum32       `json:"-"`
	Index      int         `json:"index"`
	InputLen   int         `json:"input_length"`
	IsTrigger1 bool        `json:"is_trigger1"`
	IsTrigger2 bool        `json:"is_trigger2"`
	Retry      bool        `json:"-"`
	Rh         RollingHash `json:"rolling_hash"`
	Sig1       string      `json:"sig1"`
	Sig2       string      `json:"sig2"`
}

type RollingHash struct {
	X      uint32   `json:"x"`
	Y      uint32   `json:"y"`
	Z      uint32   `json:"z"`
	C      uint32   `json:"c"`
	Size   uint32   `json:"size"`
	Window []uint32 `json:"window"`
}

func (Ctph) Algo() algoexplore.AlgoInfo {
	return algoexplore.AlgoInfo{
		Name: "ctph",
		New:  func() algoexplore.Algo { return new(Ctph) },
	}
}

const (
	b64Chars     string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	ssLength     uint32 = 64
	ssSigPattern string = "^\\d+:[0-9a-zA-Z+\\/]+:[0-9a-zA-Z+\\/]+$"
	windowSize   uint32 = 7
	blockSizeMin uint32 = 3
)

func (ctph Ctph) Init(obj interface{}, InputLen int) {
	if InputLen < 1 {
		log.Fatal("invalid input length")
	}
	p, isCtph := obj.(*Ctph)
	if !isCtph {
		log.Fatal("invalid interface provided")
	}

	p.InputLen = InputLen
	p.Retry = true
	p.Bs = calcInitBlockSize(uint32(InputLen))
	p.reset()
	ctph = *p
}

func (ctph Ctph) Step(ser string, d byte) {
	json.Unmarshal([]byte(ser), &ctph)

	ctph.Index++
	if ctph.Index >= ctph.InputLen {
		ctph.Sig1 += string(b64Chars[ctph.Hash1.Sum32()&0x3F])
		ctph.Sig2 += string(b64Chars[ctph.Hash2.Sum32()&0x3F])

		if uint32(len(ctph.Sig1)) >= ssLength/2 || ctph.Bs == blockSizeMin {
			ctph.Retry = false
			return
		}

		ctph.reset()
		ctph.Bs = ctph.Bs / 2
		return
	}

	rs := ctph.Rh.hash(d)
	ctph.Hash1.Write([]byte{d})
	ctph.Hash2.Write([]byte{d})
	ctph.IsTrigger1, ctph.IsTrigger2 = false, false

	if mod := rs % ctph.Bs; mod == ctph.Bs-1 {
		ctph.Sig1 += string(b64Chars[ctph.Hash1.Sum32()&0x3F])
		ctph.IsTrigger1 = true
		ctph.Hash1.Reset() // reinit the hash
	}

	if mod := rs % (2 * ctph.Bs); mod == (2*ctph.Bs)-1 {
		ctph.Sig2 += string(b64Chars[ctph.Hash2.Sum32()&0x3F])
		ctph.IsTrigger2 = true
		ctph.Hash2.Reset() // reinit the hash
	}
}

func (ctph Ctph) SerializeState() string {
	byteArray, err := json.Marshal(ctph)
	if err != nil {
		log.Fatal(err)
	}
	return string(byteArray)
}

func (ctph Ctph) DeserializeState(state string) {
	if err := json.Unmarshal([]byte(state), &ctph); err != nil {
		log.Fatal(err)
	}
}

// Implementation based on https://github.com/ssdeep-project/ssdeep/blob/master/fuzzy.c#L383
func calcInitBlockSize(u uint32) uint32 {
	var bi uint32
	for (blockSizeMin<<bi)*ssLength < u {
		bi++
	}

	return blockSizeMin << bi
}

/*
* Compares 2 CTPH Signatures (s1 and s2) of the form:
* 24:O7XC9FZ2LBfaW3h+XdcDljuQJtNMMqF5DjQuwM0OHC:O7S9FZ2LwWEdcM6tNMjDEuwwHC
* <blocksize>:<sigpart1>:<sigpart2>
*
* The 2006 paper recommends using levenshtein distance to compare
*
* This function will return the distance as a positive integer
 */
func compareCtphSignatures(s1, s2 string) (int, error) {
	if matched, err := regexp.MatchString(ssSigPattern, s1); err != nil || !matched {
		return -1, errors.New("invalid pattern in string 1")
	}

	if matched, err := regexp.MatchString(ssSigPattern, s2); err != nil || !matched {
		return -1, errors.New("invalid pattern in string 2")
	}

	Sig1 := strings.Split(s1, ":")
	Sig2 := strings.Split(s2, ":")

	if Sig1[0] != Sig2[0] {
		return -1, errors.New("blocksize mismatch")
	}

	firstDiff := levenshtein.ComputeDistance(Sig1[1], Sig2[1])
	secondDiff := levenshtein.ComputeDistance(Sig1[2], Sig2[2])

	if secondDiff < firstDiff {
		return secondDiff, nil
	}

	return firstDiff, nil
}

func (ctph Ctph) PrintSSDeep() string {
	return fmt.Sprintf("%d:%s:%s", ctph.Bs, ctph.Sig1, ctph.Sig2)
}

func (ctph *Ctph) reset() {
	ctph.Hash1, ctph.Hash2 = *NewFNV(), *NewFNV()
	ctph.Index = -1
	ctph.Rh = *NewRollingHash()
	ctph.Sig1, ctph.Sig2 = "", ""
}

func NewRollingHash() *RollingHash {
	return &RollingHash{
		Size:   windowSize,
		Window: make([]uint32, windowSize),
	}
}

func (rh *RollingHash) hash(d byte) uint32 {
	dint := uint32(d)
	rh.Y = rh.Y - rh.X
	rh.Y = rh.Y + rh.Size*dint
	rh.X = rh.X + dint
	rh.X = rh.X - rh.Window[rh.C%rh.Size]
	rh.Window[rh.C%rh.Size] = dint
	rh.C = rh.C + 1
	rh.Z = rh.Z << 5
	rh.Z = rh.Z ^ dint

	return (rh.X + rh.Y + rh.Z)
}
