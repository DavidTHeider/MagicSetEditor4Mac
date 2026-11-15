// **Die Spalten der Kartenliste und ihre Ordnung** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// **Gegen die ganze Kartenseite und nicht gegen den Kasten**: was hier zu
// prüfen ist, ist die Naht zwischen Wahl, Liste und Bestand — und die verläuft
// durch `Karten.tsx`.
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

describe('Die Spalten der Kartenliste', () => {
  // **Zugeklappt steht sie nicht im Weg** — MSE2 legt sie hinter einen Dialog.
  it('beginnt zugeklappt', async () => {
    belegeSeite({})
    zeige()

    const knopf = await screen.findByRole('button', {
      name: texte.ueberschrift,
    })
    expect(knopf).toHaveAttribute('aria-expanded', 'false')

    expect(screen.queryByLabelText('Seltenheit')).not.toBeInTheDocument()
  })

  // **Aufgeklappt steht jede angebotene Spalte da — auch die unsichtbare.**
  it('zeigt jede angebotene Spalte', async () => {
    belegeSeite({})
    zeige()

    const kasten = await aufklappen()

    expect(kasten.getByLabelText('Name')).toBeChecked()
    expect(kasten.getByLabelText('Seltenheit')).not.toBeChecked()
  })

  // **Ein Haken schreibt die ganze Wahl** — nicht eine Spalte.
  it('schreibt die ganze Wahl, wenn eine Spalte umgeschaltet wird', async () => {
    const setzen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeSeite({ ListenspaltenSetzen: setzen })
    zeige()

    const kasten = await aufklappen()
    await userEvent.click(kasten.getByLabelText('Seltenheit'))

    await waitFor(() => {
      expect(setzen).toHaveBeenCalledWith('1', [
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
      ])
    })
  })

  // **Die Reihenfolge der Liste ist die Reihenfolge der Spalten.**
  it('tauscht zwei Spalten über die Pfeilknöpfe', async () => {
    const gesendet: string[][] = []
    const setzen = vi.fn((_setID: string, spalten: { feld: string }[]) => {
      gesendet.push(spalten.map((spalte) => spalte.feld))

      return Promise.resolve({ fehler: null })
    })

    belegeSeite({ ListenspaltenSetzen: setzen })
    zeige()

    const kasten = await aufklappen()

    await userEvent.click(
      kasten.getByRole('button', { name: `rarity — ${texte.nachOben}` }),
    )

    await waitFor(() => {
      expect(gesendet).toEqual([['rarity', 'name']])
    })

    // **Und in die andere Richtung ebenso** — zwei Knöpfe, zwei Wege.
    await userEvent.click(
      kasten.getByRole('button', { name: `name — ${texte.nachUnten}` }),
    )

    await waitFor(() => {
      expect(gesendet).toEqual([
        ['rarity', 'name'],
        ['rarity', 'name'],
      ])
    })
  })

  // **Ein Filter, weil es 243 sein können.**
  //
  // Gemessen in der laufenden Anwendung: `magic.mse-game` bietet 243 Spalten
  // an, und der aufgeklappte Kasten war 10.361 Punkte hoch in einer Spalte von
  // 504. Wer »rarity« sucht, soll nicht an 200 Namen vorbeirollen.
  it('filtert die Spalten über Aufschrift und Feldnamen', async () => {
    belegeSeite({})
    zeige()

    const kasten = await aufklappen()

    expect(
      kasten.getByText(
        texte.gezeigt.replace('{gezeigt}', '2').replace('{alle}', '2'),
      ),
    ).toBeInTheDocument()

    // **Über die Aufschrift.**
    await userEvent.type(kasten.getByLabelText(texte.filtern), 'Selten')

    expect(kasten.queryByLabelText('Name')).not.toBeInTheDocument()
    expect(kasten.getByLabelText('Seltenheit')).toBeInTheDocument()

    expect(
      kasten.getByText(
        texte.gezeigt.replace('{gezeigt}', '1').replace('{alle}', '2'),
      ),
    ).toBeInTheDocument()

    // **Und über den Feldnamen** — das Paket beschriftet `casting cost` mit
    // »Cost«, und wer das eine sucht, soll das andere finden.
    await userEvent.clear(kasten.getByLabelText(texte.filtern))
    await userEvent.type(kasten.getByLabelText(texte.filtern), 'rarity')

    expect(kasten.getByLabelText('Seltenheit')).toBeInTheDocument()
  })

  // **Die Stelle kommt aus der ganzen Liste und nicht aus der gefilterten.**
  //
  // Sonst schöbe ein Pfeil die Spalte an den Platz ihres Nachbarn **im
  // Filter** — und das wäre irgendwo.
  it('schiebt auch im Filter an die richtige Stelle', async () => {
    const gesendet: string[][] = []
    const setzen = vi.fn((_setID: string, spalten: { feld: string }[]) => {
      gesendet.push(spalten.map((spalte) => spalte.feld))

      return Promise.resolve({ fehler: null })
    })

    belegeSeite({ ListenspaltenSetzen: setzen })
    zeige()

    const kasten = await aufklappen()
    await userEvent.type(kasten.getByLabelText(texte.filtern), 'Selten')

    await userEvent.click(
      kasten.getByRole('button', { name: `rarity — ${texte.nachOben}` }),
    )

    await waitFor(() => {
      expect(gesendet).toEqual([['rarity', 'name']])
    })
  })

  // **Die Knöpfe an den Rändern sind angehalten** — sie hätten nichts zu tun.
  it('hält die Pfeile an den Rändern an', async () => {
    belegeSeite({})
    zeige()

    const kasten = await aufklappen()

    expect(
      kasten.getByRole('button', { name: `name — ${texte.nachOben}` }),
    ).toBeDisabled()

    expect(
      kasten.getByRole('button', { name: `rarity — ${texte.nachUnten}` }),
    ).toBeDisabled()
  })

  // **Die Ordnung wird gesetzt, und „gar nicht" ist eine Wahl.**
  it('setzt die Ordnung und hebt sie wieder auf', async () => {
    const setzen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeSeite({ ListensortierungSetzen: setzen })
    zeige()

    const kasten = await aufklappen()

    // **Nach `name` und nicht mehr nach `rarity`**: seit UXR-05 (#1913)
    // bietet die Ordnung nur an, was die Liste zeigt — wie MSE2, dessen
    // Ordnung auf eine sichtbare Spalte zeigt. `rarity` ist hier ausgeblendet
    // (`listenspaltenfalten.test.tsx` hält den Fall).
    await userEvent.selectOptions(
      kasten.getByLabelText(texte.sortieren),
      'name',
    )

    await waitFor(() => {
      expect(setzen).toHaveBeenCalledWith('1', 'name', true)
    })
  })

  // **Der Richtungsknopf steht erst da, wenn sortiert wird** — ohne Ordnung
  // gäbe es keine Richtung.
  it('bietet die Richtung erst mit einer Ordnung an', async () => {
    belegeSeite({})
    zeige()

    const kasten = await aufklappen()

    expect(
      kasten.queryByRole('button', { name: texte.aufsteigend }),
    ).not.toBeInTheDocument()
  })
})
