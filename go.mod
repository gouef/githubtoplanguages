module github.com/gouef/githubtoplanguages

go 1.23.4

require (
	github.com/graphql-go/graphql v0.8.1
	github.com/joho/godotenv v1.5.1
)

replace github.com/gouef/githubtoplanguages/requests => ./requests

replace github.com/gouef/githubtoplanguages/requests/organizations => ./requests/organizations
