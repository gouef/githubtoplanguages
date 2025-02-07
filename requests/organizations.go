package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GraphQLResponse struct {
	Data Data `json:"data"`
}

type Data struct {
	Viewer Viewer `json:"viewer"`
}

type Viewer struct {
	Organizations Organizations `json:"organizations"`
}

type Organizations struct {
	Nodes    []OrganizationNode `json:"nodes"`
	PageInfo PageInfo           `json:"pageInfo"`
}

type OrganizationNode struct {
	Login         string       `json:"login"`
	CanAdminister bool         `json:"viewerCanAdminister"`
	Repositories  Repositories `json:"repositories"`
}

type Repositories struct {
	Edges []RepositoryEdge `json:"edges"`
}

type RepositoryEdge struct {
	Node RepositoryNode `json:"node"`
}

type RepositoryNode struct {
	NameWithOwner   string          `json:"nameWithOwner"`
	PrimaryLanguage PrimaryLanguage `json:"primaryLanguage"`
	Languages       Languages       `json:"languages"`
}

type PrimaryLanguage struct {
	Name string `json:"name"`
}

type Languages struct {
	Edges []LanguageEdge `json:"edges"`
}

type LanguageEdge struct {
	Node LanguageNode `json:"node"`
	Size int          `json:"size"`
}
type LanguageNode struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type ResultOrganizations struct {
	List      map[string][]string
	Languages map[string]int
}

func FetchOrganizations(loginName, token string) (*ResultOrganizations, error) {
	query := `query {
	  viewer {
		organizations(first: 100, after: null) {
		  nodes {
			login
			viewerCanAdminister
			repositories(first: 100, after: null, isFork: false, affiliations: OWNER) {
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

	var result GraphQLResponse

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

	for _, org := range result.Data.Viewer.Organizations.Nodes {
		login := org.Login
		if org.CanAdminister {
			for _, r := range org.Repositories.Edges {
				resultOrganization.List[login] = append(resultOrganization.List[login], r.Node.NameWithOwner)

				for _, l := range r.Node.Languages.Edges {
					resultOrganization.Languages[l.Node.Name] += l.Size
				}
			}
		}
	}

	return resultOrganization, nil
}
