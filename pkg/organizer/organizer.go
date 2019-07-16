package organizer

type Organizer interface {
	Organize(filePath string, resultChan chan<- string) error
}
