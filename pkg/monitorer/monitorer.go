package monitorer

type Monitorer interface {
	Monitor(value string, resultChan chan<- []byte) error
}
