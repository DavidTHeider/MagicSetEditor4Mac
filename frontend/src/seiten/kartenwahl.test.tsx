// Kartenwahl und Editorplatz.
//
// Von `karten.test.tsx` getrennt, als die Datei über 400 Zeilen wuchs
// (AGENTS.md §2). Der Schnitt folgt der Frage: dort, was die Liste zeigt,
// hier, was ein Klick daraus macht. Die **Anhakung** ist inzwischen ihrerseits
// ausgezogen — `kartenanhaken.test.tsx`.
// Die Kartenseite: Liste, Anhakung, Wahl und Editor.
//
// Von `seiten.test.tsx` getrennt, als die Datei über 400 Zeilen wuchs
// (AGENTS.md §2). Der Schnitt folgt der Frage, die eine Datei beantwortet:
// dort die Setliste, hier die Karten.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import type { KarteDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { type Api } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { ausKarten } from '../test/kartenproben'

// belegeSeite belegt zusätzlich die Stilauswahl.
//
// **Weil die Kartenseite sie zeigt** (#41). Ohne Beleg meldete die Fassade „die
// Methode ist nicht belegt", und ein Test über die Kartenliste fände zwei
// Warnungen statt einer — genau dafür ist die Vorrichtung streng (§3).
function belegeSeite(teile: Partial<Api>) {
  return belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    ...teile,
  })
}

afterEach(() => {
  raeumeApi()
})

function probeKarte(teile: Partial<KarteDTO> = {}): KarteDTO {
  return {
    id: '5',
    nummer: 0,
    werte: [{ feld: 'name', wert: 'Feuerball' }],
    ...teile,
  }
}

function zeigeKarten(setID: string) {
  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID={setID}
        karteID={undefined}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

describe('Kartenwahl und Editor', () => {
  // **Zwei Karten, und aufgeschlagen wird die zweite.** Mit einer stünde sie
  // seit #668
  // ohnehin offen, und der Klick prüfte nichts (§10).
  it('öffnet den Editor für die gewählte Karte', async () => {
    belegeSeite({
      ...ausKarten(() =>
        Promise.resolve({
          karten: [
            probeKarte(),
            probeKarte({
              id: '6',
              nummer: 1,
              werte: [{ feld: 'name', wert: 'Eiszapfen' }],
            }),
          ],
        }),
      ),
      Karten: () =>
        Promise.resolve({
          karten: [
            probeKarte(),
            probeKarte({
              id: '6',
              nummer: 1,
              werte: [{ feld: 'name', wert: 'Eiszapfen' }],
            }),
          ],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () =>
        Promise.resolve({
          vorschau: {
            stil: '',
            gezeichnet: '',
            bild: '/vorschau/AAAA.png',
            breite: 10,
            hoehe: 14,
            dpi: 150,
            befunde: [],
            flaechen: [],
            nennungen: [],
            vorgaben: [],
          },
        }),
    })

    zeigeKarten('1')

    // Die erste steht offen, sobald die Liste da ist.
    expect(await screen.findByDisplayValue('Feuerball')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /Eiszapfen/ }))

    expect(await screen.findByDisplayValue('Eiszapfen')).toBeInTheDocument()
  })

  // **Ein Wechsel des Sets räumt die gewählte Karte mit** (§7, Regel 1).
  //
  // Ohne das stünde der Editor einer Karte offen, die zum neuen Set nicht
  // gehört — und sein erstes Schreiben ginge an ein fremdes Set.
  //
  // **Seit #668
  // steht danach die erste Karte des neuen Sets offen**, nicht keine — so hält
  // es MSE2 (`SetWindow::onChangeSet` → `selectFirstCard`). Geprüft wird
  // deshalb, **wessen** Karte dasteht; dafür trägt jedes Set einen eigenen
  // Namen.
  it('schlägt beim Setwechsel die Karte des neuen Sets auf', async () => {
    belegeSeite({
      ...ausKarten((setID) =>
        Promise.resolve({
          karten: [
            probeKarte({
              id: `karte-${setID}`,
              werte: [{ feld: 'name', wert: `Feuerball ${setID}` }],
            }),
          ],
        }),
      ),
      Karten: (setID) =>
        Promise.resolve({
          karten: [
            probeKarte({
              id: `karte-${setID}`,
              werte: [{ feld: 'name', wert: `Feuerball ${setID}` }],
            }),
          ],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () =>
        Promise.resolve({
          vorschau: {
            stil: '',
            gezeichnet: '',
            bild: '/vorschau/AAAA.png',
            breite: 10,
            hoehe: 14,
            dpi: 150,
            befunde: [],
            flaechen: [],
            nennungen: [],
            vorgaben: [],
          },
        }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    expect(await screen.findByDisplayValue('Feuerball 1')).toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="2"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    expect(await screen.findByDisplayValue('Feuerball 2')).toBeInTheDocument()

    // **Und die Karte des vorigen Sets ist fort** — die eigentliche Zusage.
    expect(screen.queryByDisplayValue('Feuerball 1')).not.toBeInTheDocument()
  })

  // **Eine Vorgabe von außen schlägt die Karte auf** — so kommt ein Treffer aus
  // der Suche in den Editor, ohne dass jemand die Liste durchsucht.
  //
  // Am Baustein und nicht nur über die Anwendung: dort steigt die Kartenseite
  // mit der Vorgabe **ein**, und der Zweig, der eine **spätere** Vorgabe
  // übernimmt, bliebe ungeprüft.
  it('schlägt eine vorgegebene Karte auf, auch später', async () => {
    belegeSeite({
      ...ausKarten(() =>
        Promise.resolve({
          karten: [
            probeKarte(),
            probeKarte({
              id: 'karte-3',
              nummer: 1,
              werte: [{ feld: 'name', wert: 'Eiszapfen' }],
            }),
          ],
        }),
      ),
      Karten: () =>
        Promise.resolve({
          karten: [
            probeKarte(),
            probeKarte({
              id: 'karte-3',
              nummer: 1,
              werte: [{ feld: 'name', wert: 'Eiszapfen' }],
            }),
          ],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    // Ohne Vorgabe steht die **erste** offen.
    expect(await screen.findByDisplayValue('Feuerball')).toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID="karte-3"
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={() => undefined}
        />
      </Sprachraum>,
    )

    expect(await screen.findByDisplayValue('Eiszapfen')).toBeInTheDocument()
  })
})
