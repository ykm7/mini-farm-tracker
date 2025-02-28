import type { AGGREGATION_TYPE, CalibratedDataNames } from "@/models/Data"

export type Unit =
  | "mm" // millimeters
  | "cm" // centimeters
  | "m" // meters
  | "m³" // cubic meters
  | "L" // liters
  | "mm/hr" // millimeters per hour
  | "m/s" // meters per second
  | "℃" // degrees Celsius
  | "Pa" // Pascal (pressure)
  | "%RH" // relative humidity percentage
  | "Lux" // luminous flux per unit area
  | "" // for UV index (no unit)

export interface DisplayPoint {
  value: number
  timestamp: string
}

export interface GraphDataType {
  data: DisplayPoint[]
  unit: Unit
}

// import type { ExtendedDataPoint } from "vue-chartjs";
export type ExtendedDataPoint = {
  [key: string]: string | number | null | ExtendedDataPoint;
};

export const dynamicTimeUnit = (dataPoints: DisplayPoint[]) => {
  const oldest = dataPoints[0]
  const newest = dataPoints[dataPoints.length - 1]

  const diff = Date.parse(newest.timestamp) - Date.parse(oldest.timestamp)

  const hours = diff / (1000 * 60 * 60)

  if (hours < 1) return "minute"
  if (hours < 24) return "hour"
  if (hours < 30 * 24) return "day"
  if (hours < 365 * 24) return "month"
  return "year"
}

export type KeyOf<T> = keyof T

export class GraphData {
  Raw?: GraphDataType // Maybe??
  Volume?: GraphDataType
  AirTemperature?: GraphDataType
  AirHumidity?: GraphDataType
  LightIntensity?: GraphDataType
  UvIndex?: GraphDataType
  WindSpeed?: GraphDataType
  WindDirection?: GraphDataType
  RainGauge?: GraphDataType
  BarometricPressure?: GraphDataType
  PeakWindGust?: GraphDataType
  RainAccumulation?: GraphDataType
}

export const createGraphData = (): GraphData => {
  return {}
}

export type CalibratedDataNamesGrouping = {
  [K in keyof typeof CalibratedDataNames]: AggregatedDataMapping
}

export type AggregatedDataGrouping = {
  [K in keyof typeof AGGREGATION_TYPE]: AggregatedDataPoint[]
}

export interface AggregatedDataPoint {
  unit: string
  value: number
  date: string
}

/**
 * TODO: Give actually better name
 */
export interface AggregatedDataMapping {
  // type: CalibratedDataNames
  data: AggregatedDataGrouping
}