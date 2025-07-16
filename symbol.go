package qrcode

type symbol struct {
	module        [][]bool
	isUsed        [][]bool
	size          int
	symbolSize    int
	quietZoneSize int
}

func newSymbol(size int, quietZoneSize int) *symbol {
	var m symbol

	m.module = make([][]bool, size+2*quietZoneSize)
	m.isUsed = make([][]bool, size+2*quietZoneSize)

	for i := range m.module {
		m.module[i] = make([]bool, size+2*quietZoneSize)
		m.isUsed[i] = make([]bool, size+2*quietZoneSize)
	}

	m.size = size + 2*quietZoneSize
	m.symbolSize = size
	m.quietZoneSize = quietZoneSize

	return &m
}

func (m *symbol) get(x int, y int) (v bool) {
	v, _ = m.values(x, y)
	return
}

func (m *symbol) empty(x int, y int) bool {
	_, isUsed := m.values(x, y)
	return !isUsed
}

func (m *symbol) numEmptyModules() int {
	var count int
	for y := range m.symbolSize {
		for x := range m.symbolSize {
			if m.empty(x, y) {
				count++
			}
		}
	}

	return count
}

func (m *symbol) set(x int, y int, v bool) {
	index := func(v int) int { return v + m.quietZoneSize }
	m.module[index(y)][index(x)] = v
	m.isUsed[index(y)][index(x)] = true
}

func (m *symbol) values(x, y int) (module, isUsed bool) {
	index := func(v int) int { return v + m.quietZoneSize }
	return m.module[index(y)][index(x)], m.isUsed[index(y)][index(x)]
}

func (m *symbol) set2dPattern(x int, y int, v [][]bool) {
	for j, row := range v {
		for i, value := range row {
			m.set(x+i, y+j, value)
		}
	}
}

func (m *symbol) bitmap() [][]bool {
	module := make([][]bool, len(m.module))

	for i := range m.module {
		module[i] = m.module[i][:]
	}

	return module
}

func (m *symbol) String() string {
	var result string

	for _, row := range m.module {
		for _, value := range row {
			switch value {
			case true:
				result += "  "
			case false:
				result += "██"
			}
		}
		result += "\n"
	}

	return result
}

const (
	penaltyWeight1 = 3
	penaltyWeight2 = 3
	penaltyWeight3 = 40
	penaltyWeight4 = 10
)

func (m *symbol) penaltyScore() int {
	return m.penalty1() + m.penalty2() + m.penalty3() + m.penalty4()
}

func (m *symbol) penalty1() int {
	penalty := 0

	for x := range m.symbolSize {
		lastValue := m.get(x, 0)
		count := 1

		for y := 1; y < m.symbolSize; y++ {
			v := m.get(x, y)

			if v != lastValue {
				count = 1
				lastValue = v
			} else {
				count++
				if count == 6 {
					penalty += penaltyWeight1 + 1
				} else if count > 6 {
					penalty++
				}
			}
		}
	}

	for y := range m.symbolSize {
		lastValue := m.get(0, y)
		count := 1

		for x := 1; x < m.symbolSize; x++ {
			v := m.get(x, y)

			if v != lastValue {
				count = 1
				lastValue = v
			} else {
				count++
				if count == 6 {
					penalty += penaltyWeight1 + 1
				} else if count > 6 {
					penalty++
				}
			}
		}
	}

	return penalty
}

func (m *symbol) penalty2() int {
	penalty := 0

	for y := 1; y < m.symbolSize; y++ {
		for x := 1; x < m.symbolSize; x++ {
			topLeft := m.get(x-1, y-1)
			above := m.get(x, y-1)
			left := m.get(x-1, y)
			current := m.get(x, y)

			if current == left && current == above && current == topLeft {
				penalty++
			}
		}
	}

	return penalty * penaltyWeight2
}

func (m *symbol) penalty3() int {
	penalty := 0

	for y := range m.symbolSize {
		var bitBuffer int16 = 0x00

		for x := range m.symbolSize {
			bitBuffer <<= 1
			if v := m.get(x, y); v {
				bitBuffer |= 1
			}

			switch bitBuffer & 0x7ff {
			case 0x05d, 0x5d0:
				penalty += penaltyWeight3
				bitBuffer = 0xFF
			default:
				if x == m.symbolSize-1 && (bitBuffer&0x7f) == 0x5d {
					penalty += penaltyWeight3
					bitBuffer = 0xFF
				}
			}
		}
	}

	for x := range m.symbolSize {
		var bitBuffer int16 = 0x00

		for y := range m.symbolSize {
			bitBuffer <<= 1
			if v := m.get(x, y); v {
				bitBuffer |= 1
			}

			switch bitBuffer & 0x7ff {
			case 0x05d, 0x5d0:
				penalty += penaltyWeight3
				bitBuffer = 0xFF
			default:
				if y == m.symbolSize-1 && (bitBuffer&0x7f) == 0x5d {
					penalty += penaltyWeight3
					bitBuffer = 0xFF
				}
			}
		}
	}

	return penalty
}

func (m *symbol) penalty4() int {
	numModules := m.symbolSize * m.symbolSize
	numDarkModules := 0

	for x := range m.symbolSize {
		for y := 0; y < m.symbolSize; y++ {
			if v := m.get(x, y); v {
				numDarkModules++
			}
		}
	}

	numDarkModuleDeviation := numModules/2 - numDarkModules
	if numDarkModuleDeviation < 0 {
		numDarkModuleDeviation *= -1
	}

	return penaltyWeight4 * (numDarkModuleDeviation / (numModules / 20))
}
