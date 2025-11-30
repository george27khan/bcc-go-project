package entity

type File struct {
	Id       int
	Data     []byte
	Url      string
	Status   string
	Error    string
	LoaderId int
}
