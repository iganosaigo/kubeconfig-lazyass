package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	u "github.com/iganosaigo/kubeconfig-lazyass/internal/utils"

	fn "github.com/fsnotify/fsnotify"
	"golang.org/x/net/context"
)

func (w *watched) onEventCreateWrite(e fn.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()
	defer delete(w.timers, e.Name)
	// w.logger.Info(fmt.Sprintf("watcher: %s", e.String()))

	filename := filepath.Base(e.Name)
	ctxNameFromE := u.CleanName(filename)

	msgAdd := "added"
	if w.get(ctxNameFromE) {
		if w.overwrite {
			msgAdd = "replaced"
		}
	}

	err := w.parseAndMergeFile(e.Name)
	if err != nil {
		w.logger.Info(fmt.Sprintf("watcher: file %q skipped, %v", filename, err))
		return
	}

	err = w.writeRootConfig()
	if err != nil {
		w.logger.Error(fmt.Sprintf("watcher: failed to write context %q", ctxNameFromE))
		return
	}

	w.logger.Info(
		fmt.Sprintf(
			"watcher: context %q %s from file %s",
			ctxNameFromE, msgAdd, filename))
}

func (w *watched) onEventDelete(e fn.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// w.logger.Info(fmt.Sprintf("watcher: %s", e.String()))
	ctxNameFromE := u.CleanName(filepath.Base(e.Name))

	if w.get(ctxNameFromE) {
		w.deleteEntry(ctxNameFromE)
		err := w.writeRootConfig()
		if err != nil {
			w.logger.Error(fmt.Sprintf("watcher: failed to write context %q", ctxNameFromE))
			return
		}
		w.logger.Warn(fmt.Sprintf("watcher: context %q deleted", ctxNameFromE))
	} else {
		w.logger.Warn(
			fmt.Sprintf(
				"watcher: deleted only file %q, managed outside or wrong format?", e.Name))
	}
}

func (w *watched) watchChangeHandler(e fn.Event) {
	delay := 100 * time.Millisecond

	// Dedup Write and Create Events
	if e.Op == fn.Create || e.Op == fn.Write {
		w.mu.Lock()
		defer w.mu.Unlock()

		t, exist := w.timers[e.Name]
		if exist {
			t.Reset(delay)
		} else {
			t = time.AfterFunc(delay, func() { w.onEventCreateWrite(e) })
			w.timers[e.Name] = t
		}
	}
	// TODO: Do I need to add Rename Event?
	// For now I have only explicit unlink syscalls from upstream scripts...
	if e.Op == fn.Remove {
		w.onEventDelete(e)
	}
}

func (w *watched) watchLoop(
	ctx context.Context, inotify *fn.Watcher, watchCh chan<- struct{},
) {
	w.timers = make(map[string]*time.Timer)
	for {
		select {
		case err, ok := <-inotify.Errors:
			if !ok {
				w.logger.Info("watcher: stop inotify instance")
				return
			}
			w.logger.Error(err)
		case e, ok := <-inotify.Events:
			if !ok {
				w.logger.Info("watcher: stop inotify instance")
				return
			}
			if e.Name == w.rootConfigPath {
				switch op := e.Op; op {
				case fn.Remove, fn.Rename:
					w.logger.Fatal(GerericError, "watcher: root config was deleted from outside")
				case fn.Write:
					w.logger.Warn("watcher: root config manipulated from outside")
				}
				continue
			}
			if strings.HasSuffix(e.Name, ".lock") {
				continue
			}
			w.watchChangeHandler(e)
		case <-ctx.Done():
			w.logger.Info("watcher: signal recieved, shutting down")
			w.mu.Lock()
			defer func() { watchCh <- struct{}{} }()
			defer w.mu.Unlock()

			if err := w.writeRootConfig(); err != nil {
				w.logger.Fatal(SaveError, "watcher: failed to save config on shutting")
			} else {
				w.logger.Info("watcher: root config saved on exit")
			}
			return
		}
	}
}

func (w *watched) startWatcher(ctx context.Context) {
	watchCh := make(chan struct{})
	watcher, err := fn.NewWatcher()
	if err != nil {
		w.logger.Fatalf(
			InotifyInitError, "failed to initialize Inotify instance: %v", err)
	}
	defer watcher.Close()

	go w.watchLoop(ctx, watcher, watchCh)

	err = watcher.Add(w.workingDir)
	if err != nil {
		w.logger.Fatalf(
			InotifyInitError, "failed to watch dir %s: %v", w.workingDir, err)
	}
	<-watchCh
}
