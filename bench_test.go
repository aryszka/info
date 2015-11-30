package keyval

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"testing"
)

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

		// to how much remembering the encoding/json code,
		// a similar buffer size is used there. This value
		// can be out-of-date.
		r.BufferSize = 1 << 9

		var collect []*Entry
		for {
			entry, err := r.ReadEntry()
			if err != nil && err != io.EOF {
				b.Error(err)
				break
			}

			if entry != nil {
				collect = append(collect, entry)
				// w.WriteEntry(entry)
			}

			if err == io.EOF {
				break
			}
		}
	}
}

func BenchmarkReadJson(b *testing.B) {
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

func BenchmarkReadYaml(b *testing.B) {
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

func BenchmarkReadNoAlloc(b *testing.B) {
	all, err := ioutil.ReadFile("test.k")
	if err != nil {
		b.Error(err)
		return
	}

	buf := bytes.NewBuffer(all)
	ibuf := make([]byte, DefaultReadBufferSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := NewEntryReader(buf)
		r.buffer = ibuf

		var collect []*Entry
		for {
			entry, err := r.ReadEntry()
			if err != nil && err != io.EOF {
				b.Error(err)
				break
			}

			if entry != nil {
				collect = append(collect, entry)
			}

			if err == io.EOF {
				break
			}
		}
	}
}
