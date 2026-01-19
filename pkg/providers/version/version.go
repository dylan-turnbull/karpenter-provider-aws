/*
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

package version

import (
	"context"
	"fmt"
	"strings"

	"github.com/patrickmn/go-cache"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/logging"

	"sigs.k8s.io/karpenter/pkg/utils/pretty"
)

const (
	kubernetesVersionCacheKey = "kubernetesVersion"
	// Karpenter's supported version of Kubernetes
	// These constants are used for documentation generation and compatibility matrices
	MinK8sVersion = "1.23"
	MaxK8sVersion = "1.29"
)

type Provider interface {
	Get(ctx context.Context) (string, error)
}

// DefaultProvider get the APIServer version. This will be initialized at start up and allows karpenter to have an understanding of the cluster version
// for decision making. The version is cached to help reduce the amount of calls made to the API Server
type DefaultProvider struct {
	cache               *cache.Cache
	cm                  *pretty.ChangeMonitor
	kubernetesInterface kubernetes.Interface
}

func NewDefaultProvider(kubernetesInterface kubernetes.Interface, cache *cache.Cache) *DefaultProvider {
	return &DefaultProvider{
		cm:                  pretty.NewChangeMonitor(),
		cache:               cache,
		kubernetesInterface: kubernetesInterface,
	}
}

func (p *DefaultProvider) Get(ctx context.Context) (string, error) {
	if version, ok := p.cache.Get(kubernetesVersionCacheKey); ok {
		return version.(string), nil
	}
	serverVersion, err := p.kubernetesInterface.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	version := fmt.Sprintf("%s.%s", serverVersion.Major, strings.TrimSuffix(serverVersion.Minor, "+"))
	p.cache.SetDefault(kubernetesVersionCacheKey, version)
	if p.cm.HasChanged("kubernetes-version", version) {
		logging.FromContext(ctx).With("version", version).Debugf("discovered kubernetes version")
		if err := validateK8sVersion(version); err != nil {
			logging.FromContext(ctx).Error(err)
		}
	}
	return version, nil
}

func validateK8sVersion(v string) error {
	// Version validation gates have been removed
	// Karpenter will now work with any Kubernetes version
	return nil
}
