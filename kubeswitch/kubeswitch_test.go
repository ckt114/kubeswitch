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
	"fmt"
	"os"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

var ks *Kubeswitch

func TestNew(t *testing.T) {
	// Test using JSON config.
	os.Setenv(EnvVarConfig, "../fixtures/config.json")
	if _, err := New(); err != nil {
		t.Errorf("Expected error to be %v, got %v", nil, err)
	}

	// Test using YAML config.
	os.Setenv(EnvVarConfig, "../fixtures/config.yaml")
	if _, err := New(); err != nil {
		t.Errorf("Expected error to be %v, got %v", nil, err)
	}
}

func TestListContexts(t *testing.T) {
	ctxs := *ks.ListContexts()
	if reflect.TypeOf(ctxs) != reflect.TypeOf([]string{}) {
		t.Errorf("Expected result type of %v, got %v", reflect.TypeOf([]string{}), reflect.TypeOf(ctxs))
	}

	if len(ctxs) != 1 {
		t.Errorf("Expected length is %v, got %v", 1, len(ctxs))
	}
}

func TestIsValidContext(t *testing.T) {
	// Testing with valid context.
	if valid := ks.IsValidContext("default"); !valid {
		t.Errorf("Expected valid to be %v, got %v", true, valid)
	}

	// Testing with invalid context.
	if valid := ks.IsValidContext("invalid"); valid {
		t.Errorf("Expected valid to be %v, got %v", false, valid)
	}
}

func TestListNamespaces(t *testing.T) {
	size := 3
	loadNamespaces(ks, size)

	nss := *ks.ListNamespaces()
	if reflect.TypeOf(nss) != reflect.TypeOf([]string{}) {
		t.Errorf("Expected result type of %v, got %v", reflect.TypeOf([]string{}), reflect.TypeOf(nss))
	}

	if len(nss) != size {
		t.Errorf("Expected length is %v, got %v", size, len(nss))
	}
}

func TestIsValidNamespace(t *testing.T) {
	loadNamespaces(ks, 1)

	// Test with valid namespace.
	if valid := ks.IsValidNamespace("Namespace1"); !valid {
		t.Errorf("Expected valid to be %v, got %v", true, valid)
	}

	// Test with invalid namespace.
	if valid := ks.IsValidNamespace("invalid"); valid {
		t.Errorf("Expected valid to be %v, got %v", false, valid)
	}
}

func TestIsActive(t *testing.T) {
	// Test with active session.
	os.Setenv(EnvVarActive, "TRUE")
	if active := IsActive(); !active {
		t.Errorf("Expected active to be %v, got %v", true, active)
	}

	// Test with inactive session with env var set to nothing.
	os.Setenv(EnvVarActive, "")
	if active := IsActive(); active {
		t.Errorf("Expected active to be %v, got %v", false, active)
	}

	// Test with inactive session with no env var set.
	os.Unsetenv(EnvVarActive)
	if active := IsActive(); active {
		t.Errorf("Expected active to be %v, got %v", false, active)
	}

}

func init() {
	os.Setenv(EnvVarConfig, "../fixtures/config.yaml")
	ks, _ = New()
}

// Load sample namespaces for testing.
func loadNamespaces(k *Kubeswitch, size int) {
	var nss []corev1.Namespace
	for i := 0; i < size; i++ {
		ns := corev1.Namespace{}
		ns.Name = fmt.Sprintf("Namespace%d", i+1)
		nss = append(nss, ns)
	}
	var nsList corev1.NamespaceList
	nsList.Items = nss
	k.namespaces = &nsList
}
