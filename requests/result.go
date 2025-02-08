package requests

type Result struct {
	Repositories []*ResultRepository
}

type ResultRepository struct {
	Organization string
	Name         string
	Languages    []*ResultLanguage
}

type ResultLanguage struct {
	Name string
	Size int
}
