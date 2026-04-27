import { useApiQuery, useApiMutation } from '../../lib/swr'
import { botsApi, modulePresetsApi } from './api'
import type { CreateBotRequest, Bot } from './types'

export function useBotsQuery() {
  return useApiQuery('bots', botsApi.list)
}

export function useBotQuery(id: number) {
  return useApiQuery(id ? `bots-${id}` : null, () => botsApi.getById(id))
}

export function useCreateBotMutation() {
  return useApiMutation(
    'bots/create',
    (data: CreateBotRequest) => botsApi.create(data),
    ['bots'],
  )
}

export function useUpdateBotMutation() {
  return useApiMutation(
    'bots/update',
    ({ id, data }: { id: number; data: Partial<Bot> }) =>
      botsApi.update(id, data),
    ['bots'],
  )
}

export function useDeleteBotMutation() {
  return useApiMutation(
    'bots/delete',
    (id: number) => botsApi.remove(id),
    ['bots'],
  )
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

export function useResetPresetMutation(botId: number, moduleKey: string) {
  return useApiMutation(
    `reset-preset-${botId}-${moduleKey}`,
    () => modulePresetsApi.resetPreset(botId, moduleKey),
    [`bot-module-settings-${botId}-${moduleKey}`],
  )
}
