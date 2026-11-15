// **Der Rand für den Fokus gilt nur, solange die feste Leiste zu sehen ist**
// — AP-10 (#2367). `html:has(…)` traf auch eine Leiste, die `Activity`
// verborgen im Baum hält, und verschob das Rollen auf jeder anderen Seite.
import { render } from '@testing-library/react'
import { Activity } from 'react'
import { afterEach, expect, it } from 'vitest'

import { Ausgabefuss } from './Ausgabefuss'

const marke = () => document.documentElement.dataset.ausgabeleiste

afterEach(() => {
  delete document.documentElement.dataset.ausgabeleiste
})

it('setzt die Marke, solange eine feste Leiste steht, und nimmt sie mit ihr', () => {
  const { unmount } = render(<Ausgabefuss>Knopf</Ausgabefuss>)
  expect(marke()).toBe('fest')

  unmount()
  expect(marke()).toBeUndefined()
})

it('setzt sie für eine freie Leiste nicht', () => {
  render(<Ausgabefuss fest={false}>Knopf</Ausgabefuss>)

  expect(marke()).toBeUndefined()
})

it('folgt der Sichtbarkeit unter Activity', () => {
  const zeige = (mode: 'visible' | 'hidden') => (
    <Activity mode={mode}>
      <Ausgabefuss>Knopf</Ausgabefuss>
    </Activity>
  )
  const { rerender } = render(zeige('hidden'))
  expect(marke()).toBeUndefined()

  rerender(zeige('visible'))
  expect(marke()).toBe('fest')

  rerender(zeige('hidden'))
  expect(marke()).toBeUndefined()
})

it('behält sie, solange noch eine zweite steht', () => {
  const { rerender } = render(
    <>
      <Ausgabefuss>Eins</Ausgabefuss>
      <Ausgabefuss>Zwei</Ausgabefuss>
    </>,
  )

  rerender(<Ausgabefuss>Zwei</Ausgabefuss>)
  expect(marke()).toBe('fest')
})

// **Die Marke an `html` heißt anders als die Leiste** — sonst träfe
// `[data-ausgabefuss=fest] button` jeden Knopf der Seite (gesehen am
// 2026-10-17: das Drehbuch klickte die Seitenleiste statt „Bilder ausgeben“).
it('trägt an html einen anderen Namen als die Leiste', () => {
  render(<Ausgabefuss>Knopf</Ausgabefuss>)

  expect(document.documentElement).not.toHaveAttribute('data-ausgabefuss')
  expect(document.querySelectorAll('[data-ausgabefuss=fest]')).toHaveLength(1)
})
