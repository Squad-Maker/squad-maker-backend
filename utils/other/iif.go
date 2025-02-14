package otherUtils

// IIf is a ternary operator-like implementation
func IIf[T any](condition bool, trueValue T, falseValue T) T {
	// eu já estava de saco cheio de não ter um operador ternário na golang
	// mesmo essa implementação limitada já ajuda muito
	// ...meus tempos de VB6 ainda me assombram
	if condition {
		return trueValue
	}
	return falseValue
}
