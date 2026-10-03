package prompter

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInteractivePrompterLineInput(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	prompt := InteractivePrompter{In: strings.NewReader("maybe\nyes\ncustom\n"), Out: &out}

	confirmed, err := prompt.Confirm(t.Context(), "Continue?", false)
	require.NoError(t, err)
	assert.True(t, confirmed)
	value, err := prompt.Input(t.Context(), "Module", "fallback")
	require.NoError(t, err)
	assert.Equal(t, "custom", value)
	assert.Contains(t, out.String(), "Please answer yes or no.")
	assert.NotContains(t, out.String(), "\x1b")
}

func TestInteractivePrompterCancellationAndEOF(t *testing.T) {
	t.Parallel()

	_, err := (InteractivePrompter{In: strings.NewReader("q\n"), Out: &bytes.Buffer{}}).Confirm(t.Context(), "Continue?", true)
	require.ErrorIs(t, err, context.Canceled)

	_, err = (InteractivePrompter{In: strings.NewReader(""), Out: &bytes.Buffer{}}).Input(t.Context(), "Module", "")
	require.ErrorIs(t, err, io.EOF)
}

func TestInteractivePrompterSelections(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	prompt := InteractivePrompter{In: strings.NewReader("2\n1,3\n"), Out: &out}
	selected, err := prompt.Select(t.Context(), "Pick", []string{"one", "two", "three"}, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, selected)
	multi, err := prompt.MultiSelect(t.Context(), "Pick many", []string{"one", "two", "three"}, nil)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 2}, multi)
	assert.NotContains(t, out.String(), "\x1b")
}
