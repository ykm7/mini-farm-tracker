interface Data {
  Sensor?: string // no need to be provided
  Timestamp: string
  Data: any
}

interface LDDS45RawData {
  Bat: number
  Distance: string
  Interrupt_flag: number
  TempC_DS18B20: string
  Sensor_flag: number
}

export interface SensorData {
  LDDS45?: LDDS45RawData
}

export interface RawData extends Data {
  Valid?: boolean
  Data: SensorData // TODO: RawData should be able to take various
}

export interface CalibratedData extends Data {}
