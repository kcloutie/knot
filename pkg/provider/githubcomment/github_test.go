package githubcomment

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/kcloutie/knot/pkg/config"
	"github.com/kcloutie/knot/pkg/github"
	"github.com/kcloutie/knot/pkg/message"
	"go.uber.org/zap/zaptest"
)

func TestProvider_GetGithubConfig(t *testing.T) {
	type args struct {
		data         *message.NotificationData
		notification config.Notification
	}
	tests := []struct {
		name    string
		v       *Provider
		args    args
		want    *github.GitHubConfiguration
		wantErr bool
	}{
		{
			name: "All properties are valid",
			v:    New(),
			args: args{
				data: &message.NotificationData{
					Data: map[string]interface{}{
						"token":     "test token",
						"org":       "test org",
						"repo":      "test repo",
						"commitSha": "test commitSha",
						"prNumber":  "123",
						"apiUrl":    "test apiUrl",
					},
				},
				notification: config.Notification{
					Properties: map[string]config.PropertyAndValue{
						"token":     {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.token"}}},
						"org":       {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.org"}}},
						"repo":      {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.repo"}}},
						"commitSha": {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.commitSha"}}},
						"prNumber":  {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.prNumber"}}},
						"apiUrl":    {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.apiUrl"}}},
					},
				},
			},
			want: github.New(context.Background(), zaptest.NewLogger(t), "test org", "test repo", "test commitSha", "test token", 123, "test apiUrl"),
		},
		{
			name: "prNumber is not a valid integer",
			v:    New(),
			args: args{
				data: &message.NotificationData{
					Data: map[string]interface{}{
						"token":     "test token",
						"org":       "test org",
						"repo":      "test repo",
						"commitSha": "test commitSha",
						"prNumber":  "invalid",
						"apiUrl":    "test apiUrl",
					},
				},
				notification: config.Notification{
					Properties: map[string]config.PropertyAndValue{
						"token":     {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.token"}}},
						"org":       {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.org"}}},
						"repo":      {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.repo"}}},
						"commitSha": {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.commitSha"}}},
						"prNumber":  {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.prNumber"}}},
						"apiUrl":    {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.apiUrl"}}},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing required property",
			v:    New(),
			args: args{
				data: &message.NotificationData{
					Data: map[string]interface{}{
						"token":     "test token",
						"org":       "test org",
						"repo":      "test repo",
						"commitSha": "test commitSha",
						"prNumber":  "123",
						"apiUrl":    "test apiUrl",
					},
				},
				notification: config.Notification{
					Properties: map[string]config.PropertyAndValue{
						"org":       {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.org"}}},
						"repo":      {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.repo"}}},
						"commitSha": {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.commitSha"}}},
						"prNumber":  {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.prNumber"}}},
						"apiUrl":    {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.apiUrl"}}},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger := zaptest.NewLogger(t)
			ctx := context.Background()
			tt.v.SetLogger(testLogger)
			tt.v.SetNotification(tt.args.notification)

			got, err := tt.v.GetGithubConfig(ctx, tt.args.data, testLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.GetGithubConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got.ApiUrl, tt.want.ApiUrl) {
				t.Errorf("Provider.GetGithubConfig() ApiUrl = %v, want %v", got.ApiUrl, tt.want.ApiUrl)
			}
			if !reflect.DeepEqual(got.CommitSha, tt.want.CommitSha) {
				t.Errorf("Provider.GetGithubConfig() CommitSha = %v, want %v", got.CommitSha, tt.want.CommitSha)
			}
			if !reflect.DeepEqual(got.Org, tt.want.Org) {
				t.Errorf("Provider.GetGithubConfig() Org = %v, want %v", got.Org, tt.want.Org)
			}
			if !reflect.DeepEqual(got.PrNumber, tt.want.PrNumber) {
				t.Errorf("Provider.GetGithubConfig() PrNumber = %v, want %v", got.PrNumber, tt.want.PrNumber)
			}
			if !reflect.DeepEqual(got.Repo, tt.want.Repo) {
				t.Errorf("Provider.GetGithubConfig() Repo = %v, want %v", got.Repo, tt.want.Repo)
			}
		})
	}
}

func TestProvider_GetProviderConfig(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	tests := []struct {
		name    string
		props   map[string]config.PropertyAndValue
		want    *ProviderConfig
		wantErr bool
	}{
		{
			name: "All properties are valid",
			props: map[string]config.PropertyAndValue{
				"planTaskName": {Value: "test planTaskName"},
				"removeExistingCommentsFromAllPullRequestCommits": {Value: "true"},
				"removeExistingPullRequestComments":               {Value: "false"},
				"removeDuplicateCommitComments":                   {Value: "true"},
			},
			want: &ProviderConfig{
				PlanTaskName: "test planTaskName",
				RemoveExistingCommentsFromAllPullRequestCommits: true,
				RemoveExistingPullRequestComments:               false,
				RemoveDuplicateCommitComments:                   true,
			},
			wantErr: false,
		},
		{
			name: "Invalid boolean value",
			props: map[string]config.PropertyAndValue{
				"planTaskName": {Value: "test planTaskName"},
				"removeExistingCommentsFromAllPullRequestCommits": {Value: "invalid"},
				"removeExistingPullRequestComments":               {Value: "true"},
				"removeDuplicateCommitComments":                   {Value: "true"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "removeExistingPullRequestComments not supplied",
			props: map[string]config.PropertyAndValue{
				"planTaskName": {Value: "test planTaskName"},
				"removeExistingCommentsFromAllPullRequestCommits": {Value: "true"},
				"removeDuplicateCommitComments":                   {Value: "true"},
			},
			want: &ProviderConfig{
				PlanTaskName: "test planTaskName",
				RemoveExistingCommentsFromAllPullRequestCommits: true,
				RemoveExistingPullRequestComments:               true,
				RemoveDuplicateCommitComments:                   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				notification: config.Notification{
					Properties: tt.props,
				},
			}
			got, err := p.GetProviderConfig(context.Background(), &message.NotificationData{}, testLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.GetProviderConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.GetProviderConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_SendNotification(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		pp := req.URL.Path
		cc := req.Method
		fmt.Println(pp)
		fmt.Println(cc)

		if req.Header["Authorization"][0] != fmt.Sprintf("Bearer %v", "token") {
			t.Fatalf("Invalid token: '%v'", req.Header["Authorization"][0])
		}

		if req.URL.Path == "/api/v3/repos/org/repo/commits/sha/comments" {
			if req.Method == "GET" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
			if req.Method == "POST" {
				rw.Write([]byte(`{"id":1,"html_url":"url","body":"heading\nbody"}`))
				return
			}
		}
		if req.URL.Path == "/api/v3/repos/org/repo/comments/1" {
			if req.Method == "DELETE" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
		}

		if req.URL.Path == "/api/v3/repos/org/repo/issues/1/comments" {
			if req.Method == "GET" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
			if req.Method == "POST" {
				rw.Write([]byte(`{"id":1,"html_url":"url","body":"heading\nbody"}`))
				return
			}
		}

		if req.URL.Path == "/api/v3/repos/org/repo/issues/comments/1" {
			if req.Method == "DELETE" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
		}

		if req.URL.Path == "/api/v3/repos/org/repo/pulls/1/commits" {
			if req.Method == "GET" {
				rw.Write([]byte(`[{"sha":"1","html_url":"url"}]`))
				return
			}
		}

		if req.URL.Path == "/api/v3/repos/org/repo/commits/1/comments" {
			if req.Method == "GET" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
		}
		fmt.Printf("Unknown path: %v, %s\n", req.URL.Path, req.Method)
	}))
	defer server.Close()
	type args struct {
		data         *message.NotificationData
		notification config.Notification
	}
	tests := []struct {
		name    string
		v       *Provider
		args    args
		wantErr bool
	}{
		{
			name: "Valid properties non PR",
			v:    New(),
			args: args{
				data: &message.NotificationData{
					Data: map[string]interface{}{},
				},
				notification: config.Notification{
					Properties: map[string]config.PropertyAndValue{
						"token":     {Value: "token"},
						"org":       {Value: "org"},
						"repo":      {Value: "repo"},
						"commitSha": {Value: "sha"},
						"prNumber":  {Value: "-1"},
						"apiUrl":    {Value: server.URL},
						"heading":   {Value: "heading"},
						"body":      {Value: "body"},
					},
				},
			},
		},
		{
			name: "Valid properties PR",
			v:    New(),
			args: args{
				data: &message.NotificationData{
					Data: map[string]interface{}{},
				},
				notification: config.Notification{
					Properties: map[string]config.PropertyAndValue{
						"token":     {Value: "token"},
						"org":       {Value: "org"},
						"repo":      {Value: "repo"},
						"commitSha": {Value: "sha"},
						"prNumber":  {Value: "1"},
						"apiUrl":    {Value: server.URL},
						"heading":   {Value: "heading"},
						"body":      {Value: "body"},
					},
				},
			},
		},
		{
			name: "Valid properties PR, removeExistingCommentsFromAllPullRequestCommits true",
			v:    New(),
			args: args{
				data: &message.NotificationData{
					Data: map[string]interface{}{},
				},
				notification: config.Notification{
					Properties: map[string]config.PropertyAndValue{
						"token":     {Value: "token"},
						"org":       {Value: "org"},
						"repo":      {Value: "repo"},
						"commitSha": {Value: "sha"},
						"prNumber":  {Value: "1"},
						"apiUrl":    {Value: server.URL},
						"heading":   {Value: "heading"},
						"body":      {Value: "body"},
						"removeExistingCommentsFromAllPullRequestCommits": {Value: "true"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger := zaptest.NewLogger(t)
			ctx := context.Background()
			tt.v.SetLogger(testLogger)
			tt.v.SetNotification(tt.args.notification)

			err := tt.v.SendNotification(ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.SendNotification() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
