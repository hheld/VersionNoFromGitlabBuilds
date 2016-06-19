package VersionNoFromGitlabBuilds

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
)

type GitLabApiConnection struct {
	baseUrl string
	token   string
	client  *http.Client
}

type commitID string
type setOfCommits map[commitID]struct{}

const apiUrl = "/api/v3"

var re_nextLink = regexp.MustCompile(`<([^<>]+)>; rel="next"`)

func NewGitLabApiConnection(gitlabBaseUrl, privateToken string) *GitLabApiConnection {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &GitLabApiConnection{
		baseUrl: gitlabBaseUrl,
		token:   privateToken,
		client:  &http.Client{Transport: tr},
	}
}

func (c *GitLabApiConnection) NextVersionNo(projectName string) (int, error) {
	pid, err := c.projectIDFromName(projectName)

	if err != nil {
		return -1, err
	}

	ci, err := c.commitsPerBuild(pid)

	return len(ci) + 1, err
}

func (c *GitLabApiConnection) getRequest(endPoint string) (*http.Request, error) {
	req, err := http.NewRequest("GET", c.baseUrl+apiUrl+endPoint, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)

	return req, nil
}

func (c *GitLabApiConnection) getAbsoluteRequest(endPoint string) (*http.Request, error) {
	req, err := http.NewRequest("GET", endPoint, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)

	return req, nil
}

func (c *GitLabApiConnection) projectIDFromName(projectName string) (int, error) {
	req, err := c.getRequest("/projects")

	if err != nil {
		return -1, err
	}

	res, err := c.client.Do(req)

	if err != nil {
		return -1, err
	}

	defer res.Body.Close()

	var projectInfo interface{}

	err = json.NewDecoder(res.Body).Decode(&projectInfo)

	if err != nil {
		return -1, err
	}

	projects := projectInfo.([]interface{})

	for _, project := range projects {
		m := project.(map[string]interface{})

		for k, v := range m {
			if k == "name" {
				if v.(string) == projectName {
					return int(m["id"].(float64)), nil
				}
			}
		}
	}

	return -1, errors.New("No project named '" + projectName + "' found")
}

func (c *GitLabApiConnection) commitsPerBuildWithReq(req *http.Request, s setOfCommits) error {
	res, err := c.client.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	var buildInfo interface{}

	err = json.NewDecoder(res.Body).Decode(&buildInfo)

	if err != nil {
		return err
	}

	var builds []interface{}

	switch buildInfo.(type) {
	case []interface{}:
		builds = buildInfo.([]interface{})
	case map[string]interface{}:
		return errors.New(buildInfo.(map[string]interface{})["message"].(string))
	}

	allLinks := re_nextLink.FindAllStringSubmatch(res.Header.Get("link"), -1)

	if allLinks != nil {
		nextReq, err := c.getAbsoluteRequest(allLinks[0][1])

		if err != nil {
			return err
		}

		c.commitsPerBuildWithReq(nextReq, s)
	}

	for _, build := range builds {
		m := build.(map[string]interface{})

		for k, v := range m {
			if k == "commit" {
				c := v.(map[string]interface{})

				ciID := commitID(c["id"].(string))
				s[ciID] = struct{}{}
			}
		}
	}

	return nil
}

func (c *GitLabApiConnection) commitsPerBuild(projectID int) (setOfCommits, error) {
	var s setOfCommits = make(setOfCommits)

	req, err := c.getRequest("/projects/" + strconv.Itoa(projectID) + "/builds")

	if err != nil {
		return nil, err
	}

	err = c.commitsPerBuildWithReq(req, s)

	if err != nil {
		return nil, err
	}

	return s, nil
}
