package otherUtils

type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

type MapRand[K Numeric] struct {
	m map[K]struct{}
}

func (m *MapRand[K]) InitializeInterval(start, end K) {
	m.m = make(map[K]struct{}, int(end-start)+1)
	for i := start; i <= end; i++ {
		m.m[i] = struct{}{}
	}
}

func (m *MapRand[K]) GetRandomAndPop() (K, bool) {
	if len(m.m) == 0 {
		return 0, false
	}

	var key K
	for k := range m.m {
		key = k
		break
	}
	delete(m.m, key)
	return key, true
}
