package models

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"strings"
)

type MixtapeProject struct {
	Id      string
	Mixtape string
	Name    string
	Blurb   string
	Owners  []string
	Links   []string
	Tags    []string
}

func ListMixtapeProjects(ctx context.Context) ([]MixtapeProject, error) {
	var projectsJSON []string
	if err := redis.Do(ctx, redis.Cmd(&projectsJSON, "HVALS", "mixtape_projects")); err != nil {
		return nil, errors.RedisFailure(err)
	}

	projects := make([]MixtapeProject, 0, len(projectsJSON))
	for _, projectJSON := range projectsJSON {
		var project MixtapeProject
		if err := json.NewDecoder(strings.NewReader(projectJSON)).Decode(&project); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			continue
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func WriteMixtapeProjects(ctx context.Context, projects []MixtapeProject) error {
	redisArgs := []string{"mixtape_projects"}
	for _, project := range projects {
		projectJSON, err := json.Marshal(project)
		if err != nil {
			return errors.JsonEncodeFailure(err)
		}
		redisArgs = append(redisArgs, project.Id, string(projectJSON))
	}

	if err := redis.Do(ctx, redis.Cmd(nil, "HSET", redisArgs...)); err != nil {
		return errors.RedisFailure(err)
	}
	return nil
}
