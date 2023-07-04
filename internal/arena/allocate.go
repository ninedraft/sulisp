package arena

func Allocate[E any](n int) []E {
	return make([]E, n)
}
