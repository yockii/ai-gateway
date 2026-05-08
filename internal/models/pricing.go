package models

import "time"

// TextModelPricing 文本模型定价 (参考 Bifrost 结构)
type TextModelPricing struct {
	// 基础定价
	InputCostPerToken  float64  `json:"input_cost_per_token"`
	OutputCostPerToken float64  `json:"output_cost_per_token"`

	// 层级定价 (128k+ tokens)
	InputCostPerTokenAbove128k  *float64 `json:"input_cost_per_token_above_128k,omitempty"`
	OutputCostPerTokenAbove128k *float64 `json:"output_cost_per_token_above_128k,omitempty"`

	// 层级定价 (200k+ tokens)
	InputCostPerTokenAbove200k  *float64 `json:"input_cost_per_token_above_200k,omitempty"`
	OutputCostPerTokenAbove200k *float64 `json:"output_cost_per_token_above_200k,omitempty"`

	// 缓存定价
	CacheCreationInputTokenCost *float64 `json:"cache_creation_input_token_cost,omitempty"`
	CacheReadInputTokenCost     *float64 `json:"cache_read_input_token_cost,omitempty"`
}

// ImageModelPricing 图片模型定价
type ImageModelPricing struct {
	InputCostPerImage             *float64 `json:"input_cost_per_image,omitempty"`
	OutputCostPerImage            *float64 `json:"output_cost_per_image,omitempty"`
	OutputCostPerImageLowQuality  *float64 `json:"output_cost_per_image_low_quality,omitempty"`
	OutputCostPerImageHighQuality *float64 `json:"output_cost_per_image_high_quality,omitempty"`
	OutputCostPerImageAbove1024x1024Pixels *float64 `json:"output_cost_per_image_above_1024x1024_pixels,omitempty"`
	OutputCostPerImageAbove2048x2048Pixels *float64 `json:"output_cost_per_image_above_2048x2048_pixels,omitempty"`
}

// VideoAudioPricing 视频/音频定价
type VideoAudioPricing struct {
	InputCostPerVideoPerSecond  *float64 `json:"input_cost_per_video_per_second,omitempty"`
	OutputCostPerVideoPerSecond *float64 `json:"output_cost_per_video_per_second,omitempty"`
	InputCostPerAudioPerSecond  *float64 `json:"input_cost_per_audio_per_second,omitempty"`
	OutputCostPerAudioPerSecond *float64 `json:"output_cost_per_audio_per_second,omitempty"`
	InputCostPerAudioToken      *float64 `json:"input_cost_per_audio_token,omitempty"`
	OutputCostPerAudioToken     *float64 `json:"output_cost_per_audio_token,omitempty"`
}

// SupplierCostPricingExtended 供应商成本定价 (扩展现有模型)
type SupplierCostPricingExtended struct {
	ID            string            `json:"id" gorm:"primaryKey"`
	SupplierID    string            `json:"supplier_id" gorm:"index"`
	ModelID       string            `json:"model_id" gorm:"index"`

	// Pricing structures
	TextPricing     *TextModelPricing     `json:"text_pricing" gorm:"embedded;embeddedPrefix:text_"`
	ImagePricing    *ImageModelPricing    `json:"image_pricing" gorm:"embedded;embeddedPrefix:image_"`
	VideoAudioPricing *VideoAudioPricing  `json:"video_audio_pricing" gorm:"embedded;embeddedPrefix:video_"`

	EffectiveDate time.Time `json:"effective_date"`
	IsActive      bool      `json:"is_active" gorm:"index"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserGroupPricingExtended 用户群体定价 (扩展现有模型)
type UserGroupPricingExtended struct {
	ID               string            `json:"id" gorm:"primaryKey"`
	UserGroupID      string            `json:"user_group_id" gorm:"index"`
	ModelID          string            `json:"model_id" gorm:"index"`

	// Pricing structures
	TextPricing     *TextModelPricing     `json:"text_pricing" gorm:"embedded;embeddedPrefix:text_"`
	ImagePricing    *ImageModelPricing    `json:"image_pricing" gorm:"embedded;embeddedPrefix:image_"`
	VideoAudioPricing *VideoAudioPricing  `json:"video_audio_pricing" gorm:"embedded;embeddedPrefix:video_"`

	MinProfitMargin float64   `json:"min_profit_margin"`
	EffectiveDate   time.Time `json:"effective_date"`
	IsActive        bool      `json:"is_active" gorm:"index"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
