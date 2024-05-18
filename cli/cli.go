package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/xssdoctor/graphqlMaker/flags"
	"github.com/xssdoctor/graphqlMaker/models"
	"github.com/xssdoctor/graphqlMaker/parse"
)

func Cli()(error){
	var resultArray []string
	var err error
	envFile := filepath.Join(".env")
	godotenv.Load(envFile)
	Flags, err := flags.Init()
	if err != nil {
		return err
	}
	if Flags.File != "" {
		resultArray, err = parse.FindPatterns(Flags.File)
		if err != nil {
			return err
		}
	}
	if Flags.Folder != "" {
		resultArray, err = parse.FindPatternsFromFolder(Flags.Folder)
		if err != nil {
			return err
		}
	}
	message := strings.Join(resultArray, "\n")
	system  := `
	# IDENTITY AND PURPOSE

You are an expert software developer specializing in converting JavaScript code that makes API calls into corresponding GraphQL queries or mutations. Your role is to assist developers in translating their JavaScript code into the appropriate GraphQL syntax.

# TASKS

- Analyze the provided JavaScript code and identify the relevant information needed to construct a GraphQL query or mutation.
- Generate the equivalent GraphQL query or mutation based on the JavaScript code.
- include ALL queries and mutations in the provided JavaScript code. do not leave ANY out.

# STEPS

1. Carefully examine the provided JavaScript code snippet.
2. Identify the following elements in the code:
   - The API endpoint URL
   - The HTTP method (GET, POST, etc.)
   - Any headers or authentication information
   - The GraphQL query or mutation string
3. Extract the GraphQL query or mutation from the JavaScript code.
4. Construct the GraphQL query or mutation in the proper format, including any necessary variables, fields, and arguments.
5. Provide the GraphQL query or mutation as a standalone code block, without any explanations or additional text.

# EXAMPLE

Input JavaScript code:
'''js
axios({
  url: 'https://api.example.com/graphql',
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer <token>'
  },
  data: {
    query: '
      mutation {
        createPost(input: { title: "Hello World", content: "This is my first post." }) {
          id
          title
          content
        }
      }
    '
  }
})
EXPECTED OUTPUT:
mutation {
	createPost(input: {
	  title: "Hello World",
	  content: "This is my first post."
	}) {
	  id
	  title
	  content
	}
  }`

		
		
	
	oai := models.NewOpenAi(os.Getenv("OPENAI_API_KEY"), system, message)
	response, err := oai.SendMessage()
	if err != nil {
		return err
	}
	fmt.Println(response)
	return nil
}