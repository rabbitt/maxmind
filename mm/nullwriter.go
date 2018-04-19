package mm

// ffjson: skip
type NullWriter struct{}

func (nw *NullWriter) Write(data []byte) (int, error) {
	return len(data), nil
}
