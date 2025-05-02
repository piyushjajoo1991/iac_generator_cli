package nlp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/infra"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// Parser interfaces with NLP services to extract infrastructure entities
type Parser struct {
	// In a production implementation, this might include a client to an NLP service
}

// NewParser creates a new NLP parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseDescription parses a natural language description into an infrastructure model
func ParseDescription(description string) (*models.InfrastructureModel, error) {
	// Validate the input description
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}
	
	if len(description) < 5 {
		return nil, errors.New("description is too short to be meaningful")
	}

	parser := NewParser()
	entities, err := parser.ExtractEntities(description)
	if err != nil {
		return nil, err
	}

	// Validate and fill in missing information
	validationResult := ValidateEntities(entities)
	if !validationResult.Valid {
		// Log validation messages, but continue with the fixes applied
		fmt.Println("Validation:", validationResult.Message)
	}

	modelBuilder := infra.NewModelBuilder()
	err = modelBuilder.BuildFromParsedEntities(entities)
	if err != nil {
		return nil, err
	}

	return modelBuilder.GetModel(), nil
}

// ExtractEntities extracts infrastructure entities from the description
func (p *Parser) ExtractEntities(description string) (map[string]interface{}, error) {
	entities := make(map[string]interface{})
	
	// Preprocess the description
	description = strings.ToLower(description)
	
	// Extract AWS region
	region := ExtractRegion(description)
	entities["region"] = region
	
	// Extract VPC information
	vpcInfo := ExtractVPC(description)
	if len(vpcInfo) > 0 && vpcInfo["exists"] == true {
		entities["vpc"] = vpcInfo
	}
	
	// Extract subnet information
	subnetInfo := ExtractSubnets(description)
	if len(subnetInfo) > 0 {
		entities["subnets"] = subnetInfo
		
		// Generate CIDR blocks for the subnets if VPC exists
		if vpc, ok := entities["vpc"].(map[string]interface{}); ok {
			if vpcCIDR, ok := vpc["cidr_block"].(string); ok {
				publicCount := subnetInfo["public_count"].(int)
				privateCount := subnetInfo["private_count"].(int)
				
				publicCIDRs, privateCIDRs, err := infra.GenerateSubnetCIDRs(vpcCIDR, publicCount, privateCount)
				if err == nil {
					subnetInfo["public_cidrs"] = publicCIDRs
					subnetInfo["private_cidrs"] = privateCIDRs
				}
			}
		}
	}
	
	// Extract gateway information (IGW, NAT)
	gatewayInfo := ExtractGateways(description)
	if len(gatewayInfo) > 0 {
		entities["gateways"] = gatewayInfo
	}
	
	// Extract EKS cluster information
	eksInfo := ExtractEKS(description)
	if len(eksInfo) > 0 && eksInfo["exists"] == true {
		entities["eks"] = eksInfo
	}
	
	// If no entities were extracted, return an error
	if len(entities) <= 1 { // Only region is not enough
		return nil, errors.New("could not extract any infrastructure entities from the description")
	}
	
	return entities, nil
}