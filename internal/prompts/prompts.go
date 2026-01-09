package prompts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/serenitysz/serenity/internal/render"
)

var reader = bufio.NewReader(os.Stdin)

func Input(label, def string) (string, error) {
	fmt.Printf(
		"%s %s (%s)\n",
		render.Paint("?", render.Purple),
		render.Paint(label, render.Bold),
		def,
	)

	input, err := reader.ReadString('\n')

	if err != nil {
		return "", err
	}

	value := strings.TrimSpace(input)

	if value == "" {
		return def, nil
	}

	return value, nil
}

func Confirm(label string) (bool, error) {
	fmt.Printf(
		"%s %s %s\n",
		render.Paint("?", render.Purple),
		render.Paint(label, render.Bold),
		render.Paint("(y/N)", render.Gray),
	)

	fmt.Printf("%s ", render.Paint(">", render.Blue))

	for {
		input, err := reader.ReadString('\n')

		if err != nil {
			return false, err
		}

		value := strings.ToLower(strings.TrimSpace(input))

		switch value {
		case "y", "yes":
			return true, nil
		case "", "n", "no":
			return false, nil
		default:
			fmt.Print(render.Paint("Please answer y or n: ", render.Red))
		}
	}
}
