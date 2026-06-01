package requests

type Result struct {
	Repositories []*ResultRepository
}

type ResultRepository struct {
	Organization string
	Name         string
	Languages    []*ResultLanguage
	IsFork       bool
	IsPR         bool
	PRCount      int
}

type ResultLanguage struct {
	Name  string
	Color string
	Size  int
}
