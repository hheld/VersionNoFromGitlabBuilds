package main

import (
	"fmt"
	"github.com/hheld/VersionNoFromGitlabBuilds"
)

func main() {
	conn := VersionNoFromGitlabBuilds.NewGitLabApiConnection("https://ubuntults", "d2EbZhyk9g9JyLJPX_ys")
	no, err := conn.NextVersionNo("RunnerTest")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Next version: %d\n", no)
}
