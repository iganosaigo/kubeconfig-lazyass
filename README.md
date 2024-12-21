# kubeconfig-lazyass

`kubeconfig-lazyass[istant]` is an Inotify daemon that allows you to monitor
countless K8S kubeconfig files and add/remove those configurations to a single
kubeconfig in the background. Note: at moment the daemon only manage configs with
single entries, ie Context, Cluster, User.

## Goals

If you are a lazy person like me and have a huge number of configuration
files for K8S clusters that are often changed/added/removed, and you do
not want to manage those many configurations manually, this tool is
perfect for you.

The daemon monitors the appearance of configuration files in the working
directory and automatically merges these configurations into a single one.
When a file is deleted, the daemon removes the config from the root config
file. It can also overwrite existing configs in the main config file if
overwrite setting is set.

So, in a nutshell it's great for:

- Having hundreds(even thousands) of K8S kubeconfigs that are dynamically
  added/removed.
- Constantly overwriting/replacing kubeconfigs
- Deploying a single management node to access all clusters (bad idea:)
- You don't want to waste time merging configs manually.

In addition, the lazyass can work in CLI mode with merging config manually,
like many other similar utilities that you can find on the github.com. It might
not be that lazy, so you can always rename the binary to `kubeconfig-imnotlazyass`.

---

## Installing & Updating

```bash
go install github.com/iganosaigo/kubeconfig-lazyass@latest
```

### Daemon Mode

Running `kubeconfig-lazyass` in daemon mode will watch the files in the working
directory. The name for the new fields Context, Clusters and Users that will be added
will be determined based on the file name without the extension. For example,
`my.newcluster.com.yaml` will be converted to `my.newcluster.com`.

Current options for daemon mode:

- `--daemon` - Required for daemon. Without that opt will run in manual mode.
- `--working-dir` - Directory where your multiple configs are placed. If not
  specified then directory of the root kubeconfig will be used.
- `--kubeconfig-root` - The root kubeconfig file(by default ~/.kube/config, but
  overrided by KUBECONFIG environ if set), where all configurations
  will be merged. Note that you can specify a non-existent file, and the
  daemon will create a new config. It is probably better to create a new
  file to separate the automatic and manual logic.
- `--overwrite` - without this option, if an existing file is overwritten, i.e.
  there was a syscall write, the configuration in the root config will not be
  overwritten if present already.

### CLI Mode

Running in CLI mode has the folowing opts:

- `--kubeconfig-root` - The same as in daemon mode.
- `--src-config` - Config file you want to merge into root config.
- `--context-name` - Naming for new configuration. If you don't specify this
  opt then new name will be determined based on file name without extension.
- `--overwrite` - same as in daemon mode.

### Examples

To run daemon server

```bash
kubeconfig-lazyass --daemon --overwrite --kubeconfig-root ~/.kube/config_combined
```
