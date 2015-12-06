package keyval

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"testing"
)

const lookupTestN = 345

func BenchmarkReadKeyval(b *testing.B) {
	all, err := ioutil.ReadFile("test.k")
	if err != nil {
		b.Error(err)
		return
	}

	buf := bytes.NewBuffer(all)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := NewEntryReader(buf)

		for {
			_, err := r.ReadEntry()
			if err != nil && err != io.EOF {
				b.Error(err)
				break
			}

			if err == io.EOF {
				break
			}
		}
	}
}

func BenchmarkCompareReadJson(b *testing.B) {
	all, err := ioutil.ReadFile("test.json")
	if err != nil {
		b.Error(err)
		return
	}

	buf := bytes.NewBuffer(all)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d := json.NewDecoder(buf)
		o := make(map[string]interface{})
		if err := d.Decode(&o); err != nil && err != io.EOF {
			b.Error(err)
			break
		}
	}
}

func BenchmarkCompareReadYaml(b *testing.B) {
	all, err := ioutil.ReadFile("test.yaml")
	if err != nil {
		b.Error(err)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		o := make(map[string]interface{})
		if err := yaml.Unmarshal(all, &o); err != nil && err != io.EOF {
			b.Error(err)
			break
		}
	}
}

func lookupInit() (*Buffer, [][]string) {
	g := newGen(genOptions{})

	entries := g.n(lookupTestN)
	buf := &Buffer{}
	buf.AppendEntry(entries...)

	lookupKeys := buf.Keys()
	notfoundKeys := make([][]string, lookupTestN)
	for i := 0; i < lookupTestN; i++ {
		notfoundKeys[i] = g.strs(g.minKeyLength, g.maxKeyLength, g.keyChars)
	}

	lookupKeys = append(lookupKeys, notfoundKeys...)
	return buf, lookupKeys
}

func benchmarkLookup(b *testing.B, lookup func([]string) []*Entry, lookupKeys [][]string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, k := range lookupKeys {
			lookup(k)
		}
	}
}

func BenchmarkBufferLookupString(b *testing.B) {
	buf, lookupKeys := lookupInit()
	benchmarkLookup(b, buf.lookupString, lookupKeys)
}

func BenchmarkBufferLookupSlice(b *testing.B) {
	buf, lookupKeys := lookupInit()
	benchmarkLookup(b, buf.lookupSlice, lookupKeys)
}

func BenchmarkBufferLookupMap(b *testing.B) {
	buf, lookupKeys := lookupInit()
	benchmarkLookup(b, buf.lookupMap, lookupKeys)
}

func readInit() (*EntryReader, error) {
	entries := newGen(genOptions{}).n(lookupTestN)

	buf := bytes.NewBuffer(nil)
	w := NewEntryWriter(buf)
	for _, e := range entries {
		if err := w.WriteEntry(e); err != nil {
			return nil, err
		}
	}

	return NewEntryReader(buf), nil
}

func benchmarkBufferRead(b *testing.B, r *EntryReader, read func(*EntryReader) error) {
	for i := 0; i < b.N; i++ {
		if err := read(r); err != nil && err != io.EOF {
			b.Error(err)
			return
		}
	}
}

func BenchmarkRead(b *testing.B) {
	r, err := readInit()
	if err != nil {
		b.Error(err)
		return
	}

	buf := &Buffer{}
	benchmarkBufferRead(b, r, buf.read)
}

func BenchmarkReadCached(b *testing.B) {
	r, err := readInit()
	if err != nil {
		b.Error(err)
		return
	}

	buf := &Buffer{}
	benchmarkBufferRead(b, r, buf.readCached)
}
