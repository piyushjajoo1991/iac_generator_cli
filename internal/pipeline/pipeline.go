package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/riptano/iac_generator_cli/internal/utils"
	"go.uber.org/zap"
)

// BasePipeline provides a basic implementation of the Pipeline interface
type BasePipeline struct {
	stages        []Stage
	errorHandler  func(error) error
	reporter      ProgressReporter
	mu            sync.Mutex
	logger        *zap.SugaredLogger
}

// NewBasePipeline creates a new pipeline with the specified name
func NewBasePipeline() *BasePipeline {
	return &BasePipeline{
		stages:       make([]Stage, 0),
		errorHandler: defaultErrorHandler,
		logger:       utils.GetLogger(),
	}
}

// defaultErrorHandler is the default handler for pipeline errors
func defaultErrorHandler(err error) error {
	// By default, just return the error
	return err
}

// Execute implements the Pipeline interface
func (p *BasePipeline) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	p.mu.Lock()
	stages := make([]Stage, len(p.stages))
	copy(stages, p.stages)
	reporter := p.reporter
	p.mu.Unlock()

	var result interface{} = input

	for i, stage := range stages {
		// Check if context has been canceled
		if ctx.Err() != nil {
			p.logger.Warnw("Pipeline execution canceled by context", "stage", stage.Name())
			return nil, fmt.Errorf("pipeline execution canceled: %w", ctx.Err())
		}

		stageName := stage.Name()
		stageNum := i + 1
		totalStages := len(stages)
		
		// Report stage start if we have a reporter
		if reporter != nil {
			reporter.StartStage(stageName)
		}
		
		p.logger.Infow("Starting pipeline stage", 
			"stage", stageName,
			"number", fmt.Sprintf("%d/%d", stageNum, totalStages),
		)
		
		// Execute the stage with timeout handling
		resultCh := make(chan struct {
			res interface{}
			err error
		})
		
		// Run the stage execution in a separate goroutine
		go func() {
			stageResult, err := stage.Execute(ctx, result)
			resultCh <- struct {
				res interface{}
				err error
			}{stageResult, err}
		}()
		
		// Wait for stage completion or context cancellation
		select {
		case <-ctx.Done():
			// Context was canceled during stage execution
			if reporter != nil {
				reporter.FailStage(stageName, ctx.Err())
			}
			
			p.logger.Warnw("Pipeline stage interrupted", 
				"stage", stageName, 
				"error", ctx.Err())
			
			return nil, fmt.Errorf("pipeline stage %s interrupted: %w", stageName, ctx.Err())
			
		case res := <-resultCh:
			// Stage completed
			if res.err != nil {
				// Report stage failure
				if reporter != nil {
					reporter.FailStage(stageName, res.err)
				}
				
				p.logger.Errorw("Pipeline stage failed", 
					"stage", stageName, 
					"number", fmt.Sprintf("%d/%d", stageNum, totalStages),
					"error", res.err)
				
				// Wrap the error with stage context
				stageErr := fmt.Errorf("stage %s failed: %w", stageName, res.err)
				
				// Handle the error with custom handler if provided
				if p.errorHandler != nil {
					return nil, p.errorHandler(stageErr)
				}
				
				return nil, stageErr
			}
			
			// Stage succeeded, update result
			result = res.res
		}
		
		// Report stage completion
		if reporter != nil {
			reporter.CompleteStage(stageName)
		}
		
		p.logger.Infow("Completed pipeline stage", 
			"stage", stageName, 
			"number", fmt.Sprintf("%d/%d", stageNum, totalStages))
	}

	p.logger.Info("Pipeline execution completed successfully")
	return result, nil
}

// AddStage adds a stage to the pipeline
func (p *BasePipeline) AddStage(stage Stage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stages = append(p.stages, stage)
}

// SetErrorHandler sets a custom error handler for the pipeline
func (p *BasePipeline) SetErrorHandler(handler func(error) error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.errorHandler = handler
}

// SetProgressReporter sets a progress reporter for the pipeline
func (p *BasePipeline) SetProgressReporter(reporter ProgressReporter) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reporter = reporter
}

// BaseStage provides a common implementation for pipeline stages
type BaseStage struct {
	name       string
	stageFn    func(ctx context.Context, input interface{}) (interface{}, error)
	logger     *zap.SugaredLogger
}

// NewBaseStage creates a new stage with the given name and execution function
func NewBaseStage(name string, fn func(ctx context.Context, input interface{}) (interface{}, error)) *BaseStage {
	return &BaseStage{
		name:    name,
		stageFn: fn,
		logger:  utils.GetLogger(),
	}
}

// Execute runs the stage
func (s *BaseStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	s.logger.Debugw("Executing stage", "stage", s.name, "input_type", fmt.Sprintf("%T", input))
	return s.stageFn(ctx, input)
}

// Name returns the stage name
func (s *BaseStage) Name() string {
	return s.name
}

// ConsoleProgressReporter is a simple progress reporter that writes to the console
type ConsoleProgressReporter struct {
	output         chan string
	currentStage   string
	completedSteps int
	totalSteps     int
	mu             sync.Mutex
	logger         *zap.SugaredLogger
}

// NewConsoleProgressReporter creates a new console progress reporter
func NewConsoleProgressReporter(totalSteps int) *ConsoleProgressReporter {
	return &ConsoleProgressReporter{
		output:     make(chan string, 10),
		totalSteps: totalSteps,
		logger:     utils.GetLogger(),
	}
}

// StartStage implements ProgressReporter
func (r *ConsoleProgressReporter) StartStage(stageName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.currentStage = stageName
	
	// Provide more detailed messages based on stage name
	var message string
	switch stageName {
	case "NLPProcessing":
		message = "Processing natural language description..."
	case "ModelBuilding":
		message = "Building infrastructure model..."
	case "IaCGeneration":
		message = "Generating infrastructure code..."
	case "OutputWriting":
		message = "Writing output files..."
	default:
		message = fmt.Sprintf("Starting %s...", stageName)
	}
	
	r.output <- message
}

// CompleteStage implements ProgressReporter
func (r *ConsoleProgressReporter) CompleteStage(stageName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.currentStage == stageName {
		r.completedSteps++
		percentage := (r.completedSteps * 100) / r.totalSteps
		
		// Provide more detailed completion messages
		var message string
		switch stageName {
		case "NLPProcessing":
			message = "Successfully processed description"
		case "ModelBuilding":
			message = "Infrastructure model created successfully"
		case "IaCGeneration":
			message = "Infrastructure code generated successfully"
		case "OutputWriting":
			message = "Output files written successfully"
		default:
			message = fmt.Sprintf("Completed %s", stageName)
		}
		
		r.output <- fmt.Sprintf("%s (%d%%)", message, percentage)
		r.currentStage = ""
	}
}

// FailStage implements ProgressReporter
func (r *ConsoleProgressReporter) FailStage(stageName string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.currentStage == stageName {
		r.output <- fmt.Sprintf("Failed %s: %v", stageName, err)
		r.currentStage = ""
	}
}

// UpdateProgress implements ProgressReporter
func (r *ConsoleProgressReporter) UpdateProgress(message string, percentage int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.output <- fmt.Sprintf("%s (%d%%)", message, percentage)
}

// OutputChannel returns the channel where progress messages are sent
func (r *ConsoleProgressReporter) OutputChannel() <-chan string {
	return r.output
}

// Close closes the output channel
func (r *ConsoleProgressReporter) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	close(r.output)
}