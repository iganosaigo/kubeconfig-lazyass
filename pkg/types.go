package pkg

import (
	"os"
	"sync"
	"time"

	"github.com/iganosaigo/kubeconfig-lazyass/internal/logs"
	kapi "k8s.io/client-go/tools/clientcmd/api"
)

type kConfig struct {
	file   string
	config *kapi.Config
}

type watched struct {
	workingDir     string
	rootConfigPath string
	rootConfig     *kapi.Config
	configs        map[string]*kConfig
	overwrite      bool
	logger         *logs.Logger
	timers         map[string]*time.Timer
	stopCh         chan os.Signal
	mu             sync.Mutex
}
