// **Escape und der Fokus an der Spalte über dem Zentrum** — angestrichen
// vom Zweitblick über Phase 6.3 (M1): jedes Escape auf der Seite schloss
// sie, auch in einem Feld des Zentrums und bei einer ruhenden Überlagerung,
// und ⌥⌘[ hängte sie aus, ohne den Fokus zurückzugeben.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import { effekteNachholen } from '../test/effekte'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Sprachraum } from '../sprache/kontext'

let raeumeBreite: () => void = () => undefined

beforeEach(() => {
  raeumeBreite = gibJsdomEineBreite(1000)
})

afterEach(() => {
  raeumeBreite()
})

function Probe() {
  return (
    <Sprachraum sprache="de">
      <Arbeitsplatz
        bezeichnung="Kartenarbeitsplatz"
        navigator={{
          bezeichnung: 'Karten',
          inhalt: <button type="button">Eine Karte</button>,
        }}
        zentrum={<input aria-label="Suche im Zentrum" />}
        inspector={{ bezeichnung: 'Felder', inhalt: <p>Das Formular</p> }}
      />
    </Sprachraum>
  )
}

async function offen() {
  render(<Probe />)
  await effekteNachholen()
  await userEvent.click(
    screen.getByRole('button', { name: 'Navigator einblenden' }),
  )
  expect(screen.getByRole('complementary', { name: 'Karten' })).toHaveAttribute(
    'data-ueber',
    'zentrum',
  )
}

describe('Escape an der Spalte über dem Zentrum', () => {
  it('gehört einem Feld des Zentrums, nicht der Spalte', async () => {
    await offen()
    const feld = screen.getByRole('textbox', { name: 'Suche im Zentrum' })
    feld.focus()

    fireEvent.keyDown(feld, { key: 'Escape' })

    expect(screen.getByRole('complementary', { name: 'Karten' })).toBeVisible()
    expect(feld).toHaveFocus()
  })

  it('lässt ein Escape, das schon jemand behandelt hat', async () => {
    await offen()
    ;(document.activeElement as HTMLElement | null)?.blur()
    const vorher = (ereignis: Event) => {
      ereignis.preventDefault()
    }
    window.addEventListener('keydown', vorher, { capture: true, once: true })

    fireEvent.keyDown(window, { key: 'Escape' })

    expect(screen.getByRole('complementary', { name: 'Karten' })).toBeVisible()
  })

  it('schließt sie, wenn der Fokus in ihr steht, und gibt ihn dem Knopf', async () => {
    await offen()
    const inDer = screen.getByRole('button', { name: 'Eine Karte' })
    inDer.focus()

    fireEvent.keyDown(inDer, { key: 'Escape' })

    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
    expect(
      screen.getByRole('button', { name: 'Navigator einblenden' }),
    ).toHaveFocus()
  })

  it('gibt den Fokus dem Knopf, wenn ⌥⌘[ sie schließt, während er in ihr steht', async () => {
    await offen()
    const inDer = screen.getByRole('button', { name: 'Eine Karte' })
    inDer.focus()

    fireEvent.keyDown(inDer, {
      key: '„',
      code: 'BracketLeft',
      metaKey: true,
      altKey: true,
    })

    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
    expect(
      screen.getByRole('button', { name: 'Navigator einblenden' }),
    ).toHaveFocus()
  })
})
