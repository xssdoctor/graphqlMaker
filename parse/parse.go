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
	resultArray := []string{}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	folderPath := filepath.Join(cwd, folderName)
	fileEntries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	resultsChan := make(chan []string, len(fileEntries))
	for _, entry := range fileEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".js") {
			wg.Add(1)
			go func(entry fs.DirEntry) {
				defer wg.Done()
				filePath := filepath.Join(folderPath, entry.Name())
				patternResults, err := FindPatterns(filePath)
				if err != nil {
					fmt.Printf("Error finding patterns in file: %v\n", err)
					return
				}
				mu.Lock()
				resultsChan <- patternResults
				mu.Unlock()
			}(entry)
		}
	}
	wg.Wait()
	close(resultsChan)
	for chanresult := range resultsChan {
		resultArray = append(resultArray, chanresult...)
	}
	return resultArray, nil
}
