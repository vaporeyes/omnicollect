// ABOUTME: Dashboard metrics distinguish missing, invalid, zero and overflowing prices.
// ABOUTME: Count-only grouping and valid calendar-month labels remain deterministic.
import {expect,it} from 'vitest'
import {computeDashboardMetrics} from './useDashboardMetrics'
import type {Item} from '../api/types'
const item=(price?:number|null,moduleId='a',createdAt='2026-01-01T00:00:00Z')=>({id:moduleId,title:moduleId,moduleId,purchasePrice:price,createdAt}) as Item
const compute=(items:Item[])=>computeDashboardMetrics(items,id=>id)
it('does not present missing or invalid prices as recorded zero prices',()=>{
 const metrics=compute([item(null),item(undefined),item(NaN),item(Infinity),item(-2),item(0),item(15)])
 expect(metrics.pricedItems).toBe(2);expect(metrics.invalidPrices).toBe(3)
 expect(metrics.totalValue).toBe(15);expect(metrics.mostValuableItem?.price).toBe(15)
 expect(compute([item(null)]).mostValuableItem).toBeNull()
 expect(compute([item(0)]).mostValuableItem?.price).toBe(0)
})
it('flags totals that cannot be represented instead of drawing infinite values',()=>{
 const metrics=compute([item(Number.MAX_VALUE),item(Number.MAX_VALUE)])
 expect(metrics.valueAvailable).toBe(false)
 expect(metrics.moduleBreakdown.every(segment=>segment.percentage===0)).toBe(true)
})
it('sorts the count fallback and excludes malformed month labels',()=>{
 const metrics=compute([item(null,'a','bad'),item(null,'b','2026-13-01'),item(null,'b')])
 expect(metrics.moduleBreakdown[0].moduleId).toBe('b')
 expect(metrics.acquisitionTimeline).toEqual([{key:'2026-01',label:'Jan 2026',count:1}])
})
