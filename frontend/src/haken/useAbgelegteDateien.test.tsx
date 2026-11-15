// **Eine gezogene Datei landet im Feld, über dem sie losgelassen wurde** —
// #246 —
// und **ein Setpaket ist kein Bild**
// (#891).
//
// # Warum die Pfade aus Go kommen
//
// Der Browser sieht bei einem Ablegen nur einen **Namen**, keinen Pfad — das
// ist eine Sicherheitszusage jeder Webansicht. Wails hebt sie für die eigene
// Anwendung auf und meldet die **absoluten** Pfade über `runtime.OnFileDrop`
// (`EnableFileDrop` in den Optionen).
//
// **Über welchem Feld losgelassen wurde, weiß dagegen nur das Fenster**: Go
// bekommt Koordinaten, und die in ein Feld umzurechnen hieße, den Aufbau der
// Seite ein zweites Mal zu kennen.
import { cleanup, renderHook } from '@testing-library/react'
import { type ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Ablageraum, useAblage } from '../ablage'
import { Sprachraum } from '../sprache/kontext'
import { useBildfeld } from './useBildfeld'
import { useAbgelegteDateien } from './useAbgelegteDateien'

interface MitLaufzeit {
  runtime?: {
    OnFileDrop?: (
      rueckruf: (x: number, y: number, pfade: string[]) => void,
      ueberZiel: boolean,
    ) => void
    OnFileDropOff?: () => void
  }
}

afterEach(() => {
  delete (globalThis as MitLaufzeit).runtime
})

// belegeLaufzeit legt eine Wails-Laufzeit ins Fenster und gibt den Rückruf.
function belegeLaufzeit() {
  let gemerkt: ((x: number, y: number, pfade: string[]) => void) | undefined
  const ab = vi.fn()

  ;(globalThis as MitLaufzeit).runtime = {
    OnFileDrop: (rueckruf) => {
      gemerkt = rueckruf
    },
    OnFileDropOff: ab,
  }

  return { ablegen: (pfade: string[]) => gemerkt?.(0, 0, pfade), ab }
}

// gehaenge stellt den Ablageraum bereit — so wie `App` es tut.
function gehaenge({ children }: { children: ReactNode }) {
  return <Ablageraum>{children}</Ablageraum>
}

// stelleAn meldet den Horcher an und gibt die Ablagestelle.
//
// **Über einen Haken und nicht über einen Verweis von außen**: die Stelle
// entsteht im Raum, und ein Test, der sie danebenlegte, prüfte seine eigene.
function stelleAn(
  aufSet: (pfad: string) => void,
  aufFeld: (feld: string, pfad: string) => void,
) {
  return renderHook(
    () => {
      const ablage = useAblage()

      useAbgelegteDateien(aufSet)
      ablage.setzeBildweg(aufFeld)

      return ablage
    },
    { wrapper: gehaenge },
  )
}

describe('useAbgelegteDateien', () => {
  it('reicht die erste Datei an das gemerkte Feld', () => {
    const { ablegen } = belegeLaufzeit()
    const einsetzen = vi.fn()

    const { result } = stelleAn(vi.fn(), einsetzen)

    result.current.merkeFeld('kunst')
    ablegen(['/Users/david/urlaub.jpg', '/Users/david/zweites.png'])

    // **Die erste und nur die erste.** Ein Feld trägt ein Bild; wer fünf
    // Dateien zieht, bekäme sonst fünfmal dasselbe Feld überschrieben, und am
    // Ende stünde die letzte drin — ohne dass jemand das gewollt hat.
    expect(einsetzen).toHaveBeenCalledTimes(1)
    expect(einsetzen).toHaveBeenCalledWith('kunst', '/Users/david/urlaub.jpg')
  })

  // **Ohne gemerktes Feld geschieht nichts.** Wer neben die Karte zieht, hat
  // kein Feld getroffen — und eine Datei in ein beliebiges Feld zu setzen wäre
  // schlimmer als gar nichts zu tun.
  it('tut nichts, wenn kein Feld gemerkt ist', () => {
    const { ablegen } = belegeLaufzeit()
    const einsetzen = vi.fn()

    stelleAn(vi.fn(), einsetzen)
    ablegen(['/Users/david/urlaub.jpg'])

    expect(einsetzen).not.toHaveBeenCalled()
  })

  // **Und das gemerkte Feld gilt nur einmal.** Wer danach neben die Karte
  // zieht, setzte sonst ein Bild in das Feld von vorhin.
  it('vergisst das Feld nach dem Ablegen', () => {
    const { ablegen } = belegeLaufzeit()
    const einsetzen = vi.fn()

    const { result } = stelleAn(vi.fn(), einsetzen)

    result.current.merkeFeld('kunst')
    ablegen(['/Users/david/eins.jpg'])
    ablegen(['/Users/david/zwei.jpg'])

    expect(einsetzen).toHaveBeenCalledTimes(1)
  })

  // **Ein Ablegen ohne Datei ist keins.** Wails meldet auch das Loslassen
  // eines Ordners; die Liste ist dann leer.
  it('tut nichts ohne Datei', () => {
    const { ablegen } = belegeLaufzeit()
    const einsetzen = vi.fn()

    const { result } = stelleAn(vi.fn(), einsetzen)

    result.current.merkeFeld('kunst')
    ablegen([])

    expect(einsetzen).not.toHaveBeenCalled()
  })

  // **Ohne Laufzeit tut es nichts, und das ist kein Fehler**: im Browser der
  // Vorschau (`npm run vorschau`) gibt es keine, und dort lässt sich ohnehin
  // keine Datei ablegen.
  it('kommt ohne Laufzeit aus', () => {
    const einsetzen = vi.fn()

    expect(() => stelleAn(vi.fn(), einsetzen)).not.toThrow()
    expect(einsetzen).not.toHaveBeenCalled()
  })

  // **Und es meldet sich wieder ab.** Zwei Karteneditoren nacheinander
  // hinterließen sonst zwei Rückrufe, und der erste zeigte auf eine Karte, die
  // niemand mehr sieht.
  it('meldet sich beim Abräumen ab', () => {
    const { ab } = belegeLaufzeit()

    const { unmount } = stelleAn(vi.fn(), vi.fn())
    unmount()

    expect(ab).toHaveBeenCalled()
  })

  // **Ein Setpaket wird geöffnet, nicht eingesetzt** (#891).
  //
  // Und zwar **ohne Rücksicht** darauf, über welchem Feld es losgelassen
  // wurde: wer eine `.mse-set` zieht, meint kein Bild.
  it('gibt ein Setpaket weiter, statt es einzusetzen', () => {
    const { ablegen } = belegeLaufzeit()
    const einsetzen = vi.fn()
    const aufSet = vi.fn()

    const { result } = stelleAn(aufSet, einsetzen)

    result.current.merkeFeld('kunst')
    ablegen(['/Users/david/Vanguard.mse-set'])

    expect(aufSet).toHaveBeenCalledWith('/Users/david/Vanguard.mse-set')
    expect(einsetzen).not.toHaveBeenCalled()
  })

  // **Auch in Großbuchstaben.** Der Finder zeigt sie klein, das Dateisystem
  // von macOS unterscheidet sie nicht — ein Vergleich, der es täte, hinge an
  // einer Schreibweise, die niemand sieht.
  it('erkennt die Endung unabhängig von der Schreibweise', () => {
    const { ablegen } = belegeLaufzeit()
    const aufSet = vi.fn()

    stelleAn(aufSet, vi.fn())
    ablegen(['/Users/david/Vanguard.MSE-SET'])

    expect(aufSet).toHaveBeenCalled()
  })

  // **Und das gemerkte Feld gilt auch dann nur einmal.** Wer ein Set zieht und
  // danach ein Bild, setzte es sonst in das Feld von vorhin.
  it('vergisst das Feld auch nach einem Setpaket', () => {
    const { ablegen } = belegeLaufzeit()
    const einsetzen = vi.fn()

    const { result } = stelleAn(vi.fn(), einsetzen)

    result.current.merkeFeld('kunst')
    ablegen(['/Users/david/Vanguard.mse-set'])
    ablegen(['/Users/david/urlaub.jpg'])

    expect(einsetzen).not.toHaveBeenCalled()
  })
})

// **Ein Zug mit mehreren Dateien verliert keine** — Befund 6 des Zweitblicks
// über Phase Y.
//
// Bis zum 2026-08-25 zählte allein `pfade[0]`. Die Begründung galt für Bilder;
// derselbe Weg behandelt aber auch Setpakete, und die gingen **wortlos**
// verloren.
describe('Ein Zug mit mehreren Dateien', () => {
  it('öffnet jedes gezogene Set, nicht nur das erste', () => {
    const { ablegen } = belegeLaufzeit()
    const geoeffnet: string[] = []

    stelleAn((pfad) => geoeffnet.push(pfad), vi.fn())
    ablegen(['/a/erstes.mse-set', '/a/zweites.mse-set'])

    expect(geoeffnet).toEqual(['/a/erstes.mse-set', '/a/zweites.mse-set'])
  })

  // **Ein Bild neben einem Set verliert keines von beiden.**
  //
  // Beide Reihenfolgen, weil bis dahin je nach Reihenfolge das eine oder das
  // andere still verschwand.
  it('setzt das Bild und öffnet das Set, in jeder Reihenfolge', () => {
    for (const zug of [
      ['/a/bild.jpg', '/a/set.mse-set'],
      ['/a/set.mse-set', '/a/bild.jpg'],
    ]) {
      const { ablegen } = belegeLaufzeit()
      const aufSet = vi.fn()
      const einsetzen = vi.fn()

      const { result } = stelleAn(aufSet, einsetzen)

      result.current.merkeFeld('kunst')
      ablegen(zug)

      expect(aufSet).toHaveBeenCalledWith('/a/set.mse-set')
      expect(einsetzen).toHaveBeenCalledWith('kunst', '/a/bild.jpg')
    }
  })
})

// **Ein abgebrochener Zug trifft nicht die nächste Karte** — Befund 2 des
// Zweitblicks über Phase Y.
//
// `onDragOver` merkt das Feld, `nimmFeld` vergisst es beim **Ablegen**. Wer
// über `kunst` zieht und den Zug abbricht, ließ es bis zum 2026-08-25 stehen:
// nach einem Kartenwechsel zeigte `bildweg` schon auf die neue Karte, und ein
// Ablegen **neben** der Karte setzte das Bild in deren `kunst`.
describe('Ein Zug, der nicht abgelegt wurde', () => {
  it('lässt das gemerkte Feld nicht über den Kartenwechsel hinweg stehen', () => {
    const { ablegen } = belegeLaufzeit()
    const einsetzen = vi.fn()

    const { result, unmount } = stelleAn(vi.fn(), einsetzen)

    // Über das Feld gezogen, dann den Zug abgebrochen — kein Ablegen.
    result.current.merkeFeld('kunst')

    // **Die Kartenseite wird abgeräumt** — genau das tut ein Kartenwechsel mit
    // dem Bildweg (`useBildfeld`).
    result.current.setzeBildweg(undefined)
    result.current.merkeFeld(undefined)

    ablegen(['/a/urlaub.jpg'])

    expect(einsetzen).not.toHaveBeenCalled()

    // **Und der Beleg, dass überhaupt jemand zuhört** (§3): derselbe Weg
    // trifft, sobald ein Feld wieder gemerkt ist.
    result.current.setzeBildweg(einsetzen)
    result.current.merkeFeld('kunst')
    ablegen(['/a/urlaub.jpg'])

    expect(einsetzen).toHaveBeenCalledWith('kunst', '/a/urlaub.jpg')

    unmount()
  })
})

// **Und der Bildweg räumt es beim Abbau selbst** — dieselbe Zusage, an der
// Stelle geprüft, an der sie eingelöst wird.
//
// Ohne diesen Lauf stünde die Regel darüber als Verabredung da: der Test
// oben räumt von Hand, was `useBildfeld` tun soll. Wer die Zeile dort
// entfernt, wird hier rot.
describe('Der Bildweg beim Abbau', () => {
  it('vergisst das gemerkte Feld', () => {
    const gehaenge = ({ children }: { children: ReactNode }) => (
      <Sprachraum sprache="de">
        <Ablageraum>{children}</Ablageraum>
      </Sprachraum>
    )

    const { result } = renderHook(
      () => {
        const ablage = useAblage()
        const bildfeld = useBildfeld({
          setID: '1',
          karteID: '7',
          vorgang: { starte: async (_schluessel, tun) => tun() },
          melden: vi.fn(),
          geglueckt: vi.fn(),
        })

        return { ablage, bildfeld }
      },
      { wrapper: gehaenge },
    )

    result.current.ablage.merkeFeld('kunst')

    // **Der Beleg, dass es überhaupt gemerkt wird** — sonst prüfte die Zeile
    // darunter nichts (§10).
    expect(result.current.ablage.nimmFeld()).toBe('kunst')

    result.current.ablage.merkeFeld('kunst')
    cleanup()

    expect(result.current.ablage.nimmFeld()).toBeUndefined()
  })
})
