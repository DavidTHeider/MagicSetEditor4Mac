// **Die gewählten Spalten in der Kartenliste** — Werkstück 3 von Phase 3.7
// (#1168).
//
// **Eine zweite Datei**, weil `listenspalten.test.tsx` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt einer Frage: dort steht, **wie man wählt**,
// hier, **was die Liste danach zeigt**.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO, KartenlisteAntwort } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const zwei: KarteDTO[] = [
  {
    id: '1',
    nummer: 0,
    werte: [
      { feld: 'name', wert: 'Feuerball' },
      { feld: 'rarity', wert: 'common' },
    ],
  },
  {
    id: '2',
    nummer: 1,
    werte: [
      { feld: 'name', wert: 'Eisball' },
      { feld: 'rarity', wert: 'rare' },
    ],
  },
]

// aufbau baut eine Antwort mit zwei Spalten.
function aufbau(
  abweichung: Partial<KartenlisteAntwort> = {},
): KartenlisteAntwort {
  return {
    spalten: [
      {
        feld: 'name',
        aufschrift: 'Name',
        sichtbar: true,
        ausrichtung: 'links',
      },
      {
        feld: 'rarity',
        aufschrift: 'Seltenheit',
        sichtbar: false,
        ausrichtung: 'links',
      },
    ],
    zeilen: zwei.map((karte) => ({
      karteID: karte.id,
      nummer: karte.nummer + 1,
      werte: Object.fromEntries(
        karte.werte.map((wert) => [wert.feld, wert.wert]),
      ),
      marken: [],
    })),
    sortierfeld: '',
    aufsteigend: true,
    ohneVorlage: false,
    fehler: null,
    ...abweichung,
  }
}

function belegeSeite(teile: Partial<Api>) {
  return belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    ...ausKarten(() => Promise.resolve({ karten: zwei })),
    Karten: () => Promise.resolve({ karten: zwei }),
    Kartenlistenaufbau: () => Promise.resolve(aufbau()),
    ...teile,
  })
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID={undefined}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={vi.fn()}
      />
    </Sprachraum>,
  )
}

describe('Die gewählten Spalten in der Liste', () => {
  // **Eine sichtbare Spalte steht unter dem Namen.**
  it('zeigt eine sichtbare Spalte unter dem Kartennamen', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(
          aufbau({
            spalten: [
              {
                feld: 'name',
                aufschrift: 'Name',
                sichtbar: true,
                ausrichtung: 'links',
              },
              {
                feld: 'rarity',
                aufschrift: 'Seltenheit',
                sichtbar: true,
                ausrichtung: 'links',
              },
            ],
          }),
        ),
    })
    zeige()

    expect(await screen.findByText('common')).toBeInTheDocument()
    expect(screen.getByText('rare')).toBeInTheDocument()
  })

  // **Die Zeile antwortet dem Zeiger** — AP-4 (#2361), Befund B-11: bis
  // dahin stand die Kartenzeile ohne Hover da.
  it('hebt den Grund einer Kartenzeile unter dem Zeiger', async () => {
    belegeSeite({})
    zeige()

    const zeile = await screen.findByRole('row', { name: 'Feuerball' })
    expect(zeile).toHaveClass(
      'hover:bg-betont',
      'transition-colors',
      'duration-(--dauer-kurz)',
    )
  })

  // **Der Name steht nicht zweimal da** — er **ist** die Zeile.
  //
  // # Die Vorrichtung war großzügiger als der Betrieb
  //
  // Hier stand `feld: 'name'`, und `useListengestalt` filterte genau darauf.
  // **`magic.mse-game` nennt die Spalte aber `full name`** — am 2026-08-28 im
  // laufenden Fenster über `Kartenlistenaufbau` abgefragt —, und damit war der
  // Filter für jedes echte Paket **tot**: im Streifen stand
  // „Name: Fireball CC: XR CMC: 1 …", und „Name: Fireball" nahm das erste
  // Fünftel von 161 sichtbaren Punkten.
  //
  // Genau die Klasse, vor der §3 warnt. Der Fall trägt jetzt den Feldnamen des
  // echten Packs, und gefiltert wird über den **Wert**
  // (#1271).
  it('zeigt den Namen nicht ein zweites Mal', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(
          aufbau({
            spalten: [
              {
                feld: 'full name',
                aufschrift: 'Name',
                sichtbar: true,
                ausrichtung: 'links',
              },
              {
                feld: 'rarity',
                aufschrift: 'Seltenheit',
                sichtbar: true,
                ausrichtung: 'links',
              },
            ],
            zeilen: zwei.map((karte) => ({
              karteID: karte.id,
              nummer: karte.nummer + 1,
              werte: Object.fromEntries(
                karte.werte.map((wert) => [
                  // **Der echte Feldname des Packs**: `full name`, nicht
                  // `name`.
                  wert.feld === 'name' ? 'full name' : wert.feld,
                  wert.wert,
                ]),
              ),
              marken: [],
            })),
          }),
        ),
    })
    zeige()

    await screen.findByRole('button', { name: /Feuerball/ })

    expect(screen.queryByText('Name:')).not.toBeInTheDocument()

    // **Und die Spalte daneben steht sehr wohl da.** Ohne diese Hälfte wäre
    // der Lauf auch grün, wenn der Streifen ganz verschwände.
    expect(screen.getAllByText('Seltenheit:')).toHaveLength(zwei.length)
  })

  // **Ohne Aufschrift steht der Feldname da** — der Rückfall gehört ins
  // Fenster, nicht in den Dienst.
  it('nimmt den Feldnamen, wenn das Paket keine Aufschrift nennt', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(
          aufbau({
            spalten: [
              {
                feld: 'name',
                aufschrift: '',
                sichtbar: false,
                ausrichtung: 'links',
              },
              {
                feld: 'rarity',
                aufschrift: '',
                sichtbar: true,
                ausrichtung: 'links',
              },
            ],
          }),
        ),
    })
    zeige()

    // **Zwei Zeilen, zwei Aufschriften** — je Karte eine.
    expect(
      await screen.findAllByText((text) => text.trim() === 'rarity:'),
    ).toHaveLength(2)
  })

  // **Eine Karte ohne diesen Wert zeigt die Spalte gar nicht.**
  //
  // **»Seltenheit:« ohne etwas dahinter ist keine Auskunft, sondern Lärm** —
  // und der Vorschlag von `magic.mse-game` enthält Felder wie `full name`,
  // die eine Karte oft nicht führt. In der laufenden Anwendung war jede
  // Kartenzeile damit voll leerer Aufschriften.
  it('lässt eine Spalte weg, wenn die Karte das Feld nicht führt', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(
          aufbau({
            spalten: [
              {
                feld: 'rarity',
                aufschrift: 'Seltenheit',
                sichtbar: true,
                ausrichtung: 'links',
              },
            ],
            zeilen: [
              // **Die erste Karte führt das Feld nicht** …
              { karteID: '1', nummer: 1, werte: {}, marken: [] },
              {
                karteID: '2',
                nummer: 2,
                werte: { rarity: 'rare' },
                marken: [],
              },
            ],
          }),
        ),
    })
    zeige()

    expect(await screen.findByText('rare')).toBeInTheDocument()

    // **Einmal und nicht zweimal** — die Karte ohne Wert zeigt sie nicht.
    expect(
      screen.getAllByText((text) => text.trim() === 'Seltenheit:'),
    ).toHaveLength(1)
  })

  // **Und eine Karte, die der Aufbau gar nicht kennt, ebenso.**
  //
  // Liste und Gestalt sind **zwei** Ladevorgänge; wer eine Karte anlegt, hat
  // sie in der einen, bevor die andere nachgezogen ist.
  it('lässt eine Spalte weg, wenn der Aufbau die Karte nicht kennt', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(
          aufbau({
            spalten: [
              {
                feld: 'rarity',
                aufschrift: 'Seltenheit',
                sichtbar: true,
                ausrichtung: 'links',
              },
            ],
            zeilen: [
              {
                karteID: '1',
                nummer: 1,
                werte: { rarity: 'common' },
                marken: [],
              },
            ],
          }),
        ),
    })
    zeige()

    // Die zweite Karte steht in der Liste, im Aufbau aber nicht — und zeigt
    // die Spalte deshalb nicht.
    expect(await screen.findByText('common')).toBeInTheDocument()
    expect(
      screen.getAllByText((text) => text.trim() === 'Seltenheit:'),
    ).toHaveLength(1)
  })
})
