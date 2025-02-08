package requests

import (
	"encoding/json"
	"fmt"
	"github.com/gouef/utils"
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

func FetchUser(loginName, token string, ignored ...string) (*Result, error) {
	query := `query {
		  viewer {
			repositories(first: 100, isFork: false, affiliations: OWNER, ownerAffiliations: OWNER) {
			  edges {
				node {
				  name
				  nameWithOwner
				  primaryLanguage {
					name
				  }
				  languages(first: 3, after: null) {
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

	/*var resultOrganization = &ResultOrganizations{
		List:      map[string][]string{},
		Languages: make(map[string]int),
	}*/

	var result2 = &Result{}

	for _, r := range result.Data.Viewer.Repositories.Edges {

		if utils.InArray(r.Node.Name, ignored) {
			continue
		}
		//resultOrganization.List[r.Node.NameWithOwner] = append(resultOrganization.List[r.Node.NameWithOwner], r.Node.NameWithOwner)
		resultRepository := &ResultRepository{Name: r.Node.NameWithOwner}

		var languages []*ResultLanguage
		for _, l := range r.Node.Languages.Edges {
			//resultOrganization.Languages[l.Node.Name] += l.Size
			languages = append(languages, &ResultLanguage{Name: l.Node.Name, Size: l.Size})
		}
		resultRepository.Languages = languages

		result2.Repositories = append(result2.Repositories, resultRepository)
	}

	return result2, nil
	//return resultOrganization, nil
}
