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
  RainfallHourly?: GraphDataType
  BarometricPressure?: GraphDataType
}

export const createGraphData = (): GraphData => {
  return {}
}
