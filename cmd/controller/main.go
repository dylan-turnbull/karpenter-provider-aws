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

package main

import (
	"os"
	"strings"
	"time"

	"github.com/samber/lo"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"k8s.io/klog/v2"

	"github.com/aws/karpenter-provider-aws/pkg/cloudprovider"
	"github.com/aws/karpenter-provider-aws/pkg/controllers"
	"github.com/aws/karpenter-provider-aws/pkg/operator"
	"github.com/aws/karpenter-provider-aws/pkg/webhooks"

	"sigs.k8s.io/karpenter/pkg/cloudprovider/metrics"
	corecontrollers "sigs.k8s.io/karpenter/pkg/controllers"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	coreoperator "sigs.k8s.io/karpenter/pkg/operator"
	corewebhooks "sigs.k8s.io/karpenter/pkg/webhooks"
)

func getSocketPath() string {
	agentURL := os.Getenv("DD_TRACE_AGENT_URL")
	if strings.HasPrefix(agentURL, "unix://") {
		return strings.TrimPrefix(agentURL, "unix://")
	}
	return ""
}

func waitForDDSocket(socketPath string, maxWait time.Duration) bool {
	if socketPath == "" {
		return true
	}

	deadline := time.Now().Add(maxWait)
	klog.Infof("Waiting for Datadog APM socket at %s", socketPath)

	for time.Now().Before(deadline) {
		if info, err := os.Stat(socketPath); err == nil && (info.Mode()&os.ModeSocket) != 0 {
			klog.Infof("Datadog APM socket is ready")
			return true
		}
		time.Sleep(1 * time.Second)
	}

	klog.Warningf("Datadog APM socket not available after %v, starting without tracing", maxWait)
	return false
}

func main() {
	socketPath := getSocketPath()

	if os.Getenv("DD_TRACE_ENABLED") == "true" && waitForDDSocket(socketPath, 60*time.Second) {
		tracer.Start(
			tracer.WithService("karpenter"),
		)
		defer tracer.Stop()
	}

	ctx, op := operator.NewOperator(coreoperator.NewOperator())
	awsCloudProvider := cloudprovider.New(
		op.InstanceTypesProvider,
		op.InstanceProvider,
		op.EventRecorder,
		op.GetClient(),
		op.AMIProvider,
		op.SecurityGroupProvider,
		op.SubnetProvider,
	)
	lo.Must0(op.AddHealthzCheck("cloud-provider", awsCloudProvider.LivenessProbe))
	cloudProvider := metrics.Decorate(awsCloudProvider)

	op.
		WithControllers(ctx, corecontrollers.NewControllers(
			op.Clock,
			op.GetClient(),
			state.NewCluster(op.Clock, op.GetClient(), cloudProvider),
			op.EventRecorder,
			cloudProvider,
		)...).
		WithWebhooks(ctx, corewebhooks.NewWebhooks()...).
		WithControllers(ctx, controllers.NewControllers(
			ctx,
			op.Session,
			op.Clock,
			op.GetClient(),
			op.EventRecorder,
			op.UnavailableOfferingsCache,
			cloudProvider,
			op.SubnetProvider,
			op.SecurityGroupProvider,
			op.InstanceProfileProvider,
			op.InstanceProvider,
			op.PricingProvider,
			op.AMIProvider,
			op.LaunchTemplateProvider,
		)...).
		WithWebhooks(ctx, webhooks.NewWebhooks()...).
		Start(ctx)
}
