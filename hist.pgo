package jpegbw

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
)

// IntHist holds uint16 -> int histogram
type IntHist map[uint16]int64

// FloatHist holds uint16 -> percent histogram (cumulative)
type FloatHist map[uint16]float64

// FileHist holds full histogram data for a file
type FileHist struct {
	Hist    [4]IntHist   `json:"hist"`
	HistCum [4]FloatHist `json:"hist_cum"`
	Fn      string       `json:"file_name"`
}

// HintData holds moving histogram data for a given file
type HintData struct {
	From   int        `json:"from"`
	To     int        `json:"to"`
	Curr   int        `json:"curr"`
	Min    [4]uint16  `json:"min"`
	Max    [4]uint16  `json:"max"`
	LoPerc [4]float64 `json:"low_percent"`
	HiPerc [4]float64 `json:"high_percent"`
	LoIdx  [4]uint16  `json:"low_idx"`
	HiIdx  [4]uint16  `json:"high_idx"`
	Mult   [4]float64 `json:"mult"`
}

// WriteHist - writes histogram to file
func (fh *FileHist) WriteHist() error {
	fn := fh.Fn + ".hist"
	jsonBytes, err := json.Marshal(fh)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fn, jsonBytes, 0644)
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
