package requests

import (
	"encoding/json"
	"fmt"
	"github.com/gouef/utils"
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
	Name            string          `json:"name"`
	NameWithOwner   string          `json:"nameWithOwner"`
	PrimaryLanguage PrimaryLanguage `json:"primaryLanguage"`
	Languages       Languages       `json:"languages"`
}

type PrimaryLanguage struct {
	Name string `json:"Name"`
}

type Languages struct {
	Edges []LanguageEdge `json:"edges"`
}

type LanguageEdge struct {
	Node LanguageNode `json:"node"`
	Size int          `json:"Size"`
}
type LanguageNode struct {
	Name  string `json:"Name"`
	Color string `json:"color"`
}

type ResultOrganizations struct {
	List      map[string][]string
	Languages map[string]int
}

func FetchOrganizations(loginName, token string, ignored ...string) (*Result, error) {
	query := `query {
	  viewer {
		organizations(first: 100, after: null) {
		  nodes {
			login
			viewerCanAdminister
			repositories(first: 100, after: null, isFork: false, affiliations: OWNER, ownerAffiliations: OWNER) {
				edges {
					node {
						name
						nameWithOwner
						primaryLanguage {
							name
						}
						languages(first: 2, after: null) {
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

	var result2 = &Result{}

	for _, org := range result.Data.Viewer.Organizations.Nodes {
		if utils.InArray(org.Login, ignored) {
			continue
		}
		if org.CanAdminister {
			for _, r := range org.Repositories.Edges {
				if utils.InArray(r.Node.Name, ignored) {
					continue
				}
				resultRepository := &ResultRepository{Name: r.Node.NameWithOwner, Organization: r.Node.Name}

				var languages []*ResultLanguage
				for _, l := range r.Node.Languages.Edges {
					languages = append(languages, &ResultLanguage{Name: l.Node.Name, Size: l.Size, Color: l.Node.Color})
				}
				resultRepository.Languages = languages

				result2.Repositories = append(result2.Repositories, resultRepository)
			}
		}
	}

	return result2, nil
}
