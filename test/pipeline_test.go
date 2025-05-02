package test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/riptano/iac_generator_cli/internal/pipeline"
)

// TestBasicPipelineExecution tests the end-to-end flow of the pipeline
func TestBasicPipelineExecution(t *testing.T) {
	// Create a buffer to capture output
	var outputBuffer bytes.Buffer

	// Create pipeline parameters with a basic AWS infrastructure description
	description := "Create a VPC with 2 public subnets and 2 private subnets in us-west-2."
	
	params := &pipeline.ProcessingParams{
		Description:    description,
		OutputFormat:   "terraform",
		OutputDir:      ".", // Output to current directory
		Region:         "us-west-2",
		UseTemplates:   true,
		Debug:          true,
		ProgressWriter: &outputBuffer,
	}

	// Create pipeline coordinator
	coordinator := pipeline.NewPipelineCoordinator()
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	// Initialize pipeline
	err := coordinator.InitializePipeline(ctx, params)
	if err != nil {
		t.Fatalf("Pipeline initialization failed: %v", err)
	}
	
	// Run pipeline
	result, err := coordinator.RunPipeline(ctx, params)
	if err != nil {
		t.Fatalf("Pipeline execution failed: %v", err)
	}
	
	// Verify that a result was returned
	if result == "" {
		t.Error("Pipeline returned empty result")
	}
	
	fmt.Printf("Pipeline execution successful. Result: %s\n", result)
	fmt.Printf("Output captured: %s\n", outputBuffer.String())
}

// TestPipelineStages tests each stage of the pipeline independently
func TestPipelineStages(t *testing.T) {
	// Define test cases for different pipeline stages
	testCases := []struct {
		name        string
		description string
		expectError bool
	}{
		{
			name:        "Valid VPC Description",
			description: "Create a VPC with CIDR 10.0.0.0/16 in us-east-1",
			expectError: false,
		},
		{
			name:        "Valid EC2 Description",
			description: "Create an EC2 instance with t2.micro size in us-west-2",
			expectError: false,
		},
		{
			name:        "Valid EKS Description",
			description: "Create an EKS cluster with 2 node groups in us-east-2",
			expectError: false,
		},
		{
			name:        "Empty Description",
			description: "",
			expectError: true,
		},
		{
			name:        "Too Short Description",
			description: "VPC",
			expectError: true,
		},
	}

	// Loop through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create NLP processor
			processor := pipeline.NewNLPProcessor()
			
			// Create context
			ctx := context.Background()
			
			// Test NLP processor
			model, err := processor.ParseDescription(ctx, tc.description)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for description: %s, but got none", tc.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for description: %s - %v", tc.description, err)
				}
				
				if model == nil {
					t.Errorf("Expected non-nil model for description: %s", tc.description)
				} else {
					t.Logf("Model has %d resources", len(model.Resources))
				}
			}
		})
	}
}

// TestProgressReporting tests the progress reporting functionality
func TestProgressReporting(t *testing.T) {
	// Create progress reporter
	reporter := pipeline.NewConsoleProgressReporter(3)
	
	// Create pipeline
	p := pipeline.NewBasePipeline()
	p.SetProgressReporter(reporter)
	
	// Add test stages
	p.AddStage(pipeline.NewBaseStage("Stage1", func(ctx context.Context, input interface{}) (interface{}, error) {
		return "Stage1 Output", nil
	}))
	
	p.AddStage(pipeline.NewBaseStage("Stage2", func(ctx context.Context, input interface{}) (interface{}, error) {
		return "Stage2 Output", nil
	}))
	
	p.AddStage(pipeline.NewBaseStage("Stage3", func(ctx context.Context, input interface{}) (interface{}, error) {
		return "Stage3 Output", nil
	}))
	
	// Create a channel to receive progress messages
	outputCh := reporter.OutputChannel()
	
	// Start goroutine to collect progress messages
	messageCount := 0
	go func() {
		for range outputCh {
			messageCount++
		}
	}()
	
	// Execute pipeline
	result, err := p.Execute(context.Background(), "Initial Input")
	if err != nil {
		t.Fatalf("Pipeline execution failed: %v", err)
	}
	
	// Check result
	if result != "Stage3 Output" {
		t.Errorf("Expected 'Stage3 Output', got %v", result)
	}
	
	// Close reporter and check message count
	reporter.Close()
	
	// Wait for goroutine to finish
	time.Sleep(100 * time.Millisecond)
	
	// We expect at least 6 messages (3 stages × 2 messages per stage: start and complete)
	if messageCount < 6 {
		t.Errorf("Expected at least 6 progress messages, got %d", messageCount)
	}
}

// TestErrorHandling tests that errors in the pipeline are properly handled
func TestErrorHandling(t *testing.T) {
	// Create pipeline
	p := pipeline.NewBasePipeline()
	
	// Add test stages
	p.AddStage(pipeline.NewBaseStage("SuccessStage", func(ctx context.Context, input interface{}) (interface{}, error) {
		return "Success", nil
	}))
	
	p.AddStage(pipeline.NewBaseStage("ErrorStage", func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, fmt.Errorf("simulated error")
	}))
	
	p.AddStage(pipeline.NewBaseStage("NeverReachedStage", func(ctx context.Context, input interface{}) (interface{}, error) {
		t.Error("This stage should never be executed")
		return nil, nil
	}))
	
	// Execute pipeline
	_, err := p.Execute(context.Background(), "Initial Input")
	
	// Check error
	if err == nil {
		t.Fatal("Expected pipeline to fail, but it succeeded")
	}
	
	// Verify error message
	expectedErrStr := "stage ErrorStage failed: simulated error"
	if err.Error() != expectedErrStr {
		t.Errorf("Expected error message %q, got %q", expectedErrStr, err.Error())
	}
}