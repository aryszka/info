package keyval

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"testing"
)

const lookupTestN = 12429

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
