// ABOUTME: Builds independent drafts while preserving fields from older schema versions.
// ABOUTME: New drafts never inherit prior item keys, arrays, or optimistic edit versions.
import type {Item, ModuleSchema} from '../api/types'

export function createItemDraft(schema: ModuleSchema, item?: Item | null): Item {
  const attributes: Record<string, any> = Object.create(null)
  for (const attr of schema.attributes) {
    attributes[attr.name] = attr.type === 'boolean' ? false : attr.type === 'number' ? null : ''
  }
  Object.assign(attributes, JSON.parse(JSON.stringify(item?.attributes ?? {})))
  return {
    id: item?.id ?? '', moduleId: schema.id, title: item?.title ?? '',
    purchasePrice: item?.purchasePrice ?? null,
    images: [...(item?.images ?? [])], tags: [...(item?.tags ?? [])], attributes,
    createdAt: item?.createdAt ?? '', updatedAt: item?.updatedAt ?? '',
  }
}
