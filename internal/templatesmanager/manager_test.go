package templatesmanager_test

import (
	"testing"

	"github.com/alvii147/flagger-api/internal/templatesmanager"
	"github.com/stretchr/testify/require"
)

func TestManagerLoadSuccess(t *testing.T) {
	t.Parallel()

	tmplManager := templatesmanager.NewManager()
	textTmpl, htmlTmpl, err := tmplManager.Load("activation")

	require.NoError(t, err)
	require.NotNil(t, textTmpl)
	require.NotNil(t, htmlTmpl)
}

func TestManagerLoadError(t *testing.T) {
	t.Parallel()

	tmplManager := templatesmanager.NewManager()
	_, _, err := tmplManager.Load("deadbeef")

	require.Error(t, err)
}
