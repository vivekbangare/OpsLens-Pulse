package metadata

import (
	"opslense-pulse/shared"
)

func Build(
	agentID string,
	hostname string,
	osName string,
	version string,
	privateIP string,
	publicIP string,
	k8sIP string,
	provider string,
	region string,
	userTags map[string]string,
) shared.AgentInfo {

	return shared.AgentInfo{
		AgentID:   agentID,
		Hostname:  hostname,
		OS:        osName,
		Version:   version,
		PrivateIP: privateIP,
		PublicIP:  publicIP,
		K8sNodeIP: k8sIP,
		Tags:      userTags,
		SystemTags: map[string]string{
			"cloud_provider": provider,
			"cloud_region":   region,
		},
	}
}
