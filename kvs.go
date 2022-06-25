package kvs

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var isDebug = os.Getenv("DEBUG") == "true"

func debugLog(a ...any) {
	if isDebug {
		log.Println(a...)
	}
}

func init() {
	debugLog("enable debug mode")
}

const KeyLimit = 512

type Pair struct {
	Key      *[KeyLimit]byte
	Value    *[]byte
	IsDelete bool
	Expire   Expire
}

type Transaction struct {
	Pair []Pair
}

type Expire struct {
	Time     time.Time
	NoExpire bool
}

func (e *Expire) Expire(t time.Time) bool {
	if e.NoExpire {
		return false
	}
	return t.After(e.Time)
}

type KV map[[KeyLimit]byte]*Pair

type KVS struct {
	mtx    sync.RWMutex
	data   KV
	expire KV
}

func NewKVS() *KVS {
	s := &KVS{}
	s.data = map[[KeyLimit]byte]*Pair{}
	s.expire = map[[KeyLimit]byte]*Pair{}
	return s
}

var ErrEncode = errors.New("failed data encode")

func EncodePair(p Pair) ([]byte, error) {
	b := &bytes.Buffer{}
	e := gob.NewEncoder(b)

	if err := e.Encode(p); err != nil {
		return nil, ErrEncode
	}

	return b.Bytes(), nil
}

func DecodePair(b []byte) (Pair, error) {
	var pair Pair
	bv := bytes.NewBuffer(b)
	d := gob.NewDecoder(bv)

	if err := d.Decode(&pair); err != nil {
		return Pair{}, fmt.Errorf("failed restore: %w", err)
	}

	return pair, nil
}

func EncodeTrans(p Pair) ([]byte, error) {
	b := &bytes.Buffer{}
	e := gob.NewEncoder(b)

	if err := e.Encode(p); err != nil {
		return nil, ErrEncode
	}

	return b.Bytes(), nil
}

func DecodeTrans(b []byte) (Transaction, error) {
	var transaction Transaction
	bv := bytes.NewBuffer(b)
	d := gob.NewDecoder(bv)

	if err := d.Decode(&transaction); err != nil {
		return Transaction{}, fmt.Errorf("failed restore: %w", err)
	}

	return transaction, nil
}

func cloneKV(kv KV) KV {
	cloned := KV{}

	for k, v := range kv {
		cloned[k] = v
	}

	return cloned
}
