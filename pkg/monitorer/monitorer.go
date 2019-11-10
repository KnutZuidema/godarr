package monitorer

type Monitorer interface {
	Monitor(value string) error
}
