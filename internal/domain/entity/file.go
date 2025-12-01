package entity

type File struct {
	Id       int
	Data     []byte
	Url      string
	Error    error
	LoaderId int
}

func NewFile(url string) File {
	return File{Url: url}
}
