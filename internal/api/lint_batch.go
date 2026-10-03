package api

import "github.com/easyp-tech/easyp/internal/fs/fs"

// selectedLintWalker enumerates only checked files while retaining full root
// access for imports and RootPath for module-relative naming rules.
type selectedLintWalker struct {
	*fs.FSWalker
	files []string
}

func newSelectedLintWalker(root string, files []string) *selectedLintWalker {
	return &selectedLintWalker{FSWalker: fs.NewFSWalker(root, "."), files: files}
}

func (w *selectedLintWalker) WalkDir(callback func(string, error) error) error {
	for _, file := range w.files {
		if err := callback(file, nil); err != nil {
			return err
		}
	}
	return nil
}
