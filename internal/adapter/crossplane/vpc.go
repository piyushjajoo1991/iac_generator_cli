package crossplane

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// VPCGenerator generates Crossplane YAML for VPC and related networking components
type VPCGenerator struct {
	baseDir string
	vpcDir  string
}

// NewVPCGenerator creates a new VPC Generator
func NewVPCGenerator(baseDir string) *VPCGenerator {
	return &VPCGenerator{
		baseDir: baseDir,
		vpcDir:  filepath.Join(baseDir, "vpc"),
	}
}

// GenerateVPC generates a Crossplane VPC resource
func (g *VPCGenerator) GenerateVPC(name, cidrBlock string, enableDnsSupport, enableDnsHostnames bool) K8sObject {
	vpc := NewK8sObject("ec2.aws.crossplane.io/v1beta1", "VPC", name)
	
	// Add VPC specific properties
	vpc.AddNestedSpecField([]string{"forProvider", "cidrBlock"}, cidrBlock)
	vpc.AddNestedSpecField([]string{"forProvider", "enableDnsSupport"}, enableDnsSupport)
	vpc.AddNestedSpecField([]string{"forProvider", "enableDnsHostnames"}, enableDnsHostnames)
	
	// Add required provider config reference
	vpc.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	vpc.AddLabel("app.kubernetes.io/part-of", "network")
	vpc.AddLabel("app.kubernetes.io/component", "vpc")
	
	// Add useful annotations
	vpc.AddAnnotation("crossplane.io/external-name", name)
	
	return vpc
}

// GenerateSubnet generates a Crossplane Subnet resource
func (g *VPCGenerator) GenerateSubnet(name, vpcName, cidrBlock, availabilityZone string, isPublic bool) K8sObject {
	subnet := NewK8sObject("ec2.aws.crossplane.io/v1beta1", "Subnet", name)
	
	// Add Subnet specific properties
	subnet.AddNestedSpecField([]string{"forProvider", "cidrBlock"}, cidrBlock)
	subnet.AddNestedSpecField([]string{"forProvider", "availabilityZone"}, availabilityZone)
	
	// Reference the VPC
	subnet.AddNestedSpecField([]string{"forProvider", "vpcIdRef", "name"}, vpcName)
	
	// Set if public or private
	subnet.AddNestedSpecField([]string{"forProvider", "mapPublicIpOnLaunch"}, isPublic)
	
	// Add provider config reference
	subnet.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	subnet.AddLabel("app.kubernetes.io/part-of", "network")
	subnet.AddLabel("app.kubernetes.io/component", "subnet")
	if isPublic {
		subnet.AddLabel("network.aws.crossplane.io/subnet-type", "public")
	} else {
		subnet.AddLabel("network.aws.crossplane.io/subnet-type", "private")
	}
	
	return subnet
}

// GenerateInternetGateway generates a Crossplane Internet Gateway resource
func (g *VPCGenerator) GenerateInternetGateway(name, vpcName string) K8sObject {
	igw := NewK8sObject("ec2.aws.crossplane.io/v1beta1", "InternetGateway", name)
	
	// Reference the VPC
	igw.AddNestedSpecField([]string{"forProvider", "vpcIdRef", "name"}, vpcName)
	
	// Add provider config reference
	igw.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	igw.AddLabel("app.kubernetes.io/part-of", "network")
	igw.AddLabel("app.kubernetes.io/component", "internet-gateway")
	
	return igw
}

// GenerateNATGateway generates a Crossplane NAT Gateway resource
func (g *VPCGenerator) GenerateNATGateway(name, subnetName, allocationId string) K8sObject {
	natgw := NewK8sObject("ec2.aws.crossplane.io/v1beta1", "NATGateway", name)
	
	// Add NAT Gateway specific properties
	natgw.AddNestedSpecField([]string{"forProvider", "subnetIdRef", "name"}, subnetName)
	
	// If we have an allocation ID (for the Elastic IP), reference it
	if allocationId != "" {
		natgw.AddNestedSpecField([]string{"forProvider", "allocationIdRef", "name"}, allocationId)
	}
	
	// Add provider config reference
	natgw.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	natgw.AddLabel("app.kubernetes.io/part-of", "network")
	natgw.AddLabel("app.kubernetes.io/component", "nat-gateway")
	
	return natgw
}

// GenerateElasticIP generates a Crossplane Elastic IP resource
func (g *VPCGenerator) GenerateElasticIP(name string) K8sObject {
	eip := NewK8sObject("ec2.aws.crossplane.io/v1beta1", "ElasticIP", name)
	
	// Add Elastic IP specific properties
	eip.AddNestedSpecField([]string{"forProvider", "domain"}, "vpc")
	
	// Add provider config reference
	eip.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	eip.AddLabel("app.kubernetes.io/part-of", "network")
	eip.AddLabel("app.kubernetes.io/component", "elastic-ip")
	
	return eip
}

// GenerateRouteTable generates a Crossplane Route Table resource
func (g *VPCGenerator) GenerateRouteTable(name, vpcName string, isPublic bool) K8sObject {
	rt := NewK8sObject("ec2.aws.crossplane.io/v1beta1", "RouteTable", name)
	
	// Reference the VPC
	rt.AddNestedSpecField([]string{"forProvider", "vpcIdRef", "name"}, vpcName)
	
	// Add provider config reference
	rt.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	rt.AddLabel("app.kubernetes.io/part-of", "network")
	rt.AddLabel("app.kubernetes.io/component", "route-table")
	if isPublic {
		rt.AddLabel("network.aws.crossplane.io/route-table-type", "public")
	} else {
		rt.AddLabel("network.aws.crossplane.io/route-table-type", "private")
	}
	
	return rt
}

// GenerateRoute generates a Crossplane Route resource
func (g *VPCGenerator) GenerateRoute(name, routeTableName, destinationCidr string, gatewayType, gatewayName string) K8sObject {
	route := NewK8sObject("ec2.aws.crossplane.io/v1beta1", "Route", name)
	
	// Add Route specific properties
	route.AddNestedSpecField([]string{"forProvider", "destinationCidrBlock"}, destinationCidr)
	
	// Reference the Route Table
	route.AddNestedSpecField([]string{"forProvider", "routeTableIdRef", "name"}, routeTableName)
	
	// Reference the target Gateway based on type
	switch strings.ToLower(gatewayType) {
	case "igw":
		route.AddNestedSpecField([]string{"forProvider", "gatewayIdRef", "name"}, gatewayName)
	case "nat":
		route.AddNestedSpecField([]string{"forProvider", "natGatewayIdRef", "name"}, gatewayName)
	}
	
	// Add provider config reference
	route.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	route.AddLabel("app.kubernetes.io/part-of", "network")
	route.AddLabel("app.kubernetes.io/component", "route")
	
	return route
}

// GenerateSubnetRouteTableAssociation generates a Crossplane Subnet Route Table Association resource
func (g *VPCGenerator) GenerateSubnetRouteTableAssociation(name, subnetName, routeTableName string) K8sObject {
	assoc := NewK8sObject("ec2.aws.crossplane.io/v1beta1", "RouteTableAssociation", name)
	
	// Reference the Subnet and Route Table
	assoc.AddNestedSpecField([]string{"forProvider", "subnetIdRef", "name"}, subnetName)
	assoc.AddNestedSpecField([]string{"forProvider", "routeTableIdRef", "name"}, routeTableName)
	
	// Add provider config reference
	assoc.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	assoc.AddLabel("app.kubernetes.io/part-of", "network")
	assoc.AddLabel("app.kubernetes.io/component", "route-table-association")
	
	return assoc
}

// GenerateNetworkResources generates all Crossplane VPC networking resources from an infrastructure model
func (g *VPCGenerator) GenerateNetworkResources(model *models.InfrastructureModel) error {
	var (
		vpc          K8sObject
		vpcName      string
		publicSubnets  []K8sObject
		privateSubnets []K8sObject
		igw          K8sObject
		natGateways  []K8sObject
		eips         []K8sObject
		publicRT     K8sObject
		privateRTs   []K8sObject
		routes       []K8sObject
		associations []K8sObject
	)
	
	// Find the VPC
	for _, resource := range model.Resources {
		if resource.Type == models.ResourceVPC {
			vpcName = resource.Name
			
			// Extract VPC properties
			cidrBlock := "10.0.0.0/16"
			enableDnsSupport := true
			enableDnsHostnames := true
			
			for _, prop := range resource.Properties {
				switch prop.Name {
				case "cidr_block":
					if val, ok := prop.Value.(string); ok {
						cidrBlock = val
					}
				case "enable_dns_support":
					if val, ok := prop.Value.(bool); ok {
						enableDnsSupport = val
					}
				case "enable_dns_hostnames":
					if val, ok := prop.Value.(bool); ok {
						enableDnsHostnames = val
					}
				}
			}
			
			vpc = g.GenerateVPC(vpcName, cidrBlock, enableDnsSupport, enableDnsHostnames)
			break
		}
	}
	
	// If no VPC found, create one with default values
	if vpcName == "" {
		vpcName = "main-vpc"
		vpc = g.GenerateVPC(vpcName, "10.0.0.0/16", true, true)
	}
	
	// Find subnets
	for _, resource := range model.Resources {
		if resource.Type == models.ResourceSubnet {
			// Extract subnet properties
			name := resource.Name
			cidrBlock := "10.0.0.0/24"
			az := "us-east-1a"
			isPublic := false
			
			for _, prop := range resource.Properties {
				switch prop.Name {
				case "cidr_block":
					if val, ok := prop.Value.(string); ok {
						cidrBlock = val
					}
				case "availability_zone":
					if val, ok := prop.Value.(string); ok {
						az = val
					}
				case "map_public_ip_on_launch":
					if val, ok := prop.Value.(bool); ok {
						isPublic = val
					}
				}
			}
			
			subnet := g.GenerateSubnet(name, vpcName, cidrBlock, az, isPublic)
			
			if isPublic {
				publicSubnets = append(publicSubnets, subnet)
			} else {
				privateSubnets = append(privateSubnets, subnet)
			}
		}
	}
	
	// Find Internet Gateway
	for _, resource := range model.Resources {
		if resource.Type == models.ResourceIGW {
			igw = g.GenerateInternetGateway(resource.Name, vpcName)
			break
		}
	}
	
	// If no IGW found, create one if we have public subnets
	if igw.APIVersion == "" && len(publicSubnets) > 0 {
		igw = g.GenerateInternetGateway("main-igw", vpcName)
	}
	
	// Create public route table if needed
	if len(publicSubnets) > 0 && igw.APIVersion != "" {
		publicRT = g.GenerateRouteTable("public-rt", vpcName, true)
		
		// Create route to the internet
		internetRoute := g.GenerateRoute(
			"public-internet-route",
			publicRT.Metadata.Name,
			"0.0.0.0/0",
			"igw",
			igw.Metadata.Name,
		)
		
		routes = append(routes, internetRoute)
		
		// Create route table associations for public subnets
		for i, subnet := range publicSubnets {
			assoc := g.GenerateSubnetRouteTableAssociation(
				fmt.Sprintf("public-rt-assoc-%d", i+1),
				subnet.Metadata.Name,
				publicRT.Metadata.Name,
			)
			associations = append(associations, assoc)
		}
	}
	
	// Find NAT Gateways or create them if we have private subnets
	if len(privateSubnets) > 0 {
		natCount := 0
		
		// Find existing NAT gateways
		for _, resource := range model.Resources {
			if resource.Type == models.ResourceNATGateway {
				natCount++
			}
		}
		
		// Create NAT gateway for each AZ or at least one
		if natCount == 0 && len(publicSubnets) > 0 {
			// Determine how many NAT gateways we need (one per AZ is best practice)
			azs := make(map[string]bool)
			for _, subnet := range privateSubnets {
				for _, field := range subnet.Spec["forProvider"].(map[string]interface{}) {
					if az, ok := field.(string); ok && strings.HasPrefix(az, "us-") {
						azs[az] = true
					}
				}
			}
			
			// Create at least one NAT gateway, or one per AZ
			natCount = 1
			if len(azs) > 0 {
				natCount = len(azs)
			}
			
			// Create NAT gateways (and EIPs)
			for i := 0; i < natCount && i < len(publicSubnets); i++ {
				// Create EIP for NAT gateway
				eipName := fmt.Sprintf("nat-eip-%d", i+1)
				eip := g.GenerateElasticIP(eipName)
				eips = append(eips, eip)
				
				// Create NAT gateway
				natName := fmt.Sprintf("nat-gateway-%d", i+1)
				nat := g.GenerateNATGateway(natName, publicSubnets[i].Metadata.Name, eipName)
				natGateways = append(natGateways, nat)
				
				// Create private route table for this NAT gateway
				privateRTName := fmt.Sprintf("private-rt-%d", i+1)
				privateRT := g.GenerateRouteTable(privateRTName, vpcName, false)
				privateRTs = append(privateRTs, privateRT)
				
				// Create route to the internet via NAT
				natRoute := g.GenerateRoute(
					fmt.Sprintf("private-internet-route-%d", i+1),
					privateRTName,
					"0.0.0.0/0",
					"nat",
					natName,
				)
				routes = append(routes, natRoute)
				
				// Distribute private subnets across NAT gateways
				for j, subnet := range privateSubnets {
					if j % natCount == i {
						assoc := g.GenerateSubnetRouteTableAssociation(
							fmt.Sprintf("private-rt-assoc-%d-%d", i+1, j+1),
							subnet.Metadata.Name,
							privateRTName,
						)
						associations = append(associations, assoc)
					}
				}
			}
		}
	}
	
	// Write VPC YAML
	vpcFilePath := filepath.Join(g.vpcDir, "vpc.yaml")
	if err := WriteYAML(vpc, vpcFilePath); err != nil {
		return fmt.Errorf("failed to write VPC YAML: %w", err)
	}
	
	// Write Subnets YAML
	var allSubnets []K8sObject
	allSubnets = append(allSubnets, publicSubnets...)
	allSubnets = append(allSubnets, privateSubnets...)
	if len(allSubnets) > 0 {
		subnetsFilePath := filepath.Join(g.vpcDir, "subnets.yaml")
		if err := WriteMultiYAML(allSubnets, subnetsFilePath); err != nil {
			return fmt.Errorf("failed to write Subnets YAML: %w", err)
		}
	}
	
	// Write Gateways YAML (IGW, NAT, EIP)
	var gateways []K8sObject
	if igw.APIVersion != "" {
		gateways = append(gateways, igw)
	}
	gateways = append(gateways, eips...)
	gateways = append(gateways, natGateways...)
	if len(gateways) > 0 {
		gatewaysFilePath := filepath.Join(g.vpcDir, "gateways.yaml")
		if err := WriteMultiYAML(gateways, gatewaysFilePath); err != nil {
			return fmt.Errorf("failed to write Gateways YAML: %w", err)
		}
	}
	
	// Write Routing YAML (Route tables, routes, associations)
	var routing []K8sObject
	if publicRT.APIVersion != "" {
		routing = append(routing, publicRT)
	}
	routing = append(routing, privateRTs...)
	routing = append(routing, routes...)
	routing = append(routing, associations...)
	if len(routing) > 0 {
		routingFilePath := filepath.Join(g.vpcDir, "routing.yaml")
		if err := WriteMultiYAML(routing, routingFilePath); err != nil {
			return fmt.Errorf("failed to write Routing YAML: %w", err)
		}
	}
	
	// Update Kustomization with the routing file if needed
	if len(routing) > 0 {
		kustomizationPath := filepath.Join(g.vpcDir, "kustomization.yaml")
		content, err := utils.ReadFromFile(kustomizationPath)
		if err != nil {
			return fmt.Errorf("failed to read VPC kustomization: %w", err)
		}
		
		if !strings.Contains(content, "routing.yaml") {
			content = strings.Replace(content, "- gateways.yaml", "- gateways.yaml\n- routing.yaml", 1)
			if err := utils.WriteToFile(kustomizationPath, content); err != nil {
				return fmt.Errorf("failed to update VPC kustomization: %w", err)
			}
		}
	}
	
	return nil
}