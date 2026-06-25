import { useApiQuery, useApiMutation } from '../../lib/swr'
import { botsApi, modulePresetsApi } from './api'

export function useBotsQuery() {
  return useApiQuery('bots', botsApi.list)
}

export function useBotQuery(id: number) {
  return useApiQuery(id ? `bots-${id}` : null, () => botsApi.getById(id))
}

// Module presets

export function useModulePresetsQuery(moduleKey: string) {
  return useApiQuery(
    `module-presets-${moduleKey}`,
    () => modulePresetsApi.listPresets(moduleKey),
    { revalidateOnFocus: false },
  )
}

export function useBotModuleSettingsQuery(botId: number, moduleKey: string) {
  return useApiQuery(
    botId ? `bot-module-settings-${botId}-${moduleKey}` : null,
    () => modulePresetsApi.getBotModuleSettings(botId, moduleKey),
  )
}

export function useSelectPresetMutation(botId: number, moduleKey: string) {
  return useApiMutation(
    `select-preset-${botId}-${moduleKey}`,
    (presetKey: string) => modulePresetsApi.selectPreset(botId, moduleKey, presetKey),
    [`bot-module-settings-${botId}-${moduleKey}`],
  )
}

export function useUpdateCustomizationsMutation(botId: number, moduleKey: string) {
  return useApiMutation(
    `update-customizations-${botId}-${moduleKey}`,
    (customizations: Record<string, unknown>) =>
      modulePresetsApi.updateCustomizations(botId, moduleKey, customizations),
    [`bot-module-settings-${botId}-${moduleKey}`],
  )
}

export function useResetPresetMutation(botId: number, moduleKey: string) {
  return useApiMutation(
    `reset-preset-${botId}-${moduleKey}`,
    () => modulePresetsApi.resetPreset(botId, moduleKey),
    [`bot-module-settings-${botId}-${moduleKey}`],
  )
}
