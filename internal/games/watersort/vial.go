package watersort

import "slices"

// Vial - абстракция флакона с цветной водой
// Всего в флаконе может быть до четырех разноцветных сегментов воды
// Для игры достаточно количество цветов до 15, поэтому флакон мы можем хранить как 16-битное число, по 4 бита на каждый сегмент
type Vial uint16

// NewVial - создает новую флакон из сегментов
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
