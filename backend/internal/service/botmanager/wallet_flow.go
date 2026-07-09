package botmanager

import (
	"context"
	"fmt"

	"revisitr/internal/entity"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// walletButtonRow returns an inline keyboard row with the "add to wallet" button
// for the given platform, if that platform is configured and enabled for the org.
func (h *handler) walletButtonRow(ctx context.Context, platform, label, callbackData string) [][]entity.InlineButton {
	if h.mgr.wallet == nil {
		return nil
	}
	cfg, err := h.mgr.wallet.GetConfig(ctx, h.info.OrgID, platform)
	if err != nil || cfg == nil || !cfg.IsEnabled {
		return nil
	}
	return [][]entity.InlineButton{
		{{Text: label, Data: callbackData}},
	}
}

func (h *handler) appleWalletButtonRow(ctx context.Context) [][]entity.InlineButton {
	return h.walletButtonRow(ctx, "apple", "🎫 Добавить карту в Apple Wallet", callbackWalletAddApple)
}

func (h *handler) googleWalletButtonRow(ctx context.Context) [][]entity.InlineButton {
	return h.walletButtonRow(ctx, "google", "📲 Добавить карту в Google Wallet", callbackWalletAddGoogle)
}

func (h *handler) sendTextWithInlineKeyboard(chatID int64, text string, markup *telego.InlineKeyboardMarkup) {
	clean := stripEmojiMarkers(text)
	if clean == "" {
		return
	}
	msg := tu.Message(tu.ID(chatID), clean)
	if markup != nil {
		msg = msg.WithReplyMarkup(markup)
	}
	if _, err := h.bot.SendMessage(context.Background(), msg); err != nil {
		h.logger.Error("send message with inline keyboard", "error", err, "chat_id", chatID)
	}
}

func (h *handler) handleWalletAddApple(ctx context.Context, query *telego.CallbackQuery) {
	chatID := query.Message.GetChat().ID

	if h.mgr.wallet == nil {
		h.sendText(chatID, "Функция временно недоступна.")
		return
	}

	client, err := h.mgr.clientsRepo.GetByTelegramID(ctx, h.info.ID, query.From.ID)
	if err != nil {
		h.sendText(chatID, "Сначала напишите /start")
		return
	}

	cfg, err := h.mgr.wallet.GetConfig(ctx, h.info.OrgID, "apple")
	if err != nil || cfg == nil || !cfg.IsEnabled {
		h.sendText(chatID, "Apple Wallet сейчас не подключён.")
		return
	}

	pass, err := h.getOrIssuePass(ctx, client.ID, "apple")
	if err != nil {
		h.logger.Error("wallet: get or issue pass", "error", err, "client_id", client.ID)
		h.sendText(chatID, "Не удалось выпустить карту. Попробуйте позже.")
		return
	}
	if pass.Status != "active" {
		h.sendText(chatID, "Ваша карта лояльности деактивирована. Обратитесь в заведение.")
		return
	}

	balance, level := h.currentLoyaltySnapshot(ctx, client.ID)
	pass.LastBalance = balance
	pass.LastLevel = level
	if err := h.mgr.wallet.RefreshPassBalance(ctx, client.ID, balance, level); err != nil {
		h.logger.Warn("wallet: refresh pass balance", "error", err, "client_id", client.ID)
	}

	clientQR, err := h.mgr.wallet.GetClientsQRCode(ctx, client.ID)
	if err != nil {
		h.logger.Warn("wallet: get client qr code", "error", err, "client_id", client.ID)
	}
	orgName, err := h.mgr.wallet.GetOrgName(ctx, h.info.OrgID)
	if err != nil {
		h.logger.Warn("wallet: get org name", "error", err, "org_id", h.info.OrgID)
	}

	webServiceURL := cfg.Design.WebServiceURL
	if webServiceURL == "" && h.mgr.baseURL != "" {
		webServiceURL = h.mgr.baseURL + "/api/v1/wallet"
	}

	pkpassData, err := h.mgr.passGen.GeneratePass(pass, cfg, clientQR, orgName, webServiceURL)
	if err != nil {
		h.logger.Error("wallet: generate pass", "error", err, "client_id", client.ID)
		h.sendText(chatID, "Не удалось сформировать карту. Попробуйте позже.")
		return
	}

	h.sendWalletPassDocument(chatID, pkpassData)
}

func (h *handler) handleWalletAddGoogle(ctx context.Context, query *telego.CallbackQuery) {
	chatID := query.Message.GetChat().ID

	if h.mgr.wallet == nil {
		h.sendText(chatID, "Функция временно недоступна.")
		return
	}

	client, err := h.mgr.clientsRepo.GetByTelegramID(ctx, h.info.ID, query.From.ID)
	if err != nil {
		h.sendText(chatID, "Сначала напишите /start")
		return
	}

	cfg, err := h.mgr.wallet.GetConfig(ctx, h.info.OrgID, "google")
	if err != nil || cfg == nil || !cfg.IsEnabled {
		h.sendText(chatID, "Google Wallet сейчас не подключён.")
		return
	}

	pass, err := h.getOrIssuePass(ctx, client.ID, "google")
	if err != nil {
		h.logger.Error("wallet: get or issue pass", "error", err, "client_id", client.ID)
		h.sendText(chatID, "Не удалось выпустить карту. Попробуйте позже.")
		return
	}
	if pass.Status != "active" {
		h.sendText(chatID, "Ваша карта лояльности деактивирована. Обратитесь в заведение.")
		return
	}

	balance, level := h.currentLoyaltySnapshot(ctx, client.ID)
	pass.LastBalance = balance
	pass.LastLevel = level
	if err := h.mgr.wallet.RefreshPassBalance(ctx, client.ID, balance, level); err != nil {
		h.logger.Warn("wallet: refresh pass balance", "error", err, "client_id", client.ID)
	}

	saveURL, err := h.mgr.wallet.GenerateGoogleSaveURL(ctx, h.info.OrgID, pass)
	if err != nil {
		h.logger.Error("wallet: generate google save url", "error", err, "client_id", client.ID)
		h.sendText(chatID, "Не удалось сформировать карту. Попробуйте позже.")
		return
	}

	markup := h.inlineKeyboard([][]entity.InlineButton{
		{{Text: "📲 Открыть в Google Wallet", URL: saveURL}},
	})
	h.sendTextWithInlineKeyboard(chatID, "🎫 Ваша карта лояльности готова.", markup)
}

// getOrIssuePass returns the client's existing wallet pass for the platform, issuing a new one if absent.
func (h *handler) getOrIssuePass(ctx context.Context, clientID int, platform string) (*entity.WalletPass, error) {
	passes, err := h.mgr.wallet.GetClientPasses(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("get client passes: %w", err)
	}
	for _, p := range passes {
		if p.Platform == platform {
			pass := p
			return &pass, nil
		}
	}

	pass, err := h.mgr.wallet.IssuePass(ctx, h.info.OrgID, entity.IssueWalletPassRequest{
		ClientID: clientID,
		Platform: platform,
	})
	if err != nil {
		return nil, fmt.Errorf("issue pass: %w", err)
	}
	return pass, nil
}

// currentLoyaltySnapshot returns the balance/level of the client's first active loyalty program.
func (h *handler) currentLoyaltySnapshot(ctx context.Context, clientID int) (balance int, level string) {
	programs, err := h.mgr.loyaltyRepo.GetProgramsByOrgID(ctx, h.info.OrgID)
	if err != nil {
		return 0, ""
	}
	for _, prog := range programs {
		if !prog.IsActive {
			continue
		}
		cl, err := h.mgr.loyaltyRepo.GetClientLoyalty(ctx, clientID, prog.ID)
		if err != nil || cl == nil {
			continue
		}
		levelName := ""
		if cl.LevelID != nil {
			levels, _ := h.mgr.loyaltyRepo.GetLevelsByProgramID(ctx, prog.ID)
			for _, l := range levels {
				if l.ID == *cl.LevelID {
					levelName = l.Name
					break
				}
			}
		}
		return int(cl.Balance), levelName
	}
	return 0, ""
}

func (h *handler) sendWalletPassDocument(chatID int64, pkpassData []byte) {
	file := tu.FileFromBytes(pkpassData, "card.pkpass")
	doc := tu.Document(tu.ID(chatID), file).
		WithCaption("🎫 Ваша карта лояльности. Откройте файл, чтобы добавить в Apple Wallet.")
	if _, err := h.bot.SendDocument(context.Background(), doc); err != nil {
		h.logger.Error("send wallet pass document", "error", err, "chat_id", chatID)
		h.sendText(chatID, "Не удалось отправить карту. Попробуйте позже.")
	}
}
