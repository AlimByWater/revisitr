import { api } from '@/lib/api'
import type {
  LunchAvailabilitySlot,
  LunchCourse,
  LunchFormat,
  LunchOrder,
  LunchProgram,
  SaveLunchCourseRequest,
  SaveLunchFormatRequest,
  UpsertLunchProgramRequest,
} from './types'

export const lunchApi = {
  getProgram: async (botId: number): Promise<LunchProgram> => {
    const response = await api.get<LunchProgram>(`/lunch/bots/${botId}`)
    return response.data
  },

  updateProgram: async (botId: number, data: UpsertLunchProgramRequest): Promise<LunchProgram> => {
    const response = await api.put<LunchProgram>(`/lunch/bots/${botId}`, data)
    return response.data
  },

  createCourse: async (botId: number, data: SaveLunchCourseRequest): Promise<LunchCourse> => {
    const response = await api.post<LunchCourse>(`/lunch/bots/${botId}/courses`, data)
    return response.data
  },

  updateCourse: async (courseId: number, data: SaveLunchCourseRequest): Promise<void> => {
    await api.put(`/lunch/courses/${courseId}`, data)
  },

  deleteCourse: async (courseId: number): Promise<void> => {
    await api.delete(`/lunch/courses/${courseId}`)
  },

  createFormat: async (botId: number, data: SaveLunchFormatRequest): Promise<LunchFormat> => {
    const response = await api.post<LunchFormat>(`/lunch/bots/${botId}/formats`, data)
    return response.data
  },

  updateFormat: async (formatId: number, data: SaveLunchFormatRequest): Promise<void> => {
    await api.put(`/lunch/formats/${formatId}`, data)
  },

  deleteFormat: async (formatId: number): Promise<void> => {
    await api.delete(`/lunch/formats/${formatId}`)
  },

  setAvailability: async (botId: number, slots: LunchAvailabilitySlot[]): Promise<void> => {
    await api.put(`/lunch/bots/${botId}/availability`, { slots })
  },

  listOrders: async (botId: number, status?: string): Promise<LunchOrder[]> => {
    const response = await api.get<LunchOrder[]>(`/lunch/bots/${botId}/orders`, {
      params: status ? { status } : undefined,
    })
    return response.data
  },

  updateOrderStatus: async (orderId: number, status: string): Promise<void> => {
    await api.patch(`/lunch/orders/${orderId}`, { status })
  },
}
