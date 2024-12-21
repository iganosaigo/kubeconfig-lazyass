package cmd

import (
	"flag"
	"os"

	"github.com/iganosaigo/kubeconfig-lazyass/internal/logs"
	u "github.com/iganosaigo/kubeconfig-lazyass/internal/utils"
	p "github.com/iganosaigo/kubeconfig-lazyass/pkg"
)

type CLIArgs struct {
	kConfigRootPath string
	kConfigNewPath  string
	workingDir      string
	overwrite       bool
	daemon          bool
	kubeCtxName     string
}

func parse_args() CLIArgs {
	args := CLIArgs{}
	flag.StringVar(
		&args.kConfigRootPath, "kubeconfig-root", p.DefaultKubeconfigPath,
		"Root kubeconfig all others kubeconfigs are merged into")
	flag.BoolVar(
		&args.overwrite, "overwrite", false,
		"You have to specify this flag if you want overwrite existing context")
	flag.StringVar(&args.kConfigNewPath, "src-config", "",
		"kubeconfig to merge from")
	flag.BoolVar(&args.daemon, "daemon", false, "Daemonize with inotify watcher")
	flag.StringVar(&args.workingDir, "working-dir", "",
		"Config dir to watch for (Note: works with 'daemon' mode only)")
	flag.StringVar(&args.kubeCtxName, "context-name", "",
		"Context name for the merged file")
	flag.Parse()

	return args
}

// Override DEFAULT kubeconfig path if KUBECONFIG environ is set
func (c *CLIArgs) setKubeconfigPath() error {
	var err error
	if c.kConfigRootPath == p.DefaultKubeconfigPath {
		env := os.Getenv(p.DefaultKubeconfigEnv)
		if len(env) > 0 {
			c.kConfigRootPath = env
		}
	}
	c.kConfigRootPath, err = u.AbsolutePath(c.kConfigRootPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *CLIArgs) setWorkingDir() error {
	var err error
	if c.workingDir == "" {
		c.workingDir = u.GetDir(c.kConfigRootPath)
	} else {
		c.workingDir, err = u.AbsolutePath(c.workingDir)
		if err != nil {
			return err
		}
		if err = u.Stat(c.workingDir); err != nil {
			return err
		}
	}
	return nil
}

func Run() {
	logger := logs.NewLogger()
	args := parse_args()
	err := args.setKubeconfigPath()
	if err != nil {
		logger.Fatalf(p.SetKubeconfigError, "failed to create root kubeconfig path: %v", err)
	}

	if args.kConfigNewPath != "" {
		if args.kubeCtxName == "" {
			args.kubeCtxName = u.CleanName(args.kConfigNewPath)
			logger.Warn("no context provided for naming, using basename of file")
		}
		exitCode, err := p.Cmd(
			args.kConfigRootPath, args.kConfigNewPath,
			args.kubeCtxName, args.overwrite, logger)
		if err != nil || exitCode != p.Success {
			logger.Fatal(exitCode, err)
		}
		os.Exit(p.Success)
	}

	if args.daemon {
		err = args.setWorkingDir()
		if err != nil {
			logger.Fatalf(p.SetWorkingDirError, "failed to set working dir: %v", err)
		}
		exitCode, err := p.Daemon(args.kConfigRootPath,
			args.workingDir, args.overwrite, logger)
		if err != nil || exitCode != p.Success {
			logger.Fatal(exitCode, err)
		}
		os.Exit(p.Success)
	}

	logger.Fatal(p.GerericError, "don't know what to do, do you?")
}
