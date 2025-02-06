export type Unit = 'mm' | 'cm' | 'm' | 'm³' | 'L';

export interface DisplayPoint {
  value: number
  timestamp: string
  unit: Unit
}
