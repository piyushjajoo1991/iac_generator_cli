package infra

import (
	"fmt"
	"net"
	"strings"
)

// Infrastructure represents the top-level container for all AWS resources
type Infrastructure struct {
	Name      string        `json:"name" yaml:"name"`
	Region    string        `json:"region" yaml:"region"`
	VPCs      []*VPC        `json:"vpcs,omitempty" yaml:"vpcs,omitempty"`
	Resources []interface{} `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// NewInfrastructure creates a new Infrastructure with reasonable defaults
func NewInfrastructure(name string) *Infrastructure {
	return &Infrastructure{
		Name:      name,
		Region:    "us-east-1", // Default to us-east-1
		VPCs:      []*VPC{},
		Resources: []interface{}{},
	}
}

// Validate ensures the Infrastructure configuration is consistent
func (i *Infrastructure) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("infrastructure name cannot be empty")
	}
	
	for _, vpc := range i.VPCs {
		if err := vpc.Validate(); err != nil {
			return fmt.Errorf("vpc %s validation failed: %w", vpc.Name, err)
		}
	}
	
	return nil
}

// AddVPC adds a VPC to the infrastructure
func (i *Infrastructure) AddVPC(vpc *VPC) {
	i.VPCs = append(i.VPCs, vpc)
}

// AddResource adds a generic resource to the infrastructure
func (i *Infrastructure) AddResource(resource interface{}) {
	i.Resources = append(i.Resources, resource)
}

// String returns a string representation of the Infrastructure
func (i *Infrastructure) String() string {
	return fmt.Sprintf("Infrastructure{Name: %s, Region: %s, VPCs: %d, Resources: %d}",
		i.Name, i.Region, len(i.VPCs), len(i.Resources))
}

// VPC represents an AWS Virtual Private Cloud
type VPC struct {
	Name              string              `json:"name" yaml:"name"`
	CIDR              string              `json:"cidr" yaml:"cidr"`
	Region            string              `json:"region" yaml:"region"`
	EnableDNSSupport  bool                `json:"enable_dns_support" yaml:"enable_dns_support"`
	EnableDNSHostname bool                `json:"enable_dns_hostname" yaml:"enable_dns_hostname"`
	Subnets           []*Subnet           `json:"subnets,omitempty" yaml:"subnets,omitempty"`
	InternetGateways  []*InternetGateway  `json:"internet_gateways,omitempty" yaml:"internet_gateways,omitempty"`
	NATGateways       []*NATGateway       `json:"nat_gateways,omitempty" yaml:"nat_gateways,omitempty"`
	Tags              map[string]string   `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// NewVPC creates a new VPC with reasonable defaults
func NewVPC(name string, cidr string, region string) *VPC {
	return &VPC{
		Name:              name,
		CIDR:              cidr,
		Region:            region,
		EnableDNSSupport:  true,
		EnableDNSHostname: true,
		Subnets:           []*Subnet{},
		InternetGateways:  []*InternetGateway{},
		NATGateways:       []*NATGateway{},
		Tags:              map[string]string{"Name": name},
	}
}

// Validate ensures the VPC configuration is consistent
func (v *VPC) Validate() error {
	if v.Name == "" {
		return fmt.Errorf("vpc name cannot be empty")
	}
	
	// Validate CIDR block
	_, _, err := net.ParseCIDR(v.CIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR block format: %w", err)
	}
	
	// Validate subnets
	for _, subnet := range v.Subnets {
		if err := subnet.Validate(); err != nil {
			return fmt.Errorf("subnet %s validation failed: %w", subnet.Name, err)
		}
		
		// Ensure subnet CIDR is within VPC CIDR
		if !CIDRContains(v.CIDR, subnet.CIDR) {
			return fmt.Errorf("subnet CIDR %s is not within VPC CIDR %s", subnet.CIDR, v.CIDR)
		}
	}
	
	return nil
}

// AddSubnet adds a subnet to the VPC
func (v *VPC) AddSubnet(subnet *Subnet) {
	v.Subnets = append(v.Subnets, subnet)
}

// AddInternetGateway adds an internet gateway to the VPC
func (v *VPC) AddInternetGateway(gateway *InternetGateway) {
	v.InternetGateways = append(v.InternetGateways, gateway)
}

// AddNATGateway adds a NAT gateway to the VPC
func (v *VPC) AddNATGateway(gateway *NATGateway) {
	v.NATGateways = append(v.NATGateways, gateway)
}

// GetPublicSubnets returns all public subnets in the VPC
func (v *VPC) GetPublicSubnets() []*Subnet {
	var publicSubnets []*Subnet
	for _, subnet := range v.Subnets {
		if subnet.IsPublic {
			publicSubnets = append(publicSubnets, subnet)
		}
	}
	return publicSubnets
}

// GetPrivateSubnets returns all private subnets in the VPC
func (v *VPC) GetPrivateSubnets() []*Subnet {
	var privateSubnets []*Subnet
	for _, subnet := range v.Subnets {
		if !subnet.IsPublic {
			privateSubnets = append(privateSubnets, subnet)
		}
	}
	return privateSubnets
}

// String returns a string representation of the VPC
func (v *VPC) String() string {
	return fmt.Sprintf("VPC{Name: %s, CIDR: %s, Region: %s, Subnets: %d}",
		v.Name, v.CIDR, v.Region, len(v.Subnets))
}

// Subnet represents an AWS subnet within a VPC
type Subnet struct {
	Name             string            `json:"name" yaml:"name"`
	CIDR             string            `json:"cidr" yaml:"cidr"`
	AvailabilityZone string            `json:"availability_zone" yaml:"availability_zone"`
	IsPublic         bool              `json:"is_public" yaml:"is_public"`
	Tags             map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// NewSubnet creates a new subnet with reasonable defaults
func NewSubnet(name string, cidr string, az string, isPublic bool) *Subnet {
	return &Subnet{
		Name:             name,
		CIDR:             cidr,
		AvailabilityZone: az,
		IsPublic:         isPublic,
		Tags:             map[string]string{"Name": name},
	}
}

// Validate ensures the Subnet configuration is consistent
func (s *Subnet) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("subnet name cannot be empty")
	}
	
	// Validate CIDR block
	_, _, err := net.ParseCIDR(s.CIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR block format: %w", err)
	}
	
	// Validate AZ format (not comprehensive, just a basic check)
	if !strings.HasPrefix(s.AvailabilityZone, "us-") && 
	   !strings.HasPrefix(s.AvailabilityZone, "eu-") && 
	   !strings.HasPrefix(s.AvailabilityZone, "ap-") && 
	   !strings.HasPrefix(s.AvailabilityZone, "sa-") && 
	   !strings.HasPrefix(s.AvailabilityZone, "ca-") && 
	   !strings.HasPrefix(s.AvailabilityZone, "af-") {
		return fmt.Errorf("invalid availability zone format: %s", s.AvailabilityZone)
	}
	
	return nil
}

// String returns a string representation of the Subnet
func (s *Subnet) String() string {
	subnetType := "Private"
	if s.IsPublic {
		subnetType = "Public"
	}
	return fmt.Sprintf("Subnet{Name: %s, CIDR: %s, AZ: %s, Type: %s}",
		s.Name, s.CIDR, s.AvailabilityZone, subnetType)
}

// CIDRContains checks if the parent CIDR contains the child CIDR
func CIDRContains(parentCIDR, childCIDR string) bool {
	_, parentNet, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return false
	}
	
	childIP, _, err := net.ParseCIDR(childCIDR)
	if err != nil {
		return false
	}
	
	return parentNet.Contains(childIP)
}

// InternetGateway represents an AWS Internet Gateway
type InternetGateway struct {
	Name   string            `json:"name" yaml:"name"`
	VPC    string            `json:"vpc" yaml:"vpc"`     // Reference to VPC name
	Tags   map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// NewInternetGateway creates a new Internet Gateway with reasonable defaults
func NewInternetGateway(name string, vpcName string) *InternetGateway {
	return &InternetGateway{
		Name:   name,
		VPC:    vpcName,
		Tags:   map[string]string{"Name": name},
	}
}

// Validate ensures the Internet Gateway configuration is consistent
func (ig *InternetGateway) Validate() error {
	if ig.Name == "" {
		return fmt.Errorf("internet gateway name cannot be empty")
	}
	
	if ig.VPC == "" {
		return fmt.Errorf("internet gateway must be associated with a VPC")
	}
	
	return nil
}

// String returns a string representation of the Internet Gateway
func (ig *InternetGateway) String() string {
	return fmt.Sprintf("InternetGateway{Name: %s, VPC: %s}", ig.Name, ig.VPC)
}

// NATGateway represents an AWS NAT Gateway
type NATGateway struct {
	Name            string            `json:"name" yaml:"name"`
	Subnet          string            `json:"subnet" yaml:"subnet"`         // Reference to subnet name
	AllocationID    string            `json:"allocation_id,omitempty" yaml:"allocation_id,omitempty"` // EIP allocation ID
	ConnectivityType string           `json:"connectivity_type" yaml:"connectivity_type"`         // "public" or "private"
	Tags            map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// NewNATGateway creates a new NAT Gateway with reasonable defaults
func NewNATGateway(name string, subnetName string) *NATGateway {
	return &NATGateway{
		Name:             name,
		Subnet:           subnetName,
		ConnectivityType: "public", // Default to public connectivity
		Tags:             map[string]string{"Name": name},
	}
}

// Validate ensures the NAT Gateway configuration is consistent
func (ng *NATGateway) Validate() error {
	if ng.Name == "" {
		return fmt.Errorf("nat gateway name cannot be empty")
	}
	
	if ng.Subnet == "" {
		return fmt.Errorf("nat gateway must be associated with a subnet")
	}
	
	if ng.ConnectivityType != "public" && ng.ConnectivityType != "private" {
		return fmt.Errorf("nat gateway connectivity type must be 'public' or 'private'")
	}
	
	// For public NAT Gateways, an EIP allocation ID is required
	if ng.ConnectivityType == "public" && ng.AllocationID == "" {
		return fmt.Errorf("public nat gateway requires an elastic IP allocation ID")
	}
	
	return nil
}

// String returns a string representation of the NAT Gateway
func (ng *NATGateway) String() string {
	return fmt.Sprintf("NATGateway{Name: %s, Subnet: %s, ConnectivityType: %s}",
		ng.Name, ng.Subnet, ng.ConnectivityType)
}

// EKSCluster represents an AWS EKS (Elastic Kubernetes Service) Cluster
type EKSCluster struct {
	Name                 string            `json:"name" yaml:"name"`
	Version              string            `json:"version" yaml:"version"`
	RoleARN              string            `json:"role_arn" yaml:"role_arn"`
	SubnetIDs            []string          `json:"subnet_ids" yaml:"subnet_ids"`     // References to subnet names
	EndpointPublicAccess bool              `json:"endpoint_public_access" yaml:"endpoint_public_access"`
	EndpointPrivateAccess bool             `json:"endpoint_private_access" yaml:"endpoint_private_access"`
	SecurityGroupIDs     []string          `json:"security_group_ids,omitempty" yaml:"security_group_ids,omitempty"`
	NodePools            []*NodePool       `json:"node_pools,omitempty" yaml:"node_pools,omitempty"`
	KubernetesNetworkConfig *KubernetesNetworkConfig `json:"kubernetes_network_config,omitempty" yaml:"kubernetes_network_config,omitempty"`
	Tags                 map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// KubernetesNetworkConfig represents the Kubernetes network configuration for an EKS cluster
type KubernetesNetworkConfig struct {
	ServiceCIDR string `json:"service_cidr,omitempty" yaml:"service_cidr,omitempty"`
	IPFamily    string `json:"ip_family,omitempty" yaml:"ip_family,omitempty"` // ipv4 or ipv6
}

// NewEKSCluster creates a new EKS Cluster with reasonable defaults
func NewEKSCluster(name string, version string, roleARN string, subnetIDs []string) *EKSCluster {
	return &EKSCluster{
		Name:                  name,
		Version:               version,
		RoleARN:               roleARN,
		SubnetIDs:             subnetIDs,
		EndpointPublicAccess:  true,
		EndpointPrivateAccess: false,
		NodePools:             []*NodePool{},
		KubernetesNetworkConfig: &KubernetesNetworkConfig{
			ServiceCIDR: "172.20.0.0/16", // Default service CIDR
			IPFamily:    "ipv4",          // Default to IPv4
		},
		Tags:                  map[string]string{"Name": name},
	}
}

// Validate ensures the EKS Cluster configuration is consistent
func (e *EKSCluster) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("eks cluster name cannot be empty")
	}
	
	if e.Version == "" {
		return fmt.Errorf("eks cluster version cannot be empty")
	}
	
	if e.RoleARN == "" {
		return fmt.Errorf("eks cluster role ARN cannot be empty")
	}
	
	if len(e.SubnetIDs) < 2 {
		return fmt.Errorf("eks cluster requires at least 2 subnets")
	}
	
	// Validate node pools if any
	for _, nodePool := range e.NodePools {
		if err := nodePool.Validate(); err != nil {
			return fmt.Errorf("node pool %s validation failed: %w", nodePool.Name, err)
		}
	}
	
	// Validate network config if provided
	if e.KubernetesNetworkConfig != nil {
		if e.KubernetesNetworkConfig.ServiceCIDR != "" {
			_, _, err := net.ParseCIDR(e.KubernetesNetworkConfig.ServiceCIDR)
			if err != nil {
				return fmt.Errorf("invalid service CIDR format: %w", err)
			}
		}
		
		if e.KubernetesNetworkConfig.IPFamily != "" && 
		   e.KubernetesNetworkConfig.IPFamily != "ipv4" && 
		   e.KubernetesNetworkConfig.IPFamily != "ipv6" {
			return fmt.Errorf("ip family must be 'ipv4' or 'ipv6'")
		}
	}
	
	return nil
}

// AddNodePool adds a node pool to the EKS cluster
func (e *EKSCluster) AddNodePool(nodePool *NodePool) {
	e.NodePools = append(e.NodePools, nodePool)
}

// String returns a string representation of the EKS Cluster
func (e *EKSCluster) String() string {
	return fmt.Sprintf("EKSCluster{Name: %s, Version: %s, Subnets: %d, NodePools: %d}",
		e.Name, e.Version, len(e.SubnetIDs), len(e.NodePools))
}

// NodePool represents an AWS EKS Node Group
type NodePool struct {
	Name              string            `json:"name" yaml:"name"`
	InstanceTypes     []string          `json:"instance_types" yaml:"instance_types"`
	DiskSize          int               `json:"disk_size" yaml:"disk_size"`
	DesiredSize       int               `json:"desired_size" yaml:"desired_size"`
	MinSize           int               `json:"min_size" yaml:"min_size"`
	MaxSize           int               `json:"max_size" yaml:"max_size"`
	SubnetIDs         []string          `json:"subnet_ids" yaml:"subnet_ids"`     // References to subnet names
	NodeRoleARN       string            `json:"node_role_arn" yaml:"node_role_arn"`
	AMIType           string            `json:"ami_type,omitempty" yaml:"ami_type,omitempty"`     // AL2_x86_64, AL2_x86_64_GPU, etc.
	CapacityType      string            `json:"capacity_type,omitempty" yaml:"capacity_type,omitempty"` // ON_DEMAND or SPOT
	Labels            map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Tags              map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// NewNodePool creates a new Node Pool with reasonable defaults
func NewNodePool(name string, nodeRoleARN string, subnetIDs []string, instanceTypes []string, desiredSize int) *NodePool {
	return &NodePool{
		Name:          name,
		InstanceTypes: instanceTypes,
		DiskSize:      20,            // Default to 20 GB
		DesiredSize:   desiredSize,
		MinSize:       desiredSize,   // Default to same as desired
		MaxSize:       desiredSize*2, // Default to twice the desired size
		SubnetIDs:     subnetIDs,
		NodeRoleARN:   nodeRoleARN,
		AMIType:       "AL2_x86_64",  // Default to Amazon Linux 2
		CapacityType:  "ON_DEMAND",   // Default to On-Demand instances
		Labels:        map[string]string{},
		Tags:          map[string]string{"Name": name},
	}
}

// Validate ensures the Node Pool configuration is consistent
func (np *NodePool) Validate() error {
	if np.Name == "" {
		return fmt.Errorf("node pool name cannot be empty")
	}
	
	if len(np.InstanceTypes) == 0 {
		return fmt.Errorf("node pool must have at least one instance type")
	}
	
	if np.NodeRoleARN == "" {
		return fmt.Errorf("node pool role ARN cannot be empty")
	}
	
	if len(np.SubnetIDs) == 0 {
		return fmt.Errorf("node pool must have at least one subnet")
	}
	
	if np.MinSize < 0 {
		return fmt.Errorf("node pool min size cannot be negative")
	}
	
	if np.MaxSize < np.MinSize {
		return fmt.Errorf("node pool max size cannot be less than min size")
	}
	
	if np.DesiredSize < np.MinSize || np.DesiredSize > np.MaxSize {
		return fmt.Errorf("node pool desired size must be between min size and max size")
	}
	
	if np.AMIType != "" && 
	   np.AMIType != "AL2_x86_64" && 
	   np.AMIType != "AL2_x86_64_GPU" && 
	   np.AMIType != "AL2_ARM_64" && 
	   np.AMIType != "CUSTOM" {
		return fmt.Errorf("invalid AMI type: %s", np.AMIType)
	}
	
	if np.CapacityType != "" && 
	   np.CapacityType != "ON_DEMAND" && 
	   np.CapacityType != "SPOT" {
		return fmt.Errorf("capacity type must be 'ON_DEMAND' or 'SPOT'")
	}
	
	if np.DiskSize < 20 {
		return fmt.Errorf("disk size must be at least 20 GB")
	}
	
	return nil
}

// String returns a string representation of the Node Pool
func (np *NodePool) String() string {
	return fmt.Sprintf("NodePool{Name: %s, InstanceTypes: %v, Size: %d-%d-%d, Subnets: %d}",
		np.Name, np.InstanceTypes, np.MinSize, np.DesiredSize, np.MaxSize, len(np.SubnetIDs))
}