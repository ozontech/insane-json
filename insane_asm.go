package insaneJSON

//go:noescape
func InsaneSkipWC(a []byte, b byte) int

//go:noescape
func InsaneSkipWC_(a []byte, b []byte) (int, int)
