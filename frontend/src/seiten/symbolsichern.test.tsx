// **Was beim zweiten Mal geschieht, und was beim leeren Blatt** —
// #890,
// Befunde 1 und 3 des Nachlaufs zum Zweitblick über Phase X.
//
// # Warum eine eigene Datei
//
// In `symbolliste.test.tsx` steht, **dass** „Ins Set legen" das Feld setzt.
// Der Nachlauf hat angestrichen, dass damit die Klasse nicht geschlossen ist:
// geprüft war ein einzelner, geglückter Klick. Hier steht, was danach kommt —
// der zweite Klick und das leergeräumte Blatt. Der Schnitt folgt der Frage
// (§2), nicht der Zeilenzahl.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { leererSetkopf } from '../test/setkopfprobe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

function dreieck() {
  return {
    teile: [
      {
        art: 'Form',
        name: 'Dreieck',
        kombination: 'overlap',
        punkte: [
          {
            x: 0.5,
            y: 0.1,
            griffDavorX: 0,
            griffDavorY: 0,
            griffDanachX: 0,
            griffDanachY: 0,
            kurveDanach: false,
            sperre: 'free',
          },
          {
            x: 0.9,
            y: 0.8,
            griffDavorX: 0,
            griffDavorY: 0,
            griffDanachX: 0,
            griffDanachY: 0,
            kurveDanach: false,
            sperre: 'free',
          },
          {
            x: 0.1,
            y: 0.8,
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
    name: '',
    fehler: null,
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Symbol setID="7" />
    </Sprachraum>,
  )
}

describe('Ein Symbol zum zweiten Mal ins Set legen', () => {
  // **Der zweite Klick überschreibt die Datei des ersten.**
  //
  // # Der Befund
  //
  // `Ablegen` gab jedes Mal einen freien Namen. Zeichnen, ablegen,
  // weiterzeichnen, wieder ablegen — das gab `symbol2.mse-symbol`, das Feld
  // zeigte darauf, und `symbol1.mse-symbol` blieb liegen. Beim Sichern
  // kopiert `ausAblageKopieren` **jede** Datei des Ablageverzeichnisses; die
  // verwaiste ging als Beiwerk ins Paket.
  it('nennt den Namen von vorhin', async () => {
    const SymbolSetzen = vi
      .fn()
      .mockResolvedValue({ ...dreieck(), name: 'symbol1.mse-symbol' })

    const SetwertSetzen = vi.fn().mockResolvedValue({ fehler: null })

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      SymbolSetzen,
      SetwertSetzen,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )
    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    await waitFor(() => {
      expect(SymbolSetzen).toHaveBeenCalledWith('7', expect.anything(), '')
    })

    // **Auf das Ende des ersten Sicherns gewartet, nicht auf den Ruf**
    // (#1809): den Namen kennt die Seite erst mit der Antwort, und solange
    // sie aussteht, ruht der Knopf.
    await waitFor(() => {
      expect(screen.getByRole('button', { name: texte.sichern })).toBeEnabled()
    })

    // **Der zweite Klick**, und diesmal steht der Name dabei.
    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    await waitFor(() => {
      expect(SymbolSetzen).toHaveBeenLastCalledWith(
        '7',
        expect.anything(),
        'symbol1.mse-symbol',
      )
    })
  })

  // **Und der Name bleibt auch stehen, wenn das Feldschreiben scheitert.**
  //
  // Sonst legte der nächste Versuch eine zweite Datei an, und die erste wäre
  // dieselbe Verwaisung — nur eine Stufe später.
  it('behält ihn auch nach einem gescheiterten Feldschreiben', async () => {
    const SymbolSetzen = vi
      .fn()
      .mockResolvedValue({ ...dreieck(), name: 'symbol1.mse-symbol' })

    const SetwertSetzen = vi
      .fn()
      .mockRejectedValueOnce(new Error('Das Set ist gesperrt.'))
      .mockResolvedValue({ fehler: null })

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      SymbolSetzen,
      SetwertSetzen,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.dreieck }),
    )
    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Das Set ist gesperrt.',
    )

    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    await waitFor(() => {
      expect(SymbolSetzen).toHaveBeenLastCalledWith(
        '7',
        expect.anything(),
        'symbol1.mse-symbol',
      )
    })
  })
})

describe('Ein leergeräumtes Blatt', () => {
  // **Wer alles löscht, will kein Symbol mehr — und das lässt sich sichern.**
  //
  // # Der Befund
  //
  // „Ins Set legen" war bei null Teilen angehalten. Das gelöschte Symbol blieb
  // damit im Set stehen, und nach dem nächsten Öffnen war die Zeichnung wieder
  // da. Der Test davor endete bei der sichtbaren leeren Liste und versuchte
  // gar nicht zu sichern — angestrichen vom Nachlauf, Befund 3.
  it('räumt den Feldwert und schreibt keine Datei', async () => {
    const SymbolSetzen = vi.fn()
    const SetwertSetzen = vi.fn().mockResolvedValue({ fehler: null })

    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [{ feld: 'symbol', wert: 'symbol1.mse-symbol' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
      Symbol: vi.fn().mockResolvedValue({
        teile: dreieck().teile,
        name: 'symbol1.mse-symbol',
        fehler: null,
      }),
      SymbolSetzen,
      SetwertSetzen,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Dreieck löschen' }),
    )

    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    await waitFor(() => {
      expect(SetwertSetzen).toHaveBeenCalledWith('7', 'symbol', '', null)
    })

    expect(SymbolSetzen).not.toHaveBeenCalled()
  })

  // **Ohne Symbol im Set bleibt der Knopf angehalten.**
  //
  // Es gibt nichts zu räumen und nichts zu schreiben; ein Klick, der eine
  // leere Datei anlegte, wäre die Verwaisung von Befund 1 mit anderem Anlass.
  it('bleibt angehalten, wenn das Set keines hat', async () => {
    belegeApi({
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
    })

    zeige()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: texte.sichern })).toBeDisabled()
    })
  })
})
