package nlp

import (
	"regexp"
	"strconv"
	"strings"
)

// RegionPattern matches AWS region references
var RegionPattern = regexp.MustCompile(`(?i)(us|eu|ap|sa|ca|me|af)-(east|west|north|south|central|northeast|northwest|southeast|southwest)-\d+`)

// VPCPattern matches VPC references with optional CIDR ranges
var VPCPattern = regexp.MustCompile(`(?i)vpc(?:\s+with\s+CIDR\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})|\s+with\s+cidr\s+block\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2}))?`)

// CIDRPattern matches CIDR blocks
var CIDRPattern = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})\b`)

// SubnetPattern matches subnet references with type and count
var SubnetPattern = regexp.MustCompile(`(?i)(\d+)\s+(public|private)\s+subnet`)

// AZPattern matches availability zone references
var AZPattern = regexp.MustCompile(`(?i)(\d+)\s*az`)

// IGWPattern matches internet gateway references
var IGWPattern = regexp.MustCompile(`(?i)(\d+)?\s*(internet\s*gateway|igw)`)

// NATPattern matches NAT gateway references with optional count
var NATPattern = regexp.MustCompile(`(?i)(\d+)?\s*(nat\s*gateway)(?:\s+per\s+az)?`)

// EKSPattern matches EKS cluster references
var EKSPattern = regexp.MustCompile(`(?i)eks\s+cluster(?:\s+with\s+(public|private|public\s+and\s+private)\s+api\s+access)?(?:\s+version\s+([\d\.]+))?(?:\s+with\s+version\s+([\d\.]+))?`)

// NodePoolPattern matches node pool references with optional instance type and count
var NodePoolPattern = regexp.MustCompile(`(?i)(?:node\s*pool|nodepool)(?:\s+with\s+(\d+)\s+nodes?)?(?:\s+of\s+(\d+)\s+nodes?)?(?:\s+on\s+(t\d+\.[a-z]+|m\d+\.[a-z]+|c\d+\.[a-z]+))?`)

// InstanceTypePattern matches instance type references
var InstanceTypePattern = regexp.MustCompile(`(?i)(t\d+\.[a-z]+|m\d+\.[a-z]+|c\d+\.[a-z]+)`)

// NumberPattern extracts standalone numbers
var NumberPattern = regexp.MustCompile(`\b(\d+)\b`)

// ExtractRegion extracts the AWS region from the description
func ExtractRegion(description string) string {
	match := RegionPattern.FindString(description)
	if match != "" {
		return strings.ToLower(match)
	}
	// Default to us-east-1 if no region specified
	return "us-east-1"
}

// ExtractVPC extracts VPC details from the description
func ExtractVPC(description string) map[string]interface{} {
	vpc := make(map[string]interface{})
	
	// Always assume VPC is needed even if not specifically mentioned
	vpc["exists"] = true
	vpc["cidr_block"] = "10.0.0.0/16" // Default CIDR
	vpc["enable_dns_support"] = true
	vpc["enable_dns_hostnames"] = true
	
	// Check if VPC is mentioned
	vpcMatch := VPCPattern.FindStringSubmatch(description)
	if len(vpcMatch) > 0 {
		// Extract CIDR if specified
		if len(vpcMatch) > 1 && vpcMatch[1] != "" {
			vpc["cidr_block"] = vpcMatch[1]
		} else if len(vpcMatch) > 2 && vpcMatch[2] != "" {
			vpc["cidr_block"] = vpcMatch[2]
		}
	}
	
	// Also look for any CIDR block in the description
	if vpc["cidr_block"] == "10.0.0.0/16" {
		cidrMatch := CIDRPattern.FindStringSubmatch(description)
		if len(cidrMatch) > 0 && cidrMatch[1] != "" {
			vpc["cidr_block"] = cidrMatch[1]
		}
	}
	
	return vpc
}

// ExtractSubnets extracts subnet information from the description
func ExtractSubnets(description string) map[string]interface{} {
	subnets := make(map[string]interface{})
	
	// Initialize default counts
	publicCount := 0
	privateCount := 0
	
	// Extract subnet counts
	subnetMatches := SubnetPattern.FindAllStringSubmatch(description, -1)
	for _, match := range subnetMatches {
		if len(match) >= 3 {
			count, err := strconv.Atoi(match[1])
			if err != nil {
				continue
			}
			
			subnetType := strings.ToLower(match[2])
			if subnetType == "public" {
				publicCount = count
			} else if subnetType == "private" {
				privateCount = count
			}
		}
	}
	
	// Special case for "X public and Y private subnets" pattern
	combinedPattern := regexp.MustCompile(`(?i)(\d+)\s+public\s+and\s+(\d+)\s+private\s+subnet`)
	combinedMatch := combinedPattern.FindStringSubmatch(description)
	if len(combinedMatch) >= 3 {
		if publicCount == 0 {
			pCount, err := strconv.Atoi(combinedMatch[1])
			if err == nil && pCount > 0 {
				publicCount = pCount
			}
		}
		
		if privateCount == 0 {
			prCount, err := strconv.Atoi(combinedMatch[2])
			if err == nil && prCount > 0 {
				privateCount = prCount
			}
		}
	}
	
	// If no subnet counts found, check for AZ count and assume 1 public and 1 private per AZ
	if publicCount == 0 && privateCount == 0 {
		azMatches := AZPattern.FindStringSubmatch(description)
		if len(azMatches) >= 2 {
			azCount, err := strconv.Atoi(azMatches[1])
			if err == nil && azCount > 0 {
				publicCount = azCount
				privateCount = azCount
			}
		}
	}
	
	// Default to 1 public and 1 private if no counts found
	if publicCount == 0 {
		publicCount = 1
	}
	if privateCount == 0 {
		privateCount = 1
	}
	
	subnets["public_count"] = publicCount
	subnets["private_count"] = privateCount
	
	return subnets
}

// ExtractGateways extracts Internet Gateway and NAT Gateway details
func ExtractGateways(description string) map[string]interface{} {
	gateways := make(map[string]interface{})
	
	// Extract IGW count
	igwMatches := IGWPattern.FindStringSubmatch(description)
	igwCount := 1 // Default to 1 IGW
	if len(igwMatches) >= 2 && igwMatches[1] != "" {
		count, err := strconv.Atoi(igwMatches[1])
		if err == nil && count > 0 {
			igwCount = count
		}
	}
	
	// Extract NAT count
	natMatches := NATPattern.FindStringSubmatch(description)
	natCount := 0 // Default to 0 NAT Gateways
	if len(natMatches) >= 2 && natMatches[1] != "" {
		count, err := strconv.Atoi(natMatches[1])
		if err == nil && count > 0 {
			natCount = count
		}
	} else if strings.Contains(strings.ToLower(description), "nat gateway per az") {
		// If "NAT gateway per AZ" is mentioned, extract AZ count
		azMatches := AZPattern.FindStringSubmatch(description)
		if len(azMatches) >= 2 {
			azCount, err := strconv.Atoi(azMatches[1])
			if err == nil && azCount > 0 {
				natCount = azCount
			} else {
				// Default to 1 NAT Gateway if no count specified
				natCount = 1
			}
		}
	} else if natMatches != nil {
		// If NAT Gateway is mentioned but no count specified, default to 1
		natCount = 1
	}
	
	gateways["igw_count"] = igwCount
	gateways["nat_count"] = natCount
	
	return gateways
}

// ExtractEKS extracts EKS cluster details from the description
func ExtractEKS(description string) map[string]interface{} {
	eks := make(map[string]interface{})
	
	// Check if EKS is mentioned
	eksMatches := EKSPattern.FindStringSubmatch(description)
	if len(eksMatches) > 0 {
		eks["exists"] = true
		
		// Default API access
		eks["endpoint_public_access"] = true
		eks["endpoint_private_access"] = false
		
		// Extract API access mode if specified
		if len(eksMatches) > 1 && eksMatches[1] != "" {
			accessMode := strings.ToLower(eksMatches[1])
			
			if accessMode == "private" {
				eks["endpoint_public_access"] = false
				eks["endpoint_private_access"] = true
			} else if accessMode == "public and private" || accessMode == "private and public" {
				eks["endpoint_public_access"] = true
				eks["endpoint_private_access"] = true
			}
		}
		
		// Additional check for private access using full string check
		if strings.Contains(strings.ToLower(description), "private api access") {
			eks["endpoint_public_access"] = false
			eks["endpoint_private_access"] = true
		}
		
		// Extract version if specified (check both version patterns)
		if len(eksMatches) > 2 && eksMatches[2] != "" {
			eks["version"] = eksMatches[2]
		} else if len(eksMatches) > 3 && eksMatches[3] != "" {
			eks["version"] = eksMatches[3]
		} else {
			// Default version
			eks["version"] = "1.27"
		}
		
		// Extract node pool details
		nodePoolMatches := NodePoolPattern.FindStringSubmatch(description)
		nodeCount := 2 // Default node count
		instanceType := "t3.medium" // Default instance type
		
		// Check for node count with "with X nodes" pattern
		if len(nodePoolMatches) > 1 && nodePoolMatches[1] != "" {
			count, err := strconv.Atoi(nodePoolMatches[1])
			if err == nil && count > 0 {
				nodeCount = count
			}
		}
		
		// Check for node count with "of X nodes" pattern
		if len(nodePoolMatches) > 2 && nodePoolMatches[2] != "" {
			count, err := strconv.Atoi(nodePoolMatches[2])
			if err == nil && count > 0 {
				nodeCount = count
			}
		}
		
		// Also look for generic number mentions related to nodes
		if nodeCount == 2 && strings.Contains(description, "node") {
			// Look for patterns like "3 nodes" or "with 3 nodes"
			simpleNodeCountPattern := regexp.MustCompile(`(\d+)\s+nodes?`)
			simpleMatches := simpleNodeCountPattern.FindStringSubmatch(description)
			if len(simpleMatches) > 1 {
				count, err := strconv.Atoi(simpleMatches[1])
				if err == nil && count > 0 {
					nodeCount = count
				}
			}
		}
		
		// Extract instance type
		if len(nodePoolMatches) > 3 && nodePoolMatches[3] != "" {
			instanceType = nodePoolMatches[3]
		} else {
			// Try to find instance type elsewhere in the description
			instanceTypeMatch := InstanceTypePattern.FindString(description)
			if instanceTypeMatch != "" {
				instanceType = instanceTypeMatch
			}
		}
		
		eks["node_count"] = nodeCount
		eks["instance_type"] = instanceType
	}
	
	return eks
}

// Note: The GenerateSubnetCIDRs function is now defined in the infra package to avoid circular imports