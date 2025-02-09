module github.com/gouef/githubtoplanguages

go 1.23.4

require (
	github.com/gouef/utils v1.9.4
	github.com/joho/godotenv v1.5.1
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gouef/githubtoplanguages/requests => ./requests

replace github.com/gouef/githubtoplanguages/requests/organizations => ./requests/organizations
