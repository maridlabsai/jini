package app

type adapterDescriptor struct {
	ID                 string
	Label              string
	ProviderMode       string
	CostTier           string
	Locality           string
	Modalities         []string
	SupportsEffort     bool
	SupportsBenchmark  bool
	SupportsAutoRoute  bool
	LocalProfileTarget string
}

func adapterRegistry() map[string]adapterDescriptor {
	return map[string]adapterDescriptor{
		"codex": {
			ID:                "codex",
			Label:             "Codex CLI handoff",
			ProviderMode:      "cli-handoff",
			CostTier:          "external",
			Locality:          "cli",
			Modalities:        []string{"text", "code"},
			SupportsAutoRoute: true,
		},
		"claude-code": {
			ID:                "claude-code",
			Label:             "Claude Code CLI handoff",
			ProviderMode:      "cli-handoff",
			CostTier:          "external",
			Locality:          "cli",
			Modalities:        []string{"text", "code"},
			SupportsAutoRoute: true,
		},
		"gemini-cli": {
			ID:                "gemini-cli",
			Label:             "Gemini CLI handoff",
			ProviderMode:      "cli-handoff",
			CostTier:          "external",
			Locality:          "cli",
			Modalities:        []string{"text", "code"},
			SupportsAutoRoute: true,
		},
		"aider": {
			ID:                "aider",
			Label:             "Aider CLI handoff",
			ProviderMode:      "cli-handoff",
			CostTier:          "external",
			Locality:          "cli",
			Modalities:        []string{"text", "code"},
			SupportsAutoRoute: true,
		},
		"opencode": {
			ID:                "opencode",
			Label:             "OpenCode CLI handoff",
			ProviderMode:      "cli-handoff",
			CostTier:          "external",
			Locality:          "cli",
			Modalities:        []string{"text", "code"},
			SupportsAutoRoute: true,
		},
		"claude-api": {
			ID:                "claude-api",
			Label:             "Claude API route",
			ProviderMode:      "anthropic",
			CostTier:          "premium",
			Locality:          "remote",
			Modalities:        []string{"text", "code"},
			SupportsEffort:    true,
			SupportsAutoRoute: true,
		},
		"bedrock-sonnet": {
			ID:                "bedrock-sonnet",
			Label:             "Bedrock Sonnet",
			ProviderMode:      "bedrock",
			CostTier:          "premium",
			Locality:          "remote",
			Modalities:        []string{"text", "code"},
			SupportsEffort:    true,
			SupportsAutoRoute: true,
		},
		"chatgpt": {
			ID:                "chatgpt",
			Label:             "Azure writing route",
			ProviderMode:      "azure-openai",
			CostTier:          "standard",
			Locality:          "remote",
			Modalities:        []string{"text"},
			SupportsEffort:    true,
			SupportsAutoRoute: true,
		},
		"azure-code": {
			ID:                "azure-code",
			Label:             "Azure code route",
			ProviderMode:      "azure-openai",
			CostTier:          "standard",
			Locality:          "remote",
			Modalities:        []string{"text", "code"},
			SupportsEffort:    true,
			SupportsAutoRoute: true,
		},
		"azure-openai": {
			ID:                "azure-openai",
			Label:             "Azure OpenAI",
			ProviderMode:      "azure-openai",
			CostTier:          "standard",
			Locality:          "remote",
			Modalities:        []string{"text", "code"},
			SupportsEffort:    true,
			SupportsAutoRoute: true,
		},
		"local-fast": {
			ID:                 "local-fast",
			Label:              "Local SLM fast",
			ProviderMode:       "local-slm",
			CostTier:           "local-cheap",
			Locality:           "local",
			Modalities:         []string{"text"},
			SupportsEffort:     true,
			SupportsBenchmark:  true,
			SupportsAutoRoute:  true,
			LocalProfileTarget: "fast",
		},
		"local-workhorse": {
			ID:                 "local-workhorse",
			Label:              "Local SLM workhorse",
			ProviderMode:       "local-slm",
			CostTier:           "local-cheap",
			Locality:           "local",
			Modalities:         []string{"text", "code"},
			SupportsEffort:     true,
			SupportsBenchmark:  true,
			SupportsAutoRoute:  true,
			LocalProfileTarget: "workhorse",
		},
		"local-deep": {
			ID:                 "local-deep",
			Label:              "Local SLM deep",
			ProviderMode:       "local-slm",
			CostTier:           "local-cheap",
			Locality:           "local",
			Modalities:         []string{"text", "code"},
			SupportsEffort:     true,
			SupportsBenchmark:  true,
			SupportsAutoRoute:  true,
			LocalProfileTarget: "deep",
		},
		"local-multimodal": {
			ID:                 "local-multimodal",
			Label:              "Local SLM multimodal",
			ProviderMode:       "local-slm",
			CostTier:           "local-cheap",
			Locality:           "local",
			Modalities:         []string{"text", "image", "audio", "pdf"},
			SupportsEffort:     true,
			SupportsBenchmark:  true,
			SupportsAutoRoute:  true,
			LocalProfileTarget: "multimodal",
		},
		"local-preview": {
			ID:                "local-preview",
			Label:             "Local preview",
			ProviderMode:      "local-preview",
			CostTier:          "free",
			Locality:          "local",
			Modalities:        []string{"text"},
			SupportsEffort:    false,
			SupportsAutoRoute: true,
		},
	}
}

func adapterDescriptorForMode(mode string) (adapterDescriptor, bool) {
	descriptor, ok := adapterRegistry()[mode]
	return descriptor, ok
}

func localBenchmarkableAdapterModes() []string {
	modes := []string{}
	for _, candidate := range []string{"local-fast", "local-workhorse", "local-deep", "local-multimodal"} {
		descriptor, ok := adapterDescriptorForMode(candidate)
		if !ok || descriptor.ProviderMode != "local-slm" || !descriptor.SupportsBenchmark {
			continue
		}
		modes = append(modes, descriptor.ID)
	}
	return modes
}
