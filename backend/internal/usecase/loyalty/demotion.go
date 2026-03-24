package loyalty

import (
	"context"
	"fmt"
)

// DemoteClients checks all clients with levels and demotes if total_earned
// no longer meets the threshold. Called by scheduler daily.
func (uc *Usecase) DemoteClients(ctx context.Context) error {
	clients, err := uc.repo.GetClientsWithLevels(ctx)
	if err != nil {
		return fmt.Errorf("DemoteClients: %w", err)
	}

	for _, cl := range clients {
		newLevelID := uc.determineLevelID(ctx, cl.ProgramID, cl.TotalEarned)
		if !equalIntPtr(cl.LevelID, newLevelID) {
			cl.LevelID = newLevelID
			if err := uc.repo.UpsertClientLoyalty(ctx, &cl); err != nil {
				uc.logger.Error("demotion failed",
					"client_id", cl.ClientID,
					"program_id", cl.ProgramID,
					"error", err,
				)
				continue
			}
			uc.logger.Info("level changed by demotion check",
				"client_id", cl.ClientID,
				"program_id", cl.ProgramID,
			)
		}
	}
	return nil
}

func equalIntPtr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
