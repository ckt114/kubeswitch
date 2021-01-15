/*
Copyright © 2020 Chung Tran <chung.k.tran@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kubeswitch

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/client-go/tools/clientcmd/api"
)

const (
	// EnvVarActive is the env var that will be used
	// to determine if shell is started by kubeswitch.
	EnvVarActive = "KUBESWITCH_ACTIVE"

	// EnvVarConfig is the env var that points to a
	// session's kube config.
	EnvVarConfig = "KUBECONFIG"
)

var (
	// kubeDir returns the default kube folder.
	kubeDir = func() string {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return home + "/.kube"
	}

	// sessionDir stores kubeswitch copied config session files.
	sessionDir = func() string {
		return kubeDir() + "/tmp"
	}
)

// Kubeswitch holds loaded kube config and loaded namespaces.
type Kubeswitch struct {
	// config contains the content of loaded config
	config *api.Config

	// namespaces contains namespaces from Kubernetes
	// for current context.
	namespaces *corev1.NamespaceList
}

// New returns an instance of Kubeswitch after loading the config
// file from passed in path, KUBECONFIG env var, or default location.
func New() (*Kubeswitch, error) {
	// Load config files.
	po := clientcmd.NewDefaultPathOptions()
	config, err := po.GetStartingConfig()
	if err != nil {
		return nil, err
	}

	// Flatten config files into single file.
	if err := api.FlattenConfig(config); err != nil {
		return nil, err
	}

	return &Kubeswitch{config: config}, nil
}

// ListContexts return context names in loaded config.
func (k *Kubeswitch) ListContexts() *[]string {
	var ctxs []string

	for ctx := range k.config.Contexts {
		ctxs = append(ctxs, ctx)
	}

	sort.Strings(ctxs)
	return &ctxs
}

// SetContext set context as current context.
func (k *Kubeswitch) SetContext(ctx string) error {
	// Error out if context is not valid.
	if !k.IsValidContext(ctx) {
		return fmt.Errorf("invalid context, %s", ctx)
	}

	// Set current context to chosen context.
	k.config.CurrentContext = ctx

	// Create/update session config.
	if err := k.setupSession(); err != nil {
		return err
	}

	return nil
}

// setupSession creates a Kubeswitch session by merging all the kubeconfigs and
// write it to a temporary file and set KUBECONFIG to that file's path if not in
// a Kubeswitch sessions. Otherwise, just write the changes to the path defined in
// KUBECONFIG env var.
func (k *Kubeswitch) setupSession() error {
	// Just write the config to KUBECONFIG if in Kubeswitch session.
	if IsActive() {
		if err := k.writeConfig(os.Getenv(EnvVarConfig)); err != nil {
			return err
		}
	} else {
		// Construct temporary timestamped kubeconfig session file.
		now := time.Now()
		kubePath := fmt.Sprintf("%s/config_%d", sessionDir(), now.UnixNano())

		// Write config to temp path for new session.
		if err := k.writeConfig(kubePath); err != nil {
			return err
		}

		// Set env vars that will be visible when running new shell below.
		os.Setenv(EnvVarActive, "TRUE")
		os.Setenv(EnvVarConfig, kubePath)

		// Run a shell with new config path set as env var above.
		syscall.Exec(os.Getenv("SHELL"), []string{os.Getenv("SHELL")}, syscall.Environ())
	}

	return nil

}

// IsValidContext return true if context is one of the contexts.
func (k *Kubeswitch) IsValidContext(ctx string) bool {
	for _, c := range *k.ListContexts() {
		if ctx == c {
			return true
		}
	}
	return false
}

// LoadNamespaces loads list of namespaces for current context live from Kubernetes.
func (k *Kubeswitch) LoadNamespaces() error {
	// Convert config into []bytes.
	cfgBytes, err := clientcmd.Write(*k.config)
	if err != nil {
		return err
	}

	// Create REST config from config []bytes.
	restCfg, err := clientcmd.RESTConfigFromKubeConfig(cfgBytes)
	if err != nil {
		return err
	}

	// Create kube REST client from REST config.
	kube, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return err
	}

	// Fetch list of namespaces from Kubernetes.
	k.namespaces, err = kube.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	return nil
}

// ListNamespaces return namespaces live from Kubernetes.
func (k *Kubeswitch) ListNamespaces() *[]string {
	var nss []string

	for _, n := range k.namespaces.Items {
		nss = append(nss, n.Name)
	}

	sort.Strings(nss)
	return &nss
}

// SetNamespace sets default namespace for current context.
func (k *Kubeswitch) SetNamespace(ns string) error {
	// Error out if namespace is not valid.
	if !k.IsValidNamespace(ns) {
		return fmt.Errorf("invalid namespace, %s", ns)
	}

	// Find the current context and set its default namespace.
	for name, ctx := range k.config.Contexts {
		if name == k.config.CurrentContext {
			ctx.Namespace = ns
		}
	}

	// Create/update session config.
	if err := k.setupSession(); err != nil {
		return err
	}

	return nil
}

// IsValidNamespace return true if namespace is one of the namespaces.
func (k *Kubeswitch) IsValidNamespace(ns string) bool {
	for _, n := range k.namespaces.Items {
		if n.Name == ns {
			return true
		}
	}
	return false
}

// IsActive returns true if inside kubeswitch session.
// It uses EnvVarActive value to determine if in kubeswitch session.
func IsActive() bool {
	if a := strings.ToUpper(os.Getenv(EnvVarActive)); a == "TRUE" {
		return true
	}

	return false
}

// Purge deletes temporary session files older than `days`.
func Purge(days int) {
	delTime := time.Now().AddDate(0, 0, days*-1)

	// Delete files that are older than `days` in session folder.
	dir, _ := ioutil.ReadDir(sessionDir())
	for _, i := range dir {
		if i.ModTime().Before(delTime) {
			if err := os.Remove(sessionDir() + "/" + i.Name()); err != nil {
				fmt.Println(err)
			}
		}
	}
}

// writeConfig writes the unmarshaled config to disk.
func (k *Kubeswitch) writeConfig(path string) error {
	// Write session config file.
	if err := clientcmd.WriteToFile(*k.config, path); err != nil {
		return err
	}

	return nil
}

func init() {
	// Create temporary session folder on startup if not exists.
	if _, err := os.Stat(sessionDir()); err != nil {
		if err := os.MkdirAll(sessionDir(), 0700); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
