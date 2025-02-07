package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GraphQLUserResponse struct {
	Data DataUser `json:"data"`
}

type DataUser struct {
	Viewer ViewerUser `json:"viewer"`
}

type ViewerUser struct {
	Repositories Repositories `json:"repositories"`
	PageInfo     PageInfo     `json:"pageInfo"`
}

func FetchUser(loginName, token string) (*ResultOrganizations, error) {
	query := `query {
		  viewer {
			repositories(first: 100, isFork: false) {
			  edges {
				node {
				  nameWithOwner
				  primaryLanguage {
					name
				  }
				  languages(first: 5, after: null) {
					edges {
					  node {
						name
						color
					  }
					  size
					}
				  }
				}
			  } 
			  pageInfo {
				hasNextPage
				hasPreviousPage
				endCursor
				startCursor
			  }
			}
		  }
		}`

	var result GraphQLUserResponse

	resp, err := Request(token, query)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned non-200 status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var resultOrganization = &ResultOrganizations{
		List:      map[string][]string{},
		Languages: make(map[string]int),
	}

	for _, r := range result.Data.Viewer.Repositories.Edges {
		resultOrganization.List[r.Node.NameWithOwner] = append(resultOrganization.List[r.Node.NameWithOwner], r.Node.NameWithOwner)

		for _, l := range r.Node.Languages.Edges {
			resultOrganization.Languages[l.Node.Name] += l.Size
		}
	}

	return resultOrganization, nil
}
