// **Ein Zug, der endet, ohne die Fläche zu verlassen** —
// #1112.
//
// `dragleave` kommt nicht in jedem Fall: Escape über der Fläche, ein Zug aus
// dem Fenster heraus, überlappende Bildflächen. Ob dann noch eines kommt,
// entscheidet die Plattform — `dragend` kommt immer.
//
// **Die Folge war dieselbe wie beim gemeldeten Fall**, nur seltener: das
// nächste Ablegen **irgendwo** traf das Feld von vorhin.
//
// Eigene Datei, weil `bildflaeche.test.tsx` mit diesen beiden Läufen über 400
// Zeilen ginge (§2).
import { fireEvent, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { type MitLaufzeit, zeigeBildflaeche } from '../test/bildflaeche'

const texte = kataloge.de

afterEach(() => {
  delete (globalThis as MitLaufzeit).runtime
  raeumeApi()
})

describe('Ein Zug, der ohne dragleave endet', () => {
  it('vergisst das gemerkte Bildfeld', async () => {
    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    ;(globalThis as MitLaufzeit).runtime = {
      OnFileDrop: (rueckruf) => {
        ablegen = rueckruf
      },
      OnFileDropOff: () => undefined,
    }

    const { BildAusDatei } = zeigeBildflaeche()

    const flaeche = await screen.findByRole('button', {
      name: `kunst — ${texte.seiten.karten.bildEinsetzen}`,
    })

    await effekteNachholen()

    fireEvent.dragOver(flaeche)
    fireEvent.dragEnd(flaeche)
    ablegen?.(0, 0, ['/Users/david/urlaub.jpg'])

    await waitFor(() => {
      expect(BildAusDatei).not.toHaveBeenCalled()
    })

    // **Der Beleg, dass dieser Weg trägt** (§3): ohne ihn wäre die Verneinung
    // auch grün, wenn `ablegen` gar nichts mehr auslöste.
    fireEvent.dragOver(flaeche)
    ablegen?.(0, 0, ['/Users/david/urlaub.jpg'])

    await waitFor(() => {
      expect(BildAusDatei).toHaveBeenCalledWith(
        '5',
        'kunst',
        '/Users/david/urlaub.jpg',
      )
    })
  })

  // **Und ein Textfeld räumt die Erinnerung auch am Zugende nicht weg.**
  it('lässt ein Textfeld das gemerkte Bildfeld in Ruhe', async () => {
    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    ;(globalThis as MitLaufzeit).runtime = {
      OnFileDrop: (rueckruf) => {
        ablegen = rueckruf
      },
      OnFileDropOff: () => undefined,
    }

    const { BildAusDatei } = zeigeBildflaeche()

    const bildfeld = await screen.findByRole('button', {
      name: `kunst — ${texte.seiten.karten.bildEinsetzen}`,
    })

    await effekteNachholen()

    const textfeld = await screen.findByRole('button', {
      name: `name — ${texte.seiten.karten.feldWaehlen}`,
    })

    fireEvent.dragOver(bildfeld)
    fireEvent.dragEnd(textfeld)
    ablegen?.(0, 0, ['/Users/david/urlaub.jpg'])

    await waitFor(() => {
      expect(BildAusDatei).toHaveBeenCalledWith(
        '5',
        'kunst',
        '/Users/david/urlaub.jpg',
      )
    })
  })
})
