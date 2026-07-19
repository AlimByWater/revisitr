import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/features/menus/queries', () => ({
  useMenusQuery: vi.fn(),
  useCopyMenuMutation: vi.fn(),
  useCreateMenuMutation: vi.fn(),
  useDeleteMenuMutation: vi.fn(),
  useUpdateMenuMutation: vi.fn(),
}))

vi.mock('@/features/pos/queries', () => ({
  usePOSQuery: vi.fn(),
}))

import {
  useCreateMenuMutation,
  useCopyMenuMutation,
  useDeleteMenuMutation,
  useMenusQuery,
  useUpdateMenuMutation,
} from '@/features/menus/queries'
import { usePOSQuery } from '@/features/pos/queries'
import MenusPage from './index'

const mockUseMenusQuery = vi.mocked(useMenusQuery)
const mockUseCreateMenuMutation = vi.mocked(useCreateMenuMutation)
const mockUseCopyMenuMutation = vi.mocked(useCopyMenuMutation)
const mockUseDeleteMenuMutation = vi.mocked(useDeleteMenuMutation)
const mockUseUpdateMenuMutation = vi.mocked(useUpdateMenuMutation)
const mockUsePOSQuery = vi.mocked(usePOSQuery)

function renderPage(initialEntry = '/dashboard/menus?botId=12') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/dashboard/menus" element={<MenusPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

function mutationStub() {
  return {
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
    isPending: false,
    isError: false,
    isSuccess: false,
    isMutating: false,
    reset: vi.fn(),
    trigger: vi.fn(),
    data: undefined,
    error: undefined,
  }
}

describe('MenusPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    mockUseCreateMenuMutation.mockReturnValue({
      ...mutationStub(),
    } as unknown as ReturnType<typeof useCreateMenuMutation>)

    mockUseCopyMenuMutation.mockReturnValue({
      ...mutationStub(),
    } as unknown as ReturnType<typeof useCopyMenuMutation>)

    mockUseDeleteMenuMutation.mockReturnValue({
      ...mutationStub(),
    } as unknown as ReturnType<typeof useDeleteMenuMutation>)

    mockUseUpdateMenuMutation.mockReturnValue({
      ...mutationStub(),
    } as unknown as ReturnType<typeof useUpdateMenuMutation>)

    mockUsePOSQuery.mockReturnValue({
      data: [
        {
          id: 10,
          org_id: 1,
          name: 'Маросейка',
          address: 'Маросейка, 12',
          phone: '+7 000 000-00-00',
          schedule: {},
          is_active: true,
          created_at: '2026-04-18T00:00:00Z',
          updated_at: '2026-04-18T00:00:00Z',
        },
      ],
      isLoading: false,
      isError: false,
      error: undefined,
      mutate: vi.fn(),
      isValidating: false,
    } as unknown as ReturnType<typeof usePOSQuery>)
  })

  it('shows conflict warning and back link to bot modules', () => {
    mockUseMenusQuery.mockReturnValue({
      data: [
        {
          id: 1,
          org_id: 1,
          name: 'Основное меню',
          source: 'manual',
          created_at: '2026-04-18T00:00:00Z',
          updated_at: '2026-04-18T00:00:00Z',
          categories: [],
          bindings: [
            {
              menu_id: 1,
              pos_id: 10,
              pos_name: 'Маросейка',
              is_active: true,
              created_at: '2026-04-18T00:00:00Z',
            },
          ],
        },
        {
          id: 2,
          org_id: 1,
          name: 'Барное меню',
          source: 'manual',
          created_at: '2026-04-18T00:00:00Z',
          updated_at: '2026-04-18T00:00:00Z',
          categories: [],
          bindings: [
            {
              menu_id: 2,
              pos_id: 10,
              pos_name: 'Маросейка',
              is_active: true,
              created_at: '2026-04-18T01:00:00Z',
            },
          ],
        },
      ],
      isLoading: false,
      isError: false,
      mutate: vi.fn(),
      error: undefined,
      isValidating: false,
    } as unknown as ReturnType<typeof useMenusQuery>)

    renderPage()

    expect(screen.getByRole('link', { name: /Назад к модулям бота/i })).toHaveAttribute(
      'href',
      '/dashboard/bots/12?tab=modules',
    )
    expect(
      screen.getByText(/для точки продаж Маросейка выбрано более одного меню/i),
    ).toBeInTheDocument()
    expect(screen.getByText(/Сейчас показывается: Основное меню/i)).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: /Создать копию/i })).toHaveLength(2)
  })

  it('copies menu composition without POS bindings notice', async () => {
    const copyMutate = vi.fn().mockResolvedValue(undefined)
    const refetchMenus = vi.fn()
    mockUseCopyMenuMutation.mockReturnValue({
      ...mutationStub(),
      mutate: copyMutate,
    } as unknown as ReturnType<typeof useCopyMenuMutation>)
    mockUseMenusQuery.mockReturnValue({
      data: [
        {
          id: 1,
          org_id: 1,
          name: 'Основное меню',
          source: 'manual',
          created_at: '2026-04-18T00:00:00Z',
          updated_at: '2026-04-18T00:00:00Z',
          categories: [],
          bindings: [
            {
              menu_id: 1,
              pos_id: 10,
              pos_name: 'Маросейка',
              is_active: true,
              created_at: '2026-04-18T00:00:00Z',
            },
          ],
        },
      ],
      isLoading: false,
      isError: false,
      mutate: refetchMenus,
      error: undefined,
      isValidating: false,
    } as unknown as ReturnType<typeof useMenusQuery>)

    renderPage()
    fireEvent.click(screen.getByRole('button', { name: /Создать копию/i }))

    await waitFor(() => {
      expect(copyMutate).toHaveBeenCalledWith({
        id: 1,
        data: { name: 'Основное меню (копия)' },
      })
      expect(refetchMenus).toHaveBeenCalled()
    })
    expect(screen.getByRole('status')).toHaveTextContent('Копия создана. Точки продаж не привязаны.')
  })

  it('requires confirmation before deleting a menu', async () => {
    const deleteMutate = vi.fn().mockResolvedValue(undefined)
    const refetchMenus = vi.fn()
    mockUseDeleteMenuMutation.mockReturnValue({
      ...mutationStub(),
      mutate: deleteMutate,
    } as unknown as ReturnType<typeof useDeleteMenuMutation>)
    mockUseMenusQuery.mockReturnValue({
      data: [
        {
          id: 1,
          org_id: 1,
          name: 'Основное меню',
          source: 'manual',
          created_at: '2026-04-18T00:00:00Z',
          updated_at: '2026-04-18T00:00:00Z',
          categories: [],
          bindings: [],
        },
      ],
      isLoading: false,
      isError: false,
      mutate: refetchMenus,
      error: undefined,
      isValidating: false,
    } as unknown as ReturnType<typeof useMenusQuery>)

    renderPage()

    fireEvent.click(screen.getByRole('button', { name: /Удалить меню/i }))
    expect(deleteMutate).not.toHaveBeenCalled()

    fireEvent.click(screen.getByRole('button', { name: /^Удалить$/i }))
    await waitFor(() => {
      expect(deleteMutate).toHaveBeenCalledWith(1)
    })
  })
})
