// **Der Symboleditor arbeitet auch an einem Kartenfeld** — Werkstück A von
// Phase 4.11
// (#1339).
//
// # Was er dafür brauchte
//
// **Nicht viel.** `Symbol(setID, name)` liest seit jeher einen **benannten**
// Symbolstand, und `SymbolSetzen(setID, teile, unter)` schreibt ihn unter
// einem Namen — die Go-Seite war nie ans Setsymbol gebunden. Was fehlte, war
// der Weg von der Kartenzeile hierher und die Frage, **welchen Feldwert** die
// Seite liest und schreibt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

// zeichnung ist ein Symbolstand mit einem einzigen Punkt.
function zeichnung() {
  return {
    teile: [
      {
        art: 'Form',
        name: 'Kreis',
        kombination: 'overlap',
        punkte: [
          {
            x: 0.5,
            y: 0.5,
            griffDavorX: 0,
            griffDavorY: 0,
            griffDanachX: 0,
            griffDanachY: 0,
            kurveDanach: false,
            sperre: 'free',
          },
        ],
        spiegelart: '',
        kopien: 0,
        mitteX: 0,
        mitteY: 0,
        griffX: 0,
        griffY: 0,
        teile: [],
      },
    ],
    name: 'eigen.mse-symbol',
    fehler: null,
  }
}

describe('Der Symboleditor mit einem Kartenfeld als Ziel', () => {
  it('liest den Dateinamen aus dem Kartenfeld, nicht aus den Setangaben', async () => {
    const Symbolstand = vi.fn().mockResolvedValue(zeichnung())
    const Setangaben = vi.fn()

    belegeApi({
      Symbol: Symbolstand,
      Setangaben,
      Karte: vi.fn().mockResolvedValue({
        karte: {
          id: '5',
          nummer: 0,
          werte: [{ feld: 'kartensymbol', wert: 'eigen.mse-symbol' }],
        },
        fehler: null,
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" ziel={{ karteID: '5', feld: 'kartensymbol' }} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(Symbolstand).toHaveBeenCalledWith('7', 'eigen.mse-symbol')
    })

    // **Und die Setangaben werden gar nicht erst gefragt.** Ohne diese
    // Zusicherung wäre der Lauf auch dann grün, wenn beide Wege liefen — und
    // dann stünde bei einem Set ohne Symbol der leere Name im Kartenfeld.
    expect(Setangaben).not.toHaveBeenCalled()
  })

  // **Ohne Ziel bleibt es das Setsymbol.** Die Gegenprobe: ohne sie wäre auch
  // ein Editor grün, der immer die Karte fragte — und der hätte kein Set mehr.
  it('liest ohne Ziel weiter die Setangaben', async () => {
    const Setangaben = vi
      .fn()
      .mockResolvedValue({ felder: [], werte: [], fehler: null })
    const Karte = vi.fn()

    belegeApi({ Setangaben, Karte })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(Setangaben).toHaveBeenCalledWith('7')
    })

    expect(Karte).not.toHaveBeenCalled()
  })

  // **Ein Kartenfeld ohne Wert lädt nichts.** Derselbe Fall wie beim Set ohne
  // Symbol: ein Aufruf mit leerem Namen wäre eine Frage nach einer Datei, die
  // niemand gemeint hat.
  it('lädt nichts, wenn das Kartenfeld leer ist', async () => {
    const Symbolstand = vi.fn()

    const Karte = vi.fn().mockResolvedValue({
      karte: { id: '5', nummer: 0, werte: [] },
      fehler: null,
    })

    belegeApi({ Symbol: Symbolstand, Karte })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" ziel={{ karteID: '5', feld: 'kartensymbol' }} />
      </Sprachraum>,
    )

    // **Gewartet wird auf ein Abschlusssignal**: die Karte wird gefragt, und
    // erst danach steht fest, dass kein Symbolstand geholt wird. Eine
    // Verneinung ohne dieses Signal träfe schon vor dem Laden zu (§3).
    await waitFor(() => {
      expect(Karte).toHaveBeenCalledWith('5')
    })

    expect(Symbolstand).not.toHaveBeenCalled()
  })

  // **Und gespeichert wird in das Kartenfeld, nicht in die Setangabe.**
  //
  // Das ist die andere Hälfte: ein Editor, der das Kartenfeld läse und in
  // `set.symbol` schriebe, hätte beim ersten Speichern das Symbol des Sets
  // überschrieben — mit einer Zeichnung, die einer Karte gehört.
  it('schreibt den Dateinamen in das Kartenfeld', async () => {
    const SymbolSetzen = vi.fn().mockResolvedValue({
      teile: [],
      name: 'symbol1.mse-symbol',
      fehler: null,
    })

    const FeldSetzen = vi.fn().mockResolvedValue({ fehler: null })
    const SetwertSetzen = vi.fn()

    belegeApi({
      Grundform: vi.fn().mockResolvedValue(zeichnung()),
      Karte: vi.fn().mockResolvedValue({
        karte: { id: '5', nummer: 0, werte: [] },
        fehler: null,
      }),
      SymbolSetzen,
      FeldSetzen,
      SetwertSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" ziel={{ karteID: '5', feld: 'kartensymbol' }} />
      </Sprachraum>,
    )

    const texte = kataloge.de.seiten.karten.symbol

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )
    await screen.findByRole('button', { name: 'Punkt 1 von Kreis' })
    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith(
        '7',
        '5',
        'kartensymbol',
        'symbol1.mse-symbol',
        null,
      )
    })

    // **Und die Setangabe bleibt unberührt.**
    expect(SetwertSetzen).not.toHaveBeenCalled()
  })
})
