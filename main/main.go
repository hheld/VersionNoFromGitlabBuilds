package main

import (
	"fmt"
	"github.com/hheld/VersionNoFromGitlabBuilds"
)

func main() {
	conn := VersionNoFromGitlabBuilds.NewGitLabAPIConnection("https://ubuntults", "d2EbZhyk9g9JyLJPX_ys")
	no, err := conn.NextVersionNo("RunnerTest")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Next version: %d\n", no)

	err = conn.CreateTag("RunnerTest", "bcd3098b54bcdb5b864f6299a80890a3740bafdb", "Test3")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
