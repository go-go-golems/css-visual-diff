package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/ai"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

type AIReviewResult struct {
	Sections []AISectionReview `json:"sections"`
}

type AISectionReview struct {
	Name     string     `json:"name"`
	Selector string     `json:"selector"`
	Question string     `json:"question"`
	Original AIResponse `json:"original"`
	React    AIResponse `json:"react"`
}

type AIResponse struct {
	Screenshot string    `json:"screenshot"`
	Answer     ai.Answer `json:"answer"`
}

func RunAIReview(ctx context.Context, cfg *config.Config) error {
	capturePath := filepath.Join(cfg.Output.Dir, "capture.json")
	data, err := os.ReadFile(capturePath)
	if err != nil {
		return fmt.Errorf("ai-review requires capture.json: %w", err)
	}

	var capture CaptureResult
	if err := json.Unmarshal(data, &capture); err != nil {
		return err
	}

	client := ai.NoopClient{}
	result := AIReviewResult{}

	for i, section := range cfg.Sections {
		if section.OCRQuestion == "" {
			continue
		}

		review := AISectionReview{
			Name:     section.Name,
			Selector: section.Selector,
			Question: section.OCRQuestion,
		}

		if i < len(capture.Original.Sections) {
			origShot := capture.Original.Sections[i].Screenshot
			review.Original.Screenshot = origShot
			ans, err := client.AnswerQuestion(ctx, origShot, section.OCRQuestion)
			if err != nil {
				ans.Error = err.Error()
			}
			review.Original.Answer = ans
		}
		if i < len(capture.React.Sections) {
			reactShot := capture.React.Sections[i].Screenshot
			review.React.Screenshot = reactShot
			ans, err := client.AnswerQuestion(ctx, reactShot, section.OCRQuestion)
			if err != nil {
				ans.Error = err.Error()
			}
			review.React.Answer = ans
		}

		result.Sections = append(result.Sections, review)
	}

	if cfg.Output.WriteJSON {
		if err := writeJSON(filepath.Join(cfg.Output.Dir, "ai-review.json"), result); err != nil {
			return err
		}
	}
	if cfg.Output.WriteMarkdown {
		if err := writeAIMarkdown(filepath.Join(cfg.Output.Dir, "ai-review.md"), result); err != nil {
			return err
		}
	}

	return nil
}

func writeAIMarkdown(path string, result AIReviewResult) error {
	content := "# css-visual-diff AI Review Report\n\n"
	for _, s := range result.Sections {
		content += fmt.Sprintf("## %s\n\n", s.Name)
		content += fmt.Sprintf("Question: %s\n\n", s.Question)
		content += "### Original\n\n"
		content += fmt.Sprintf("Screenshot: %s\n\n", s.Original.Screenshot)
		content += fmt.Sprintf("Answer: %s\n\n", s.Original.Answer.Text)
		if s.Original.Answer.Error != "" {
			content += fmt.Sprintf("Error: %s\n\n", s.Original.Answer.Error)
		}
		content += "### React\n\n"
		content += fmt.Sprintf("Screenshot: %s\n\n", s.React.Screenshot)
		content += fmt.Sprintf("Answer: %s\n\n", s.React.Answer.Text)
		if s.React.Answer.Error != "" {
			content += fmt.Sprintf("Error: %s\n\n", s.React.Answer.Error)
		}
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
