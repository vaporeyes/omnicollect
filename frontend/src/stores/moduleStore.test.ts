// ABOUTME: Unit tests for the module Pinia store.
// ABOUTME: Verifies fetchModules success and error handling with mocked fetch.
import {describe, it, expect, vi, beforeEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {useModuleStore} from './moduleStore'

beforeEach(() => {
  setActivePinia(createPinia())
  vi.restoreAllMocks()
})

describe('fetchModules', () => {
  it('sets modules on success', async () => {
    const modules = [
      {id: 'comics', displayName: 'Comics', attributes: []},
      {id: 'coins', displayName: 'Coins', attributes: [{name: 'year', type: 'number'}]},
    ]
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(modules),
    } as unknown as Response)

    const store = useModuleStore()
    await store.fetchModules()

    expect(store.modules).toEqual(modules)
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('sets error on failure', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.resolve({error: 'server error'}),
    } as unknown as Response)

    const store = useModuleStore()
    await store.fetchModules()

    expect(store.modules).toEqual([])
    expect(store.error).toBe('server error')
    expect(store.loading).toBe(false)
  })

  it('getModuleById returns correct module', async () => {
    const modules = [
      {id: 'comics', displayName: 'Comics', attributes: []},
      {id: 'coins', displayName: 'Coins', attributes: []},
    ]
    global.fetch = vi.fn().mockResolvedValue({
      ok: true, status: 200, json: () => Promise.resolve(modules),
    } as unknown as Response)

    const store = useModuleStore()
    await store.fetchModules()

    expect(store.getModuleById('comics')?.displayName).toBe('Comics')
    expect(store.getModuleById('nonexistent')).toBeNull()
  })
})
