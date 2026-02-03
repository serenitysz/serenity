package version

import (
	"context"
	"fmt"
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
		return exception.InternalError("we could not found the latest version (%w)", err)
	}

	if !found || latest.LessOrEqual(Version) {
		fmt.Printf("You'are already running in the latest version of Serenity (%s)\n", render.Paint(Version, render.Gray, noColor))

		return nil
	}

	exe, err := os.Executable()

	if err != nil {
		return exception.InternalError("failed to locate executable (%w)", err)
	}

	fmt.Printf("Updating Serenity from %s to %s...\n", render.Paint(Version, render.Gray, noColor), render.Paint(latest.Version(), render.Gray, noColor))

	if err := selfupdate.UpdateTo(ctx, latest.AssetURL, latest.AssetName, exe); err != nil {
		return exception.InternalError("update failed (%w)", err)
	}

	fmt.Printf("You'are now on the %s version of Serenity!\n", render.Paint(latest.Version(), render.Gray, noColor))

	return nil
}
