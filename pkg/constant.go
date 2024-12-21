package pkg

const (
	DefaultKubeconfigEnv  = "KUBECONFIG"
	DefaultKubeconfigPath = "~/.kube/config"
)

const (
	Success int = iota
	GerericError
	SetWorkingDirError
	ChangeDirError
	SetKubeconfigError
	ParseError
	MergeError
	SaveError
	InotifyInitError
)
