package github

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/v57/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	DefaultBaseURL       = "https://api.github.com/"
	DefaultUploadBaseURL = "https://uploads.github.com/"
)

type GitHubConfiguration struct {
	context     context.Context
	logger      *zap.Logger
	accessToken string
	Org         string
	Repo        string
	CommitSha   string
	ApiUrl      string
	PrNumber    int
}

func New(context context.Context, log *zap.Logger, org, repo, commitSha, accessToken string, prNumber int, apiUrl string) *GitHubConfiguration {
	logger := log.With(zap.String("provider", "githubcomment"), zap.String("org", org), zap.String("repo", repo), zap.String("commitSha", commitSha))
	return &GitHubConfiguration{
		context:     context,
		logger:      logger,
		accessToken: accessToken,
		Org:         org,
		Repo:        repo,
		CommitSha:   commitSha,
		PrNumber:    prNumber,
		ApiUrl:      apiUrl,
	}
}

func (c *GitHubConfiguration) WriteCommitComment(body string, commentHeading string, removeDuplicateCommitComment bool) (*github.RepositoryComment, error) {
	client, err := c.NewClient()
	if err != nil {
		return nil, err
	}

	comment := &github.RepositoryComment{
		Body: &body,
	}

	if removeDuplicateCommitComment {
		c.logger.Info("RemoveDuplicateCommitComments was true, removing existing comments on the commit")
		r, _ := regexp.Compile(fmt.Sprintf("^%v", strings.Trim(commentHeading, " ")))
		err = c.removeCommitComments(client, r, c.CommitSha)
		if err != nil {
			c.logger.Error(fmt.Sprintf("failed to cleanup existing comments on commit '%v'", c.CommitSha), zap.Error(err))
		}
	} else {
		c.logger.Info("RemoveDuplicateCommitComments was false, skipping the removal of existing comments on the commit")
	}

	newComment, resp, err := client.Repositories.CreateComment(c.context, c.Org, c.Repo, c.CommitSha, comment)
	body, err = c.checkHttpResponse(newComment, resp, err)
	return newComment, err

}

func (c *GitHubConfiguration) NewClient() (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.accessToken},
	)
	tc := oauth2.NewClient(c.context, ts)
	client := github.NewClient(tc)
	var err error
	if c.ApiUrl != "" {
		client, err = client.WithEnterpriseURLs(c.ApiUrl, DefaultUploadBaseURL)
		if err != nil {
			return nil, err
		}
	}
	return client, err
}

func (c *GitHubConfiguration) WritePullRequestComment(body string) (*github.IssueComment, error) {
	client, err := c.NewClient()
	if err != nil {
		return nil, err
	}

	comment := &github.IssueComment{
		Body: &body,
	}

	newComment, resp, err := client.Issues.CreateComment(c.context, c.Org, c.Repo, c.PrNumber, comment)
	body, err = c.checkHttpResponse(newComment, resp, err)

	return newComment, err

}

func (c *GitHubConfiguration) checkHttpResponse(_ interface{}, resp *github.Response, err error) (string, error) {
	respBodyString := ""
	if err != nil {
		if resp != nil && resp.Body != nil {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			respBodyString = buf.String()
		}
		return respBodyString, fmt.Errorf("github api error. Response Body: %v", respBodyString)
	}

	if resp.StatusCode <= 199 || resp.StatusCode >= 400 {
		return respBodyString, fmt.Errorf("github API call failed. Status Code: %v. Response Body: %v", resp.StatusCode, respBodyString)
	}

	return respBodyString, nil
}

func (c *GitHubConfiguration) CleanExistingCommentsOnAllPullRequestCommits(commentHeading string) error {
	client, err := c.NewClient()
	if err != nil {
		return err
	}

	r, _ := regexp.Compile(fmt.Sprintf("^%v", strings.Trim(commentHeading, " ")))

	if c.PrNumber == -1 || c.PrNumber == 0 {
		comments, _, err := client.Repositories.ListCommitComments(c.context, c.Org, c.Repo, c.CommitSha, nil)

		if err != nil {
			return fmt.Errorf("an error occurred attempting to list the commits on comment '%v'. Error: %v", c.CommitSha, err)
		}
		for _, comment := range comments {
			c.logger.Info("found commit comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL))
			c.removeCommitComment(*comment.Body, *comment.ID, r, client)
		}

	} else {
		commits, _, err := client.PullRequests.ListCommits(c.context, c.Org, c.Repo, c.PrNumber, nil)

		if err != nil {
			return fmt.Errorf("an error occurred attempting to list commits from pull request '%v'. Error: %v", c.PrNumber, err)
		}
		for _, commit := range commits {
			sha := *commit.SHA
			err := c.removeCommitComments(client, r, sha)
			if err != nil {
				return fmt.Errorf("an error occurred attempting to remove comments on commit '%v'. Error: %v", sha, err)
			}
		}
	}
	return nil
}

func (c *GitHubConfiguration) CleanExistingCommentsOnPullRequest(commentHeading string) error {
	client, err := c.NewClient()
	if err != nil {
		return err
	}

	r, _ := regexp.Compile(fmt.Sprintf("^%v", strings.Trim(commentHeading, " ")))

	comments, _, _ := client.Issues.ListComments(c.context, c.Org, c.Repo, c.PrNumber, nil)

	if err != nil {
		return fmt.Errorf("an error occurred attempting to list the commits on pr '%v'. Error: %v", c.PrNumber, err)
	}
	for _, comment := range comments {
		c.logger.Info("found pull request comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL))
		body := *comment.Body
		if r.MatchString(body) {
			_, err := client.Issues.DeleteComment(c.context, c.Org, c.Repo, *comment.ID)

			if err != nil {
				c.logger.Error("failed to remove github pull request comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL), zap.Error(err))
			} else {
				c.logger.Info("removed github pull request comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL))
			}
		}
	}
	return nil
}

func (c *GitHubConfiguration) CleanExistingCommentsOnCommit(commentHeading string) error {
	client, err := c.NewClient()
	if err != nil {
		return err
	}

	r, _ := regexp.Compile(fmt.Sprintf("^%v", strings.Trim(commentHeading, " ")))

	err = c.removeCommitComments(client, r, c.CommitSha)
	if err != nil {
		return fmt.Errorf("an error occurred attempting to remove comments on commit '%v'. Error: %v", c.CommitSha, err)
	}

	return nil
}

func (c *GitHubConfiguration) removeCommitComments(client *github.Client, r *regexp.Regexp, sha string) error {
	comments, _, err := client.Repositories.ListCommitComments(c.context, c.Org, c.Repo, sha, nil)

	if err != nil {
		return fmt.Errorf("an error occurred attempting to list the comments on commit '%v'. Error: %v", c.CommitSha, err)
	}
	for _, comment := range comments {
		c.logger.Info("found commit comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL))
		c.removeCommitComment(*comment.Body, *comment.ID, r, client)
	}
	return nil
}

func (c *GitHubConfiguration) removeCommitComment(body string, commentId int64, r *regexp.Regexp, client *github.Client) {
	// body := *comment.Body
	if r.MatchString(body) {
		_, err := client.Repositories.DeleteComment(c.context, c.Org, c.Repo, commentId)
		if err != nil {
			c.logger.Error("failed to remove github commit comment", zap.String("commentId", fmt.Sprintf("%v", commentId)), zap.Error(err))
		} else {
			c.logger.Info("removed github commit comment", zap.String("commentId", fmt.Sprintf("%v", commentId)))
		}
	}
}
