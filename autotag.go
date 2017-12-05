package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gobwas/glob"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func indexFile(name string) string {
	return name + ".files"
}

func walker(contents *[]string, wildcards []string) filepath.WalkFunc {
	var globs []glob.Glob

	for i := range wildcards {
		globs = append(globs, glob.MustCompile(wildcards[i]))
	}
	return func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			for i := range globs {
				if globs[i].Match(path) {
					*contents = append(*contents, path)
					break
				}
			}
		}
		return nil
	}
}

func createIndex(name string, path string, wildcards []string) {
	f, err := os.Create(name)
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	indexContents := []string{}
	filepath.Walk(path, walker(&indexContents, wildcards))
	for i := range indexContents {
		s := fmt.Sprintln(indexContents[i])
		bc, err := w.WriteString(s)
		check(err)
		if bc < len(s) {
			panic(fmt.Sprintf("Couldn't write to %s", name))
		}
	}
	w.Flush()
	return
}

func main() {
	config, err := os.Open("conf.json")
	if err != nil {
		log.Println(err)
		return
	}
	dec := json.NewDecoder(config)
	var v map[string]interface{}
	if err := dec.Decode(&v); err != nil {
		log.Println(err)
		return
	}
	config.Close()
	tagExe := v["executable"].(string)
	var args []string
	projects := v["projects"].([]interface{})
	for _, p := range projects {
		var projectWildcards []string
		project := p.(map[string]interface{})
		projectName := project["name"].(string)
		projectIndex := indexFile(projectName)
		projectTagpath := project["tagpath"].(string)
		projectFlags := project["flags"].([]interface{})
		args = make([]string, 0, len(projectFlags)+3)
		args = append(args, "-e")
		args = append(args, "-L")
		args = append(args, projectIndex)
		for _, q := range projectFlags {
			args = append(args, q.(string))
		}
		wildcards := project["wildcards"].([]interface{})
		for _, w := range wildcards {
			projectWildcards = append(projectWildcards, w.(string))
		}
		createIndex(projectIndex, projectTagpath, projectWildcards)
		fmt.Println("Creating", projectIndex)
		tagCmd := exec.Command(tagExe, args...)
		tagOut, err := tagCmd.CombinedOutput()
		fmt.Println("> UcTags Output")
		fmt.Println(string(tagOut))
		fmt.Println(err)
	}
}
