import { describe, expect, it } from 'vitest'
import { headwordOnSubmit } from './search'

describe('headwordOnSubmit', () => {
  it('looks up the typed headword on Enter even when a suggestion is highlighted', () => {
    expect(headwordOnSubmit('hello', ['hell', 'hello', 'helmet'], 0)).toBe('hello')
  })

  it('looks up the typed headword when no suggestion is highlighted', () => {
    expect(headwordOnSubmit('hello', ['hell', 'hello'], -1)).toBe('hello')
  })
})
