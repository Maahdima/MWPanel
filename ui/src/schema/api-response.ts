import { z } from 'zod'

export const apiStatusEnum = z.enum(['success', 'error'])

export const createApiResponseSchema = <T extends z.ZodTypeAny>(
  dataSchema: T
) =>
  z.object({
    statusCode: z.number(),
    status: apiStatusEnum,
    data: dataSchema,
  })
