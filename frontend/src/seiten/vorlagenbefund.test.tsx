// **Der Befund steht unter dem Editor** — FMT-02, S10.NAT-02 und Befund F12
// (#1691).
//
// Nach dem Artboard `Werkstatt` steht er im Zentrum und nicht im Inspector:
// der weicht bei engem Fenster, und dann hätte niemand gelesen, warum eine
// Stelle angestrichen ist.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Vorlagenbefund } from './Vorlagenbefund'

import type { Skriptbefund } from '../haken/usePufferbefund'

const texte = kataloge.de.seiten.vorlagen

function zeige(befund: Skriptbefund | undefined, sprache: 'de' | 'en' = 'de') {
  const zurStelle = vi.fn()

  render(
    <Sprachraum sprache={sprache}>
      <Vorlagenbefund befund={befund} zurStelle={zurStelle} />
    </Sprachraum>,
  )

  return zurStelle
}

describe('Der Befund der Werkstatt', () => {
  it('sagt ohne Befund leise, dass nichts zu finden ist', () => {
    zeige({ stelle: null, meldung: '', schluessel: '', werte: [] })

    expect(screen.getByText(texte.diagnoseFrei)).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.zurStelle }),
    ).not.toBeInTheDocument()
  })

  it('schweigt, solange noch nicht geprüft ist', () => {
    zeige(undefined)

    expect(screen.queryByText(texte.diagnoseFrei)).not.toBeInTheDocument()
  })

  it('nennt den Befund mit Zeile und Spalte und springt dorthin', async () => {
    const zurStelle = zeige({
      stelle: { zeile: 4, spalte: 7 },
      meldung: 'Die Feldart »bild« kennt MSE2 nicht',
      schluessel: 'unbekannterSchluessel',
      werte: [],
    })

    expect(screen.getByRole('status')).toHaveTextContent(
      'Die Feldart »bild« kennt MSE2 nicht',
    )
    expect(screen.getByText('Zeile 4, Spalte 7')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.zurStelle }))

    expect(zurStelle).toHaveBeenCalledWith({ zeile: 4, spalte: 7 })
  })
})

// **Im englischen Fenster steht der Befund auf Englisch** (§9) — Go schickt
// neben dem deutschen Satz Schlüssel und Werte, und bis FMT-02 stand trotzdem
// der deutsche Satz da.
describe('Der Befund im englischen Fenster', () => {
  const englisch = kataloge.en.seiten.vorlagen

  it.each([
    [
      'einen Syntaxfehler',
      { schluessel: 'fehltNach', werte: ['`if`', '`then`'] },
      'after `if`, `then` is missing',
    ],
    [
      'eine unbekannte Feldart',
      { schluessel: 'feldartUnbekannt', werte: ['bild', 'boolean, choice'] },
      englisch.feldartUnbekannt
        .replace('{0}', 'bild')
        .replace('{1}', 'boolean, choice'),
    ],
    [
      'eine fehlende Feldart',
      { schluessel: 'feldartFehlt', werte: [] },
      englisch.feldartFehlt,
    ],
    [
      'einen unbekannten Schlüssel mit dem deutschen Satz',
      { schluessel: 'gibtEsNicht', werte: [] },
      'der deutsche Satz',
    ],
  ])('übersetzt %s', (_art, schluessel, satz) => {
    zeige(
      {
        stelle: { zeile: 1, spalte: 1 },
        meldung: 'der deutsche Satz',
        ...schluessel,
      },
      'en',
    )

    expect(screen.getByRole('status')).toHaveTextContent(satz)
    expect(satz).not.toContain('{')
  })
})
