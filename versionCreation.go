package VersionNoFromGitlabBuilds

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
)

// GitLabAPIConnection collects all data needed for a connection to GitLab's REST api.
type GitLabAPIConnection struct {
    baseURL string
    token   string
    client  *http.Client
}

type commitID string
type setOfCommits map[commitID]struct{}

const apiURL = "/api/v3"

var reNextLink = regexp.MustCompile(`<([^<>]+)>; rel="next"`)

// NewGitLabAPIConnection creates a connection to a GitLab server such that its REST api functions can be used.
func NewGitLabAPIConnection(gitlabBaseURL, privateToken string) *GitLabAPIConnection {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &GitLabAPIConnection{
		baseURL: gitlabBaseURL,
		token:   privateToken,
		client:  &http.Client{Transport: tr},
	}
}

// NextVersionNo looks at all builds on the connected server associated with the given project and creates a new build number from that information.
// This results in automatically increasing build numbers per project.
func (c *GitLabAPIConnection) NextVersionNo(projectName string) (int, error) {
	pid, err := c.projectIDFromName(projectName)

	if err != nil {
		return -1, err
	}

	ci, err := c.commitsPerBuild(pid)

	return len(ci) + 1, err
}

func (c *GitLabAPIConnection) getRequest(endPoint string) (*http.Request, error) {
	req, err := http.NewRequest("GET", c.baseURL + apiURL +endPoint, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)

	return req, nil
}

func (c *GitLabAPIConnection) getAbsoluteRequest(endPoint string) (*http.Request, error) {
	req, err := http.NewRequest("GET", endPoint, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)

	return req, nil
}

func (c *GitLabAPIConnection) projectIDFromName(projectName string) (int, error) {
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

func (c *GitLabAPIConnection) commitsPerBuildWithReq(req *http.Request, s setOfCommits) error {
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

	allLinks := reNextLink.FindAllStringSubmatch(res.Header.Get("link"), -1)

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

func (c *GitLabAPIConnection) commitsPerBuild(projectID int) (setOfCommits, error) {
	var s = make(setOfCommits)

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
