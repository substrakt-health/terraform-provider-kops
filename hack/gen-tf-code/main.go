package main

import (
	"fmt"
	"log"
	"path"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/eddycharly/terraform-provider-kops/pkg/api/config"
	"github.com/eddycharly/terraform-provider-kops/pkg/api/datasources"
	"github.com/eddycharly/terraform-provider-kops/pkg/api/kube"
	"github.com/eddycharly/terraform-provider-kops/pkg/api/resources"
	"github.com/eddycharly/terraform-provider-kops/pkg/api/utils"
	"k8s.io/kops/pkg/apis/kops"
)

var mappings = map[string]string{
	"github.com/eddycharly/terraform-provider-kops/pkg/api/config":      "config",
	"github.com/eddycharly/terraform-provider-kops/pkg/api/datasources": "datasources",
	"github.com/eddycharly/terraform-provider-kops/pkg/api/kube":        "kube",
	"github.com/eddycharly/terraform-provider-kops/pkg/api/resources":   "resources",
	"github.com/eddycharly/terraform-provider-kops/pkg/api/utils":       "utils",
	"k8s.io/kops/pkg/apis/kops":                                         "kops",
}

func buildDoc(t reflect.Type, p, header, footer string, funcMaps ...template.FuncMap) {
	fileName := toSnakeCase(fieldName(t.Name())) + ".md"
	executeTemplate(t, fmt.Sprintf(docs, header, footer), p, fileName, funcMaps...)
}

func buildSchema(t reflect.Type, p, scope string, funcMaps ...template.FuncMap) {
	fileName := fmt.Sprintf("%s_%s.generated.go", scope, t.Name())
	executeTemplate(t, schemas, p, fileName, funcMaps...)
}

func buildTests(t reflect.Type, p, scope string, funcMaps ...template.FuncMap) {
	fileName := fmt.Sprintf("%s_%s.generated_test.go", scope, t.Name())
	executeTemplate(t, tests, p, fileName, funcMaps...)
}

type generated struct {
	t reflect.Type
	o *options
}

func generate(i interface{}, opts ...func(o *options)) generated {
	t := reflect.TypeOf(i)
	o := newOptions()
	for _, opt := range opts {
		opt(o)
	}
	if err := o.verify(t); err != nil {
		panic(err)
	}
	return generated{
		t: t,
		o: o,
	}
}

func build(scope, docs string, parser *parser, g ...generated) {
	o := map[reflect.Type]*options{}
	for _, gen := range g {
		o[gen.t] = gen.o
	}
	for _, gen := range g {
		funcMaps := []template.FuncMap{
			reflectFuncs(gen.t, mappings, parser),
			optionFuncs(scope == "DataSource", o),
			schemaFuncs(scope),
			docFuncs(parser, o),
			sprig.TxtFuncMap(),
		}
		buildSchema(gen.t, path.Join("pkg/schemas", mappings[gen.t.PkgPath()]), scope, funcMaps...)
		buildTests(gen.t, path.Join("pkg/schemas", mappings[gen.t.PkgPath()]), scope, funcMaps...)
		if gen.o.doc != nil {
			buildDoc(gen.t, docs, gen.o.doc.header, gen.o.doc.footer, funcMaps...)
		}
	}
}

func main() {
	log.Println("loading packages...")
	parser, err := initParser(
		"github.com/eddycharly/terraform-provider-kops/pkg/api/config",
		"github.com/eddycharly/terraform-provider-kops/pkg/api/datasources",
		"github.com/eddycharly/terraform-provider-kops/pkg/api/kube",
		"github.com/eddycharly/terraform-provider-kops/pkg/api/resources",
	)
	if err != nil {
		panic(err)
	}
	log.Println("generating schemas, expanders and flatteners...")
	build(
		"Resource",
		"docs/resources/",
		parser,
		generate(resources.Cluster{},
			version(1),
			required("Name", "AdminSshKey"),
			computedOnly("Revision"),
			sensitive("AdminSshKey"),
			forceNew("Name"),
			doc(resourceClusterHeader, resourceClusterFooter),
		),
		generate(resources.InstanceGroup{},
			version(1),
			required("ClusterName", "Name"),
			forceNew("ClusterName", "Name"),
			computedOnly("Revision"),
			doc(resourceInstanceGroupHeader, resourceInstanceGroupFooter),
		),
		generate(resources.ClusterUpdater{},
			required("ClusterName"),
			computedOnly("Revision"),
			doc(resourceClusterUpdaterHeader, ""),
		),
		generate(utils.RollingUpdateOptions{},
			noSchema(),
		),
		generate(resources.ClusterSecrets{},
			sensitive("DockerConfig"),
		),
		generate(resources.ValidateOptions{}),
		generate(utils.ValidateOptions{},
			noSchema(),
		),
		generate(resources.ApplyOptions{}),
		generate(kops.ClusterSpec{},
			noSchema(),
			exclude("GossipConfig", "DNSControllerGossipConfig", "Target"),
			rename("Subnets", "Subnet"),
			rename("EtcdClusters", "EtcdCluster"),
			required("CloudProvider", "Subnets", "NetworkID", "Topology", "EtcdClusters", "Networking"),
			computed("MasterPublicName", "MasterInternalName", "ConfigBase", "NetworkCIDR", "NonMasqueradeCIDR", "IAM"),
		),
		generate(kops.InstanceMetadataOptions{}),
		generate(kops.NodeTerminationHandlerConfig{}),
		generate(kops.MetricsServerConfig{}),
		generate(kops.ClusterAutoscalerConfig{}),
		generate(kops.AddonSpec{},
			required("Manifest"),
		),
		generate(kops.EgressProxySpec{},
			required("HTTPProxy"),
		),
		generate(kops.HTTPProxy{},
			required("Host", "Port"),
		),
		generate(kops.ContainerdConfig{}),
		generate(kops.PackagesConfig{}),
		generate(kops.DockerConfig{}),
		generate(kops.KubeDNSConfig{}),
		generate(kops.KubeAPIServerConfig{},
			nullable("AnonymousAuth"),
		),
		generate(kops.KubeControllerManagerConfig{}),
		generate(kops.CloudControllerManagerConfig{}),
		generate(kops.KubeSchedulerConfig{}),
		generate(kops.KubeProxyConfig{}),
		generate(kops.KubeletConfigSpec{},
			nullable("AnonymousAuth"),
		),
		generate(kops.CloudConfiguration{}),
		generate(kops.ExternalDNSConfig{}),
		generate(kops.OpenstackConfiguration{}),
		generate(kops.OpenstackLoadbalancerConfig{}),
		generate(kops.OpenstackMonitor{}),
		generate(kops.OpenstackRouter{}),
		generate(kops.OpenstackBlockStorageConfig{}),
		generate(kops.LeaderElectionConfiguration{}),
		generate(kops.NodeLocalDNSConfig{}),
		generate(kops.AuthenticationSpec{}),
		generate(kops.AuthorizationSpec{}),
		generate(kops.NodeAuthorizationSpec{}),
		generate(kops.Assets{}),
		generate(kops.IAMSpec{}),
		generate(kops.KopeioAuthenticationSpec{}),
		generate(kops.AwsAuthenticationSpec{}),
		generate(kops.AlwaysAllowAuthorizationSpec{}),
		generate(kops.RBACAuthorizationSpec{}),
		generate(kops.NodeAuthorizerSpec{}),
		generate(kops.InstanceGroupSpec{},
			noSchema(),
			required("Role", "MinSize", "MaxSize", "MachineType", "Subnets"),
			computed("Image"),
		),
		generate(kops.AccessSpec{}),
		generate(kops.DNSAccessSpec{}),
		generate(kops.LoadBalancerAccessSpec{},
			required("Type"),
		),
		generate(kops.EtcdClusterSpec{},
			required("Name", "Members"),
			rename("Members", "Member"),
		),
		generate(kops.EtcdBackupSpec{},
			required("BackupStore", "Image"),
		),
		generate(kops.EtcdManagerSpec{}),
		generate(kops.EtcdMemberSpec{},
			required("Name", "InstanceGroup"),
		),
		generate(kops.EnvVar{},
			required("Name"),
		),
		generate(kops.ClusterSubnetSpec{},
			required("Name", "ProviderID", "Type", "Zone"),
			computed("CIDR"),
		),
		generate(kops.TopologySpec{},
			required("Masters", "Nodes", "DNS"),
		),
		generate(kops.BastionSpec{},
			required("BastionPublicName"),
		),
		generate(kops.BastionLoadBalancerSpec{},
			required("AdditionalSecurityGroups"),
		),
		generate(kops.DNSSpec{},
			required("Type"),
		),
		generate(kops.NetworkingSpec{}),
		generate(kops.ClassicNetworkingSpec{}),
		generate(kops.KubenetNetworkingSpec{}),
		generate(kops.ExternalNetworkingSpec{}),
		generate(kops.CNINetworkingSpec{}),
		generate(kops.KopeioNetworkingSpec{}),
		generate(kops.WeaveNetworkingSpec{}),
		generate(kops.FlannelNetworkingSpec{}),
		generate(kops.CalicoNetworkingSpec{}),
		generate(kops.CanalNetworkingSpec{}),
		generate(kops.KuberouterNetworkingSpec{}),
		generate(kops.RomanaNetworkingSpec{}),
		generate(kops.AmazonVPCNetworkingSpec{}),
		generate(kops.CiliumNetworkingSpec{}),
		generate(kops.HubbleSpec{}),
		generate(kops.LyftVPCNetworkingSpec{}),
		generate(kops.GCENetworkingSpec{}),
		generate(kops.VolumeSpec{},
			required("Device"),
		),
		generate(kops.VolumeMountSpec{},
			required("Device", "Filesystem", "Path"),
		),
		generate(kops.MixedInstancesPolicySpec{},
			nullable("OnDemandBase", "OnDemandAboveBase"),
		),
		generate(kops.UserData{},
			required("Name", "Type", "Content"),
		),
		generate(kops.LoadBalancer{}),
		generate(kops.IAMProfileSpec{},
			required("Profile"),
		),
		generate(kops.HookSpec{},
			required("Name"),
		),
		generate(kops.ExecContainerAction{},
			required("Image"),
		),
		generate(kops.FileAssetSpec{},
			required("Name", "Path", "Content"),
		),
		generate(kops.RollingUpdate{}),
		// 1.20
		generate(resources.RollingUpdateOptions{}),
		generate(kops.AzureConfiguration{}),
		generate(kops.AWSEBSCSIDriver{}),
		generate(kops.NTPConfig{}),
		generate(kops.CertManagerConfig{}),
		generate(kops.AWSLoadBalancerControllerConfig{}),
		generate(kops.GossipConfigSecondary{}),
		generate(kops.LoadBalancerSubnetSpec{}),
		generate(kops.DNSControllerGossipConfigSecondary{}),
		generate(kops.OpenstackNetwork{}),
		// 1.21
		generate(kops.WarmPoolSpec{}),
		generate(kops.ServiceAccountIssuerDiscoveryConfig{}),
		generate(kops.SnapshotControllerConfig{}),
		generate(kops.ServiceAccountExternalPermission{}),
		generate(kops.AWSPermission{}),
	)
	build(
		"Config",
		"docs/guides/",
		parser,
		generate(config.Provider{},
			required("StateStore"),
			doc(configProviderHeader, ""),
		),
		generate(config.Aws{}),
		generate(config.AwsAssumeRole{}),
		generate(config.Openstack{}),
		generate(config.Klog{},
			nullable("Verbosity"),
		),
	)
	build(
		"DataSource",
		"docs/data-sources/",
		parser,
		generate(datasources.KubeConfig{},
			required("ClusterName"),
			computed("Admin", "Internal"),
			doc(dataKubeConfigHeader, ""),
		),
		generate(datasources.ClusterStatus{},
			required("ClusterName"),
			doc(dataClusterStatusHeader, ""),
		),
		generate(resources.Cluster{},
			version(1),
			required("Name"),
			exclude("Revision"),
			doc(dataClusterHeader, ""),
		),
		generate(resources.InstanceGroup{},
			version(1),
			required("ClusterName", "Name"),
			exclude("Revision"),
			doc(dataInstanceGroupHeader, ""),
		),
		generate(resources.ClusterSecrets{},
			sensitive("DockerConfig"),
		),
		generate(kube.Config{},
			noSchema(),
			sensitive("KubeBearerToken", "KubePassword", "CaCert", "ClientCert", "ClientKey"),
		),
		generate(kops.ClusterSpec{},
			exclude("GossipConfig", "DNSControllerGossipConfig", "Target"),
			rename("Subnets", "Subnet"),
			rename("EtcdClusters", "EtcdCluster"),
		),
		generate(kops.InstanceMetadataOptions{}),
		generate(kops.NodeTerminationHandlerConfig{}),
		generate(kops.MetricsServerConfig{}),
		generate(kops.ClusterAutoscalerConfig{}),
		generate(kops.AddonSpec{}),
		generate(kops.GossipConfig{}),
		generate(kops.ClusterSubnetSpec{}),
		generate(kops.TopologySpec{}),
		generate(kops.DNSControllerGossipConfig{}),
		generate(kops.EgressProxySpec{}),
		generate(kops.EtcdClusterSpec{},
			rename("Members", "Member"),
		),
		generate(kops.ContainerdConfig{}),
		generate(kops.PackagesConfig{}),
		generate(kops.DockerConfig{}),
		generate(kops.KubeDNSConfig{}),
		generate(kops.KubeAPIServerConfig{},
			nullable("AnonymousAuth"),
		),
		generate(kops.KubeControllerManagerConfig{}),
		generate(kops.CloudControllerManagerConfig{}),
		generate(kops.KubeSchedulerConfig{}),
		generate(kops.KubeProxyConfig{}),
		generate(kops.CloudConfiguration{}),
		generate(kops.ExternalDNSConfig{}),
		generate(kops.NetworkingSpec{}),
		generate(kops.AccessSpec{}),
		generate(kops.AuthenticationSpec{}),
		generate(kops.DNSAccessSpec{}),
		generate(kops.LoadBalancerAccessSpec{}),
		generate(kops.KopeioAuthenticationSpec{}),
		generate(kops.AwsAuthenticationSpec{}),
		generate(kops.OpenstackConfiguration{}),
		generate(kops.LeaderElectionConfiguration{}),
		generate(kops.AuthorizationSpec{}),
		generate(kops.NodeAuthorizationSpec{}),
		generate(kops.Assets{}),
		generate(kops.IAMSpec{}),
		generate(kops.AlwaysAllowAuthorizationSpec{}),
		generate(kops.RBACAuthorizationSpec{}),
		generate(kops.HTTPProxy{}),
		generate(kops.EtcdMemberSpec{}),
		generate(kops.EtcdBackupSpec{}),
		generate(kops.EtcdManagerSpec{}),
		generate(kops.NodeLocalDNSConfig{}),
		generate(kops.ClassicNetworkingSpec{}),
		generate(kops.KubenetNetworkingSpec{}),
		generate(kops.ExternalNetworkingSpec{}),
		generate(kops.CNINetworkingSpec{}),
		generate(kops.EnvVar{}),
		generate(kops.KopeioNetworkingSpec{}),
		generate(kops.WeaveNetworkingSpec{}),
		generate(kops.FlannelNetworkingSpec{}),
		generate(kops.CalicoNetworkingSpec{}),
		generate(kops.CanalNetworkingSpec{}),
		generate(kops.KuberouterNetworkingSpec{}),
		generate(kops.RomanaNetworkingSpec{}),
		generate(kops.AmazonVPCNetworkingSpec{}),
		generate(kops.CiliumNetworkingSpec{}),
		generate(kops.HubbleSpec{}),
		generate(kops.LyftVPCNetworkingSpec{}),
		generate(kops.GCENetworkingSpec{}),
		generate(kops.NodeAuthorizerSpec{}),
		generate(kops.OpenstackLoadbalancerConfig{}),
		generate(kops.OpenstackMonitor{}),
		generate(kops.OpenstackRouter{}),
		generate(kops.OpenstackBlockStorageConfig{}),
		generate(kops.BastionSpec{}),
		generate(kops.DNSSpec{}),
		generate(kops.BastionLoadBalancerSpec{}),
		generate(kops.InstanceGroupSpec{},
			noSchema(),
		),
		generate(kops.VolumeSpec{}),
		generate(kops.VolumeMountSpec{}),
		generate(kops.HookSpec{}),
		generate(kops.FileAssetSpec{}),
		generate(kops.KubeletConfigSpec{},
			nullable("AnonymousAuth"),
		),
		generate(kops.MixedInstancesPolicySpec{},
			nullable("OnDemandBase", "OnDemandAboveBase"),
		),
		generate(kops.UserData{}),
		generate(kops.LoadBalancer{}),
		generate(kops.IAMProfileSpec{}),
		generate(kops.RollingUpdate{}),
		generate(kops.ExecContainerAction{}),
		// 1.20
		generate(kops.AzureConfiguration{}),
		generate(kops.AWSEBSCSIDriver{}),
		generate(kops.NTPConfig{}),
		generate(kops.CertManagerConfig{}),
		generate(kops.AWSLoadBalancerControllerConfig{}),
		generate(kops.GossipConfigSecondary{}),
		generate(kops.LoadBalancerSubnetSpec{}),
		generate(kops.DNSControllerGossipConfigSecondary{}),
		generate(kops.OpenstackNetwork{}),
		// 1.21
		generate(kops.WarmPoolSpec{}),
		generate(kops.ServiceAccountIssuerDiscoveryConfig{}),
		generate(kops.SnapshotControllerConfig{}),
		generate(kops.ServiceAccountExternalPermission{}),
		generate(kops.AWSPermission{}),
	)
}
