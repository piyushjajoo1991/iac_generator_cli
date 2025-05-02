package nlp

import (
	"fmt"
	"strings"
	
	"github.com/riptano/iac_generator_cli/internal/infra"
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid   bool
	Message string
	Fixes   map[string]interface{}
}

// NewValidationResult creates a new empty validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:   true,
		Message: "",
		Fixes:   make(map[string]interface{}),
	}
}

// ValidateEntities checks if extracted entities are complete and consistent
func ValidateEntities(entities map[string]interface{}) *ValidationResult {
	result := NewValidationResult()
	messages := []string{}

	// Check if region exists
	if _, ok := entities["region"]; !ok {
		entities["region"] = "us-east-1"
		result.Fixes["region"] = "us-east-1"
		messages = append(messages, "Added default region (us-east-1)")
	}

	// Check if VPC exists
	vpcExists := false
	if vpc, ok := entities["vpc"].(map[string]interface{}); ok {
		vpcExists = true
		
		// Check if VPC has CIDR
		if _, ok := vpc["cidr_block"]; !ok {
			vpc["cidr_block"] = "10.0.0.0/16"
			result.Fixes["vpc_cidr"] = "10.0.0.0/16"
			messages = append(messages, "Added default VPC CIDR (10.0.0.0/16)")
		}
	} else {
		// Create a default VPC
		vpc := make(map[string]interface{})
		vpc["exists"] = true
		vpc["cidr_block"] = "10.0.0.0/16"
		vpc["enable_dns_support"] = true
		vpc["enable_dns_hostnames"] = true
		
		entities["vpc"] = vpc
		vpcExists = true
		result.Fixes["vpc"] = vpc
		messages = append(messages, "Added default VPC configuration")
	}

	// Check if subnets exist and are consistent
	if vpcExists {
		if subnets, ok := entities["subnets"].(map[string]interface{}); ok {
			// Ensure public and private subnet counts exist
			if _, ok := subnets["public_count"]; !ok {
				// Default to 1 public subnet
				subnets["public_count"] = 1
				result.Fixes["public_subnet_count"] = 1
				messages = append(messages, "Added default public subnet count (1)")
			}
			
			if _, ok := subnets["private_count"]; !ok {
				// Default to 1 private subnet
				subnets["private_count"] = 1
				result.Fixes["private_subnet_count"] = 1
				messages = append(messages, "Added default private subnet count (1)")
			}
			
			// Ensure CIDRs are generated
			_, publicOk := subnets["public_cidrs"]
			_, privateOk := subnets["private_cidrs"]
			if !publicOk || !privateOk {
				// Get VPC CIDR
				vpc := entities["vpc"].(map[string]interface{})
				cidr := vpc["cidr_block"].(string)
				publicCount := subnets["public_count"].(int)
				privateCount := subnets["private_count"].(int)
				
				// Generate subnet CIDRs
				publicCIDRs, privateCIDRs, err := infra.GenerateSubnetCIDRs(cidr, publicCount, privateCount)
				if err == nil {
					subnets["public_cidrs"] = publicCIDRs
					subnets["private_cidrs"] = privateCIDRs
					result.Fixes["subnet_cidrs"] = map[string]interface{}{
						"public":  publicCIDRs,
						"private": privateCIDRs,
					}
					messages = append(messages, "Generated subnet CIDRs")
				}
			}
		} else {
			// Create default subnets
			subnets := make(map[string]interface{})
			subnets["public_count"] = 1
			subnets["private_count"] = 1
			
			// Generate subnet CIDRs
			vpc := entities["vpc"].(map[string]interface{})
			cidr := vpc["cidr_block"].(string)
			publicCIDRs, privateCIDRs, err := infra.GenerateSubnetCIDRs(cidr, 1, 1)
			if err == nil {
				subnets["public_cidrs"] = publicCIDRs
				subnets["private_cidrs"] = privateCIDRs
			}
			
			entities["subnets"] = subnets
			result.Fixes["subnets"] = subnets
			messages = append(messages, "Added default subnet configuration")
		}
	}

	// Check if gateways exist when subnets do
	if _, ok := entities["subnets"]; ok {
		if _, ok := entities["gateways"]; !ok {
			// Create default gateways
			gateways := make(map[string]interface{})
			gateways["igw_count"] = 1
			gateways["nat_count"] = 0
			
			entities["gateways"] = gateways
			result.Fixes["gateways"] = gateways
			messages = append(messages, "Added default gateway configuration (1 IGW, 0 NAT)")
		}
	}

	// Check if EKS configuration is complete
	if eks, ok := entities["eks"].(map[string]interface{}); ok {
		// Ensure EKS version is set
		if _, ok := eks["version"]; !ok {
			eks["version"] = "1.27"
			result.Fixes["eks_version"] = "1.27"
			messages = append(messages, "Added default EKS version (1.27)")
		}
		
		// Ensure node count is set
		if _, ok := eks["node_count"]; !ok {
			eks["node_count"] = 2
			result.Fixes["node_count"] = 2
			messages = append(messages, "Added default node count (2)")
		}
		
		// Ensure instance type is set
		if _, ok := eks["instance_type"]; !ok {
			eks["instance_type"] = "t3.medium"
			result.Fixes["instance_type"] = "t3.medium"
			messages = append(messages, "Added default instance type (t3.medium)")
		}
		
		// Ensure API access mode is set
		if _, ok := eks["endpoint_public_access"]; !ok {
			eks["endpoint_public_access"] = true
			result.Fixes["endpoint_public_access"] = true
			messages = append(messages, "Added default public API access (true)")
		}
		
		if _, ok := eks["endpoint_private_access"]; !ok {
			eks["endpoint_private_access"] = false
			result.Fixes["endpoint_private_access"] = false
			messages = append(messages, "Added default private API access (false)")
		}
	}

	// Set validation result
	if len(messages) > 0 {
		// In this case, the validation is still successful, but we've made modifications
		// We leave Valid as true since we're returning a fixed, usable entity map
		result.Message = fmt.Sprintf("Validation added default values: %s", strings.Join(messages, ", "))
		
		// Log the validation message
		fmt.Println(result.Message)
	}

	return result
}