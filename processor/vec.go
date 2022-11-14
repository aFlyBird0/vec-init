package processor

type vector string

func (v vector) string() string {
	return string(v)
}
