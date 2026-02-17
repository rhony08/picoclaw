package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/caarlos0/env/v11"
)

// FlexibleStringSlice is a []string that also accepts JSON numbers,
// so allow_from can contain both "123" and 123.
type FlexibleStringSlice []string

func (f *FlexibleStringSlice) UnmarshalJSON(data []byte) error {
	// Try []string first
	var ss []string
	if err := json.Unmarshal(data, &ss); err == nil {
		*f = ss
		return nil
	}

	// Try []interface{} to handle mixed types
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	result := make([]string, 0, len(raw))
	for _, v := range raw {
		switch val := v.(type) {
		case string:
			result = append(result, val)
		case float64:
			result = append(result, fmt.Sprintf("%.0f", val))
		default:
			result = append(result, fmt.Sprintf("%v", val))
		}
	}
	*f = result
	return nil
}

type Config struct {
	Agents    AgentsConfig    `json:"agents"`
	Channels  ChannelsConfig  `json:"channels"`
	Providers ProvidersConfig `json:"providers"`
	Gateway   GatewayConfig   `json:"gateway"`
	Tools     ToolsConfig     `json:"tools"`
	Heartbeat HeartbeatConfig `json:"heartbeat"`
	Devices   DevicesConfig   `json:"devices"`
	mu        sync.RWMutex
}

type AgentsConfig struct {
	Defaults AgentDefaults           `json:"defaults"`
	Profiles map[string]AgentProfile `json:"profiles,omitempty"`
	Routing  []RoutingRule           `json:"routing,omitempty"`
}

type ModelSpec struct {
	Model    string `json:"model,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func (m ModelSpec) ResolvedModel() string {
	// Return just the model name without provider prefix.
	// The provider field is used for routing/config purposes only,
	// not for the actual API call. The API expects just the model name.
	name := strings.TrimSpace(m.Model)
	return name
}

// AgentProfile defines a named agent configuration that can be selected per session
type AgentProfile struct {
	Workspace           string      `json:"workspace,omitempty" env:"PICOCLAW_AGENTS_PROFILES_{{.Name}}_WORKSPACE"`
	RestrictToWorkspace bool        `json:"restrict_to_workspace,omitempty" env:"PICOCLAW_AGENTS_PROFILES_{{.Name}}_RESTRICT_TO_WORKSPACE"`
	Provider            string      `json:"provider,omitempty" env:"PICOCLAW_AGENTS_PROFILES_{{.Name}}_PROVIDER"`
	Model               string      `json:"model,omitempty" env:"PICOCLAW_AGENTS_PROFILES_{{.Name}}_MODEL"`
	MaxTokens           int         `json:"max_tokens,omitempty" env:"PICOCLAW_AGENTS_PROFILES_{{.Name}}_MAX_TOKENS"`
	Temperature         float64     `json:"temperature,omitempty" env:"PICOCLAW_AGENTS_PROFILES_{{.Name}}_TEMPERATURE"`
	MaxToolIterations   int         `json:"max_tool_iterations,omitempty" env:"PICOCLAW_AGENTS_PROFILES_{{.Name}}_MAX_TOOL_ITERATIONS"`
	SystemPrompt        string      `json:"system_prompt,omitempty" env:"PICOCLAW_AGENTS_PROFILES_{{.Name}}_SYSTEM_PROMPT"`
	AllowedTools        []string    `json:"allowed_tools,omitempty" env:"PICOCLAW_AGENTS_PROFILES_{{.Name}}_ALLOWED_TOOLS"`
	Models              []ModelSpec `json:"models,omitempty"`
	ResolvedModels      []string    `json:"-"`
}

func (ap *AgentProfile) prepareModels() {
	ap.ResolvedModels = buildResolvedModelList(ap.Model, ap.Models)
	if len(ap.ResolvedModels) > 0 {
		ap.Model = ap.ResolvedModels[0]
	}
}

func (ap AgentProfile) ModelCandidates() []string {
	if len(ap.ResolvedModels) > 0 {
		return ap.ResolvedModels
	}
	if ap.Model != "" {
		return []string{ap.Model}
	}
	return nil
}

// RoutingRule defines a rule to automatically assign agents based on channel and user
type RoutingRule struct {
	Channel string   `json:"channel" env:"PICOCLAW_AGENTS_ROUTING_{{.Index}}_CHANNEL"`
	Agent   string   `json:"agent" env:"PICOCLAW_AGENTS_ROUTING_{{.Index}}_AGENT"`
	UserIDs []string `json:"user_ids,omitempty" env:"PICOCLAW_AGENTS_ROUTING_{{.Index}}_USER_IDS"`
	UserID  string   `json:"user_id,omitempty" env:"PICOCLAW_AGENTS_ROUTING_{{.Index}}_USER_ID"`
}

// EffectiveUserIDs returns the configured user IDs for this rule, including deprecated single user ID support.
func (r *RoutingRule) EffectiveUserIDs() []string {
	if len(r.UserIDs) == 0 && r.UserID != "" {
		return []string{r.UserID}
	}
	return r.UserIDs
}

func buildResolvedModelList(base string, specs []ModelSpec) []string {
	resolved := make([]string, 0, len(specs))
	for _, spec := range specs {
		name := spec.ResolvedModel()
		if name == "" {
			continue
		}
		resolved = append(resolved, name)
	}
	if len(resolved) == 0 && base != "" {
		resolved = append(resolved, base)
	}
	return resolved
}

type AgentDefaults struct {
	Workspace           string      `json:"workspace" env:"PICOCLAW_AGENTS_DEFAULTS_WORKSPACE"`
	RestrictToWorkspace bool        `json:"restrict_to_workspace" env:"PICOCLAW_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE"`
	Provider            string      `json:"provider" env:"PICOCLAW_AGENTS_DEFAULTS_PROVIDER"`
	Model               string      `json:"model" env:"PICOCLAW_AGENTS_DEFAULTS_MODEL"`
	MaxTokens           int         `json:"max_tokens" env:"PICOCLAW_AGENTS_DEFAULTS_MAX_TOKENS"`
	Temperature         float64     `json:"temperature" env:"PICOCLAW_AGENTS_DEFAULTS_TEMPERATURE"`
	MaxToolIterations   int         `json:"max_tool_iterations" env:"PICOCLAW_AGENTS_DEFAULTS_MAX_TOOL_ITERATIONS"`
	Models              []ModelSpec `json:"models,omitempty"`
	ResolvedModels      []string    `json:"-"`
}

func (d *AgentDefaults) prepareModels() {
	d.ResolvedModels = buildResolvedModelList(d.Model, d.Models)
	if len(d.ResolvedModels) > 0 {
		d.Model = d.ResolvedModels[0]
	}
}

func (d AgentDefaults) ModelCandidates() []string {
	if len(d.ResolvedModels) > 0 {
		return d.ResolvedModels
	}
	if d.Model != "" {
		return []string{d.Model}
	}
	return nil
}

type ChannelsConfig struct {
	WhatsApp WhatsAppConfig `json:"whatsapp"`
	Telegram TelegramConfig `json:"telegram"`
	Feishu   FeishuConfig   `json:"feishu"`
	Discord  DiscordConfig  `json:"discord"`
	MaixCam  MaixCamConfig  `json:"maixcam"`
	QQ       QQConfig       `json:"qq"`
	DingTalk DingTalkConfig `json:"dingtalk"`
	Slack    SlackConfig    `json:"slack"`
	LINE     LINEConfig     `json:"line"`
	OneBot   OneBotConfig   `json:"onebot"`
}

type WhatsAppConfig struct {
	Enabled   bool                `json:"enabled" env:"PICOCLAW_CHANNELS_WHATSAPP_ENABLED"`
	BridgeURL string              `json:"bridge_url" env:"PICOCLAW_CHANNELS_WHATSAPP_BRIDGE_URL"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_WHATSAPP_ALLOW_FROM"`
}

type TelegramConfig struct {
	Enabled   bool                `json:"enabled" env:"PICOCLAW_CHANNELS_TELEGRAM_ENABLED"`
	Token     string              `json:"token" env:"PICOCLAW_CHANNELS_TELEGRAM_TOKEN"`
	Proxy     string              `json:"proxy" env:"PICOCLAW_CHANNELS_TELEGRAM_PROXY"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_TELEGRAM_ALLOW_FROM"`
}

type FeishuConfig struct {
	Enabled           bool                `json:"enabled" env:"PICOCLAW_CHANNELS_FEISHU_ENABLED"`
	AppID             string              `json:"app_id" env:"PICOCLAW_CHANNELS_FEISHU_APP_ID"`
	AppSecret         string              `json:"app_secret" env:"PICOCLAW_CHANNELS_FEISHU_APP_SECRET"`
	EncryptKey        string              `json:"encrypt_key" env:"PICOCLAW_CHANNELS_FEISHU_ENCRYPT_KEY"`
	VerificationToken string              `json:"verification_token" env:"PICOCLAW_CHANNELS_FEISHU_VERIFICATION_TOKEN"`
	AllowFrom         FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_FEISHU_ALLOW_FROM"`
}

type DiscordConfig struct {
	Enabled   bool                `json:"enabled" env:"PICOCLAW_CHANNELS_DISCORD_ENABLED"`
	Token     string              `json:"token" env:"PICOCLAW_CHANNELS_DISCORD_TOKEN"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_DISCORD_ALLOW_FROM"`
}

type MaixCamConfig struct {
	Enabled   bool                `json:"enabled" env:"PICOCLAW_CHANNELS_MAIXCAM_ENABLED"`
	Host      string              `json:"host" env:"PICOCLAW_CHANNELS_MAIXCAM_HOST"`
	Port      int                 `json:"port" env:"PICOCLAW_CHANNELS_MAIXCAM_PORT"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_MAIXCAM_ALLOW_FROM"`
}

type QQConfig struct {
	Enabled   bool                `json:"enabled" env:"PICOCLAW_CHANNELS_QQ_ENABLED"`
	AppID     string              `json:"app_id" env:"PICOCLAW_CHANNELS_QQ_APP_ID"`
	AppSecret string              `json:"app_secret" env:"PICOCLAW_CHANNELS_QQ_APP_SECRET"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_QQ_ALLOW_FROM"`
}

type DingTalkConfig struct {
	Enabled      bool                `json:"enabled" env:"PICOCLAW_CHANNELS_DINGTALK_ENABLED"`
	ClientID     string              `json:"client_id" env:"PICOCLAW_CHANNELS_DINGTALK_CLIENT_ID"`
	ClientSecret string              `json:"client_secret" env:"PICOCLAW_CHANNELS_DINGTALK_CLIENT_SECRET"`
	AllowFrom    FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_DINGTALK_ALLOW_FROM"`
}

type SlackConfig struct {
	Enabled   bool                `json:"enabled" env:"PICOCLAW_CHANNELS_SLACK_ENABLED"`
	BotToken  string              `json:"bot_token" env:"PICOCLAW_CHANNELS_SLACK_BOT_TOKEN"`
	AppToken  string              `json:"app_token" env:"PICOCLAW_CHANNELS_SLACK_APP_TOKEN"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_SLACK_ALLOW_FROM"`
}

type LINEConfig struct {
	Enabled            bool                `json:"enabled" env:"PICOCLAW_CHANNELS_LINE_ENABLED"`
	ChannelSecret      string              `json:"channel_secret" env:"PICOCLAW_CHANNELS_LINE_CHANNEL_SECRET"`
	ChannelAccessToken string              `json:"channel_access_token" env:"PICOCLAW_CHANNELS_LINE_CHANNEL_ACCESS_TOKEN"`
	WebhookHost        string              `json:"webhook_host" env:"PICOCLAW_CHANNELS_LINE_WEBHOOK_HOST"`
	WebhookPort        int                 `json:"webhook_port" env:"PICOCLAW_CHANNELS_LINE_WEBHOOK_PORT"`
	WebhookPath        string              `json:"webhook_path" env:"PICOCLAW_CHANNELS_LINE_WEBHOOK_PATH"`
	AllowFrom          FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_LINE_ALLOW_FROM"`
}

type OneBotConfig struct {
	Enabled            bool                `json:"enabled" env:"PICOCLAW_CHANNELS_ONEBOT_ENABLED"`
	WSUrl              string              `json:"ws_url" env:"PICOCLAW_CHANNELS_ONEBOT_WS_URL"`
	AccessToken        string              `json:"access_token" env:"PICOCLAW_CHANNELS_ONEBOT_ACCESS_TOKEN"`
	ReconnectInterval  int                 `json:"reconnect_interval" env:"PICOCLAW_CHANNELS_ONEBOT_RECONNECT_INTERVAL"`
	GroupTriggerPrefix []string            `json:"group_trigger_prefix" env:"PICOCLAW_CHANNELS_ONEBOT_GROUP_TRIGGER_PREFIX"`
	AllowFrom          FlexibleStringSlice `json:"allow_from" env:"PICOCLAW_CHANNELS_ONEBOT_ALLOW_FROM"`
}

type HeartbeatConfig struct {
	Enabled  bool `json:"enabled" env:"PICOCLAW_HEARTBEAT_ENABLED"`
	Interval int  `json:"interval" env:"PICOCLAW_HEARTBEAT_INTERVAL"` // minutes, min 5
}

type DevicesConfig struct {
	Enabled    bool `json:"enabled" env:"PICOCLAW_DEVICES_ENABLED"`
	MonitorUSB bool `json:"monitor_usb" env:"PICOCLAW_DEVICES_MONITOR_USB"`
}

type ProvidersConfig struct {
	Anthropic     ProviderConfig `json:"anthropic"`
	OpenAI        ProviderConfig `json:"openai"`
	OpenRouter    ProviderConfig `json:"openrouter"`
	Groq          ProviderConfig `json:"groq"`
	Zhipu         ProviderConfig `json:"zhipu"`
	VLLM          ProviderConfig `json:"vllm"`
	Gemini        ProviderConfig `json:"gemini"`
	Nvidia        ProviderConfig `json:"nvidia"`
	Ollama        ProviderConfig `json:"ollama"`
	Moonshot      ProviderConfig `json:"moonshot"`
	ShengSuanYun  ProviderConfig `json:"shengsuanyun"`
	DeepSeek      ProviderConfig `json:"deepseek"`
	GitHubCopilot ProviderConfig `json:"github_copilot"`
}

type ProviderConfig struct {
	APIKey      string `json:"api_key" env:"PICOCLAW_PROVIDERS_{{.Name}}_API_KEY"`
	APIBase     string `json:"api_base" env:"PICOCLAW_PROVIDERS_{{.Name}}_API_BASE"`
	Proxy       string `json:"proxy,omitempty" env:"PICOCLAW_PROVIDERS_{{.Name}}_PROXY"`
	AuthMethod  string `json:"auth_method,omitempty" env:"PICOCLAW_PROVIDERS_{{.Name}}_AUTH_METHOD"`
	ConnectMode string `json:"connect_mode,omitempty" env:"PICOCLAW_PROVIDERS_{{.Name}}_CONNECT_MODE"` //only for Github Copilot, `stdio` or `grpc`
}

type GatewayConfig struct {
	Host string `json:"host" env:"PICOCLAW_GATEWAY_HOST"`
	Port int    `json:"port" env:"PICOCLAW_GATEWAY_PORT"`
}

type BraveConfig struct {
	Enabled    bool   `json:"enabled" env:"PICOCLAW_TOOLS_WEB_BRAVE_ENABLED"`
	APIKey     string `json:"api_key" env:"PICOCLAW_TOOLS_WEB_BRAVE_API_KEY"`
	MaxResults int    `json:"max_results" env:"PICOCLAW_TOOLS_WEB_BRAVE_MAX_RESULTS"`
}

type DuckDuckGoConfig struct {
	Enabled    bool `json:"enabled" env:"PICOCLAW_TOOLS_WEB_DUCKDUCKGO_ENABLED"`
	MaxResults int  `json:"max_results" env:"PICOCLAW_TOOLS_WEB_DUCKDUCKGO_MAX_RESULTS"`
}

type WebToolsConfig struct {
	Brave      BraveConfig      `json:"brave"`
	DuckDuckGo DuckDuckGoConfig `json:"duckduckgo"`
}

type ToolsConfig struct {
	Web WebToolsConfig `json:"web"`
}

// PrepareAgentModels resolves configured model specs for defaults and profiles.
func (c *Config) PrepareAgentModels() {
	c.Agents.Defaults.prepareModels()
	for name, profile := range c.Agents.Profiles {
		profile.prepareModels()
		c.Agents.Profiles[name] = profile
	}
}

func DefaultConfig() *Config {
	return &Config{
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Workspace:           "~/.picoclaw/workspace",
				RestrictToWorkspace: true,
				Provider:            "",
				Model:               "glm-4.7",
				MaxTokens:           8192,
				Temperature:         0.7,
				MaxToolIterations:   20,
			},
		},
		Channels: ChannelsConfig{
			WhatsApp: WhatsAppConfig{
				Enabled:   false,
				BridgeURL: "ws://localhost:3001",
				AllowFrom: FlexibleStringSlice{},
			},
			Telegram: TelegramConfig{
				Enabled:   false,
				Token:     "",
				AllowFrom: FlexibleStringSlice{},
			},
			Feishu: FeishuConfig{
				Enabled:           false,
				AppID:             "",
				AppSecret:         "",
				EncryptKey:        "",
				VerificationToken: "",
				AllowFrom:         FlexibleStringSlice{},
			},
			Discord: DiscordConfig{
				Enabled:   false,
				Token:     "",
				AllowFrom: FlexibleStringSlice{},
			},
			MaixCam: MaixCamConfig{
				Enabled:   false,
				Host:      "0.0.0.0",
				Port:      18790,
				AllowFrom: FlexibleStringSlice{},
			},
			QQ: QQConfig{
				Enabled:   false,
				AppID:     "",
				AppSecret: "",
				AllowFrom: FlexibleStringSlice{},
			},
			DingTalk: DingTalkConfig{
				Enabled:      false,
				ClientID:     "",
				ClientSecret: "",
				AllowFrom:    FlexibleStringSlice{},
			},
			Slack: SlackConfig{
				Enabled:   false,
				BotToken:  "",
				AppToken:  "",
				AllowFrom: FlexibleStringSlice{},
			},
			LINE: LINEConfig{
				Enabled:            false,
				ChannelSecret:      "",
				ChannelAccessToken: "",
				WebhookHost:        "0.0.0.0",
				WebhookPort:        18791,
				WebhookPath:        "/webhook/line",
				AllowFrom:          FlexibleStringSlice{},
			},
			OneBot: OneBotConfig{
				Enabled:            false,
				WSUrl:              "ws://127.0.0.1:3001",
				AccessToken:        "",
				ReconnectInterval:  5,
				GroupTriggerPrefix: []string{},
				AllowFrom:          FlexibleStringSlice{},
			},
		},
		Providers: ProvidersConfig{
			Anthropic:    ProviderConfig{},
			OpenAI:       ProviderConfig{},
			OpenRouter:   ProviderConfig{},
			Groq:         ProviderConfig{},
			Zhipu:        ProviderConfig{},
			VLLM:         ProviderConfig{},
			Gemini:       ProviderConfig{},
			Nvidia:       ProviderConfig{},
			Moonshot:     ProviderConfig{},
			ShengSuanYun: ProviderConfig{},
		},
		Gateway: GatewayConfig{
			Host: "0.0.0.0",
			Port: 18790,
		},
		Tools: ToolsConfig{
			Web: WebToolsConfig{
				Brave: BraveConfig{
					Enabled:    false,
					APIKey:     "",
					MaxResults: 5,
				},
				DuckDuckGo: DuckDuckGoConfig{
					Enabled:    true,
					MaxResults: 5,
				},
			},
		},
		Heartbeat: HeartbeatConfig{
			Enabled:  true,
			Interval: 30, // default 30 minutes
		},
		Devices: DevicesConfig{
			Enabled:    false,
			MonitorUSB: true,
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) WorkspacePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return expandHome(c.Agents.Defaults.Workspace)
}

func (c *Config) GetAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Providers.OpenRouter.APIKey != "" {
		return c.Providers.OpenRouter.APIKey
	}
	if c.Providers.Anthropic.APIKey != "" {
		return c.Providers.Anthropic.APIKey
	}
	if c.Providers.OpenAI.APIKey != "" {
		return c.Providers.OpenAI.APIKey
	}
	if c.Providers.Gemini.APIKey != "" {
		return c.Providers.Gemini.APIKey
	}
	if c.Providers.Zhipu.APIKey != "" {
		return c.Providers.Zhipu.APIKey
	}
	if c.Providers.Groq.APIKey != "" {
		return c.Providers.Groq.APIKey
	}
	if c.Providers.VLLM.APIKey != "" {
		return c.Providers.VLLM.APIKey
	}
	if c.Providers.ShengSuanYun.APIKey != "" {
		return c.Providers.ShengSuanYun.APIKey
	}
	return ""
}

func (c *Config) GetAPIBase() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Providers.OpenRouter.APIKey != "" {
		if c.Providers.OpenRouter.APIBase != "" {
			return c.Providers.OpenRouter.APIBase
		}
		return "https://openrouter.ai/api/v1"
	}
	if c.Providers.Zhipu.APIKey != "" {
		return c.Providers.Zhipu.APIBase
	}
	if c.Providers.VLLM.APIKey != "" && c.Providers.VLLM.APIBase != "" {
		return c.Providers.VLLM.APIBase
	}
	return ""
}

func expandHome(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		if len(path) > 1 && path[1] == '/' {
			return home + path[1:]
		}
		return home
	}
	return path
}

// GetAgentProfile returns the agent profile by name, merging with defaults
// If profile doesn't exist, returns defaults
func (c *Config) GetAgentProfile(name string) AgentProfile {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Start with defaults
	profile := AgentProfile{
		Workspace:           c.Agents.Defaults.Workspace,
		RestrictToWorkspace: c.Agents.Defaults.RestrictToWorkspace,
		Provider:            c.Agents.Defaults.Provider,
		Model:               c.Agents.Defaults.Model,
		Models:              c.Agents.Defaults.Models,
		ResolvedModels:      c.Agents.Defaults.ResolvedModels,
		MaxTokens:           c.Agents.Defaults.MaxTokens,
		Temperature:         c.Agents.Defaults.Temperature,
		MaxToolIterations:   c.Agents.Defaults.MaxToolIterations,
	}

	// If name is empty or "default", return defaults
	if name == "" || name == "default" {
		profile.prepareModels()
		return profile
	}

	// Merge with profile if exists
	if p, ok := c.Agents.Profiles[name]; ok {
		if p.Workspace != "" {
			profile.Workspace = p.Workspace
		}
		if p.Provider != "" {
			profile.Provider = p.Provider
		}
		if p.Model != "" {
			profile.Model = p.Model
		}
		if p.MaxTokens != 0 {
			profile.MaxTokens = p.MaxTokens
		}
		if p.Temperature != 0 {
			profile.Temperature = p.Temperature
		}
		if p.MaxToolIterations != 0 {
			profile.MaxToolIterations = p.MaxToolIterations
		}
		if p.SystemPrompt != "" {
			profile.SystemPrompt = p.SystemPrompt
		}
		if len(p.AllowedTools) > 0 {
			profile.AllowedTools = p.AllowedTools
		}
		// Bool fields - always use profile value if explicitly set (not default false)
		profile.RestrictToWorkspace = p.RestrictToWorkspace
		if len(p.Models) > 0 {
			profile.Models = p.Models
		}
		if len(p.ResolvedModels) > 0 {
			profile.ResolvedModels = p.ResolvedModels
		}
	}
	profile.prepareModels()

	return profile
}

// GetRoutedAgent returns the agent name for a given channel and user ID
// Returns empty string if no routing rule matches
func (c *Config) GetRoutedAgent(channel, userID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, rule := range c.Agents.Routing {
		if rule.Channel != channel {
			continue
		}
		userIDs := rule.EffectiveUserIDs()
		if len(userIDs) == 0 {
			return rule.Agent
		}
		for _, uid := range userIDs {
			if uid == "*" && userID != "" {
				return rule.Agent
			}
			if uid == userID {
				return rule.Agent
			}
		}
	}
	return ""
}

// ListAgentProfiles returns all available agent profile names
func (c *Config) ListAgentProfiles() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	profiles := make([]string, 0, len(c.Agents.Profiles)+1)
	profiles = append(profiles, "default")
	for name := range c.Agents.Profiles {
		profiles = append(profiles, name)
	}
	return profiles
}

// ProfileExists checks if an agent profile exists
func (c *Config) ProfileExists(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if name == "default" || name == "" {
		return true
	}
	_, ok := c.Agents.Profiles[name]
	return ok
}
