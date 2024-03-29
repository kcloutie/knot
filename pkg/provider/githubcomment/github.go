package githubcomment

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kcloutie/knot/pkg/config"
	"github.com/kcloutie/knot/pkg/github"
	"github.com/kcloutie/knot/pkg/message"
	"github.com/kcloutie/knot/pkg/provider"
	"github.com/kcloutie/knot/pkg/template"
	"go.uber.org/zap"
)

var _ provider.ProviderInterface = (*Provider)(nil)

type Provider struct {
	Log          *zap.Logger
	providerName string
	notification config.Notification
}

type ProviderConfig struct {
	// Terraform plan task name, set the value when using terraform and additional data will be displayed in the github comment
	PlanTaskName string `json:"planTaskName,omitempty"`

	// True or false whether existing pipeline comments on all PR commits will be removed.
	// A pull request can contain multiple commits. Depending on how the pull request is pushed up, a pipeline may execute on each commit
	// and in turn each commit will contain a comment. When this property is true, every comment on every commit related to the pull request
	// will be removed. The default should be set to false in order to keep the pipelineRun history of each pipeline execution
	RemoveExistingCommentsFromAllPullRequestCommits bool `json:"removeExistingCommentsFromAllPullRequestCommits,omitempty"`

	// True or false whether existing pipeline comments on PR should be removed.
	// Each time the pipeline executes it will write a new issue comment to the pull request of the results of the pipelineRun
	// When this is set to true, all existing issue comments created by the pipeline will be removed. This is to make it easier
	// for people reviewing to find the results of the latest pipelineRun.
	// The default for this should be set to true so that the pull request only contains the latest pipelineRun result comment.
	// NOTE: The same comment is written to the latest commit and can be viewed by looking at the commit
	RemoveExistingPullRequestComments bool `json:"removeExistingPullRequestComments,omitempty"`

	// True or false to remove duplicate pipeline comments on a single (latest) commit
	// Each time the pipeline executes, this operator will write a comment on the latest commit. When the pipeline executes for a second time
	// on the same commit, a second comment will be written to that commit.
	// When this is set to true, any existing comments (created by this operator) on the latest commit will be removed keeping only the latest comment
	RemoveDuplicateCommitComments bool `json:"removeDuplicateCommitComments,omitempty"`
}

func New() *Provider {
	return &Provider{
		providerName: "github/comment",
	}
}

func (v *Provider) SetLogger(logger *zap.Logger) {
	v.Log = logger
}
func (v *Provider) GetName() string {
	return v.providerName
}

func (v *Provider) GetDescription() string {
	return ""
}

func (v *Provider) SetNotification(notification config.Notification) {
	v.notification = notification
}

func (v *Provider) SendNotification(ctx context.Context, data *message.NotificationData) error {
	v.Log = v.Log.With(zap.String("provider", v.providerName))
	_, err := provider.HasRequiredProperties(v.notification.Properties, v.GetRequiredPropertyNames())
	if err != nil {
		return err
	}

	templateConfig := template.NewRenderTemplateOptions()
	provider.SetGoTemplateOptionValues(ctx, v.Log, &templateConfig, v.notification.Properties)

	ghConfig, err := v.GetServiceConfig(ctx, data, v.Log)
	if err != nil {
		return err
	}

	v.Log = v.Log.With(zap.String("org", ghConfig.Org), zap.String("repo", ghConfig.Repo), zap.String("commitSha", ghConfig.CommitSha), zap.Int("pr", ghConfig.PrNumber), zap.String("enterpriseUrl", ghConfig.EnterpriseUrl))

	providerConfig, err := v.GetProviderConfig(ctx, data, v.Log)
	if err != nil {
		return err
	}

	heading, err := v.notification.Properties["heading"].GetValue(ctx, v.Log, data)
	if err != nil {
		return err
	}
	renderedHeading, err := template.RenderTemplateValues(ctx, heading, fmt.Sprintf("%s_%s/heading", data.ID, v.providerName), data.AsMap(), []string{}, templateConfig)
	if err != nil {
		return err
	}

	body, err := v.notification.Properties["body"].GetValue(ctx, v.Log, data)
	if err != nil {
		return err
	}
	renderedBody, err := template.RenderTemplateValues(ctx, body, fmt.Sprintf("%s_%s/body", data.ID, v.providerName), data.AsMap(), []string{}, templateConfig)
	if err != nil {
		return err
	}

	if providerConfig.RemoveExistingCommentsFromAllPullRequestCommits {
		v.Log.Info("Cleaning up existing commit comments")
		ghConfig.CleanExistingCommentsOnAllPullRequestCommits(string(renderedHeading))
		// v.log.Info("Finished cleaning up existing comments on all commits of the pull request")
	} else {
		v.Log.Info("RemoveExistingCommentsFromAllPullRequestCommits was set to false, skipping the deletion of existing comments")
	}

	if ghConfig.PrNumber > 0 {
		if providerConfig.RemoveExistingPullRequestComments {
			v.Log.Info("Cleaning up existing pull request comments")
			ghConfig.CleanExistingCommentsOnPullRequest(string(renderedHeading))
			// v.log.Info("Finished cleaning up existing comments on the pull request")
		} else {
			v.Log.Info("RemoveExistingPullRequestComments was set to false, skipping the deletion of existing comments")
		}
	} else {
		v.Log.Info("Pull request number was not greater than 0, skipping the deletion of existing comments")
	}

	// Would normally generate the comment body here, but the body is not generated using templates
	v.Log.Info("Creating commit comment")
	newComment, err := ghConfig.WriteCommitComment(string(renderedBody), string(renderedHeading), providerConfig.RemoveDuplicateCommitComments)
	if err != nil {
		// v.log.Error("failed to write the github commit comment", zap.Error(err))
		return fmt.Errorf("unable to write github commit comment. Error: %v", err)
	}

	v.Log = v.Log.With(zap.String("commitCommentUrl", newComment.GetHTMLURL()))
	v.Log.Info("github commit comment has been created")

	if ghConfig.PrNumber > 0 {
		v.Log.Info("Creating pull request comment")
		newComment, err := ghConfig.WritePullRequestComment(string(renderedBody))
		if err != nil {
			// return githubToken, fmt.Errorf("unable to write github pull request comment. Error: %v", err)
			return fmt.Errorf("unable to write github pull request comment. Error: %v", err)
		}
		// r.EventEmitter.EmitMessage(ctx, &notification, zap.InfoLevel, "GithubComment", fmt.Sprintf("github pull request comment has been created here %s", *newComment.HTMLURL))
		v.Log = v.Log.With(zap.String("PrCommentUrl", newComment.GetHTMLURL()))
		v.Log.Info("github pull request comment has been created")
	} else {
		v.Log.Info("Pull request number was not greater than 0, skipping the creation of the pull request comment")
	}

	return nil
}

func (v *Provider) GetProviderConfig(ctx context.Context, data *message.NotificationData, log *zap.Logger) (*ProviderConfig, error) {
	planTaskName, err := v.notification.Properties["planTaskName"].GetValue(ctx, v.Log, data)
	if err != nil {
		planTaskName = ""
	}

	removeExistingCommentsFromAllPullRequestCommits := false
	removeExistingCommentsFromAllPullRequestCommitsStr, err := v.notification.Properties["removeExistingCommentsFromAllPullRequestCommits"].GetValue(ctx, v.Log, data)
	if err == nil && removeExistingCommentsFromAllPullRequestCommitsStr != "" {
		removeExistingCommentsFromAllPullRequestCommits, err = strconv.ParseBool(removeExistingCommentsFromAllPullRequestCommitsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied removeExistingCommentsFromAllPullRequestCommits '%v' to a boolean. Error: %v", removeExistingCommentsFromAllPullRequestCommitsStr, err)
		}
	}

	removeExistingPullRequestComments := true
	removeExistingPullRequestCommentsStr, err := v.notification.Properties["removeExistingPullRequestComments"].GetValue(ctx, v.Log, data)
	if err == nil && removeExistingPullRequestCommentsStr != "" {
		removeExistingPullRequestComments, err = strconv.ParseBool(removeExistingPullRequestCommentsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied removeExistingPullRequestComments '%v' to a boolean. Error: %v", removeExistingPullRequestCommentsStr, err)
		}
	}

	removeDuplicateCommitComments := true
	removeDuplicateCommitCommentsStr, err := v.notification.Properties["removeDuplicateCommitComments"].GetValue(ctx, v.Log, data)
	if err == nil && removeDuplicateCommitCommentsStr != "" {
		removeDuplicateCommitComments, err = strconv.ParseBool(removeDuplicateCommitCommentsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied removeDuplicateCommitComments '%v' to a boolean. Error: %v", removeDuplicateCommitCommentsStr, err)
		}
	}

	config := ProviderConfig{
		PlanTaskName: planTaskName,
		RemoveExistingCommentsFromAllPullRequestCommits: removeExistingCommentsFromAllPullRequestCommits,
		RemoveExistingPullRequestComments:               removeExistingPullRequestComments,
		RemoveDuplicateCommitComments:                   removeDuplicateCommitComments,
	}
	return &config, nil
}

func (v *Provider) GetServiceConfig(ctx context.Context, data *message.NotificationData, log *zap.Logger) (*github.GitHubConfiguration, error) {
	token, err := v.notification.Properties["token"].GetValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("the github token property was not supplied or was empty")
	}

	org, err := v.notification.Properties["org"].GetValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if org == "" {
		return nil, fmt.Errorf("the github org property was not supplied or was empty")
	}

	repo, err := v.notification.Properties["repo"].GetValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if repo == "" {
		return nil, fmt.Errorf("the github repo property was not supplied or was empty")
	}

	commitSha, err := v.notification.Properties["commitSha"].GetValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if commitSha == "" {
		return nil, fmt.Errorf("the github commitSha property was not supplied or was empty")
	}

	prNumberStr, err := v.notification.Properties["prNumber"].GetValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if prNumberStr == "" {
		return nil, fmt.Errorf("the github prNumber property was not supplied or was empty")
	}

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {

		return nil, fmt.Errorf("failed to convert the supplied pr number '%v' to an integer. Error: %v", prNumberStr, err)
	}

	isEnterprise := false
	enterpriseUrl, err := v.notification.Properties["enterpriseUrl"].GetValue(ctx, v.Log, data)
	if err != nil {
		enterpriseUrl = github.DefaultBaseURL
	} else {
		isEnterprise = true
	}

	ghConfig := github.New(ctx, log, org, repo, commitSha, token, prNumber, enterpriseUrl, isEnterprise)
	return ghConfig, nil
}

func (v *Provider) GetHelp() string {
	return ""
}

func (v *Provider) GetProperties() []config.NotificationProperty {
	return []config.NotificationProperty{
		{
			Name:        "heading",
			Description: "The heading of the comment. This field supports go templating. The heading is also used to find previous comments to remove",
			Required:    config.AsBoolPointer(true),
		},
		{
			Name:        "body",
			Description: "The body of the comment. This field supports go templating",
			Required:    config.AsBoolPointer(true),
		},
		{
			Name:        "token",
			Description: "The github token to use for authentication. This token should have the necessary permissions to write comments to the repository",
			Required:    config.AsBoolPointer(true),
		},
		{
			Name:        "org",
			Description: "The github organization",
			Required:    config.AsBoolPointer(true),
		},
		{
			Name:        "repo",
			Description: "The github repository",
			Required:    config.AsBoolPointer(true),
		},
		{
			Name:        "commitSha",
			Description: "The commit sha where the comment will be written",
			Required:    config.AsBoolPointer(true),
		},
		{
			Name:        "prNumber",
			Description: "The pull request number where the comment will be written. If the value is less than 0, the comment will not be written to the pull request",
			Required:    config.AsBoolPointer(true),
		},
	}
}

func (v *Provider) GetRequiredPropertyNames() []string {
	return provider.GetRequiredPropertyNames(v)
}
