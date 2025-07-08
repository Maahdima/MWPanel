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
      ipv4: z.string(),
      isp: z.string(),
    })
  ),
  DNSConfig: z.nullable(
    z.object({
      dns_servers: z.string(),
    })
  ),
})

export const deviceDataResponseSchema =
  createApiResponseSchema(deviceDataSchema)

export type DeviceData = z.infer<typeof deviceDataSchema>
export type DeviceInfoData = z.infer<typeof deviceDataSchema>['DeviceInfo']
