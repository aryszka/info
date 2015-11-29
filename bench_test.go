package keyval

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
	// "os"
	"encoding/json"
	"gopkg.in/yaml.v2"
)

func BenchmarkReadKeyval(b *testing.B) {
	all, err := ioutil.ReadFile("test.k")
	if err != nil {
		b.Error(err)
		return
	}

	buf := bytes.NewBuffer(all)
	// bufOut := bytes.NewBuffer(nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := NewReader(buf)
		// w := NewWriter(bufOut)
		for {
			// entry, err := r.ReadEntry()
			_, err := r.ReadEntry()
			if err != nil && err != io.EOF {
				b.Error(err)
				break
			}

			// if entry != nil {
			//     w.WriteEntry(entry)
			// }

			if err == io.EOF {
				break
			}
		}
	}

	// b.StopTimer()
	// if err := ioutil.WriteFile("test-check.k", bufOut.Bytes(), os.ModePerm); err != nil {
	//     b.Error(err)
	// }
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

	buf := bytes.NewBuffer(all)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		o := make(map[string]interface{})
		if err := yaml.Unmarshal(buf.Bytes(), &o); err != nil && err != io.EOF {
			b.Error(err)
			break
		}
	}
}
