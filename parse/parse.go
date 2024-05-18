package parse

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/xssdoctor/graphqlMaker/models"
)

var patterns = []string{
	`fetch\s*\(['"][^'"]+['"],\s*{[^}]*method:\s*['"]POST['"][^}]*body:\s*{["']query["']:\s*['"][^'"]+['"][^}]*}`,
	`axios\s*\({[^}]*url:\s*['"][^'"]+['"],\s*method:\s*['"]POST['"][^}]*data:\s*{["']query["']:\s*['"][^'"]+['"][^}]*}\s*\)`,
	`['"]{["']query["']\s*:\s*['"]\s*(mutation|query)\s+[^{]+?{`,
	`client.query\s*{[^}]*query:\s*[^}]+}\s*`,
	`client.mutate\s*{[^}]*mutation:\s*[^}]+}\s*`,
	`Relay.Environment\s*\.sendQuery\s*\({[^}]*text:\s*['"]\s*query\s+[^{]+?{`,
	`Relay.Environment\s*\.commitMutation\s*\({[^}]*text:\s*['"]\s*mutation\s+[^{]+?{`,
	`apolloClient\.query\s*\({[^}]*query:\s*['"]\s*query\s+[^{]+?{`,
	`apolloClient\.mutate\s*\({[^}]*mutation:\s*['"]\s*mutation\s+[^{]+?{`,
	`gql\s*\(\s*['"]\s*(mutation|query)\s+[^{]+?{`,
	`graphql\s*\(\s*['"]\s*(mutation|query)\s+[^{]+?{`,
	`useQuery\s*\(\s*['"]\s*(mutation|query)\s+[^{]+?{`,
	`useMutation\s*\(\s*['"]\s*(mutation|query)\s+[^{]+?{`,
	`useLazyQuery\s*\(\s*['"]\s*(mutation|query)\s+[^{]+?{`,
	`useSubscription\s*\(\s*['"]\s*(mutation|query)\s+[^{]+?{`,
	`useQuery\s*\(\s*['"]\s*(mutation|query)\s+[^{]+?{`,
	`useMutation\s*\(\s*['"]\s*(mutation|query)\s+[^{]+?{`,
	`query`,
	`mutation`,
	`Query`,
	`Mutation`,
}
var system string = `
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

func readLines(file *os.File) ([]string, error) {
	var lines []string
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func FindPatterns(filename string) ([]string, error) {
	var result []string
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if strings.Contains("/", filename) {
		filename = strings.Split(filename, "/")[0]
		filename = filepath.Join(cwd, filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	lineList, err := readLines(file)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	resultsChan := make(chan string, len(lineList))
	for _, pattern := range patterns {
		wg.Add(1)
		go func(pattern string) {
			defer wg.Done()
			regex, err := regexp.Compile(pattern)
			if err != nil {
				fmt.Printf("Error compiling regex: %v\n", err)
				return
			}
			for i, line := range lineList {
				if regex.MatchString(line) {
					var start, end int
					if i < 10 {
						start = 0
					} else {
						start = i - 10
					}
					if i > len(lineList)-11 {
						end = len(lineList)
					} else {
						end = i + 11
					}
					aboveAndBelow := lineList[start:end]
					compiledList := strings.Join(aboveAndBelow, "\n")
					mu.Lock()
					resultsChan <- compiledList
					mu.Unlock()
				}
			}
		}(pattern)
	}
	wg.Wait()
	close(resultsChan)
	for chanresult := range resultsChan {
		result = append(result, chanresult)
	}
	return result, nil
}

func FindPatternsFromFolder(folderName string) ([]string, error) {
	var resultArray []string
	var wg sync.WaitGroup
	resultsChan := make(chan []string, 100)

	err := filepath.WalkDir(folderName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".js") {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				patternResults, err := FindPatterns(path)
				if err != nil {
					fmt.Printf("Error finding patterns in file: %v\n", err)
					return
				}
				resultsChan <- patternResults
			}(path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for chanresult := range resultsChan {
		if len(chanresult) > 0 {
			message := strings.Join(chanresult, "\n")
			cwd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			envFile := filepath.Join(cwd, ".env")
			godotenv.Load(envFile)
			oai := models.NewOpenAi(os.Getenv("OPENAI_API_KEY"), system, message)
			response, err := oai.SendMessage()
			if err != nil {
				return nil, err
			}
			fmt.Println(response)
		}
	}

	return resultArray, nil
}
