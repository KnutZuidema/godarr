package organizer

type Organizer interface {
	Organize(filePath string) error
}
