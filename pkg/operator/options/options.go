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

package options

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	coreoptions "sigs.k8s.io/karpenter/pkg/operator/options"
	"sigs.k8s.io/karpenter/pkg/utils/env"

	"github.com/cloudpilot-ai/karpenter-provider-alibabacloud/pkg/utils"
)

func init() {
	coreoptions.Injectables = append(coreoptions.Injectables, &Options{})
}

type optionsKey struct{}

type Options struct {
	ClusterID               string
	RegionID                string
	AliNetwork              string
	VMMemoryOverheadPercent float64
	Interruption            bool
	TelemetryShare          bool
	APGCreationQPS          int
	ClusterType             string
}

func (o *Options) AddFlags(fs *coreoptions.FlagSet) {
	fs.StringVar(&o.ClusterID, "cluster-id", env.WithDefaultString("CLUSTER_ID", ""), "The external kubernetes cluster id for new nodes to connect with.")
	fs.StringVar(&o.RegionID, "region-id", env.WithDefaultString("REGION_ID", ""), "Region of the kubernetes cluster. If not set, automatically obtain the host region.")
	fs.StringVar(&o.AliNetwork, "alibaba-cloud-network", env.WithDefaultString("ALIBABA_CLOUD_NETWORK", ""), "Network type for AlibabaCloud endpoint. If not set, use public.")
	// TODO: for different OS, the overhead is different, find a way to fix this.
	fs.Float64Var(&o.VMMemoryOverheadPercent, "vm-memory-overhead-percent", utils.WithDefaultFloat64("VM_MEMORY_OVERHEAD_PERCENT", 0.065), "The VM memory overhead as a percent that will be subtracted from the total memory for all instance types.")
	fs.BoolVar(&o.Interruption, "interruption", env.WithDefaultBool("INTERRUPTION", true), "Enable interruption handling.")
	fs.BoolVar(&o.TelemetryShare, "telemetry-share", env.WithDefaultBool("TELEMETRY_SHARE", true), "Enable telemetry sharing.")
	fs.IntVar(&o.APGCreationQPS, "apg-qps", int(env.WithDefaultInt64("APG_CREATION_QPS", 100)), "The QPS limit for creating AutoProvisionGroup.")
	fs.StringVar(&o.ClusterType, "cluster-type", env.WithDefaultString("CLUSTER_TYPE", "ACKManaged"), "Type of cluster, which specifies the method to generate userdata. The default is ACKManaged, with an option for Custom configuration. If your cluster-type is not default or ACKManaged, you need to add taint(karpenter.sh/unregistered:NoExecute) before the node is ready")
}

func (o *Options) Parse(fs *coreoptions.FlagSet, args ...string) error {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		return fmt.Errorf("parsing flags, %w", err)
	}
	if err := o.Validate(); err != nil {
		return fmt.Errorf("validating options, %w", err)
	}
	return nil
}

func (o *Options) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, o)
}

func ToContext(ctx context.Context, opts *Options) context.Context {
	return context.WithValue(ctx, optionsKey{}, opts)
}

func FromContext(ctx context.Context) *Options {
	retval := ctx.Value(optionsKey{})
	if retval == nil {
		return nil
	}
	return retval.(*Options)
}
