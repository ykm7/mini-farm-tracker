import type { ObjectId } from "@/types/ObjectId"

interface AssetMetricsCylinderVolume {
  Volume: number
  Radius: number
  Height: number
}

interface AssetMetrics {
  Volume?: AssetMetricsCylinderVolume
}

export interface Asset {
  Id: ObjectId
  Name: string
  Description: string
  Sensors?: string[]
  Metrics?: AssetMetrics
}
