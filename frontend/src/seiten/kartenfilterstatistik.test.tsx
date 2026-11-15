// **Die Kartenliste, wenn die Statistik sie filtert** — UXR-F6
// (#2013).
//
// MSE2 zeigt unter dem Diagramm nur die Karten der gewählten Gruppe
// (`StatsFilter`, `src/gui/set/stats_panel.cpp` Zeilen 536–545, Fassung
// f9d9356 vom 2021-07-06). Hier steht die Liste auf den Karten, weit vom
// Diagramm — also gilt zweierlei: **die Liste zeigt genau die Karten des
// Balkens**, und **eine Zeile sagt, warum nicht alle dastehen**, mit den zwei
// Wegen daraus.
//
// Der zweite Teil geht durch den ganzen Arbeitsplatz: gewählt wird in der
// Schublade, gesehen auf den Karten, und der Rückweg führt dorthin zurück.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { aufheben, nichtsAufgehoben } from '../haken/useAufgehobeneSeiten'
import type { Setschluessel } from '../navigation'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import {
  Statistikfilterraum,
  useStatistikfilter,
  type Gruppenwahl,
} from '../statistikfilter'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'
import { Setbereich } from './Setbereich'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.statistik
const treffer = kataloge.de.seiten.karten.schnellfilter.treffer

const drei: KarteDTO[] = [
  { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Feuerball' }] },
  { id: '2', nummer: 1, werte: [{ feld: 'name', wert: 'Eisball' }] },
  { id: '3', nummer: 2, werte: [{ feld: 'name', wert: 'Blitz' }] },
]

// seltene sind die Karten 2 und 3 — nicht die erste.
const seltene: Gruppenwahl = {
  wahl: 'feld:rarity',
  quelle: 'rarity',
  gruppe: 'rare',
  kartenIDs: ['2', '3'],
}

function belegeKarten(zusatz = {}) {
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    ...ausKarten(() => Promise.resolve({ karten: drei })),
    Karten: () => Promise.resolve({ karten: drei }),
    Kartenlistenaufbau: () =>
      Promise.resolve({
        spalten: [],
        zeilen: drei.map((karte, stelle) => ({
          karteID: karte.id,
          nummer: stelle + 1,
          werte: { name: karte.werte[0]?.wert ?? '' },
          marken: [],
        })),
        sortierfeld: '',
        aufsteigend: true,
        ohneVorlage: false,
        fehler: null,
      }),
    ...zusatz,
  })
}

// Setzer legt eine Wahl in den Raum — so, wie die Statistik es täte.
function Setzer({ wahl }: { wahl: Gruppenwahl }) {
  const filter = useStatistikfilter()

  return (
    <button
      type="button"
      onClick={() => {
        filter.waehlen(wahl)
      }}
    >
      wählen
    </button>
  )
}

function seite(wechseln = vi.fn(), wahl = seltene): ReactNode {
  return (
    <Sprachraum sprache="de">
      <Statistikfilterraum wechseln={wechseln}>
        <Setzer wahl={wahl} />
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={vi.fn()}
          symbolZeichnen={vi.fn()}
        />
      </Statistikfilterraum>
    </Sprachraum>
  )
}

// namen sind die Karten, die die Liste zeigt — über die Aufschriften der
// Zeilen, wie in `listenordnung.test.tsx`.
async function namen() {
  const zeilen = await screen.findAllByRole('button', { name: /ball|Blitz/ })

  return zeilen.map((zeile) => zeile.textContent)
}

describe('Die Kartenliste unter dem Filter der Statistik', () => {
  it('zeigt ohne Wahl alle Karten und keine Zeile', async () => {
    belegeKarten()
    render(seite())

    expect(await namen()).toHaveLength(3)
    expect(
      screen.queryByRole('group', { name: texte.filterBezeichnung }),
    ).not.toBeInTheDocument()
  })

  it('zeigt nur die Karten des Balkens und sagt, warum', async () => {
    belegeKarten()
    render(seite())

    await namen()
    await userEvent.click(screen.getByRole('button', { name: 'wählen' }))

    await waitFor(async () => {
      expect(await namen()).toEqual([
        expect.stringContaining('Eisball'),
        expect.stringContaining('Blitz'),
      ])
    })

    const zeile = screen.getByRole('group', { name: texte.filterBezeichnung })
    expect(zeile).toHaveTextContent(
      einsetzen(texte.filterKopf, { quelle: 'rarity', gruppe: 'rare' }),
    )

    // **Die Zahl kommt aus der Liste**: zwei von drei.
    expect(
      within(zeile).getByText(
        einsetzen(treffer, { gezeigt: '2', gesamt: '3' }),
      ),
    ).toBeInTheDocument()
  })

  it('hebt den Filter auf Verlangen auf', async () => {
    belegeKarten()
    render(seite())

    await namen()
    await userEvent.click(screen.getByRole('button', { name: 'wählen' }))
    await userEvent.click(
      await screen.findByRole('button', { name: texte.filterAufheben }),
    )

    await waitFor(async () => {
      expect(await namen()).toHaveLength(3)
    })
    expect(
      screen.queryByRole('group', { name: texte.filterBezeichnung }),
    ).not.toBeInTheDocument()
  })

  // **Der Rückweg**: die Wahl bleibt, die Schublade geht auf.
  it('führt zur Statistik zurück, ohne die Wahl zu lösen', async () => {
    const wechseln = vi.fn()

    belegeKarten()
    render(seite(wechseln))

    await namen()
    await userEvent.click(screen.getByRole('button', { name: 'wählen' }))
    await userEvent.click(
      await screen.findByRole('button', { name: texte.filterZurStatistik }),
    )

    expect(wechseln).toHaveBeenLastCalledWith('statistik')
    expect(
      screen.getByRole('group', { name: texte.filterBezeichnung }),
    ).toBeInTheDocument()
  })

  // **Eine Wahl mit Karten, die dieses Set nicht kennt, zeigt nichts** und
  // stürzt nicht ab — die Liste ist eine Schnittmenge.
  it('zeigt keine Karte, wenn keine der Wahl in der Liste steht', async () => {
    belegeKarten()
    render(seite(vi.fn(), { ...seltene, kartenIDs: ['99'] }))

    await namen()
    await userEvent.click(screen.getByRole('button', { name: 'wählen' }))

    expect(
      await screen.findByText(
        einsetzen(treffer, { gezeigt: '0', gesamt: '3' }),
      ),
    ).toBeInTheDocument()
  })
})

// **Durch den ganzen Arbeitsplatz**: gewählt wird in der Schublade, gesehen
// wird auf den Karten.
describe('Die Wahl durch den Arbeitsplatz', () => {
  const rare = {
    name: 'rare',
    karten: 2,
    kartenIDs: ['2', '3'],
  }

  function bereich(gang: readonly Setschluessel[], wechseln = vi.fn()) {
    const stand = gang.reduce(
      (bisher, ziel) => aufheben(bisher, ziel, '7'),
      nichtsAufgehoben,
    )

    return (
      <Sprachraum sprache="de">
        <Setbereich
          seite={gang[gang.length - 1] ?? 'karten'}
          aufgehoben={stand}
          setID="7"
          wechseln={wechseln}
          karteID={undefined}
          oeffneKarte={vi.fn()}
          oeffneSymbol={vi.fn()}
          bestandGeaendert={vi.fn()}
          verlaufsstand={0}
          verlaufsfehler={undefined}
        />
      </Sprachraum>
    )
  }

  function belegeAlles() {
    belegeKarten({
      Sets: () => Promise.resolve({ sets: [] }),
      Statistik: () =>
        Promise.resolve({
          statistik: {
            karten: 3,
            felder: 1,
            leere: 0,
            belegung: [{ name: 'rarity', karten: 3 }],
          },
        }),
      Verteilung: () =>
        Promise.resolve({
          verteilung: [{ name: 'common', karten: 1, kartenIDs: ['1'] }, rare],
        }),
      Auszaehldimensionen: () =>
        Promise.resolve({ dimensionen: [], fehler: null }),
    })
  }

  it('filtert die Karten hinter der Schublade und nach dem Schließen', async () => {
    const wechseln = vi.fn()

    belegeAlles()

    const { rerender } = render(bereich(['karten', 'statistik'], wechseln))

    await userEvent.selectOptions(
      await screen.findByRole('combobox', { name: texte.worueber }),
      'feld:rarity',
    )
    await userEvent.click(await screen.findByRole('button', { name: /rare/ }))

    // Der Knopf in der Schublade ist der Weg zur Liste.
    await userEvent.click(
      screen.getByRole('button', { name: texte.gruppeZeigen }),
    )
    expect(wechseln).toHaveBeenLastCalledWith('karten')

    // Die Schublade schließt sich, wie der Arbeitsplatz es tut.
    rerender(bereich(['karten', 'statistik', 'karten'], wechseln))

    const zeile = await screen.findByRole('group', {
      name: texte.filterBezeichnung,
    })
    expect(zeile).toHaveTextContent(
      einsetzen(texte.filterKopf, { quelle: 'rarity', gruppe: 'rare' }),
    )
    expect(
      within(zeile).getByText(
        einsetzen(treffer, { gezeigt: '2', gesamt: '3' }),
      ),
    ).toBeInTheDocument()

    // Und der Rückweg öffnet die Statistik, in der der Balken gewählt bleibt.
    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.filterZurStatistik }),
    )
    expect(wechseln).toHaveBeenLastCalledWith('statistik')

    rerender(bereich(['karten', 'statistik', 'karten', 'statistik'], wechseln))

    expect(
      await screen.findByRole('button', { name: /rare/, pressed: true }),
    ).toBeInTheDocument()
  })

  // **Ein anderes Set erbt keine Wahl** (§7, Regel 1): `Seiteninhalt` baut den
  // Bereich mit dem Set neu, und der Raum geht mit.
  it('trägt in einem neuen Bereich keine Wahl des alten', async () => {
    belegeAlles()

    const { rerender } = render(bereich(['karten', 'statistik']))

    await userEvent.selectOptions(
      await screen.findByRole('combobox', { name: texte.worueber }),
      'feld:rarity',
    )
    await userEvent.click(await screen.findByRole('button', { name: /rare/ }))

    rerender(
      <Sprachraum sprache="de">
        <Setbereich
          key="8"
          seite="karten"
          aufgehoben={aufheben(nichtsAufgehoben, 'karten', '8')}
          setID="8"
          wechseln={vi.fn()}
          karteID={undefined}
          oeffneKarte={vi.fn()}
          oeffneSymbol={vi.fn()}
          bestandGeaendert={vi.fn()}
          verlaufsstand={0}
          verlaufsfehler={undefined}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(
        screen.queryByRole('group', { name: texte.filterBezeichnung }),
      ).not.toBeInTheDocument()
    })
  })
})
