package downloader

type Downloader interface {
	Download(file []byte, resultChan chan<- string) error
}
