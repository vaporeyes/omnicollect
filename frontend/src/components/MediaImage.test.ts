// ABOUTME: Authenticated image component lifecycle and stale-response regressions.
// ABOUTME: Verifies lazy visibility, cancellation, and object URL release without a browser network.
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {createApp, h, nextTick, reactive, type App} from 'vue'
import MediaImage from './MediaImage.vue'
import {getMedia} from '../api/client'

vi.mock('../api/client', () => ({getMedia: vi.fn()}))
const media = vi.mocked(getMedia)
let app: App | undefined
let host: HTMLElement
const flush = async () => { for (let i = 0; i < 5; i++) await nextTick() }

beforeEach(() => {
  host = document.createElement('div')
  document.body.appendChild(host)
  media.mockReset()
  vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:loaded')
  vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})
})
afterEach(() => { app?.unmount(); app = undefined; host.remove(); vi.restoreAllMocks(); vi.unstubAllGlobals() })

function mount(props: {src: string; loading?: 'lazy' | 'eager'}) {
  const state = reactive(props)
  app = createApp({render: () => h(MediaImage, {...state, alt: 'Collection image'})})
  app.mount(host)
  return state
}

describe('MediaImage', () => {
  it('loads authenticated blobs and releases them on replacement and unmount', async () => {
    media.mockResolvedValue(new Blob(['image'], {type: 'image/png'}))
    const state = mount({src: '/originals/first.png'})
    await flush()
    expect(host.querySelector('img')?.getAttribute('src')).toBe('blob:loaded')
    expect(host.querySelector('img')?.alt).toBe('Collection image')
    const firstSignal = media.mock.calls[0][1]
    state.src = '/originals/second.png'
    await flush()
    expect(firstSignal?.aborted).toBe(true)
    expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:loaded')
    app!.unmount(); app = undefined
    expect(URL.revokeObjectURL).toHaveBeenCalledTimes(2)
  })

  it('ignores a stale fetch even when it completes after cancellation', async () => {
    let finish!: (blob: Blob) => void
    media.mockImplementationOnce(() => new Promise(resolve => { finish = resolve }))
    media.mockResolvedValueOnce(new Blob(['new'], {type: 'image/png'}))
    const state = mount({src: '/originals/old.png'})
    state.src = '/originals/new.png'
    await flush()
    finish(new Blob(['old'], {type: 'image/png'}))
    await flush()
    expect(URL.createObjectURL).toHaveBeenCalledTimes(1)
  })

  it('waits for visibility before fetching lazy images', async () => {
    let intersect!: IntersectionObserverCallback
    const disconnect = vi.fn()
    vi.stubGlobal('IntersectionObserver', class {
      constructor(callback: IntersectionObserverCallback) { intersect = callback }
      observe() {}
      disconnect = disconnect
    })
    media.mockResolvedValue(new Blob(['image'], {type: 'image/png'}))
    mount({src: '/thumbnails/photo.png', loading: 'lazy'})
    expect(media).not.toHaveBeenCalled()
    intersect([{isIntersecting: true} as IntersectionObserverEntry], {} as IntersectionObserver)
    await flush()
    expect(media).toHaveBeenCalledTimes(1)
    expect(disconnect).toHaveBeenCalled()
  })
})
