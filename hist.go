package jpegbw

import (
	"fmt"
	"io/ioutil"
	"math"

	yaml "gopkg.in/yaml.v2"
)

// IntHist holds uint16 -> int histogram
type IntHist map[uint16]int

// FloatHist holds uint16 -> percent histogram (cumulative)
type FloatHist map[uint16]float64

// FileHist holds full histogram data for a file
type FileHist struct {
	Hist    [4]IntHist   `yaml:"hist"`
	HistCum [4]FloatHist `yaml:"hist_cum"`
	Fn      string       `yaml:"file_name"`
}

// WriteHist - writes histogram to file
func (fh *FileHist) WriteHist() error {
	fn := fh.Fn + ".hist"
	yamlBytes, err := yaml.Marshal(fh)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fn, yamlBytes, 0644)
	return err
}

// Str - display histogram in human readable form
func (m IntHist) Str() string {
	s := ""
	for i := uint16(0); true; i++ {
		v := m[i]
		if v > 0 {
			s += fmt.Sprintf("%d => %d\n", i, m[i])
		}
		if i == 0xffff {
			break
		}
	}
	return s
}

// Str - display histogram in human readable form
func (m FloatHist) Str() string {
	s := ""
	prev := -1.0
	for i := uint16(0); true; i++ {
		v := m[i]
		if v > 0.00001 && v < 99.99999 && math.Abs(v-prev) > 0.00001 {
			s += fmt.Sprintf("%d => %.5f%%\n", i, m[i])
		}
		prev = v
		if i == 0xffff {
			break
		}
	}
	return s
}
