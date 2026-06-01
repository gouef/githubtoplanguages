package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gouef/utils"
	"gopkg.in/yaml.v3"
)

type GraphQLUserResponse struct {
	Data DataUser `json:"data"`
}

type DataUser struct {
	Viewer ViewerUser `json:"viewer"`
}

type ViewerUser struct {
	Repositories UserRepositories `json:"repositories"`
	PullRequests UserPullRequests `json:"pullRequests"`
}

type UserRepositories struct {
	Edges    []UserRepositoryEdge `json:"edges"`
	PageInfo PageInfo             `json:"pageInfo"`
}

type UserRepositoryEdge struct {
	Node UserRepositoryNode `json:"node"`
}

type UserRepositoryNode struct {
	Name            string          `json:"name"`
	NameWithOwner   string          `json:"nameWithOwner"`
	IsFork          bool            `json:"isFork"`
	PrimaryLanguage PrimaryLanguage `json:"primaryLanguage"`
	Languages       Languages       `json:"languages"`
}

type UserPullRequests struct {
	Edges    []UserPullRequestEdge `json:"edges"`
	PageInfo PageInfo              `json:"pageInfo"`
}

type UserPullRequestEdge struct {
	Node UserPullRequestNode `json:"node"`
}

type UserPullRequestNode struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	Repository PRRepositoryDetail `json:"repository"`
	Files      PRFiles            `json:"files"`
}

type PRRepositoryDetail struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
}

type PRFiles struct {
	Nodes    []PRFileNode `json:"nodes"`
	PageInfo PageInfo     `json:"pageInfo"`
}

type PRFileNode struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
}

type LinguistLanguage struct {
	Color      string   `yaml:"color"`
	Extensions []string `yaml:"extensions"`
}

type GraphQLPRFilesResponse struct {
	Data struct {
		Node struct {
			Files PRFiles `json:"files"`
		} `json:"node"`
	} `json:"data"`
}

func LoadLinguistLanguages() (map[string]struct{ Name, Color string }, error) {
	url := "https://raw.githubusercontent.com/github-linguist/linguist/master/lib/linguist/languages.yml"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download linguist languages, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rawData map[string]LinguistLanguage
	if err := yaml.Unmarshal(body, &rawData); err != nil {
		return nil, err
	}

	extensionMap := make(map[string]struct{ Name, Color string })
	for langName, langInfo := range rawData {
		color := langInfo.Color
		if color == "" {
			color = "#cccccc"
		}

		for _, ext := range langInfo.Extensions {
			extensionMap[strings.ToLower(ext)] = struct{ Name, Color string }{
				Name:  langName,
				Color: color,
			}
		}
	}

	return extensionMap, nil
}

func FetchUser(loginName, token string, withFork bool, ignored ...string) (*Result, error) {
	var isForkStr string
	if withFork {
		isForkStr = "true"
	} else {
		isForkStr = "false"
	}

	query := fmt.Sprintf(`query($after: String) {
		  viewer {
			repositories(first: 100, isFork: %s, affiliations: OWNER, ownerAffiliations: OWNER, after: $after) {
			  edges {
				node {
				  name
				  nameWithOwner
				  isFork
				  primaryLanguage {
					name
				  }
				  languages(first: 10, after: null) {
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
				endCursor
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

			resultRepository := &ResultRepository{
				Name:   r.Node.NameWithOwner,
				IsFork: r.Node.IsFork,
			}

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
}

func FetchUserPRLanguages(token string, extensionMap map[string]struct{ Name, Color string }, ignored ...string) (*Result, error) {
	mainQuery := `query($after: String) {
		viewer {
			pullRequests(first: 50, after: $after, states: [OPEN, MERGED]) {
				edges {
					node {
						id
						title
						repository {
							name
							nameWithOwner
						}
						files(first: 100) {
							nodes {
								path
								additions
							}
							pageInfo {
								hasNextPage
								endCursor
							}
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	}`

	filePaginationQuery := `query($prId: ID!, $fileCursor: String) {
		node(id: $prId) {
			... on PullRequest {
				files(first: 100, after: $fileCursor) {
					nodes {
						path
						additions
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}`

	finalResult := &Result{}
	var prCursor interface{} = nil

	for {
		variables := map[string]interface{}{
			"after": prCursor,
		}

		resp, err := Request(token, mainQuery, variables)
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

		prData := result.Data.Viewer.PullRequests

		for _, edge := range prData.Edges {
			prNode := edge.Node
			repo := prNode.Repository
			if repo.Name == "" || utils.InArray(repo.Name, ignored) {
				continue
			}

			allFiles := append([]PRFileNode{}, prNode.Files.Nodes...)
			currentFilesPageInfo := prNode.Files.PageInfo

			for currentFilesPageInfo.HasNextPage {
				fileVariables := map[string]interface{}{
					"prId":       prNode.ID,
					"fileCursor": currentFilesPageInfo.EndCursor,
				}

				fileResp, err := Request(token, filePaginationQuery, fileVariables)
				if err != nil {
					return nil, err
				}

				if fileResp.StatusCode != http.StatusOK {
					fileResp.Body.Close()
					break
				}

				var fileResult GraphQLPRFilesResponse
				if err := json.NewDecoder(fileResp.Body).Decode(&fileResult); err != nil {
					fileResp.Body.Close()
					return nil, err
				}
				fileResp.Body.Close()

				allFiles = append(allFiles, fileResult.Data.Node.Files.Nodes...)
				currentFilesPageInfo = fileResult.Data.Node.Files.PageInfo
			}

			prLanguagesMap := make(map[string]*ResultLanguage)

			for _, file := range allFiles {
				if file.Additions == 0 {
					continue
				}

				ext := strings.ToLower(filepath.Ext(file.Path))

				if langInfo, ok := extensionMap[ext]; ok {
					if _, exists := prLanguagesMap[langInfo.Name]; !exists {
						prLanguagesMap[langInfo.Name] = &ResultLanguage{
							Name:  langInfo.Name,
							Color: langInfo.Color,
							Size:  0,
						}
					}
					prLanguagesMap[langInfo.Name].Size += file.Additions
				}
			}

			if len(prLanguagesMap) > 0 {
				resultRepository := &ResultRepository{
					Name: repo.NameWithOwner,
					IsPR: true,
				}

				for _, lang := range prLanguagesMap {
					resultRepository.Languages = append(resultRepository.Languages, lang)
				}

				finalResult.Repositories = append(finalResult.Repositories, resultRepository)
			}
		}

		if !prData.PageInfo.HasNextPage {
			break
		}
		prCursor = prData.PageInfo.EndCursor
	}

	return finalResult, nil
}
