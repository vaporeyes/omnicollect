// ABOUTME: Centralized fetch-based HTTP client for the OmniCollect REST API.
// ABOUTME: Supports optional Auth0 Bearer token injection via setTokenGetter.

import {captureSession} from '../auth/session'
import type {TagCount, AIAnalysisResult, AIStatus, Showcase} from './types'

const BASE_URL = (import.meta as any).env?.VITE_API_URL || ''

// Token getter function set by the Auth0 plugin when auth is configured.
// Returns an access token string, or null in local mode.
type TokenGetter = () => Promise<string>
let getToken: TokenGetter | null = null

// setTokenGetter is called by the auth plugin to wire up token retrieval.
export function setTokenGetter(fn: TokenGetter | null) {
  getToken = fn
}

async function authHeaders(): Promise<Record<string, string>> {
  if (!getToken) return {}
  const token = await getToken()
  if (typeof token !== 'string' || !token.trim()) throw new Error('Authentication did not provide an access token')
  return {Authorization: `Bearer ${token}`}
}

export class APIError extends Error {
  constructor(message: string, public readonly status: number) {super(message); this.name = 'APIError'}
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const body = await res.json()
      if (body.error) msg = body.error
    } catch { /* use status text */ }
    throw new APIError(msg, res.status)
  }
  if (res.status === 204 || res.status === 205) return undefined as T
  const text = await res.text()
  return text.trim() ? JSON.parse(text) as T : undefined as T
}

function withAbort<T>(promise: Promise<T>, signal: AbortSignal): Promise<T> {
  return new Promise((resolve, reject) => {
    const abort = () => {signal.removeEventListener('abort', abort); reject(new DOMException('Request cancelled', 'AbortError'))}
    signal.addEventListener('abort', abort, {once: true})
    promise.then(value => {signal.removeEventListener('abort', abort); resolve(value)}, error => {signal.removeEventListener('abort', abort); reject(error)})
    if (signal.aborted) abort()
  })
}

async function perform<T>(path: string, options: RequestInit, consume: (res: Response, check: () => void) => Promise<T> = handleResponse): Promise<T> {
  const session = captureSession()
  session.assertCurrent()
  const controller = new AbortController()
  const abort = () => controller.abort()
  const check = () => {session.assertCurrent(); if (controller.signal.aborted) throw new DOMException('Request cancelled', 'AbortError')}
  const caller = options.signal
  session.signal.addEventListener('abort', abort, {once: true})
  caller?.addEventListener('abort', abort, {once: true})
  if (caller?.aborted) abort()
  try {
    const auth = await withAbort(authHeaders(), controller.signal)
    check()
    const res = await withAbort(fetch(BASE_URL + path, {...options, signal: controller.signal, headers: {...options.headers, ...auth}}), controller.signal)
    check()
    const result = await withAbort(consume(res, check), controller.signal)
    check()
    return result
  } finally {
    session.signal.removeEventListener('abort', abort)
    caller?.removeEventListener('abort', abort)
  }
}
export function get<T>(path: string, signal?: AbortSignal): Promise<T> {
  return perform<T>(path, {signal})
}
export async function getMedia(path: string, signal?: AbortSignal): Promise<Blob> {
  if (!/^\/(originals|thumbnails)\/[A-Za-z0-9][A-Za-z0-9_.%-]*$/.test(path)) throw new Error('Invalid media path')
  return perform(path, {signal}, async res => {
    if (!res.ok) {await handleResponse(res); throw new Error('Image unavailable')}
    const blob = await res.blob()
    if (!['image/jpeg', 'image/png', 'image/gif', 'image/webp'].includes(blob.type)) throw new Error('Invalid image response')
    return blob
  })
}
export function post<T>(path: string, body: any, signal?: AbortSignal): Promise<T> {
  return perform<T>(path, {method: 'POST', signal, headers: {'Content-Type': 'application/json'}, body: JSON.stringify(body)})
}
export function put<T>(path: string, body: any): Promise<T> {
  return perform<T>(path, {method: 'PUT', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(body)})
}
export function del(path: string): Promise<void> {return perform<void>(path, {method: 'DELETE'})}
export function postFile<T>(path: string, file: File, fieldName = 'image', signal?: AbortSignal): Promise<T> {
  const body = new FormData()
  body.append(fieldName, file)
  return perform<T>(path, {method: 'POST', body, signal})
}

export async function getAllTags(): Promise<TagCount[]> {
  return get<TagCount[]>('/api/v1/tags')
}

export async function renameTag(oldName: string, newName: string): Promise<{updated: number}> {
  return post<{updated: number}>('/api/v1/tags/rename', {oldName, newName})
}

export async function deleteTag(name: string): Promise<void> {
  return del('/api/v1/tags/' + encodeURIComponent(name))
}

export async function analyzeBackup(file: File): Promise<import('./types').ImportSummary> {
  return postFile<import('./types').ImportSummary>('/api/v1/import/analyze', file, 'backup')
}

export async function executeImport(tempId: string, mode: string): Promise<import('./types').ImportResult> {
  return post<import('./types').ImportResult>('/api/v1/import/execute', {tempId, mode})
}

export async function analyzeItem(imageFilename: string, moduleId: string, signal?: AbortSignal): Promise<AIAnalysisResult> {
  return post<AIAnalysisResult>('/api/v1/ai/analyze', {imageFilename, moduleId}, signal)
}

export async function getAIStatus(): Promise<AIStatus> {
  return get<AIStatus>('/api/v1/ai/status')
}

export async function toggleShowcase(moduleId: string, enabled: boolean): Promise<Showcase> {
  return post<Showcase>('/api/v1/showcases/toggle', {moduleId, enabled})
}

export async function listShowcases(): Promise<Showcase[]> {
  return get<Showcase[]>('/api/v1/showcases')
}

export async function downloadFile(path: string, body?: any): Promise<void> {
  const opts: RequestInit = body
    ? {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(body)}
    : {method: 'GET'}
  return perform(path, opts, async (res, check) => {
    if (!res.ok) throw new Error(`Download failed: HTTP ${res.status}`)
    const blob = await res.blob()
    check()
    const disposition = res.headers.get('Content-Disposition') || ''
    const match = disposition.match(/filename="?([^"]+)"?/)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    try {
      a.href = url
      a.download = match?.[1] || 'download'
      document.body.appendChild(a)
      a.click()
    } finally {a.remove(); URL.revokeObjectURL(url)}
  })
}
