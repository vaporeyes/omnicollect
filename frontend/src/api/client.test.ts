// ABOUTME: Unit tests for the centralized fetch-based HTTP client.
// ABOUTME: Mocks global fetch to verify request construction and response handling.
import {describe, it, expect, vi, beforeEach} from 'vitest'
import {get, post, del, postFile, downloadFile} from './client'

function mockFetch(body: any, status = 200, headers: Record<string, string> = {}) {
  const respHeaders = new Headers(headers)
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
    blob: () => Promise.resolve(new Blob(['data'])),
    headers: respHeaders,
  } as unknown as Response)
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('get', () => {
  it('fetches and returns parsed JSON', async () => {
    global.fetch = mockFetch([{id: '1', title: 'Test'}])
    const result = await get<any[]>('/api/v1/items')
    expect(result).toEqual([{id: '1', title: 'Test'}])
    expect(fetch).toHaveBeenCalledWith('/api/v1/items')
  })

  it('throws on non-ok response', async () => {
    global.fetch = mockFetch({error: 'not found'}, 404)
    await expect(get('/api/v1/items/bad')).rejects.toThrow('not found')
  })
})

describe('post', () => {
  it('sends JSON body and parses response', async () => {
    global.fetch = mockFetch({id: '1', title: 'Saved'})
    const result = await post<any>('/api/v1/items', {title: 'Saved'})
    expect(result).toEqual({id: '1', title: 'Saved'})
    const call = (fetch as any).mock.calls[0]
    expect(call[1].method).toBe('POST')
    expect(call[1].headers['Content-Type']).toBe('application/json')
    expect(JSON.parse(call[1].body)).toEqual({title: 'Saved'})
  })
})

describe('del', () => {
  it('sends DELETE request without error on 204', async () => {
    global.fetch = mockFetch(null, 204)
    await expect(del('/api/v1/items/123')).resolves.toBeUndefined()
    const call = (fetch as any).mock.calls[0]
    expect(call[1].method).toBe('DELETE')
  })

  it('throws on error response', async () => {
    global.fetch = mockFetch({error: 'item not found'}, 404)
    await expect(del('/api/v1/items/bad')).rejects.toThrow('item not found')
  })
})

describe('postFile', () => {
  it('sends FormData with file', async () => {
    global.fetch = mockFetch({filename: 'test.jpg'})
    const file = new File(['data'], 'test.jpg', {type: 'image/jpeg'})
    const result = await postFile<any>('/api/v1/images/upload', file)
    expect(result).toEqual({filename: 'test.jpg'})
    const call = (fetch as any).mock.calls[0]
    expect(call[1].method).toBe('POST')
    expect(call[1].body).toBeInstanceOf(FormData)
  })
})

describe('downloadFile', () => {
  it('creates blob download link', async () => {
    global.fetch = mockFetch(null, 200, {'Content-Disposition': 'attachment; filename="backup.zip"'})

    // Mock DOM methods
    const clickSpy = vi.fn()
    const createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue({
      href: '',
      download: '',
      click: clickSpy,
    } as unknown as HTMLAnchorElement)
    vi.spyOn(document.body, 'appendChild').mockImplementation((el) => el)
    vi.spyOn(document.body, 'removeChild').mockImplementation((el) => el)
    vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:test')
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})

    await downloadFile('/api/v1/export/backup')

    expect(clickSpy).toHaveBeenCalled()
    createElementSpy.mockRestore()
  })
})
