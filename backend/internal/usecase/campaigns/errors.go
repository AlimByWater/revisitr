package campaigns

import "fmt"

var (
	ErrCampaignNotFound    = fmt.Errorf("campaign not found")
	ErrNotCampaignOwner    = fmt.Errorf("not authorized to manage this campaign")
	ErrCampaignAlreadySent = fmt.Errorf("campaign already sent")
	ErrScenarioNotFound    = fmt.Errorf("scenario not found")
	ErrNotScenarioOwner    = fmt.Errorf("not authorized to manage this scenario")
)
