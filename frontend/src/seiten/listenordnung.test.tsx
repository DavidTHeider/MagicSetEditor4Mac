// **Die Ordnung der Kartenliste und ihre Ränder** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// **Eine zweite Datei**, weil `listenspalten.test.tsx` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt einer Frage: dort steht, **wie man Spalten
// wählt**, hier, **wonach die Liste ordnet** — und was an den Rändern gilt.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO, KartenlisteAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.listenspalten

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

// aufklappen öffnet den Wahlkasten.
async function aufklappen() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.ueberschrift }),
  )

  return within(screen.getByRole('region', { name: texte.ueberschrift }))
}

describe('Die Ordnung der Kartenliste', () => {
  // **Die berechnete Reihenfolge steht in der Liste** — Befund 1 des
  // Zweitblicks über Phase 3.7.
  //
  // # Warum dieser Lauf gefehlt hat
  //
  // Der Dienst sortiert `Aufbau.Zeilen`, die Bindung gibt die Reihenfolge
  // zurück — und `useListengestalt` machte daraus **nur** eine Zuordnung von
  // Werten nach Karten-ID. Die Liste bekam ihre Karten weiter aus
  // `Kartenzeilen`, und das ordnet nach `cards.reihenfolge`.
  //
  // **Kein Lauf konnte es sehen**, weil jede Vorrichtung für Karten und
  // Antwortzeilen dieselbe Reihenfolge nahm. Hier ist die Antwort **umgekehrt**
  // zu den Karten: sortiert die Liste nicht nach ihr, steht Feuerball oben.
  it('zeigt die Karten in der Reihenfolge der Antwort', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(
          aufbau({
            sortierfeld: 'name',
            zeilen: [...zwei].reverse().map((karte, stelle) => ({
              karteID: karte.id,
              nummer: stelle + 1,
              werte: Object.fromEntries(
                karte.werte.map((wert) => [wert.feld, wert.wert]),
              ),
              marken: [],
            })),
          }),
        ),
    })

    zeige()

    // **Über die Aufschriften der Zeilen**, nicht über eine Zahl: die Liste
    // zeichnet nur, was zu sehen ist, und eine Stellenzahl sagte nichts
    // darüber, was jemand liest.
    const zeilen = await screen.findAllByRole('button', { name: /ball/ })

    expect(zeilen.map((zeile) => zeile.textContent)).toEqual([
      expect.stringContaining('Eisball'),
      expect.stringContaining('Feuerball'),
    ])
  })

  // **Ohne Aufschrift steht der Feldname — im Wahlkasten wie in der Liste.**
  it('nimmt im Wahlkasten den Feldnamen, wenn keine Aufschrift dasteht', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(
          aufbau({
            // **Sichtbar** — seit UXR-05 (#1913) bietet die Ordnung nur an,
            // was die Liste zeigt (`listenspaltenfalten.test.tsx`).
            spalten: [
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

    const kasten = await aufklappen()

    expect(kasten.getByLabelText('rarity')).toBeInTheDocument()

    // **Und im Auswahlfeld ebenso.**
    expect(
      within(kasten.getByLabelText(texte.sortieren)).getByRole('option', {
        name: 'rarity',
      }),
    ).toBeInTheDocument()
  })

  // **Absteigend sagt es** — die Aufschrift des Knopfes nennt den Stand.
  it('nennt die absteigende Richtung', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(aufbau({ sortierfeld: 'name', aufsteigend: false })),
    })
    zeige()

    const kasten = await aufklappen()

    expect(
      kasten.getByRole('button', { name: texte.absteigend }),
    ).toBeInTheDocument()
  })

  it('kehrt die Richtung um', async () => {
    const setzen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(aufbau({ sortierfeld: 'name', aufsteigend: true })),
      ListensortierungSetzen: setzen,
    })
    zeige()

    const kasten = await aufklappen()
    await userEvent.click(
      kasten.getByRole('button', { name: texte.aufsteigend }),
    )

    await waitFor(() => {
      expect(setzen).toHaveBeenCalledWith('1', 'name', false)
    })
  })

  // **Ein Fehler des Ladens steht da, wo die Spalten fehlen** (§8).
  it('meldet, wenn die Gestalt nicht zu laden war', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve({
          ...aufbau(),
          spalten: [],
          fehler: {
            code: 'konflikt',
            meldung: 'Der Bestand ist belegt.',
            schluessel: '',
            feld: '',
          },
        }),
    })
    zeige()

    const kasten = await aufklappen()

    expect(await kasten.findByRole('alert')).toHaveTextContent(
      'Der Bestand ist belegt.',
    )
  })

  // **Ohne Spielpaket steht es dabei** (§8) — sonst sähe die Wahl aus, als
  // hätte das Spiel keine Aufschriften.
  it('sagt, wenn die Spalten aus den Karten kommen', async () => {
    belegeSeite({
      Kartenlistenaufbau: () => Promise.resolve(aufbau({ ohneVorlage: true })),
    })
    zeige()

    const kasten = await aufklappen()

    expect(kasten.getByText(texte.ohneVorlage)).toBeInTheDocument()
  })

  // **Mit Vorlage steht er nicht da** — die Gegenprobe.
  it('sagt es nicht, wenn das Spielpaket dasteht', async () => {
    belegeSeite({})
    zeige()

    const kasten = await aufklappen()

    expect(kasten.queryByText(texte.ohneVorlage)).not.toBeInTheDocument()
  })

  // **Ein Fehler beim Schreiben steht in der Handlungszeile** — und die Wahl
  // bleibt stehen.
  it('meldet einen Schreibfehler', async () => {
    belegeSeite({
      ListenspaltenSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'konflikt',
            meldung: 'Das Set ist belegt.',
            schluessel: '',
            feld: '',
          },
        }),
    })
    zeige()

    const kasten = await aufklappen()
    await userEvent.click(kasten.getByLabelText('Seltenheit'))

    expect(await kasten.findByRole('alert')).toHaveTextContent(
      'Das Set ist belegt.',
    )
  })
})

// **Die Ordnung gilt für die ganze Lasche, nicht nur für die gezeichnete
// Liste** — Befund 3 des Nachlaufs über Phase 3.7.
//
// # Warum das mehr ist als die Anzeige
//
// Der erste Anlauf stellte die Karten erst in `Kartenspalte` um, also im
// letzten Kind. Damit war die Ordnung eine **Darstellung**: die zuerst
// aufgeschlagene Karte, die Nachfolgerin nach dem Löschen und die
// Reihenfolge der Zwischenablage kamen weiter aus `cards.reihenfolge`.
//
// MSE2 führt dafür **eine** `sorted_list`, aus der Auswahl, Vorher/Nachher,
// Löschnachfolge und Kopierreihenfolge lesen (`src/gui/control/item_list.cpp`,
// Zeilen 37–64 und 87–100, Fassung `f9d9356` vom 2021-07-06).
describe('Die Ordnung gilt für die ganze Lasche', () => {
  // **Die erste aufgeschlagene Karte ist die erste der Ordnung.**
  //
  // Ohne Wahl schlägt die Lasche die erste Karte auf. Kommt die aus der
  // Datei statt aus der Ordnung, steht der Editor auf einer anderen Karte,
  // als die Liste oben zeigt.
  it('schlägt die erste Karte der Ordnung auf', async () => {
    belegeSeite({
      Kartenlistenaufbau: () =>
        Promise.resolve(
          aufbau({
            sortierfeld: 'name',
            zeilen: [...zwei].reverse().map((karte, stelle) => ({
              karteID: karte.id,
              nummer: stelle + 1,
              werte: Object.fromEntries(
                karte.werte.map((wert) => [wert.feld, wert.wert]),
              ),
              marken: [],
            })),
          }),
        ),
    })

    zeige()

    // **Über das Namensfeld des Editors** — es zeigt, welche Karte offen ist.
    // Die Liste allein sagte es nicht: sie zeigt beide.
    expect(await screen.findByDisplayValue('Eisball')).toBeInTheDocument()
  })
})
