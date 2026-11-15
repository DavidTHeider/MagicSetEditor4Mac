// Wächter über die Kataloge (AGENTS.md §9).
//
// Der Typ `Katalog` hält bereits die **Struktur**: ein fehlender englischer
// Schlüssel ist ein Übersetzungsfehler, kein Testbefund. Was er nicht hält, ist
// der **Inhalt** — eine englische Zeichenkette, die versehentlich den deutschen
// Satz trägt, ist typrichtig und trotzdem falsch. Genau das prüft dieser Test.
//
// # Was er nicht prüft
//
// Ob die Übersetzung *gut* ist. Er sieht nur, dass sie eine andere ist und dass
// keine leer bleibt. Wo ein Wort in beiden Sprachen gleich lautet — Eigennamen
// wie „MagicSetEditor4Mac" —, steht es in `gleichInBeidenSprachen`; sonst wäre
// die Prüfung eine Einladung, den Eintrag zu entfernen statt zu übersetzen.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { kataloge, sprachen, type Katalog } from './kataloge'
import { Sprachraum, useTexte } from './kontext'
import { gleichInBeidenSprachen } from './gleichinbeidensprachen'

describe('Sprachkataloge', () => {
  it('führt beide Sprachen', () => {
    expect(Object.keys(kataloge).sort()).toEqual([...sprachen].sort())
  })

  it('hat zu jedem deutschen Schlüssel einen englischen', () => {
    expect(pfade(kataloge.en)).toEqual(pfade(kataloge.de))
  })

  it('lässt keinen Eintrag leer', () => {
    for (const sprache of sprachen) {
      for (const [pfad, wert] of eintraege(kataloge[sprache])) {
        expect(wert.trim(), `${sprache}: ${pfad}`).not.toBe('')
      }
    }
  })

  it('übersetzt wirklich, statt den deutschen Satz zu wiederholen', () => {
    const deutsch = new Map(eintraege(kataloge.de))

    for (const [pfad, englisch] of eintraege(kataloge.en)) {
      if (gleichInBeidenSprachen.has(pfad)) {
        continue
      }

      expect(englisch, `${pfad} ist unübersetzt`).not.toBe(deutsch.get(pfad))
    }
  })
})

describe('useTexte', () => {
  function Probe() {
    return <span>{useTexte().allgemein.sprache.bezeichnung}</span>
  }

  it('gibt ohne Sprachraum den deutschen Katalog', () => {
    render(<Probe />)

    expect(screen.getByText('Sprache')).toBeInTheDocument()
  })

  it('folgt der eingestellten Sprache', () => {
    render(
      <Sprachraum sprache="en">
        <Probe />
      </Sprachraum>,
    )

    expect(screen.getByText('Language')).toBeInTheDocument()
  })
})

// ------------------------------------------------------------------ Hilfsmittel

function eintraege(katalog: Katalog): [string, string][] {
  const gesammelt: [string, string][] = []

  const gehe = (wert: unknown, pfad: string): void => {
    if (typeof wert === 'string') {
      gesammelt.push([pfad, wert])

      return
    }

    for (const [schluessel, kind] of Object.entries(wert as object)) {
      gehe(kind, pfad === '' ? schluessel : `${pfad}.${schluessel}`)
    }
  }

  gehe(katalog, '')

  return gesammelt
}

function pfade(katalog: Katalog): string[] {
  return eintraege(katalog)
    .map(([pfad]) => pfad)
    .sort()
}
