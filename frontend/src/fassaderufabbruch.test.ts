// **`rufAbbrechen` fasst nur an, was abbrechbar ist** — BK-07d
// (#2118).
import { describe, expect, it, vi } from 'vitest'

import { rufAbbrechen } from './fassade'

describe('rufAbbrechen', () => {
  it('ruft das cancel eines abbrechbaren Rufs', () => {
    const cancel = vi.fn()
    rufAbbrechen(Object.assign(Promise.resolve(), { cancel }))

    expect(cancel).toHaveBeenCalledTimes(1)
  })

  it('lässt einen Ruf ohne cancel und keinen Ruf in Ruhe', () => {
    expect(() => {
      rufAbbrechen(Promise.resolve())
      rufAbbrechen(undefined)
    }).not.toThrow()
  })
})
