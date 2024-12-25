package pkg

import (
	"errors"
	"fmt"
	"os"

	u "github.com/iganosaigo/kubeconfig-lazyass/internal/utils"
	kcmd "k8s.io/client-go/tools/clientcmd"
	kapi "k8s.io/client-go/tools/clientcmd/api"
)

func (w *watched) ParseOrCreateConfig() (*kapi.Config, error) {
	err := u.Stat(w.rootConfigPath)
	if err != nil && errors.Is(err, os.ErrPermission) {
		return nil, fmt.Errorf("permission denied: %w", err)
	} else if err != nil {
		return kapi.NewConfig(), nil
	}
	return w.ParseConfigFile(w.rootConfigPath)
}

func (w *watched) ParseConfigFile(path string) (*kapi.Config, error) {
	kubeConfig, err := kcmd.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	if err = kcmd.Validate(*kubeConfig); err != nil {
		return nil, err
	}

	return kubeConfig, nil
}

func (w *watched) Merge(ctxName string, cfgProvided *kConfig) (*kapi.Config, error) {
	rootConfigCopy := w.rootConfig.DeepCopy()

	cfgProvidedCount := w.isSingleConfig(cfgProvided.config)

	if len(cfgProvidedCount) == 0 {
		if _, ok := rootConfigCopy.Contexts[ctxName]; ok {
			if w.overwrite {
				w.logger.Warn(
					fmt.Sprintf("existing context %q want to be replaced", ctxName))
			} else {
				return nil, fmt.Errorf("context %q already present", ctxName)
			}
		}
		oldClusterName := u.GetSingleKey(cfgProvided.config.Clusters)
		oldCtxName := u.GetSingleKey(cfgProvided.config.Contexts)
		oldAuthName := u.GetSingleKey(cfgProvided.config.AuthInfos)

		if rootConfigCopy.CurrentContext == "" {
			rootConfigCopy.CurrentContext = ctxName
		}

		rootConfigCopy.Clusters[ctxName] = cfgProvided.config.Clusters[oldClusterName]
		rootConfigCopy.Contexts[ctxName] = cfgProvided.config.Contexts[oldCtxName]
		rootConfigCopy.AuthInfos[ctxName] = cfgProvided.config.AuthInfos[oldAuthName]
		rootConfigCopy.Contexts[ctxName].Cluster = ctxName
		rootConfigCopy.Contexts[ctxName].AuthInfo = ctxName
	} else {
		return nil, fmt.Errorf("multiple clusters provided")
	}

	if err := kcmd.Validate(*rootConfigCopy); err != nil {
		return nil, err
	}

	w.configs[ctxName] = cfgProvided
	return rootConfigCopy, nil
}

func (w *watched) isSingleConfig(config *kapi.Config) []string {
	result := make([]string, 0)
	if singleSection := u.IsSingleEntry(config.Contexts); !singleSection {
		result = append(result, "Context")
	}
	if singleSection := u.IsSingleEntry(config.Clusters); !singleSection {
		result = append(result, "Clusters")
	}
	if singleSection := u.IsSingleEntry(config.AuthInfos); !singleSection {
		result = append(result, "Users")
	}

	return result
}

func (w *watched) saveRootConfig() (int, error) {
	err := kcmd.WriteToFile(*w.rootConfig, w.rootConfigPath)
	if err != nil {
		return SaveError, fmt.Errorf("write merged config: %w", err)
	}
	return Success, nil
}
