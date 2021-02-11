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
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

var pf = rootCmd.PersistentFlags()

func execOutput() (error, string) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := rootCmd.Execute()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	return err, string(out)
}

func TestConfigFlag(t *testing.T) {
	var out string
	var vs string

	// Test default Kubeswitch config.
	execOutput()
	vs = viper.GetString("config")
	if vs != defaultCfg {
		t.Errorf("Expected config to be %s, got %s", defaultCfg, vs)
	}

	// Test non-existence Kubeswitch config.
	pf.Set("config", "/path/to/not/exists/config")
	_, out = execOutput()
	warn := fmt.Sprintf("WARN: Config file \"%s\" not exists\n", viper.ConfigFileUsed())
	if !strings.Contains(out, warn) {
		t.Errorf("Non-existence config should throw warning")
	}
}

func TestNoConfigFlag(t *testing.T) {
	var vb bool

	// Test no-config set to true.
	pf.Set("no-config", "true")
	execOutput()
	vb = viper.GetBool("noConfig")
	if vb != true {
		t.Errorf("Expected no-config to be true, got %t", vb)
	}
}

func TestKubeconfigFlag(t *testing.T) {
	var vs string

	// Test setting kubeconfig flag.
	kCfg := "config.yaml"
	pf.Set("kubeconfig", kCfg)
	execOutput()
	vs = viper.GetString("kubeconfig")
	if vs != kCfg {
		t.Errorf("Expected kubeconfig to be %s, got %s", kCfg, vs)
	}
}

func TestPromptSizeFlag(t *testing.T) {
	var vi int

	// Test setting promtSize flag.
	pf.Set("prompt-size", "15")
	execOutput()
	vi = viper.GetInt("promptSize")
	if vi != 15 {
		t.Errorf("Expected promptSize to be 15, got %d", vi)
	}
}

func TestNoPromptFlag(t *testing.T) {
	var vb bool

	// Test setting noPrompt flag.
	pf.Set("no-prompt", "true")
	execOutput()
	vb = viper.GetBool("noPrompt")
	if vb != true {
		t.Errorf("Expected noPrompt to be true, got %t", vb)
	}
}
