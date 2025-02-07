package requests

type PageInfo struct {
	EndCursor       string `json:"endCursor"`
	StartCursor     string `json:"startCursor"`
	HasNextPage     bool   `json:"hasNextPage"`     // Example
	HasPreviousPage bool   `json:"hasPreviousPage"` // Example
}
