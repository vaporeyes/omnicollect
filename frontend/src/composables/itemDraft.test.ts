// ABOUTME: Draft isolation, unknown-field preservation, and edit-version regressions.
// ABOUTME: Ensures schema changes cannot silently erase historical item metadata.
import {describe, expect, it} from 'vitest'
import {createItemDraft} from './itemDraft'
import type {Item, ModuleSchema} from '../api/types'

const schema: ModuleSchema = {id: 'books', displayName: 'Books', attributes: [{name: 'count', type: 'number'}, {name: 'read', type: 'boolean'}]}
const item: Item = {id: 'one', moduleId: 'books', title: 'Book', purchasePrice: null, images: ['image.png'], tags: ['keep'], attributes: {legacy: {nested: 'keep'}, count: 3}, createdAt: '2020-01-01T00:00:00Z', updatedAt: '2025-01-01T00:00:00.123456Z'}

describe('createItemDraft', () => {
  it('preserves unknown attributes and the exact edit version', () => {
    const draft = createItemDraft(schema, item)
    expect(draft.attributes).toEqual({count: 3, read: false, legacy: {nested: 'keep'}})
    expect(draft.updatedAt).toBe(item.updatedAt)
  })
  it('never mutates the original item through draft arrays or nested values', () => {
    const draft = createItemDraft(schema, item)
    draft.images.push('new.png')
    draft.tags.length = 0
    draft.attributes.legacy.nested = 'changed'
    expect(item.images).toEqual(['image.png'])
    expect(item.tags).toEqual(['keep'])
    expect(item.attributes.legacy.nested).toBe('keep')
  })
  it('starts clean when switching to a new item or schema', () => {
    createItemDraft(schema, item)
    const draft = createItemDraft({...schema, id: 'other', attributes: []})
    expect(draft.id).toBe('')
    expect(draft.updatedAt).toBe('')
    expect(Object.keys(draft.attributes)).toEqual([])
    expect(draft.images).toEqual([])
  })
})
