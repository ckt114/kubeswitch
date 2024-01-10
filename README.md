# Kubeswitch

Kubernetes context and namespace switching with style. Inspired by [Kubie](https://github.com/sbstp/kubie) with additional features that Kubie lacks.

# Prerequisites

- go 1.14+
- make

# Installation

## Download

You can download Kubeswitch from [releases](https://github.com/ckt114/kubeswitch/releases) page.

## MacOS
```console
$ brew install ckt114/tap/kubeswitch
```

## Build

You can build Kubeswitch by running the following commands.

```shell
$ git clone https://github.com/ckt114/kubeswitch
$ cd kubeswitch
$ make
$ sudo make install
```

## Default Configuration

Install default configuation to as `$HOME/.kubeswitch.yaml`.

```shell
$ make install-config
```

## Shell Completion

### Bash

```shell
$ make bash-completion
```

### ZSH

```shell
$ make zsh-completion
```

### Fish

```shell
$ make fish-completion
```

# Usage

## Without Shell Completion

```shell
# Switching context. Use context or ctx commands.
$ kubeswitch ctx [ENTER]
? Select context. / to search:
    aws-east1
    aws-west1
    k3s
  ▸ kind

# Switching namespace. Use namespace or ns commands.
(kind|default) $ kubeswitch ns [ENTER]
? Select namespace. / to search:
    argocd
    default
  ▸ jenkins
    kube-node-lease
    kube-public
    kube-system
(kind|jenkins) $
```

## With Shell Completion

```shell
# Switching context. Use context or ctx commands.
$ kubeswitch ctx [TAB]
aws-east1  aws-west1   k3s  kind

# Switching namespace. Use namespace or ns commands.
(kind|default) $ kubeswitch ns [TAB]
argocd  default  jenkins  kube-node-lease  kube-public  kube-system
```

## Configuration

Kubeswitch default config file is `$HOME/.kubeswitch.yaml`.
Use `-c` or `--config` flags or `KUBESWITCH_CONFIG` environment variable to
override default config file. The following keys are used by Kubeswitch

- `kubeConfig` - Kubernetes config file to merge into Kubeswitch session file `KUBESWITCH_KUBECONFIG`
- `configs` - Array list of path patterns to search for Kubernetes config files
- `promptSize` - Number of items to show for selection prompt`KUBESWITCH_PROMPTSIZE`
- `noPrompt` - Don't use selection prompt; print each item per line`KUBESWITCH_NOPROMPT`
- `purge`
  - `days` - Number of days to retain Kubeswitch session files`KUBESWITCH_PURGE_DAYS`

# Shell Prompt

Using shell prompt integration will greatly help knowing which Kubernetes
context and namespace you're currently interacting with.

## Bash and ZSH

- [kube-ps1](https://github.com/jonmosco/kube-ps1)

## Fish

- [fish-kubectl-prompt](https://github.com/vpistis/fish-kubectl-prompt)
