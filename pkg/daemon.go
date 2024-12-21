package pkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/iganosaigo/kubeconfig-lazyass/internal/logs"
	"github.com/iganosaigo/kubeconfig-lazyass/internal/utils"

	kcmd "k8s.io/client-go/tools/clientcmd"
)

func Daemon(cfgRoot, workingDir string, overwrite bool,
	logger *logs.Logger) (int, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := &watched{
		configs:        make(map[string]*kConfig),
		stopCh:         make(chan os.Signal, 1),
		logger:         logger,
		overwrite:      overwrite,
		workingDir:     workingDir,
		rootConfigPath: cfgRoot,
	}

	signal.Notify(watcher.stopCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-watcher.stopCh
		cancel()
	}()

	rootContent, err := watcher.ParseOrCreateConfig()
	if err != nil {
		return ParseError, fmt.Errorf("parsing root config: %w", err)
	}
	watcher.rootConfig = rootContent

	if err = os.Chdir(workingDir); err != nil {
		return ChangeDirError, err
	}

	allFiles, err := utils.ListFilesInDir(workingDir)
	if err != nil {
		return ParseError, fmt.Errorf("list files: %v", err)
	}

	for _, file := range allFiles {
		if file == filepath.Base(cfgRoot) {
			continue
		}
		err := watcher.parseAndMergeFile(file)
		if err != nil {
			watcher.logger.Info(fmt.Sprintf("init: file %q skipped, %v", file, err))
			continue
		}
		watcher.logger.Info(fmt.Sprintf("init: file %q added to root config", file))
	}

	watcher.writeRootConfig()
	watcher.startWatcher(ctx)
	return Success, nil
}

func (w *watched) get(ctxName string) bool {
	_, exists := w.configs[ctxName]
	return exists
}

func (w *watched) deleteEntry(ctxName string) {
	delete(w.configs, ctxName)
	delete(w.rootConfig.Clusters, ctxName)
	delete(w.rootConfig.Contexts, ctxName)
	delete(w.rootConfig.AuthInfos, ctxName)

	if w.rootConfig.CurrentContext == ctxName {
		w.rootConfig.CurrentContext = ""
	}
}

func (w *watched) parseAndMergeFile(file string) error {
	configParsed, err := w.ParseConfigFile(file)
	if err != nil {
		// return fmt.Errorf("parse error: %w", err)
		return errors.New("parse error")
	}

	newFile := &kConfig{file: file, config: configParsed}
	fileCtxName := utils.CleanName(newFile.file)
	newRootContent, err := w.Merge(fileCtxName, newFile)
	if err != nil {
		return err
	}

	w.rootConfig = newRootContent
	return nil
}

func (w *watched) writeRootConfig() error {
	err := kcmd.WriteToFile(*w.rootConfig, w.rootConfigPath)
	if err != nil {
		return err
	}
	return nil
}
