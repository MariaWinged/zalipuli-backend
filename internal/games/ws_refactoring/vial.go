package ws

import (
	"slices"
	"zalipuli/pkg/api"
)

// Vial - абстракция флакона с цветной водой
// Всего во флаконе может быть до четырех разноцветных сегментов воды
// Для игры достаточно количество цветов до 15, поэтому флакон мы можем хранить как 16-битное число, по 4 бита на каждый сегмент
type Vial uint16

type Vials []Vial

func (v Vials) Len() int           { return len(v) }
func (v Vials) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v Vials) Less(i, j int) bool { return v[i] < v[j] }

// NewVial - создает новый флакон из сегментов
func NewVial(segments []int) Vial {
	var hash uint16
	for _, segment := range segments {
		hash <<= ColorSize
		hash |= uint16(segment)
	}

	return Vial(hash)
}

// Segments - возвращает сегменты флакона
func (f Vial) Segments() []int {
	segments := make([]int, 0)
	for f > 0 {
		segment := int(f & (1<<ColorSize - 1))
		segments = append(segments, segment)
		f >>= ColorSize
	}
	slices.Reverse(segments)

	return segments
}

// LastSegment - возвращает последний сегмент флакона. Если флакон пуст, возвращает 0
func (f Vial) LastSegment() int {
	return int(f & (1<<ColorSize - 1))
}

// Len - возвращает количество сегментов в флаконе
func (f Vial) Len() int {
	var l int

	for ; f > 0; l++ {
		f >>= ColorSize
	}

	return l
}

func convertFromApiVials(apiVials api.Vials) Vials {
	vials := make([]Vial, 0, len(apiVials))
	for _, vial := range apiVials {
		vials = append(vials, NewVial(vial))
	}
	return vials
}

func convertToApiVials(vials Vials) api.Vials {
	apiVials := make(api.Vials, len(vials))
	for i, vial := range vials {
		apiVials[i] = vial.Segments()
	}
	return apiVials
}
