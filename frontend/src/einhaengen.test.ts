import { describe, expect, it } from 'vitest'

import { wurzelSuchen } from './einhaengen'

describe('wurzelSuchen', () => {
  it('findet das Element #wurzel', () => {
    document.body.innerHTML = '<div id="wurzel"></div>'

    expect(wurzelSuchen(document).id).toBe('wurzel')
  })

  it('meldet ein fehlendes Element, statt undefined weiterzureichen', () => {
    document.body.innerHTML = ''

    expect(() => wurzelSuchen(document)).toThrow(/#wurzel/)
  })
})
