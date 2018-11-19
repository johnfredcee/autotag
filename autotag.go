package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

/*
 * Given a name return the index file name
 */
func indexFile(name string) string {
	return name + ".files"
}

/*
 * Given a name return the tag file name
 */
func tagFile(name string) string {
	return name + ".tags"
}

/**
 * Return a fucntion that will recursively walk a path, filtering with the
* given wildcards, and accumulating file names in contents
*/
func walker(contents *[]string, wildcards []string) filepath.WalkFunc {
	var globs []glob.Glob

	for i := range wildcards {
		globs = append(globs, glob.MustCompile(wildcards[i]))
	}
	return func(path string, info os.FileInfo, err error) error {
		if (info != nil) && (!info.IsDir()) {
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

/*
 * Create an named index file, with filenames recursively found on the given path,
 * filtered by the given wildcards
 */
func createIndex(name string, paths []interface{}, wildcards []string) {
	f, err := os.Create(name)
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	indexContents := []string{}
	for _, path := range paths {
		p := path.(string)
		filepath.Walk(p, walker(&indexContents, wildcards))
	}
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

var sem = make(chan int, 0)

func copyOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		log.Print(line)
		//contents = append(contents, line)
	}
}

func projectDirectory(p interface{}) string {
	project := p.(map[string]interface{})
	projectTagpath := project["tagpath"].(string)
	return projectTagpath
}

func scanProject(p interface{}, tagExe string) {
	var projectWildcards []string
	var args []string
	project := p.(map[string]interface{})
	projectName := project["name"].(string)
	projectIndex := indexFile(projectName)
	outputFile := tagFile(projectName)
	projectTagpaths := project["tagpath"].([]interface{})
	projectFlags := project["flags"].([]interface{})
	args = make([]string, 0, len(projectFlags)+4)
	args = append(args, "-L")
	args = append(args, projectIndex)
	args = append(args, "-f")
	args = append(args, outputFile)
	for _, q := range projectFlags {
		args = append(args, q.(string))
	}
	wildcards := project["wildcards"].([]interface{})
	for _, w := range wildcards {
		projectWildcards = append(projectWildcards, w.(string))
	}
	fmt.Println("Creating", projectIndex)
	createIndex(projectIndex, projectTagpaths, projectWildcards)
	fmt.Println("Tagging", projectIndex)
	tagCmd := exec.Command(tagExe, args...)

	stdout, err := tagCmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := tagCmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = tagCmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Launching ", tagCmd.Args)
	copyOutput(stdout)
	copyOutput(stderr)

	tagCmd.Wait()
	fmt.Println("Done ", tagCmd.Args)
	sem <- 1
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
	projects := v["projects"].([]interface{})
	var projectCount = len(projects)
	for _, p := range projects {
		go scanProject(p, tagExe)
	}
	done := projectCount
	for done != 0 {
		inc := <-sem
		done = done - inc
	}
}
