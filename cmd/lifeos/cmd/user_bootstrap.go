package cmd

import (
	"context"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	settingsapp "github.com/valentinezhov/lifeos/internal/settings/app"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
)

type userBootstrap struct {
	settings *settingsapp.EnsureDefaults
	spheres  *spheresapp.EnsureDefaultSpheres
}

func (b *userBootstrap) Execute(ctx context.Context, userID ids.UserID) error {
	if err := b.settings.Execute(ctx, userID); err != nil {
		return err
	}
	return b.spheres.Execute(ctx, userID)
}
