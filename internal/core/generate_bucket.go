package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type ImmutableData struct {
	data []byte
}

func newImmutableData(data []byte) *ImmutableData {
	return &ImmutableData{data: data}
}

func (i *ImmutableData) Data() []byte {
	return i.data
}

type GenerateBucket struct {
	filesToWrite map[string]*ImmutableData
	lock         sync.RWMutex
}

func NewGenerateBucket() *GenerateBucket {
	return &GenerateBucket{
		filesToWrite: make(map[string]*ImmutableData),
	}
}

func (b *GenerateBucket) PutFile(_ context.Context, path string, data []byte) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.filesToWrite[path] = newImmutableData(data)
}

func (b *GenerateBucket) GetFile(_ context.Context, path string) (*ImmutableData, bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	file, ok := b.filesToWrite[path]
	return file, ok
}

func (b *GenerateBucket) RemoveFile(_ context.Context, path string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	delete(b.filesToWrite, path)
}

func (b *GenerateBucket) DumpToFs(ctx context.Context) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	// TODO: Есть возможность писать файлы асинхронно (MkdirAll до)
	for path, file := range b.filesToWrite {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("os.MkdirAll %s: %w", dir, err)
		}

		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("os.Create %s: %w", path, err)
		}
		_, err = f.Write(file.Data())
		if err != nil {
			return fmt.Errorf("f.Write %s: %w", path, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("f.Close %s: %w", path, err)
		}
	}

	return nil
}
