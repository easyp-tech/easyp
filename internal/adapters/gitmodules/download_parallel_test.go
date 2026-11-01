package gitmodules

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestInstallBoundsAcquisitionAndJoinsCancellation(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		download bool
	}{
		{name: "cache_install"},
		{name: "mod_download", download: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testInstallAcquisitionBound(t, test.download)
		})
	}
}

func testInstallAcquisitionBound(t *testing.T, download bool) {
	t.Helper()
	storage := t.TempDir()
	project := t.TempDir()
	synctest.Test(t, func(t *testing.T) {
		const count = 12
		started := make(chan string, count)
		release := make(chan struct{})
		log := installBarrierLogger{started: started, release: release}
		cache := New(storage).WithLogger(log)
		lock := v1.Lock{Version: 1}
		for index := range count {
			lock.Modules = append(lock.Modules, v1.LockedModule{
				Source:  "example.test/team/" + string(rune('a'+index)),
				Version: strings.Repeat("a", 40), Commit: strings.Repeat("a", 40),
				Hash: "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			})
		}
		var manifest strings.Builder
		manifest.WriteString("module example.test/app\n")
		for _, entry := range lock.Modules {
			manifest.WriteString("require " + entry.Source + " " + entry.Version + "\n")
		}
		require.NoError(t, os.WriteFile(filepath.Join(project, v1.ModuleFile), []byte(manifest.String()), 0o600))
		raw, err := yaml.Marshal(lock)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(project, v1.LockFile), raw, 0o600))
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan error, 1)
		go func() {
			if download {
				done <- modules.Download(ctx, project, cache)
				return
			}
			done <- cache.Install(ctx, lock)
		}()
		synctest.Wait()
		assert.Len(t, started, 8, "independent installs should fill the bounded worker pool")
		cancel()
		synctest.Wait()
		select {
		case <-done:
			t.Error("Install returned while acquisition workers were active")
		default:
		}
		close(release)
		require.ErrorIs(t, <-done, context.Canceled)
		assert.Len(t, started, 8, "cancellation must stop scheduling more installs")
	})
}

func TestVerifyCachedReturnsFirstLockErrorAfterOutOfOrderCompletion(t *testing.T) {
	t.Parallel()
	storage := t.TempDir()
	synctest.Test(t, func(t *testing.T) {
		log := installOrderLogger{secondDone: make(chan struct{})}
		cache := New(storage).WithLogger(log)
		lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{
			{Source: "a", Version: versionlessTestPin, Commit: versionlessTestPin, Hash: "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},
			{Source: "b", Version: versionlessTestPin, Commit: versionlessTestPin, Hash: "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},
		}}
		err := cache.VerifyCached(t.Context(), lock)
		require.ErrorContains(t, err, "policy module a@")
		_, statErr := os.Stat(cache.root)
		require.ErrorIs(t, statErr, os.ErrNotExist, "read-only verification must not create cache directories")
	})
}

func TestConcurrentInstallAndVerificationRespectSnapshotGate(t *testing.T) {
	t.Parallel()
	cache := New(t.TempDir())
	entry := v1.LockedModule{Source: filepath.Join(t.TempDir(), "missing"), Version: versionlessTestPin, Commit: versionlessTestPin, Hash: "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}
	lock := v1.Lock{Version: 1, Modules: []v1.LockedModule{entry}}
	synctest.Test(t, func(t *testing.T) {
		snapshot := cache.snapshotState(v1ModuleCachePath(cache.root, entry))
		require.NoError(t, snapshot.lock(t.Context()))
		defer snapshot.unlock()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan error, 2)
		go func() { done <- cache.Install(ctx, lock) }()
		go func() { done <- cache.VerifyCached(ctx, lock) }()
		synctest.Wait()
		assert.Empty(t, done, "both commands must wait for the shared snapshot gate")
		cancel()
		require.ErrorIs(t, <-done, context.Canceled)
		require.ErrorIs(t, <-done, context.Canceled)
		_, err := os.Stat(cache.root)
		require.ErrorIs(t, err, os.ErrNotExist)
	})
}

const versionlessTestPin = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type installOrderLogger struct {
	installBarrierLogger
	secondDone chan struct{}
}

func (log installOrderLogger) Debug(_ context.Context, message string, _ ...slog.Attr) {
	if log.phase != "install module" {
		return
	}
	if log.source == "a" && message == "Dependency operation started" {
		<-log.secondDone
	}
	if log.source == "b" && message == "Dependency operation finished" {
		close(log.secondDone)
	}
}

func (log installOrderLogger) With(attrs ...slog.Attr) logger.Logger {
	log.installBarrierLogger = log.installBarrierLogger.With(attrs...).(installBarrierLogger)
	return log
}

type installBarrierLogger struct {
	started chan<- string
	release <-chan struct{}
	source  string
	phase   string
}

func (log installBarrierLogger) Debug(_ context.Context, message string, _ ...slog.Attr) {
	if message == "Dependency operation started" && log.phase == "install module" {
		log.started <- log.source
		<-log.release
	}
}

func (installBarrierLogger) Info(context.Context, string, ...slog.Attr)  {}
func (installBarrierLogger) Warn(context.Context, string, ...slog.Attr)  {}
func (installBarrierLogger) Error(context.Context, string, ...slog.Attr) {}

func (log installBarrierLogger) With(attrs ...slog.Attr) logger.Logger {
	for _, attr := range attrs {
		switch attr.Key {
		case "source":
			log.source = attr.Value.String()
		case "phase":
			log.phase = attr.Value.String()
		}
	}
	return log
}
