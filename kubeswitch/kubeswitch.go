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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/ghodss/yaml"
	homedir "github.com/mitchellh/go-homedir"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/client-go/tools/clientcmd/api/v1"
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

	// Load config into memory and throw error if unsuccessful.
	buf, err := ioutil.ReadFile(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to load %s\n%s", cfg, err)
	}

	// Convert buffer into JSON format for unmarshaling.
	cfgJson, err := yaml.YAMLToJSON(buf)
	if err != nil {
		return nil, err
	}

	// Unmarshal config JSON into api.Config format.
	if err := json.Unmarshal(cfgJson, &ks.config); err != nil {
		return nil, err
	}

	return &ks, nil
}

// ListContexts return context names in loaded config.
func (k *KubeSwitch) ListContexts() *[]string {
	var ctxs []string

	for _, c := range k.config.Contexts {
		ctxs = append(ctxs, c.Name)
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

// GetContext returns requested context object from config and its slice's index.
func (k *KubeSwitch) GetContext(ctx string) (int, *api.Context, error) {
	for i, c := range k.config.Contexts {
		if c.Name == ctx {
			return i, &c.Context, nil
		}
	}
	return -1, nil, fmt.Errorf(fmt.Sprintf("context `%s` not found", ctx))
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
	// Create rest config from kube config file.
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv(configEVar))
	if err != nil {
		return err
	}

	// Create rest client from rest config.
	kApi, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// Fetch list of namespaces from Kubernetes.
	k.namespaces, err = kApi.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
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

	// Get the current context so that namespace can be set on it.
	idx, _, err := k.GetContext(k.config.CurrentContext)
	if err != nil {
		return err
	}

	// Set context's namespace to provided namespace
	k.config.Contexts[idx].Context.Namespace = ns

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

// writeConfig writes the unmarshaled config to disk.
func (k *KubeSwitch) writeConfig(path string) error {
	// Convert config object into JSON for writing to file.
	cfgFile, err := json.Marshal(k.config)
	if err != nil {
		return err
	}

	// Write session config file.
	if err := ioutil.WriteFile(path, cfgFile, 0644); err != nil {
		return err
	}

	return nil
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

func init() {
	// Create temporary session folder on startup if not exists.
	if _, err := os.Stat(sessionDir()); err != nil {
		if err := os.MkdirAll(sessionDir(), 0755); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
