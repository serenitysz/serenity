package prompts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/serenitysz/serenity/internal/exception"
	"github.com/serenitysz/serenity/internal/render"
)

var reader = bufio.NewReader(os.Stdin)

func Input(label, def string, noColor bool) (string, error) {
	fmt.Printf(
		"%s %s (%s)\n",
		render.Paint("?", render.Purple, noColor),
		render.Paint(label, render.Bold, noColor),
		def,
	)

	input, err := reader.ReadString('\n')

	if err != nil {
		return "", exception.InternalError("%v", err)
	}

	value := strings.TrimSpace(input)

	if value == "" {
		return def, nil
	}

	return value, nil
}

func Confirm(label string, noColor bool) (bool, error) {
	fmt.Printf(
		"%s %s %s\n",
		render.Paint("?", render.Purple, noColor),
		render.Paint(label, render.Bold, noColor),
		render.Paint("(y/N)", render.Gray, noColor),
	)

	fmt.Printf("%s ", render.Paint(">", render.Blue, noColor))

	for {
		input, err := reader.ReadString('\n')

		if err != nil {
			return false, exception.InternalError("%v", err)
		}

		value := strings.ToLower(strings.TrimSpace(input))

		switch value {
		case "y", "yes":
			return true, nil
		case "", "n", "no":
			return false, nil
		default:
			fmt.Print(render.Paint("Please answer y or n: ", render.Red, noColor))
		}
	}
}
