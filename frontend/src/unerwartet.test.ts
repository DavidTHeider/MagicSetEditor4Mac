// **Ein Fehler ohne Schlüssel erreicht das Fenster nicht im Rohtext** — F03
// (#2342),
// Befunde A-03 und A-24.
//
// # Was kam
//
// Die Panik einer Bindung kam als Wails-Satz an —
// `github.com/…/internal/app.Bindungen.Sets: panic: runtime error: …` —,
// mit Paketpfad und englischem Laufzeitsatz, und ein `TypeError` aus dem
// eigenen Code als `Cannot read properties of undefined`. Beides zeigte
// `meldungAus` so, wie es war.
//
// # Was die Prüfungen halten
//
// Beides wird ein Fehler der Klasse »unbekannt« mit einer Kennung; der
// Rohtext geht mit derselben Kennung über `FensterfehlerMelden` ins
// Protokoll — **einmal** je Fehler, auch wenn ihn mehrere Stellen zeigen.
import { afterEach, describe, expect, it, vi } from 'vitest'

import { vorsatz, wailsBindungen } from './anbindung'
import { Bindungsfehler } from './bindungsfehler'
import { meldungAus, vorgangAus } from './meldung'
import { kataloge } from './sprache/kataloge'
import { belegeApi, raeumeApi } from './test/api'

afterEach(() => {
  raeumeApi()
})

const panik = `github.com/DavidTHeider/MagicSetEditor4Mac/internal/app.Bindungen.Sets: panic: runtime error: invalid memory address or nil pointer dereference`

describe('Die Panik einer Bindung', () => {
  it('wird ein Fehler »unbekannt« mit Kennung, und der Rohtext geht ins Protokoll', async () => {
    const gerufen: [string, unknown[]][] = []
    const gebaut = wailsBindungen({
      ByName: (name: string, ...argumente: unknown[]) => {
        gerufen.push([name, argumente])

        return name.endsWith('FensterfehlerMelden')
          ? Promise.resolve({ fehler: null })
          : Promise.reject(new Error(panik))
      },
    }) as Record<string, () => Promise<unknown>>

    const ausgang = await gebaut.Sets?.().catch((grund: unknown) => grund)

    expect(ausgang).toBeInstanceOf(Bindungsfehler)
    const fehler = ausgang as Bindungsfehler
    expect(fehler.code).toBe('unbekannt')
    expect(fehler.werte.vorgang).toMatch(/^[0-9a-f]{4}-[0-9a-f]{4}$/)

    const satz = meldungAus(fehler, kataloge.de)
    expect(satz).toContain(fehler.werte.vorgang)
    expect(satz).not.toMatch(/github\.com|panic|runtime error/)

    expect(gerufen).toContainEqual([
      `${vorsatz}FensterfehlerMelden`,
      [
        {
          kennung: fehler.werte.vorgang,
          ebene: 'ruf',
          bereich: 'sets',
          meldung: panik,
        },
      ],
    ])
  })
})

describe('Ein Fehler mit eigenem Typ in einer Meldung', () => {
  it('zeigt den Satz mit Kennung und meldet den Rohtext einmal', () => {
    const melden = vi.fn(() => Promise.resolve({ fehler: null }))
    belegeApi({ FensterfehlerMelden: melden })
    const ursache = new TypeError(
      "Cannot read properties of undefined (reading 'name')",
    )

    const erster = meldungAus(ursache, kataloge.de)
    const zweiter = meldungAus(ursache, kataloge.en)
    const kennung = vorgangAus(ursache)

    expect(kennung).toMatch(/^[0-9a-f]{4}-[0-9a-f]{4}$/)
    expect(erster).toContain(kennung)
    expect(erster).not.toContain('Cannot read')
    expect(zweiter).toContain(kennung)
    expect(melden).toHaveBeenCalledTimes(1)
    expect(melden).toHaveBeenCalledWith({
      kennung,
      ebene: 'meldung',
      bereich: 'fenster',
      meldung:
        "TypeError: Cannot read properties of undefined (reading 'name')",
    })
  })

  // **Bestenfalls-Arbeit** (§7, Regel 3): scheitert die Meldung oder gibt
  // es keine Fassade, steht der Satz mit Kennung trotzdem da.
  it('zeigt die Kennung auch, wenn die Meldung scheitert', async () => {
    const melden = vi.fn(() => Promise.reject(new RangeError('voll')))
    belegeApi({ FensterfehlerMelden: melden })
    const ursache = new RangeError('Invalid array length')

    expect(meldungAus(ursache, kataloge.de)).toContain(vorgangAus(ursache))
    await Promise.resolve()
    expect(melden).toHaveBeenCalledTimes(1)
  })

  it('zeigt die Kennung auch ohne Fassade', () => {
    const ursache = new SyntaxError('Unexpected token')

    expect(vorgangAus(ursache)).toMatch(/^[0-9a-f]{4}-[0-9a-f]{4}$/)
  })

  it('lässt einen schlichten Fehler, wie er ist', () => {
    expect(meldungAus(new Error('network down'), kataloge.en)).toBe(
      'network down',
    )
    expect(vorgangAus(new Error('network down'))).toBeUndefined()
  })
})
