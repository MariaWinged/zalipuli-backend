package ws

const (
	// VialHeight - высота флакона (количество сегментов)
	VialHeight = 4
	// MinColorsCount - минимальное количество цветов в уровне
	MinColorsCount = 5
	// MaxColorsCount - максимальное количество цветов в уровне
	MaxColorsCount = 15
	// ColorSize - размер сегмента в битах
	ColorSize = 4
)

const gameName = "watersort"

var WaterSortGraph *Graph

// OneColoredVials - флаконы, заполненные одним цветом
var OneColoredVials = make([]Vial, 0, MaxColorsCount)

// FinalPositionsHash - хэши успешных завершений уровней
var FinalPositionsHash = make([]string, 0, MaxColorsCount)

func FillConstants() {
	vials := Vials{
		0, 0,
	}
	for i := uint8(0); i < MaxColorsCount; i++ {
		var vial Vial
		for j := 0; j < VialHeight; j++ {
			vial <<= ColorSize
			vial |= Vial(i + 1)
		}

		OneColoredVials = append(OneColoredVials, vial)
		vials = append(vials, vial)
		position := NewPosition(vials)
		FinalPositionsHash = append(FinalPositionsHash, position.Hash)
	}
}
