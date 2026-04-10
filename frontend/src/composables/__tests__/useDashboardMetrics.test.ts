// ABOUTME: Unit tests for dashboard metrics computation.
// ABOUTME: Covers empty state, price aggregation, module grouping, and timeline bucketing.

import {describe, it, expect} from 'vitest'
import {computeDashboardMetrics} from '../useDashboardMetrics'
import type {Item} from '../../api/types'

function makeItem(overrides: Partial<Item> = {}): Item {
  return {
    id: 'id-1',
    moduleId: 'coins',
    title: 'Test Item',
    purchasePrice: null,
    images: [],
    tags: [],
    attributes: {},
    createdAt: '2026-01-15T10:00:00Z',
    updatedAt: '2026-01-15T10:00:00Z',
    ...overrides,
  }
}

const nameMap: Record<string, string> = {
  coins: 'Coins',
  vinyl: 'Vinyl',
  stamps: 'Stamps',
  cards: 'Cards',
  books: 'Books',
  toys: 'Toys',
  watches: 'Watches',
}

function getModuleName(id: string): string {
  return nameMap[id] ?? id
}

describe('computeDashboardMetrics', () => {
  it('returns zeros for empty items', () => {
    const m = computeDashboardMetrics([], getModuleName)
    expect(m.totalValue).toBe(0)
    expect(m.totalItems).toBe(0)
    expect(m.mostValuableItem).toBeNull()
    expect(m.moduleBreakdown).toEqual([])
    expect(m.acquisitionTimeline).toEqual([])
  })

  it('sums only non-null prices for totalValue', () => {
    const items = [
      makeItem({id: '1', purchasePrice: 100}),
      makeItem({id: '2', purchasePrice: null}),
      makeItem({id: '3', purchasePrice: 50}),
    ]
    const m = computeDashboardMetrics(items, getModuleName)
    expect(m.totalValue).toBe(150)
    expect(m.totalItems).toBe(3)
  })

  it('finds the most valuable item', () => {
    const items = [
      makeItem({id: '1', title: 'Cheap', purchasePrice: 10}),
      makeItem({id: '2', title: 'Expensive', purchasePrice: 500}),
      makeItem({id: '3', title: 'Mid', purchasePrice: 100}),
    ]
    const m = computeDashboardMetrics(items, getModuleName)
    expect(m.mostValuableItem).toEqual({id: '2', title: 'Expensive', price: 500})
  })

  it('returns null mostValuableItem when no items have prices', () => {
    const items = [
      makeItem({id: '1', purchasePrice: null}),
      makeItem({id: '2', purchasePrice: null}),
    ]
    const m = computeDashboardMetrics(items, getModuleName)
    expect(m.mostValuableItem).toBeNull()
  })

  it('groups items by module with correct value percentages', () => {
    const items = [
      makeItem({id: '1', moduleId: 'coins', purchasePrice: 300}),
      makeItem({id: '2', moduleId: 'vinyl', purchasePrice: 100}),
      makeItem({id: '3', moduleId: 'coins', purchasePrice: 100}),
    ]
    const m = computeDashboardMetrics(items, getModuleName)
    expect(m.moduleBreakdown).toHaveLength(2)
    // Sorted by value descending
    expect(m.moduleBreakdown[0].moduleName).toBe('Coins')
    expect(m.moduleBreakdown[0].totalValue).toBe(400)
    expect(m.moduleBreakdown[0].itemCount).toBe(2)
    expect(m.moduleBreakdown[0].percentage).toBe(80)
    expect(m.moduleBreakdown[1].moduleName).toBe('Vinyl')
    expect(m.moduleBreakdown[1].percentage).toBe(20)
  })

  it('groups smallest modules into Other when more than 6', () => {
    const items = [
      makeItem({id: '1', moduleId: 'coins', purchasePrice: 700}),
      makeItem({id: '2', moduleId: 'vinyl', purchasePrice: 600}),
      makeItem({id: '3', moduleId: 'stamps', purchasePrice: 500}),
      makeItem({id: '4', moduleId: 'cards', purchasePrice: 400}),
      makeItem({id: '5', moduleId: 'books', purchasePrice: 300}),
      makeItem({id: '6', moduleId: 'toys', purchasePrice: 200}),
      makeItem({id: '7', moduleId: 'watches', purchasePrice: 100}),
    ]
    const m = computeDashboardMetrics(items, getModuleName)
    // 5 top + 1 "Other" = 6 segments
    expect(m.moduleBreakdown).toHaveLength(6)
    const other = m.moduleBreakdown.find(s => s.moduleName === 'Other')
    expect(other).toBeDefined()
    // Other = toys(200) + watches(100) = 300
    expect(other!.totalValue).toBe(300)
    expect(other!.itemCount).toBe(2)
  })

  it('does not group into Other when 6 or fewer modules', () => {
    const items = [
      makeItem({id: '1', moduleId: 'coins', purchasePrice: 100}),
      makeItem({id: '2', moduleId: 'vinyl', purchasePrice: 100}),
      makeItem({id: '3', moduleId: 'stamps', purchasePrice: 100}),
      makeItem({id: '4', moduleId: 'cards', purchasePrice: 100}),
      makeItem({id: '5', moduleId: 'books', purchasePrice: 100}),
      makeItem({id: '6', moduleId: 'toys', purchasePrice: 100}),
    ]
    const m = computeDashboardMetrics(items, getModuleName)
    expect(m.moduleBreakdown).toHaveLength(6)
    expect(m.moduleBreakdown.find(s => s.moduleName === 'Other')).toBeUndefined()
  })

  it('groups items by month in chronological order', () => {
    const items = [
      makeItem({id: '1', createdAt: '2026-03-10T00:00:00Z'}),
      makeItem({id: '2', createdAt: '2026-01-05T00:00:00Z'}),
      makeItem({id: '3', createdAt: '2026-01-20T00:00:00Z'}),
      makeItem({id: '4', createdAt: '2026-03-15T00:00:00Z'}),
    ]
    const m = computeDashboardMetrics(items, getModuleName)
    expect(m.acquisitionTimeline).toHaveLength(2)
    expect(m.acquisitionTimeline[0]).toEqual({key: '2026-01', label: 'Jan 2026', count: 2})
    expect(m.acquisitionTimeline[1]).toEqual({key: '2026-03', label: 'Mar 2026', count: 2})
  })

  it('handles single-month timeline', () => {
    const items = [
      makeItem({id: '1', createdAt: '2026-06-01T00:00:00Z'}),
      makeItem({id: '2', createdAt: '2026-06-15T00:00:00Z'}),
    ]
    const m = computeDashboardMetrics(items, getModuleName)
    expect(m.acquisitionTimeline).toHaveLength(1)
    expect(m.acquisitionTimeline[0].label).toBe('Jun 2026')
    expect(m.acquisitionTimeline[0].count).toBe(2)
  })
})
