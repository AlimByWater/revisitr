package campaigns

import "fmt"

var (
	ErrCampaignNotFound      = fmt.Errorf("campaign not found")
	ErrNotCampaignOwner      = fmt.Errorf("not authorized to manage this campaign")
	ErrCampaignAlreadySent   = fmt.Errorf("campaign already sent")
	ErrCampaignNotScheduled  = fmt.Errorf("campaign is not scheduled")
	ErrCampaignNotDraft      = fmt.Errorf("campaign must be in draft status")
	ErrCampaignNotSendable   = fmt.Errorf("campaign must be in draft or scheduled status")
	ErrScenarioNotFound      = fmt.Errorf("scenario not found")
	ErrNotScenarioOwner      = fmt.Errorf("not authorized to manage this scenario")
	ErrInvalidVariantPct     = fmt.Errorf("variant audience percentages must sum to 100")
	ErrVariantNotFound       = fmt.Errorf("variant not found")
	ErrTemplateNotFound      = fmt.Errorf("campaign template not found")
	ErrTemplateIsSystem      = fmt.Errorf("cannot modify system template")
	ErrNotTemplateOwner      = fmt.Errorf("not authorized to manage this template")
)
