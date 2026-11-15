// **Die Karte steht in der Mitte und passt ins Fenster** — UXR-06
// (#1914).
//
// # Der Befund
//
// Gemessen bei 1.280 × 800 mit dem Probebestand: im Zentrum des Arbeitsplatzes
// nahm die Vorschau **zwei Fünftel** der Breite — die Regel stammt aus der
// Zeit, als Formular und Vorschau nebeneinander standen (#1271) — und die
// Karte begann unter fünf Zeilen Bedienelemente bei y = 627. Sie war 256
// Punkte breit und stand zu sechzig Prozent unter dem Fensterrand: wer ein
// Feld im Inspector änderte, sah die Änderung nicht. §6: „Die Karte ist das
// Hauptobjekt."
//
// # Was hier gehalten wird
//
// - **Im Arbeitsplatz nimmt die Vorschau das ganze Zentrum**, und „Anpassen"
//   gibt der Karte eine Breite, die auch die Höhe unter ihr achtet
//   (`passendeBreite`, gerechnet in `zoom.test.ts`).
// - **Ohne Arbeitsplatz bleibt es, wie es war**: Formular und Vorschau
//   nebeneinander, die Vorschau zwei Fünftel (#1271).
// - **Eine Stufe gibt ihre eigene Breite**; die Höhe begrenzt nur „Anpassen".
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { vorschau } from '../test/bildflaeche'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const karte: KarteDTO = {
  id: '1',
  nummer: 0,
  werte: [{ feld: 'name', wert: 'Feuerball' }],
}

function belege() {
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    ...ausKarten(() => Promise.resolve({ karten: [karte] })),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Verlauf: () => Promise.resolve({ schritte: [] }),
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
  })
}

// buehne ist der Rahmen um das Bild — an ihm hängen die Flächen der Felder.
async function buehne(): Promise<HTMLElement> {
  const bild = await screen.findByRole('img', { name: 'Vorschau der Karte' })
  const rahmen = bild.parentElement

  if (rahmen === null) {
    throw new Error('das Bild hat keinen Rahmen')
  }

  return rahmen
}

describe('Die Karte im Zentrum', () => {
  it('nimmt das Zentrum und passt in Breite und Höhe', async () => {
    belege()
    render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID="1"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const rahmen = await buehne()

    // **Gerechnet unter jsdom**: 768 Punkte Fenster, der Rahmen bei 0, 32
    // frei zum Rand — 736 Punkte Höhe, bei 100 × 140 also 525 Punkte Breite.
    //
    // **Gewartet auf die Messung, nicht auf das Bild**: die Breite rechnet
    // ein Effekt aus der freien Höhe, und davor steht ihr Rückfall. Unter
    // Last las dieser Test am 2026-10-04 „min(100%, 171px)" — das Bild war
    // da, der Effekt noch nicht gelaufen (§3, „gewartet wird auf die Wirkung").
    await waitFor(() => {
      expect(rahmen.style.width).toBe('min(100%, 525px)')
    })
    expect(rahmen.closest('.w-2\\/5')).toBeNull()

    // **Eine Stufe gibt ihre eigene Breite** — und nicht die passende. Die
    // Vorrichtung antwortet auf jeden Faktor mit 100 Punkten; bei 200 % sind
    // das 50 Vorlagenpunkte, doppelt gezeigt (`cssBreite`).
    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: 'Größe der Vorschau' }),
      '200',
    )
    expect((await buehne()).style.width).toBe('100px')
  })

  it('bleibt ohne Arbeitsplatz zwei Fünftel breit neben dem Formular', async () => {
    belege()
    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    const rahmen = await buehne()

    expect(rahmen.style.width).toBe('')
    expect(rahmen.closest('.w-2\\/5')).not.toBeNull()
  })
})
