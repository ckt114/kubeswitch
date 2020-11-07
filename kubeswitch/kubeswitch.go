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
	// activeEVar is the env var that will be used
	// to determine if shell is started by kubeswitch.
	activeEVar = "KUBESWITCH_ACTIVE"

	// configEVar is the env var that points to a
	// session's kube config.
	configEVar = "KUBECONFIG"
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

	// defaultConfig returns value from KUBECONFIG env var
	// if defined; otherwise use default config path.
	defaultConfig = func() string {
		cfg := os.Getenv(configEVar)
		if cfg == "" {
			cfg = kubeDir() + "/config"
		}
		return cfg
	}
)

// KubeSwitch holds loaded kube config and loaded namespaces.
type KubeSwitch struct {
	// config contains the content of loaded config
	config *api.Config

	// namespaces contains namespaces from Kubernetes
	// for current context.
	namespaces *corev1.NamespaceList
}

// NewKubeSwitch returns an instance of KubeSwitch after loading the config
// file from passed in path, KUBECONFIG env var, or default location.
func NewKubeSwitch(cfg string) (*KubeSwitch, error) {
	// Use config path if passed in. Otherwise use default location.
	if cfg == "" {
		cfg = defaultConfig()
	}

	var ks KubeSwitch
	ks.config = clientcmd.GetConfigFromFileOrDie(cfg)

	return &ks, nil
}

// ListContexts return context names in loaded config.
func (k *KubeSwitch) ListContexts() *[]string {
	var ctxs []string

	for ctx := range k.config.Contexts {
		ctxs = append(ctxs, ctx)
	}

	sort.Strings(ctxs)
	return &ctxs
}

// SetContext set context as current context.
func (k *KubeSwitch) SetContext(ctx string) error {
	// Error out if context is not valid.
	if !k.IsValidContext(ctx) {
		return fmt.Errorf("invalid context, %s", ctx)
	}

	// Set current context to chosen context.
	k.config.CurrentContext = ctx

	// If in an active kubeswitch session, then just write the config
	// back into KUBECONFIG path b/c that path already is the session's kube config path.
	if k.IsActive() {
		if err := k.writeConfig(os.Getenv(configEVar)); err != nil {
			return err
		}
	} else {
		// Since we're not in a kubeswitch session, construct a path to the new
		// temp config file and write a copy of the config with updated
		// chosen context to it.
		now := time.Now()
		kubePath := fmt.Sprintf("%s/config_%d", sessionDir(), now.UnixNano())

		// Write updated config with selected context to temp path for new session.
		if err := k.writeConfig(kubePath); err != nil {
			return err
		}

		// Set env vars that will be visible when running new shell below.
		os.Setenv(activeEVar, "TRUE")
		os.Setenv(configEVar, kubePath)

		// Run a shell with new config path set as env var above.
		syscall.Exec(os.Getenv("SHELL"), []string{os.Getenv("SHELL")}, syscall.Environ())
	}

	return nil
}

// IsValidContext return true if context is one of the contexts.
func (k *KubeSwitch) IsValidContext(ctx string) bool {
	for _, c := range *k.ListContexts() {
		if ctx == c {
			return true
		}
	}
	return false
}

// LoadNamespaces loads list of namespaces for current context live from Kubernetes.
func (k *KubeSwitch) LoadNamespaces() error {
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
func (k *KubeSwitch) ListNamespaces() *[]string {
	var nss []string

	for _, n := range k.namespaces.Items {
		nss = append(nss, n.Name)
	}

	sort.Strings(nss)
	return &nss
}

// SetNamespace sets default namespace for current context.
func (k *KubeSwitch) SetNamespace(ns string) error {
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

	// Write updated config with selected namespace.
	if err := k.writeConfig(os.Getenv(configEVar)); err != nil {
		return err
	}

	return nil
}

// IsValidNamespace return true if namespace is one of the namespaces.
func (k *KubeSwitch) IsValidNamespace(ns string) bool {
	for _, n := range k.namespaces.Items {
		if n.Name == ns {
			return true
		}
	}
	return false
}

// IsActive returns true if inside kubeswitch session.
// It uses env var KUBESWITCH_ACTIVE=TRUE to determine if in kubeswitch session.
func (k *KubeSwitch) IsActive() bool {
	if active := strings.ToUpper(os.Getenv(activeEVar)); active == "TRUE" {
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
func (k *KubeSwitch) writeConfig(path string) error {
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
