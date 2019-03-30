package jpegbw

import (
	"fmt"
	"math"
)

// IntHist holds uint16 -> int histogram
type IntHist map[uint16]int

// FloatHist holds uint16 -> percent histogram (cumulative)
type FloatHist map[uint16]float64

func (m IntHist) str() string {
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

func (m FloatHist) str() string {
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
