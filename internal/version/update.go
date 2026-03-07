package version

import (
	"context"
	"os"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/render"
)

const SLUG = "serenitysz/serenity"

func Update(noColor bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel()

	latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(SLUG))

	if err != nil {
		return exception.InternalError("could not check for the latest Serenity release: %w", err)
	}

	if !found || latest.LessOrEqual(Version) {
		render.Infof("Serenity is already up to date (%s)", render.Paint(Version, render.Gray, noColor))

		return nil
	}

	exe, err := os.Executable()

	if err != nil {
		return exception.InternalError("could not locate the current Serenity executable: %w", err)
	}

	render.Infof("updating Serenity from %s to %s", render.Paint(Version, render.Gray, noColor), render.Paint(latest.Version(), render.Gray, noColor))

	if err := selfupdate.UpdateTo(ctx, latest.AssetURL, latest.AssetName, exe); err != nil {
		return exception.InternalError("could not replace the current Serenity binary: %w", err)
	}

	render.Successf("updated Serenity to %s", render.Paint(latest.Version(), render.Gray, noColor))

	return nil
}
