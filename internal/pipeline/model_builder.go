package pipeline

import (
	"context"
	"fmt"

	"github.com/riptano/iac_generator_cli/internal/infra"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"go.uber.org/zap"
)

// ModelBuilderImpl is the implementation of the ModelBuilder interface
type ModelBuilderImpl struct {
	// region is the AWS region to use for resources
	region string
	logger *zap.SugaredLogger
}

// NewModelBuilder creates a new model builder with the specified region
func NewModelBuilder(region string) *ModelBuilderImpl {
	return &ModelBuilderImpl{
		region: region,
		logger: utils.GetLogger(),
	}
}

// BuildModel implements ModelBuilder
func (b *ModelBuilderImpl) BuildModel(ctx context.Context, input interface{}) (*models.InfrastructureModel, error) {
	b.logger.Debugw("Building infrastructure model")

	// Check if the context is canceled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var model *models.InfrastructureModel

	// Handle different input types
	switch v := input.(type) {
	case *models.InfrastructureModel:
		// If we already have a model, just return it
		model = v
	case map[string]interface{}:
		// Build model from parsed entities
		builder := infra.NewModelBuilder()
		err := builder.BuildFromParsedEntities(v)
		if err != nil {
			return nil, fmt.Errorf("failed to build model from entities: %w", err)
		}
		model = builder.GetModel()
	default:
		return nil, fmt.Errorf("invalid input type for model building: %T", input)
	}

	// Enhance the model with additional information
	enhancedModel, err := b.EnhanceModel(model)
	if err != nil {
		return nil, fmt.Errorf("failed to enhance model: %w", err)
	}

	b.logger.Debugw("Model built successfully",
		"resources_count", len(enhancedModel.Resources),
	)

	return enhancedModel, nil
}

// EnhanceModel implements ModelBuilder
func (b *ModelBuilderImpl) EnhanceModel(model *models.InfrastructureModel) (*models.InfrastructureModel, error) {
	// Apply region to resources that support it
	for i, resource := range model.Resources {
		// Check resource type and update region if applicable
		switch resource.Type {
		case models.ResourceEC2Instance, models.ResourceRDSInstance, models.ResourceVPC,
			models.ResourceSubnet, models.ResourceSecurityGroup:
			// Check if region is already set
			hasRegion := false
			for _, prop := range resource.Properties {
				if prop.Name == "region" {
					hasRegion = true
					break
				}
			}

			// Add region property if not already set
			if !hasRegion && b.region != "" {
				model.Resources[i].AddProperty("region", b.region)
			}
		}
	}

	// Add resource dependencies based on resource relationships
	// This is a simplified example - in a real implementation, we would analyze
	// the resources and add appropriate dependencies
	for i := range model.Resources {
		switch model.Resources[i].Type {
		case models.ResourceSubnet:
			// Add dependency to VPC for subnets
			for _, resource := range model.Resources {
				if resource.Type == models.ResourceVPC {
					model.Resources[i].AddDependency(resource.Name)
					break
				}
			}
		case models.ResourceIGW:
			// Add dependency to VPC for Internet Gateways
			for _, resource := range model.Resources {
				if resource.Type == models.ResourceVPC {
					model.Resources[i].AddDependency(resource.Name)
					break
				}
			}
		case models.ResourceNATGateway:
			// Add dependency to Subnet for NAT Gateways
			for _, resource := range model.Resources {
				if resource.Type == models.ResourceSubnet {
					model.Resources[i].AddDependency(resource.Name)
					break
				}
			}
		}
	}

	return model, nil
}

// ModelBuildStage creates a pipeline stage that builds an infrastructure model
func (b *ModelBuilderImpl) ModelBuildStage() Stage {
	return NewBaseStage("ModelBuilding", func(ctx context.Context, input interface{}) (interface{}, error) {
		return b.BuildModel(ctx, input)
	})
}