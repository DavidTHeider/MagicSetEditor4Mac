// **Was ein Tisch lädt, gibt er beim Abräumen frei** — B04
// (#2223).
//
// PixiJS legt jede Textur aus `Assets.load` in einem **globalen** Speicher ab
// und gibt sie erst mit `Assets.unload` frei. Die Spielbilder sind
// inhaltsadressiert: wer eine Karte bearbeitet und den Tisch wieder öffnet,
// lädt eine neue Adresse — und die alte blieb liegen. Im Zwei-Stunden-Lauf
// wuchs der Webprozess so um 34 MB je Minute.
import { describe, expect, it } from 'vitest'

import { neuesTexturbuch } from './texturbuch'

describe('das Texturbuch eines Tisches', () => {
  it('gibt beim Abräumen jede geladene Adresse einmal frei', () => {
    const buch = neuesTexturbuch(new Map())

    expect(buch.geladen('a')).toBe('behalten')
    expect(buch.geladen('b')).toBe('behalten')
    expect(buch.geladen('a')).toBe('behalten')

    expect(buch.raeume().sort()).toEqual(['a', 'b'])
    expect(buch.raeume()).toEqual([])
  })

  it('gibt ein Bild, das erst nach dem Abräumen ankommt, gleich wieder frei', () => {
    const buch = neuesTexturbuch(new Map())
    buch.raeume()

    expect(buch.geladen('spät')).toBe('freigeben')
  })

  it('lässt stehen, was ein zweiter Tisch noch braucht', () => {
    const nutzer = new Map<string, number>()
    const erster = neuesTexturbuch(nutzer)
    const zweiter = neuesTexturbuch(nutzer)

    erster.geladen('geteilt')
    erster.geladen('nur-erster')
    zweiter.geladen('geteilt')

    expect(erster.raeume()).toEqual(['nur-erster'])
    expect(zweiter.raeume()).toEqual(['geteilt'])
  })

  it('gibt ein spätes Bild nicht frei, solange ein anderer Tisch es hält', () => {
    const nutzer = new Map<string, number>()
    const erster = neuesTexturbuch(nutzer)
    const zweiter = neuesTexturbuch(nutzer)

    zweiter.geladen('geteilt')
    erster.raeume()

    expect(erster.geladen('geteilt')).toBe('behalten')
    expect(zweiter.raeume()).toEqual(['geteilt'])
  })
})
