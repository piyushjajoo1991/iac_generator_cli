package infra

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
)

// CIDRInfo holds information about a CIDR block
type CIDRInfo struct {
	CIDR      string
	Mask      int
	Network   net.IP
	FirstIP   net.IP
	LastIP    net.IP
	Available int
}

// ParseCIDR parses a CIDR string and returns detailed information about it
func ParseCIDR(cidr string) (*CIDRInfo, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR format: %w", err)
	}

	// Calculate mask size
	ones, bits := ipnet.Mask.Size()

	// Calculate available IP addresses (2^host_bits - 2 for network and broadcast)
	// For /31 and /32, special handling is needed
	var available int
	if ones >= 31 {
		// For /31, RFC 3021 allows using both IPs
		// For /32, only 1 IP is available
		available = int(math.Pow(2, float64(bits-ones)))
	} else {
		// For /30 and below, subtract 2 for network and broadcast addresses
		available = int(math.Pow(2, float64(bits-ones))) - 2
	}

	// Get the first usable IP (network address + 1)
	firstIP := make(net.IP, len(ipnet.IP))
	copy(firstIP, ipnet.IP)
	if ones < 31 {
		// For /30 and below, first IP is network address + 1
		incrementIP(firstIP)
	}

	// Get the last usable IP (broadcast address - 1 or the same as broadcast for /31 and /32)
	lastIP := getLastIP(ipnet)
	if ones < 31 {
		// For /30 and below, last IP is broadcast address - 1
		decrementIP(lastIP)
	}

	return &CIDRInfo{
		CIDR:      cidr,
		Mask:      ones,
		Network:   ipnet.IP,
		FirstIP:   firstIP,
		LastIP:    lastIP,
		Available: available,
	}, nil
}

// SubdivideCIDR divides a CIDR block into smaller subnets of equal size
func SubdivideCIDR(parentCIDR string, newMask int) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR format: %w", err)
	}

	parentMask, _ := ipnet.Mask.Size()
	if newMask <= parentMask {
		return nil, fmt.Errorf("new mask must be larger than parent mask")
	}

	// Calculate the number of subnets to create
	subnetCount := int(math.Pow(2, float64(newMask-parentMask)))

	// Generate subnet CIDRs
	subnets := make([]string, subnetCount)

	// Calculate the size of each subnet
	_, bits := ipnet.Mask.Size()
	subnetSize := int(math.Pow(2, float64(bits-newMask)))

	// Start with the network address of the parent CIDR
	currentIP := make(net.IP, len(ipnet.IP))
	copy(currentIP, ipnet.IP)

	// Create each subnet
	for i := 0; i < subnetCount; i++ {
		subnets[i] = fmt.Sprintf("%s/%d", currentIP.String(), newMask)

		// Increment the IP address by the subnet size to get the next subnet
		for j := 0; j < subnetSize; j++ {
			incrementIP(currentIP)
		}
	}

	return subnets, nil
}

// AllocateSubnets allocates a specified number of subnets from a VPC CIDR block
// with options for different subnet sizes and public/private distinction
func AllocateSubnets(vpcCIDR string, publicCount, privateCount int, publicMask, privateMask int) ([]string, []string, error) {
	// Use default masks if not specified
	if publicMask == 0 {
		publicMask = 24 // Default to /24
	}

	if privateMask == 0 {
		privateMask = 24 // Default to /24
	}

	_, ipnet, err := net.ParseCIDR(vpcCIDR)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid VPC CIDR format: %w", err)
	}

	vpcMask, _ := ipnet.Mask.Size()

	// Make sure we can allocate enough subnets
	totalSubnetsNeeded := publicCount + privateCount
	maxSubnets := int(math.Pow(2, float64(privateMask-vpcMask)))

	if totalSubnetsNeeded > maxSubnets {
		return nil, nil, fmt.Errorf("not enough address space in VPC CIDR %s to allocate %d subnets with /%d mask",
			vpcCIDR, totalSubnetsNeeded, privateMask)
	}

	// First, allocate all subnets at the smaller of the two masks
	smallerMask := publicMask
	if privateMask > publicMask {
		smallerMask = privateMask
	}

	allSubnets, err := SubdivideCIDR(vpcCIDR, smallerMask)
	if err != nil {
		return nil, nil, err
	}

	// Allocate public subnets from the beginning
	publicSubnets := make([]string, publicCount)
	for i := 0; i < publicCount && i < len(allSubnets); i++ {
		publicSubnets[i] = allSubnets[i]
	}

	// Allocate private subnets after public ones
	privateSubnets := make([]string, privateCount)
	for i := 0; i < privateCount && i+publicCount < len(allSubnets); i++ {
		privateSubnets[i] = allSubnets[i+publicCount]
	}

	return publicSubnets, privateSubnets, nil
}

// GenerateSubnetName generates a meaningful name for a subnet
func GenerateSubnetName(vpcName string, isPublic bool, az string, index int) string {
	visibility := "private"
	if isPublic {
		visibility = "public"
	}

	// Extract the AZ letter (e.g., "a" from "us-east-1a")
	azParts := strings.Split(az, "-")
	azSuffix := ""
	if len(azParts) > 0 {
		lastPart := azParts[len(azParts)-1]
		if len(lastPart) > 0 {
			azSuffix = string(lastPart[len(lastPart)-1])
		}
	}

	return fmt.Sprintf("%s-%s-%s-%d", vpcName, visibility, azSuffix, index)
}

// CIDRToNetworkAndMask converts a CIDR to a network address and mask bits
// e.g., "10.0.0.0/16" -> "10.0.0.0", 16
func CIDRToNetworkAndMask(cidr string) (string, int, error) {
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid CIDR format: %s", cidr)
	}

	mask, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid mask in CIDR: %w", err)
	}

	return parts[0], mask, nil
}

// incrementIP increments an IP address by 1
func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

// decrementIP decrements an IP address by 1
func decrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] > 0 {
			ip[i]--
			break
		}
		ip[i] = 255
	}
}

// getLastIP returns the broadcast address of the network
func getLastIP(ipnet *net.IPNet) net.IP {
	// Create a copy of the network IP
	lastIP := make(net.IP, len(ipnet.IP))
	copy(lastIP, ipnet.IP)

	// Calculate the broadcast address by setting host bits to 1
	for i := 0; i < len(lastIP); i++ {
		lastIP[i] = lastIP[i] | ^ipnet.Mask[i]
	}

	return lastIP
}
