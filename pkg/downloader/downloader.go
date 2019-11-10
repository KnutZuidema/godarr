package downloader

type Downloader interface {
	Download(file []byte) error
}
