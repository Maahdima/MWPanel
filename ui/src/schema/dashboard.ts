import { z } from 'zod'
import { createApiResponseSchema } from '@/schema/api-response.ts'

const recentOnlinePeerSchema = z.object({
  name: z.string(),
  last_seen: z.string(),
})

export const deviceDataSchema = z.object({
  ServerInfo: z.object({
    total_servers: z.number(),
    active_servers: z.number(),
  }),
  InterfaceInfo: z.object({
    total_interfaces: z.number(),
    active_interfaces: z.number(),
  }),
  PeerInfo: z.object({
    recent_online_peers: z.nullable(z.array(recentOnlinePeerSchema)),
    total_peers: z.number(),
    online_peers: z.number(),
    offline_peers: z.number(),
    disabled_peers: z.number(),
  }),
  DeviceIdentity: z.object({
    identity: z.string(),
  }),
  DeviceInfo: z.object({
    board_name: z.string(),
    os_version: z.string(),
    cpu_arch: z.string(),
    uptime: z.string(),
    cpu_load: z.string(),
    total_memory: z.string(),
    free_memory: z.string(),
    total_disk: z.string(),
    free_disk: z.string(),
  }),
  DeviceIPv4Address: z.nullable(
    z.object({
      ipv4: z.string().optional(),
      isp: z.string().optional(),
    })
  ),
  DNSConfig: z.nullable(
    z.object({
      dns_servers: z.string(),
    })
  ),
})

export const trafficUsageSchema = z.object({
  interface_id: z.number(),
  date: z.string(),
  download: z.string(),
  upload: z.string(),
  total: z.string(),
})

export const dailyTrafficUsageSchema = z.array(trafficUsageSchema).nullable()

export const deviceDataResponseSchema =
  createApiResponseSchema(deviceDataSchema)

export const dailyTrafficUsageResponseSchema = createApiResponseSchema(
  dailyTrafficUsageSchema
)

export type DeviceData = z.infer<typeof deviceDataSchema>
export type DailyTrafficUsage = z.infer<typeof dailyTrafficUsageSchema>
