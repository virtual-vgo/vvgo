package models

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/models/credit"
	"github.com/virtual-vgo/vvgo/pkg/models/leader"
	"github.com/virtual-vgo/vvgo/pkg/models/part"
	"github.com/virtual-vgo/vvgo/pkg/models/project"
)

type Config struct {
	SpreadsheetID string
}

var config Config

func Initialize(newConfig Config) {
	config = newConfig
}

func ListProjects(ctx context.Context, identity *login.Identity) (project.Projects, error) {
	return project.List(ctx, identity, config.SpreadsheetID)
}

func ListCurrentProjects(ctx context.Context, identity *login.Identity) (project.Projects, error) {
	projects, err := ListProjects(ctx, identity)
	if err != nil {
		return nil, err
	}
	return projects.Current(), nil
}

func ListParts(ctx context.Context, identity *login.Identity) (part.Parts, error) {
	return part.List(ctx, identity, config.SpreadsheetID)
}

func ListCurrentParts(ctx context.Context, identity *login.Identity) (part.Parts, error) {
	parts, err := ListParts(ctx, identity)
	if err != nil {
		return nil, err
	}
	return parts.Current(), nil
}

func ListCredits(ctx context.Context) (credit.Credits, error) {
	return credit.List(ctx, config.SpreadsheetID)
}

func ListLeaders(ctx context.Context) ([]leader.Leader, error) {
	return leader.List(ctx, config.SpreadsheetID)
}
