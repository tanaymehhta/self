package services

import (
	"encoding/json"
	"fmt"
	"time"
)

type PipelineStep struct {
	Step      string      `json:"step"`
	Status    string      `json:"status"` // "started", "success", "error"
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Duration  string      `json:"duration,omitempty"`
}

type PipelineLogger struct {
	Steps     []PipelineStep `json:"steps"`
	StartTime time.Time      `json:"start_time"`
	EndTime   *time.Time     `json:"end_time,omitempty"`
	Duration  string         `json:"total_duration,omitempty"`
}

func NewPipelineLogger() *PipelineLogger {
	return &PipelineLogger{
		Steps:     make([]PipelineStep, 0),
		StartTime: time.Now(),
	}
}

func (p *PipelineLogger) LogStep(step, status, message string, data interface{}) {
	p.Steps = append(p.Steps, PipelineStep{
		Step:      step,
		Status:    status,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

func (p *PipelineLogger) LogStart(step, message string) {
	p.LogStep(step, "started", message, nil)
}

func (p *PipelineLogger) LogSuccess(step, message string, data interface{}) {
	p.LogStep(step, "success", message, data)
}

func (p *PipelineLogger) LogError(step, message string, err error) {
	p.LogStep(step, "error", fmt.Sprintf("%s: %v", message, err), nil)
}

func (p *PipelineLogger) Complete() {
	now := time.Now()
	p.EndTime = &now
	p.Duration = now.Sub(p.StartTime).String()
}

func (p *PipelineLogger) GetSummary() map[string]interface{} {
	successful := 0
	failed := 0

	for _, step := range p.Steps {
		switch step.Status {
		case "success":
			successful++
		case "error":
			failed++
		}
	}

	return map[string]interface{}{
		"total_steps":      len(p.Steps),
		"successful_steps": successful,
		"failed_steps":     failed,
		"total_duration":   p.Duration,
		"status":          func() string {
			if failed > 0 {
				return "partial_success"
			}
			if successful > 0 {
				return "success"
			}
			return "in_progress"
		}(),
	}
}

func (p *PipelineLogger) Print() {
	fmt.Println("\n=== PIPELINE EXECUTION LOG ===")
	fmt.Printf("Started: %s\n", p.StartTime.Format("15:04:05"))
	if p.EndTime != nil {
		fmt.Printf("Ended: %s\n", p.EndTime.Format("15:04:05"))
		fmt.Printf("Duration: %s\n", p.Duration)
	}
	fmt.Println()

	for i, step := range p.Steps {
		status := func() string {
			switch step.Status {
			case "success":
				return "✅"
			case "error":
				return "❌"
			case "started":
				return "🔄"
			default:
				return "⏳"
			}
		}()

		fmt.Printf("%d. %s %s: %s\n", i+1, status, step.Step, step.Message)

		if step.Data != nil {
			if dataStr := p.formatData(step.Data); dataStr != "" {
				fmt.Printf("   Data: %s\n", dataStr)
			}
		}
	}

	fmt.Println()
	summary := p.GetSummary()
	fmt.Printf("Summary: %d steps, %d successful, %d failed\n",
		summary["total_steps"], summary["successful_steps"], summary["failed_steps"])
	fmt.Println("================================")
}

func (p *PipelineLogger) formatData(data interface{}) string {
	if data == nil {
		return ""
	}

	switch v := data.(type) {
	case string:
		if len(v) > 100 {
			return fmt.Sprintf("%.100s... (truncated, %d chars total)", v, len(v))
		}
		return v
	case []string:
		if len(v) == 0 {
			return "empty array"
		}
		return fmt.Sprintf("Array with %d items: [%s, ...]", len(v),
			func() string {
				if len(v[0]) > 50 {
					return v[0][:50] + "..."
				}
				return v[0]
			}())
	case map[string]interface{}:
		jsonData, _ := json.Marshal(v)
		if len(jsonData) > 200 {
			return fmt.Sprintf("JSON object (%.200s...)", string(jsonData))
		}
		return string(jsonData)
	default:
		return fmt.Sprintf("%v", data)
	}
}