// ABOUTME: TypeScript interfaces mirroring Go structs for the REST API.
// ABOUTME: Replaces Wails auto-generated bindings with explicit type definitions.

export interface Item {
  id: string
  moduleId: string
  title: string
  purchasePrice: number | null
  images: string[]
  attributes: Record<string, any>
  createdAt: string
  updatedAt: string
}

export interface ModuleSchema {
  id: string
  displayName: string
  description?: string
  attributes: AttributeSchema[]
}

export interface AttributeSchema {
  name: string
  type: string
  required?: boolean
  options?: string[]
  display?: DisplayHints
}

export interface DisplayHints {
  label?: string
  placeholder?: string
  widget?: string
  group?: string
  order?: number
}

export interface ProcessImageResult {
  filename: string
  originalPath: string
  thumbnailPath: string
  width: number
  height: number
  format: string
}

export interface BulkDeleteResult {
  deleted: number
}

export interface BulkUpdateResult {
  updated: number
}

export interface ApiError {
  error: string
}
