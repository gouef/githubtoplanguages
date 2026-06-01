package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gouef/utils"
)

type GraphQLUserResponse struct {
	Data DataUser `json:"data"`
}

type DataUser struct {
	Viewer ViewerUser `json:"viewer"`
}

type ViewerUser struct {
	Repositories Repositories `json:"repositories"`
}

func FetchUser(loginName, token string, withFork bool, ignored ...string) (*Result, error) {
	var isForkStr string
	if withFork {
		isForkStr = "true"
	} else {
		isForkStr = "false"
	}
	query := fmt.Sprintf(`query {
		  viewer {
			repositories(first: 100, isFork: %s, affiliations: OWNER, ownerAffiliations: OWNER, after: $after) {
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
		}`, isForkStr)

	finalResult := &Result{}
	var cursor interface{} = nil

	for {
		variables := map[string]interface{}{
			"after": cursor,
		}

		resp, err := Request(token, query, variables)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API returned non-200 status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
		}

		var result GraphQLUserResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		repoData := result.Data.Viewer.Repositories

		for _, r := range repoData.Edges {
			if utils.InArray(r.Node.Name, ignored) {
				continue
			}
			resultRepository := &ResultRepository{Name: r.Node.NameWithOwner}

			var languages []*ResultLanguage
			for _, l := range r.Node.Languages.Edges {
				languages = append(languages, &ResultLanguage{Name: l.Node.Name, Size: l.Size, Color: l.Node.Color})
			}
			resultRepository.Languages = languages

			finalResult.Repositories = append(finalResult.Repositories, resultRepository)
		}

		if !repoData.PageInfo.HasNextPage {
			break
		}

		cursor = repoData.PageInfo.EndCursor
	}

	return finalResult, nil

	/*
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

	   var result2 = &Result{}

	   	for _, r := range result.Data.Viewer.Repositories.Edges {
	   		if utils.InArray(r.Node.Name, ignored) {
	   			continue
	   		}
	   		resultRepository := &ResultRepository{Name: r.Node.NameWithOwner}

	   		var languages []*ResultLanguage
	   		for _, l := range r.Node.Languages.Edges {
	   			languages = append(languages, &ResultLanguage{Name: l.Node.Name, Size: l.Size, Color: l.Node.Color})
	   		}
	   		resultRepository.Languages = languages

	   		result2.Repositories = append(result2.Repositories, resultRepository)
	   	}

	   return result2, nil
	*/
}
