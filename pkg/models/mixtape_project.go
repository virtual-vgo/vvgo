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
	Id      string   `json:"id"`
	Mixtape string   `json:"mixtape"`
	Name    string   `json:"Name"`
	Blurb   string   `json:"blurb"`
	Channel string   `json:"channel"`
	Hosts   []string `json:"hosts,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

const MixtapeProjectsRedisKey = "mixtape:projects"

func ListMixtapeProjects(ctx context.Context) ([]MixtapeProject, error) {
	var projectsJSON []string
	if err := redis.Do(ctx, redis.Cmd(&projectsJSON, "HVALS", MixtapeProjectsRedisKey)); err != nil {
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
	if len(projects) == 0 {
		logger.Infof("skipping empty write to %s", MixtapeProjectsRedisKey)
		return nil
	}

	redisArgs := []string{MixtapeProjectsRedisKey}
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
