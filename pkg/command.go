package pkg

import (
	"fmt"

	"github.com/iganosaigo/kubeconfig-lazyass/internal/logs"
)

func Cmd(cfgRootPath, cfgNewPath, cfgNewCtx string, overwrite bool, logger *logs.Logger) (int, error) {
	config := &watched{
		rootConfigPath: cfgRootPath,
		logger:         logger,
		configs:        make(map[string]*kConfig),
		overwrite:      overwrite,
	}
	cfgNew := &kConfig{file: cfgNewPath}

	var err error
	config.rootConfig, err = config.ParseOrCreateConfig()
	if err != nil {
		return ParseError, err
	}

	cfgNew.config, err = config.ParseConfigFile(cfgNew.file)
	if err != nil {
		return ParseError, err
	}

	config.rootConfig, err = config.Merge(cfgNewCtx, cfgNew)
	if err != nil {
		return MergeError, fmt.Errorf("merging configs: %w", err)
	}

	return config.saveRootConfig()
}
