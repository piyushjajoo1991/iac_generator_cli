package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/nlp"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"go.uber.org/zap"
)

// NLPProcessorImpl is the implementation of the NLPProcessor interface
type NLPProcessorImpl struct {
	// parser is the underlying NLP parser
	parser *nlp.Parser
	logger *zap.SugaredLogger
}

// NewNLPProcessor creates a new NLP processor
func NewNLPProcessor() *NLPProcessorImpl {
	return &NLPProcessorImpl{
		parser: nlp.NewParser(),
		logger: utils.GetLogger(),
	}
}

// ParseDescription implements NLPProcessor
func (p *NLPProcessorImpl) ParseDescription(ctx context.Context, description string) (*models.InfrastructureModel, error) {
	p.logger.Debugw("Parsing description", "length", len(description))

	// Check if the context is canceled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Validate the description first
	valid, message := p.ValidateDescription(description)
	if !valid {
		return nil, fmt.Errorf("invalid description: %s", message)
	}

	// Enhance the description to improve NLP parsing
	enhancedDescription := nlp.EnhanceDescription(description)

	// Parse the description
	model, err := nlp.ParseDescription(enhancedDescription)
	if err != nil {
		return nil, fmt.Errorf("failed to parse description: %w", err)
	}

	p.logger.Debugw("Description parsed successfully",
		"resources_count", len(model.Resources),
		"description_length", len(description),
	)

	return model, nil
}

// ValidateDescription implements NLPProcessor
func (p *NLPProcessorImpl) ValidateDescription(description string) (bool, string) {
	// Validate that the description is not empty
	if strings.TrimSpace(description) == "" {
		return false, "description is empty"
	}

	// Validate that the description is of a reasonable length
	if len(description) < 10 {
		return false, "description is too short"
	}

	// Add a basic validation to check if the description contains infrastructure terms
	lowercaseDesc := strings.ToLower(description)
	infraTerms := []string{
		"ec2", "instance", "vpc", "subnet", "security group", "s3", "bucket",
		"rds", "database", "lambda", "function", "dynamodb", "table",
		"cloudwatch", "alarm", "metric", "gateway", "igw", "nat",
		"eks", "kubernetes", "cluster", "node", "group",
	}

	containsInfraTerm := false
	for _, term := range infraTerms {
		if strings.Contains(lowercaseDesc, term) {
			containsInfraTerm = true
			break
		}
	}

	if !containsInfraTerm {
		return false, "description doesn't contain any recognizable infrastructure terms"
	}

	return true, ""
}

// ProcessStage creates a pipeline stage that processes a description using the NLP processor
func (p *NLPProcessorImpl) ProcessStage() Stage {
	return NewBaseStage("NLPProcessing", func(ctx context.Context, input interface{}) (interface{}, error) {
		var description string
		switch v := input.(type) {
		case string:
			description = v
		case *ProcessingParams:
			description = v.Description
		default:
			return nil, fmt.Errorf("invalid input type for NLP processing: %T", input)
		}

		return p.ParseDescription(ctx, description)
	})
}