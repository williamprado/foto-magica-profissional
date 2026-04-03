package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type PromptSection struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type AnalysisResult struct {
	Summary  string          `json:"summary"`
	Sections []PromptSection `json:"sections"`
}

type ImageResult struct {
	ContentType string
	Bytes       []byte
}

type Provider interface {
	AnalyzeReference(ctx context.Context, mimeType string, image []byte) (AnalysisResult, error)
	GeneratePrompt(ctx context.Context, analysis AnalysisResult, instructions string) ([]PromptSection, error)
	GenerateImage(ctx context.Context, prompt []PromptSection, userMime string, userImage []byte, referenceMime string, referenceImage []byte) (ImageResult, error)
}

type MockProvider struct{}

func (MockProvider) AnalyzeReference(_ context.Context, _ string, _ []byte) (AnalysisResult, error) {
	return AnalysisResult{
		Summary: "Mock analysis for professional studio portrait",
		Sections: []PromptSection{
			{Key: "style", Title: "STYLE", Content: "Visual executivo premium com iluminação de estúdio e composição limpa."},
			{Key: "pose", Title: "POSE", Content: "Postura confiante, enquadramento vertical e presença profissional."},
			{Key: "background", Title: "BACKGROUND", Content: "Fundo escuro com textura suave, contraste controlado e profundidade."},
		},
	}, nil
}

func (MockProvider) GeneratePrompt(_ context.Context, analysis AnalysisResult, instructions string) ([]PromptSection, error) {
	sections := append([]PromptSection{}, analysis.Sections...)
	if strings.TrimSpace(instructions) != "" {
		sections = append(sections, PromptSection{Key: "instructions", Title: "EXTRA", Content: instructions})
	}
	return sections, nil
}

func (MockProvider) GenerateImage(_ context.Context, _ []PromptSection, _ string, userImage []byte, _ string, _ []byte) (ImageResult, error) {
	return ImageResult{ContentType: "image/jpeg", Bytes: userImage}, nil
}

type GoogleProvider struct {
	client        *genai.Client
	analysisModel string
	promptModel   string
	imageModel    string
}

func NewGoogleProvider(ctx context.Context, apiKey, analysisModel, promptModel, imageModel string) (Provider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	return GoogleProvider{
		client:        client,
		analysisModel: analysisModel,
		promptModel:   promptModel,
		imageModel:    imageModel,
	}, nil
}

func (p GoogleProvider) AnalyzeReference(ctx context.Context, mimeType string, image []byte) (AnalysisResult, error) {
	parts := []*genai.Part{
		genai.NewPartFromText(`Analyze this portrait reference for a professional AI photography workflow. Return JSON with:
{
  "summary": "short summary",
  "sections": [{"key":"style","title":"STYLE","content":"..."},{"key":"pose","title":"POSE","content":"..."},{"key":"lighting","title":"LIGHTING","content":"..."}]
}`),
		{
			InlineData: &genai.Blob{MIMEType: mimeType, Data: image},
		},
	}
	contents := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}
	response, err := p.client.Models.GenerateContent(ctx, p.analysisModel, contents, nil)
	if err != nil {
		return AnalysisResult{}, err
	}

	text := partsToText(response)
	return parseAnalysisText(text)
}

func (p GoogleProvider) GeneratePrompt(ctx context.Context, analysis AnalysisResult, instructions string) ([]PromptSection, error) {
	payload, _ := json.Marshal(analysis)
	prompt := fmt.Sprintf(`You are generating a professional photo prompt in structured JSON.
Use this reference analysis: %s
Additional user instructions: %s
Return only a JSON array with sections, each having key, title, content.`, string(payload), instructions)

	response, err := p.client.Models.GenerateContent(ctx, p.promptModel, genai.Text(prompt), nil)
	if err != nil {
		return nil, err
	}

	var sections []PromptSection
	if err := json.Unmarshal([]byte(cleanJSON(partsToText(response))), &sections); err != nil {
		return nil, err
	}
	return sections, nil
}

func (p GoogleProvider) GenerateImage(ctx context.Context, prompt []PromptSection, userMime string, userImage []byte, referenceMime string, referenceImage []byte) (ImageResult, error) {
	var promptLines []string
	for _, section := range prompt {
		promptLines = append(promptLines, section.Title+": "+section.Content)
	}
	parts := []*genai.Part{
		genai.NewPartFromText("Create a professional portrait based on the following brief:\n" + strings.Join(promptLines, "\n")),
		{InlineData: &genai.Blob{MIMEType: userMime, Data: userImage}},
		{InlineData: &genai.Blob{MIMEType: referenceMime, Data: referenceImage}},
	}
	contents := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}
	response, err := p.client.Models.GenerateContent(ctx, p.imageModel, contents, &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
		ImageConfig: &genai.ImageConfig{
			AspectRatio: "4:5",
		},
	})
	if err != nil {
		return ImageResult{}, err
	}

	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && len(part.InlineData.Data) > 0 {
				return ImageResult{
					ContentType: part.InlineData.MIMEType,
					Bytes:       part.InlineData.Data,
				}, nil
			}
		}
	}

	return ImageResult{}, fmt.Errorf("no image returned by google provider")
}

func partsToText(response *genai.GenerateContentResponse) string {
	var builder strings.Builder
	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				builder.WriteString(part.Text)
			}
		}
	}
	return builder.String()
}

func parseAnalysisText(raw string) (AnalysisResult, error) {
	var result AnalysisResult
	if err := json.Unmarshal([]byte(cleanJSON(raw)), &result); err != nil {
		return AnalysisResult{}, err
	}
	return result, nil
}

func cleanJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}
