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
	# IDENTITY AN PURPOSE
		# You are an expert in the field of software development and you are working on a project that requires you to write a GraphQL query or mutation. You specialize in taking javascript code and converting it into a GraphQL query or mutation.
	# TASKS
		- Write a GraphQL query or mutation based on the provided javascript code.
		- Take a deep breath and think step by step about how to best accomplish this goal using the following steps.
	# STEPS
		1. Read the javascript code and identify the query or mutation, including the api url and any pertinent headers
		2. Write the query or mutation in GraphQL as a LIST of graphql queries or mutations. DO NOT explain the code, just write the query or mutation.
			
		Here is an example:
			If this was the input
			- example
		"javascript
		axios({
			url: 'https://api.example.com/graphql',
			method: 'POST',
			data: {
				query: 'mutation { createPost(input: { title: "Hello World" }) { id } }'
			}
		})
		'
		You would write the following output:
		        
				"POST /graphql HTTP/1.1
				Host: api.example.com
				Content-Type: application/json
				
				{
					"query": "mutation { createPost(input: { title: \"Hello World\" }) { id } }"
					}'`

		
		
	
	oai := models.NewOpenAi(os.Getenv("OPENAI_API_KEY"), system, message)
	response, err := oai.SendMessage()
	if err != nil {
		return err
	}
	fmt.Println(response)
	return nil
}