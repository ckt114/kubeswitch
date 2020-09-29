# Kubeswitch
Kubernetes context and namespace switching with style. Inspired by [Kubie](https://github.com/sbstp/kubie) with additional features that Kubie lacks.


# Prerequisites
- go 1.14+
- make


# Installation
## Download
You can download Kubeswitch from [releases](https://github.com/trankchung/kubeswitch/releases) page.

## Build
You can build Kubeswitch by running the following commands.
```shell
$ git clone https://github.com/trankchung/kubeswitch
$ cd kubeswitch
$ make
$ sudo make install
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
## No Shell Completion
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

# Shell Prompt
Using shell prompt integration will greatly help knowing which Kubernetes context and namespace you're currently interacting with.
## Bash and ZSH
- [kube-ps1](https://github.com/jonmosco/kube-ps1)

## Fish
- [fish-kubectl-prompt](https://github.com/vpistis/fish-kubectl-prompt)
