package duplosdk

import (
	"fmt"
)

// DuploEmrClusterRequest is a Duplo SDK object that represents a emr cluster
type DuploEmrClusterRequest struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID     string `json:"-"`
	Arn          string `json:"Arn,omitempty"`
	Name         string `json:"Name,omitempty"`
	ReleaseLabel string `json:"ReleaseLabel,omitempty"`
	Status       string `json:"Status,omitempty"`
	LogUri       string `json:"LogUri,omitempty"`

	JobFlowId    string `json:"JobFlowId,omitempty"`
	ResourceType int    `json:"ResourceType,omitempty"`
	CustomAmiId  string `json:"CustomAmiId,omitempty"`

	EbsRootVolumeSize           int    `json:"EbsRootVolumeSize,omitempty"`
	StepConcurrencyLevel        int    `json:"StepConcurrencyLevel,omitempty"`
	ScaleDownBehavior           string `json:"ScaleDownBehavior,omitempty"`
	TerminationProtection       bool   `json:"TerminationProtection,omitempty"`
	KeepJobFlowAliveWhenNoSteps bool   `json:"KeepJobFlowAliveWhenNoSteps,omitempty"`
	VisibleToAllUsers           bool   `json:"VisibleToAllUsers,omitempty"`

	//ec2
	MasterInstanceType string `json:"MasterInstanceType,omitempty"`
	SlaveInstanceType  string `json:"SlaveInstanceType,omitempty"`
	InstanceCount      int    `json:"InstanceCount,omitempty"`
	//can we use this for subnetid selection
	Zone int `json:"Zone,omitempty"`

	//JSON str
	Applications           string `json:"Applications,omitempty"`
	Steps                  string `json:"Steps,omitempty"`
	Configurations         string `json:"Configurations,omitempty"`
	BootstrapActions       string `json:"BootstrapActions,omitempty"`
	JobFlowInstancesConfig string `json:"JobFlowInstancesConfig,omitempty"`
	//JSON str
	AdditionalInfo       string `json:"AdditionalInfo,omitempty"`
	ManagedScalingPolicy string `json:"ManagedScalingPolicy,omitempty"`
	InstanceGroups       string `json:"InstanceGroups,omitempty"`
	InstanceFleets       string `json:"InstanceFleets,omitempty"`

	//== debug ec2-attributes
	MetaData string `json:"MetaData,omitempty"`
	State    string `json:"State,omitempty"`
}

type DuploEmrClusterSummary struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID  string `json:"-"`
	Name      string `json:"Name,omitempty"`
	Arn       string `json:"Arn,omitempty"`
	Status    string `json:"Status,omitempty"`
	JobFlowId string `json:"JobFlowId,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// DuploEmrClusterGet retrieves an emr cluster via the Duplo API
func (c *Client) DuploEmrClusterGet(tenantID string, name string) (*DuploEmrClusterRequest, ClientError) {
	rp := DuploEmrClusterRequest{}
	err := c.getAPI(
		fmt.Sprintf("DuploEmrClusterGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/emrCluster/%s", tenantID, name),
		&rp)
	rp.TenantID = tenantID
	return &rp, err
}

// DuploEmrClusterGetList retrieves a emr cluster via the Duplo API
func (c *Client) DuploEmrClusterGetList(tenantID string) (*[]DuploEmrClusterSummary, ClientError) {
	// todo: not tested data
	rp := []DuploEmrClusterSummary{}
	err := c.getAPI(
		fmt.Sprintf("DuploEmrClusterGet(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/emrCluster", tenantID),
		&rp)
	return &rp, err
}
