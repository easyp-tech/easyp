package prompter

import "context"

// Prompter provides an interface for interactive user interaction.
type Prompter interface {
	// Confirm asks a yes/no question. Returns true if the user answered "yes".
	Confirm(ctx context.Context, message string, defaultValue bool) (bool, error)

	// Select offers to choose one option from the list. Returns the index of the selected option.
	Select(ctx context.Context, message string, options []string, defaultIndex int) (int, error)

	// MultiSelect offers to choose multiple options. Returns the indices of the selected options.
	MultiSelect(ctx context.Context, message string, options []string, defaults []bool) ([]int, error)

	// Input requests text input from the user.
	Input(ctx context.Context, message string, defaultValue string) (string, error)
}
