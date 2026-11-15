// **Ein Auswahlfeld: nativ bis vierzehn Werte, mit Suche ab fünfzehn** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`.
//
// Die Stellen, deren Liste wachsen kann — Decks, Sets, Felder, Turniere —,
// fragen nicht selbst, ab wann eine Suche dazugehört. Sie geben dem
// `Auswahlfeld` ihre Zeilen, und das entscheidet **an einer Stelle**: unter
// `suchschwelle` bleibt es das native Feld mit dem Menü von macOS, darüber ist
// es die `Suchauswahl`.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { verstoesse } from '../test/axe'
import { Auswahlfeld } from './Auswahlfeld'
import { suchschwelle } from './suchwahl'

import type { Suchwahl } from './suchwahl'

// zeilen baut so viele Zeilen, wie verlangt sind: „Deck 01", „Deck 02" …
function zeilen(anzahl: number): Suchwahl[] {
  return Array.from({ length: anzahl }, (_, stelle) => {
    const nummer = String(stelle + 1).padStart(2, '0')

    return { wert: `d${nummer}`, name: `Deck ${nummer}` }
  })
}

interface Eigenschaften {
  anzahl: number
  aendern?: (wert: string) => void
  anfang?: string
  angehalten?: boolean
  ungueltig?: boolean
  beschriebenVon?: string
  benennen?: boolean
  klasse?: string
}

function Probe({
  anzahl,
  aendern = vi.fn(),
  anfang = 'd02',
  angehalten = false,
  ungueltig = false,
  beschriebenVon,
  benennen = false,
  klasse,
}: Eigenschaften) {
  const [wert, setzeWert] = useState(anfang)

  return (
    <Sprachraum sprache="de">
      <label htmlFor="deckwahl">Deck</label>
      <Auswahlfeld
        id="deckwahl"
        bezeichnung="Deck"
        eintraege={zeilen(anzahl)}
        wert={wert}
        aendern={(neu) => {
          aendern(neu)
          setzeWert(neu)
        }}
        angehalten={angehalten}
        ungueltig={ungueltig}
        benennen={benennen}
        {...(beschriebenVon === undefined ? {} : { beschriebenVon })}
        {...(klasse === undefined ? {} : { klasse })}
      />
    </Sprachraum>
  )
}

const feld = () => screen.getByRole('combobox', { name: 'Deck' })

describe('Auswahlfeld — die Schwelle', () => {
  it('liegt bei fünfzehn', () => {
    expect(suchschwelle).toBe(15)
  })

  it('bleibt bei vierzehn Werten ein natives Feld', async () => {
    const aendern = vi.fn()
    render(<Probe anzahl={suchschwelle - 1} aendern={aendern} />)

    expect(feld().tagName).toBe('SELECT')
    expect(feld()).toHaveValue('d02')
    expect(screen.getAllByRole('option')).toHaveLength(suchschwelle - 1)

    await userEvent.selectOptions(feld(), 'd05')

    expect(aendern).toHaveBeenCalledExactlyOnceWith('d05')
  })

  it('wird bei fünfzehn Werten eine Auswahl mit Suche', async () => {
    const aendern = vi.fn()
    render(<Probe anzahl={suchschwelle} aendern={aendern} />)

    expect(feld().tagName).toBe('INPUT')
    expect(feld()).toHaveValue('Deck 02')
    // **Die Zeilen stehen erst, wenn das Feld offen ist.**
    expect(screen.queryAllByRole('option')).toHaveLength(0)

    await userEvent.click(feld())
    await userEvent.keyboard('deck 15{Enter}')

    expect(aendern).toHaveBeenCalledExactlyOnceWith('d15')
    expect(feld()).toHaveValue('Deck 15')
  })

  it('folgt der Zahl der Werte, wenn sie über die Schwelle wächst', () => {
    const { rerender } = render(<Probe anzahl={suchschwelle - 1} />)

    expect(feld().tagName).toBe('SELECT')

    rerender(<Probe anzahl={suchschwelle} />)

    expect(feld().tagName).toBe('INPUT')
    expect(feld()).toHaveValue('Deck 02')
  })
})

describe('Auswahlfeld — was an das Feld geht', () => {
  it.each([
    ['unter der Schwelle', suchschwelle - 1, 'SELECT'],
    ['ab der Schwelle', suchschwelle, 'INPUT'],
  ])('reicht Beschreibung und Befund durch, %s', (_, anzahl, art) => {
    render(<Probe anzahl={anzahl} ungueltig beschriebenVon="hilfe" />)

    expect(feld().tagName).toBe(art)
    expect(feld()).toHaveAttribute('aria-describedby', 'hilfe')
    expect(feld()).toHaveAttribute('aria-invalid', 'true')
  })

  it.each([
    ['unter der Schwelle', suchschwelle - 1],
    ['ab der Schwelle', suchschwelle],
  ])('lässt beides weg, wo es nicht gilt, %s', (_, anzahl) => {
    render(<Probe anzahl={anzahl} />)

    expect(feld()).not.toHaveAttribute('aria-describedby')
    expect(feld()).not.toHaveAttribute('aria-invalid')
  })

  it.each([
    ['unter der Schwelle', suchschwelle - 1],
    ['ab der Schwelle', suchschwelle],
  ])('sperrt das Feld, solange ein Vorgang läuft, %s', (_, anzahl) => {
    render(<Probe anzahl={anzahl} angehalten />)

    expect(feld()).toBeDisabled()
  })

  // **Ohne Beschriftung trägt das Feld seinen Namen selbst** — die Kartenfelder
  // haben kein `<label>`, nur ein `aria-label`.
  it.each([
    ['unter der Schwelle', suchschwelle - 1],
    ['ab der Schwelle', suchschwelle],
  ])('benennt sich selbst, wenn es gebeten wird, %s', (_, anzahl) => {
    render(
      <Sprachraum sprache="de">
        <Auswahlfeld
          id="ohnelabel"
          bezeichnung="Wasserzeichen"
          eintraege={zeilen(anzahl)}
          wert="d01"
          aendern={vi.fn()}
          benennen
        />
      </Sprachraum>,
    )

    expect(
      screen.getByRole('combobox', { name: 'Wasserzeichen' }),
    ).toBeInTheDocument()
  })

  it('setzt die Klassen auf das native Feld', () => {
    render(<Probe anzahl={suchschwelle - 1} klasse="meine-klasse" />)

    expect(feld()).toHaveClass('meine-klasse')
  })

  it('hat ohne eigene Klassen eine Kante und einen sichtbaren Fokusring', () => {
    render(<Probe anzahl={suchschwelle - 1} />)

    expect(feld()).toHaveClass('border-kante')
    expect(feld().className).toContain('focus-visible:outline-ring')
  })
})

describe('Auswahlfeld — Barrierefreiheit', () => {
  it.each([
    ['unter der Schwelle', suchschwelle - 1],
    ['ab der Schwelle', suchschwelle],
  ])('hält die Regeln, %s', async (_, anzahl) => {
    const { container } = render(<Probe anzahl={anzahl} />)

    expect(await verstoesse(container)).toEqual([])

    await userEvent.click(feld())

    expect(await verstoesse(document.body)).toEqual([])
  })
})
