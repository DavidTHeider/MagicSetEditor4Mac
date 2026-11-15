// **Welche Seiten stehen bleiben** — UX-03.AK-02
// (#1675).
import { renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import {
  aufheben,
  nichtsAufgehoben,
  useAufgehobeneSeiten,
  type Aufgehoben,
} from './useAufgehobeneSeiten'

import type { Seitenschluessel } from '../navigation'

// gang geht eine Folge von Seiten im selben Set.
function gang(seiten: Seitenschluessel[], setID = 'a'): Aufgehoben {
  return seiten.reduce(
    (stand, seite) => aufheben(stand, seite, setID),
    nichtsAufgehoben,
  )
}

describe('aufheben', () => {
  it('hebt jede besuchte Anwendungsseite einmal auf', () => {
    expect(gang(['start', 'decks', 'sammlung', 'decks']).anwendung).toEqual([
      'start',
      'decks',
      'sammlung',
    ])
  })

  it('gibt denselben Stand zurück, wenn nichts dazukommt', () => {
    const stand = gang(['decks', 'karten'])

    expect(aufheben(stand, 'decks', 'a')).toBe(stand)
    expect(aufheben(stand, 'karten', 'a')).toBe(stand)
  })

  it('merkt sich die Laschen und die zuletzt besuchte Setseite', () => {
    const stand = gang(['karten', 'ausgeben', 'karten', 'decks'])

    expect(stand.laschen).toEqual(['karten', 'ausgeben'])
    expect(stand.setseite).toBe('karten')
  })

  // **Drei Seiten, ein Assistent** (UXR-09, #1917): `bilder` und
  // `druckbogen` stehen auf dem Platz von `ausgeben` — sonst stünden drei
  // Assistenten verborgen nebeneinander, jeder mit seiner eigenen Wahl.
  it('legt Bilder und Druckbogen auf den Platz des Ausgabeassistenten', () => {
    const stand = gang(['karten', 'druckbogen', 'bilder', 'ausgeben'])

    expect(stand.laschen).toEqual(['karten', 'ausgeben'])
    expect(stand.setseite).toBe('ausgeben')
    expect(gang(['bilder']).setseite).toBe('bilder')
  })

  // **Unter einer Schublade stehen die Karten** (UX-08, #1680).
  it('legt unter eine Schublade die Karten', () => {
    const stand = gang(['stil', 'setangaben', 'konsole'])

    expect(stand.laschen).toEqual(['karten'])
    expect(stand.schubladen).toEqual(['stil', 'setangaben', 'konsole'])
    expect(stand.setseite).toBe('konsole')
  })

  // **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext** (§7,
  // Regel 1) — die Anwendungsseiten bleiben, sie gehören keinem Set.
  it('vergisst die Setseiten, wenn das Set wechselt', () => {
    const vorher = gang(['decks', 'karten', 'stil'])
    const nachher = aufheben(vorher, 'decks', 'b')

    expect(nachher).toEqual({
      ...nichtsAufgehoben,
      setID: 'b',
      anwendung: ['decks'],
    })
  })

  // **Ein Zurückspielen räumt alles** (§7, Regel 1, „in ihrer größten Form")
  // — Befund M9 (a) des Zweitblicks über Phase 6.3 (#1666). Stehen bleibt nur
  // die Seite, auf der es geschah: sie zeigt, wohin der alte Bestand ging.
  it('vergisst jede verlassene Seite, wenn der Bestand ersetzt wurde', () => {
    const vorher = gang(['decks', 'sammlung', 'karten', 'einstellungen'])
    const nachher = aufheben(vorher, 'einstellungen', undefined, 1)

    expect(nachher).toEqual({
      ...nichtsAufgehoben,
      bestand: 1,
      anwendung: ['einstellungen'],
    })

    // **Und danach hebt er wieder auf**, unter dem neuen Stand.
    expect(aufheben(nachher, 'decks', undefined, 1).anwendung).toEqual([
      'einstellungen',
      'decks',
    ])
  })
})

describe('useAufgehobeneSeiten', () => {
  it('führt den Stand über die Zeichnungen', () => {
    const { result, rerender } = renderHook(
      ({ seite }: { seite: Seitenschluessel }) =>
        useAufgehobeneSeiten(seite, 'a'),
      { initialProps: { seite: 'decks' as Seitenschluessel } },
    )

    rerender({ seite: 'karten' })
    rerender({ seite: 'decks' })

    expect(result.current.anwendung).toEqual(['decks'])
    expect(result.current.laschen).toEqual(['karten'])
    expect(result.current.setseite).toBe('karten')
  })
})
