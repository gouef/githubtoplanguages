package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const githubGraphQLAPI = "https://api.github.com/graphql"

type GraphQLResponse struct {
	Data struct {
		Viewer struct {
			Repositories  RepositoryConnection   `json:"repositories"`
			Organizations OrganizationConnection `json:"organizations"`
		} `json:"viewer"`
	} `json:"data"`
}

type RepositoryConnection struct {
	Nodes    []Repository `json:"nodes"`
	PageInfo PageInfo     `json:"pageInfo"`
}

type OrganizationConnection struct {
	Nodes []struct {
		Login        string               `json:"login"`
		Name         string               `json:"name"`
		ID           string               `json:"id"`
		Repositories RepositoryConnection `json:"repositories"`
	} `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Repository struct {
	Name      string `json:"name"`
	FullName  string `json:"nameWithOwner"`
	Languages struct {
		Edges []struct {
			Node struct {
				Name string `json:"name"`
			} `json:"node"`
			Size int `json:"size"`
		} `json:"edges"`
	} `json:"languages"`
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

func fetchOrganizations(token string) ([]string, error) {
	query := `{
		viewer {
			organizations(first: 100) {
				nodes {
					login
					name
					id
				}
			}
		}
	}`
	response, err := fetchRepositories(token, query)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Response: %v\n", response)

	var organizations []string
	for _, org := range response.Data.Viewer.Organizations.Nodes {
		// Zde můžete zkontrolovat, co přesně dostáváte v odpovědi
		fmt.Printf("Organization: %s, ID: %s\n", org.Login, org.ID)
		organizations = append(organizations, org.ID)
	}

	return organizations, nil
}

func fetchRepositoriesForOrganization(token string, orgLogin string, cursor string) ([]Repository, string, error) {
	query := fmt.Sprintf(`{
		organization(id: "%s") {
			repositories(first: 100, after: %q) {
				nodes { 
					name 
					nameWithOwner 
					languages(first: 10) { 
						edges { 
							node { name } 
							size 
						} 
					} 
				}
				pageInfo { hasNextPage endCursor }
			}
		}
	}`, orgLogin, cursor)

	response, err := fetchRepositories(token, query)
	if err != nil {
		return nil, "", err
	}

	// Ověření odpovědi
	if len(response.Data.Viewer.Organizations.Nodes) == 0 {
		fmt.Printf("No organizations found or no access to repositories for organization %s\n", orgLogin)
		return nil, "", fmt.Errorf("no organizations found or no access to repositories for organization %s", orgLogin)
	}

	// Ověření, zda organizace má repozitáře
	org := response.Data.Viewer.Organizations.Nodes[0]
	if len(org.Repositories.Nodes) == 0 {
		fmt.Printf("No repositories found for organization %s\n", orgLogin)
		return nil, "", nil
	}

	// Pokud organizace má repozitáře
	var repositories []Repository
	repositories = append(repositories, org.Repositories.Nodes...)
	var nextCursor string
	if org.Repositories.PageInfo.HasNextPage {
		nextCursor = org.Repositories.PageInfo.EndCursor
	}

	return repositories, nextCursor, nil
}

func fetchAllRepositories(token string) ([]Repository, error) {
	var repositories []Repository
	var repoCursor, orgCursor string

	// Získání organizací
	organizations, err := fetchOrganizations(token)
	if err != nil {
		return nil, err
	}

	// Získání repozitářů pro organizace
	for _, org := range organizations {
		for {
			orgRepos, nextCursor, err := fetchRepositoriesForOrganization(token, org, orgCursor)
			if err != nil {
				break
			}
			repositories = append(repositories, orgRepos...)
			if nextCursor == "" {
				break
			}
			orgCursor = nextCursor
		}
	}

	// Získání repozitářů pro uživatele
	query := fmt.Sprintf(`{
		viewer {
			repositories(first: 100, after: %q) {
				nodes { name nameWithOwner languages(first: 10) { edges { node { name } size } } }
				pageInfo { hasNextPage endCursor }
			}
		}
	}`, repoCursor)

	response, err := fetchRepositories(token, query)
	if err != nil {
		return nil, err
	}

	repositories = append(repositories, response.Data.Viewer.Repositories.Nodes...)
	if response.Data.Viewer.Repositories.PageInfo.HasNextPage {
		repoCursor = response.Data.Viewer.Repositories.PageInfo.EndCursor
	}

	return repositories, nil
}

func fetchRepositories(token string, query string) (*GraphQLResponse, error) {
	requestBody, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", githubGraphQLAPI, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed: %s", resp.Status)
	}

	var result GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN is not set")
	}

	repositories, err := fetchAllRepositories(token)
	if err != nil {
		log.Fatalf("Error fetching repositories: %v", err)
	}
	totalLanguages := make(map[string]int)
	for _, repo := range repositories {
		for _, lang := range repo.Languages.Edges {
			totalLanguages[lang.Node.Name] += lang.Size
		}
	}

	fmt.Println("Top jazyky napříč všemi repozitáři a organizacemi:")
	for lang, bytes := range totalLanguages {
		fmt.Printf("%s: %d bytes\n", lang, bytes)
	}
}
