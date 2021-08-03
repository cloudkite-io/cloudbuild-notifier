package cloudbuild

import (
	"context"
	"fmt"
	"github.com/fatih/structs"
	"google.golang.org/api/cloudbuild/v1"
	"log"
)

type CloudbuildClient struct {
	ProjectID string
	client    *cloudbuild.ProjectsBuildsService
}

func New(projectId string) (*CloudbuildClient, error) {
	ctx := context.Background()
	cloudbuildService, err := cloudbuild.NewService(ctx)

	if err != nil {
		return &CloudbuildClient{}, fmt.Errorf("Error creating cloudbuild client: %s", err)
	}
	return &CloudbuildClient{
		ProjectID: projectId,
		client:    cloudbuild.NewProjectsBuildsService(cloudbuildService),
	}, nil
}

type BuildParameters struct {
	Id            string
	REPO_NAME     string
	BRANCH_NAME   string
	COMMIT_SHA    string
	REVISION_ID   string
	TRIGGER_NAME  string
	HEAD_REPO_URL string
}

func (c *CloudbuildClient) GetBuildParameters(buildId string) (BuildParameters, error) {
	buildParams := &BuildParameters{
		Id: buildId,
	}
	b := structs.New(buildParams)
	result, err := c.client.Get(c.ProjectID, buildId).Do()
	if err != nil {
		log.Printf("error getting build parameters: %s\n", err)
		return *buildParams, err
	}

	for k, v := range result.Substitutions {
		if f, ok := b.FieldOk(k); ok {
			f.Set(v)
		}
	}
	return *buildParams, nil
}
