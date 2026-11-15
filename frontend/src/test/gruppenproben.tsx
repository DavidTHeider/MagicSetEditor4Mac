// Proben für die Wahl eines Balkens — UXR-F6
// (#2013).
//
// Sie stehen hier und nicht in jeder Testdatei, weil zwei Dateien dieselbe
// Seite in demselben Raum prüfen: `gruppenwahl.test.tsx` die Wahl selbst,
// `gruppenabgleich.test.tsx`, was nach dem Neuladen aus ihr wird. Zweimal
// dieselbe Vorrichtung von Hand wäre zweimal eine, die auseinanderlaufen kann
// (§3).
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Activity, type ReactNode } from 'react'
import { vi } from 'vitest'

import type { WertzahlDTO } from '../api'
import { Statistik } from '../seiten/Statistik'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Statistikfilterraum, useStatistikfilter } from '../statistikfilter'
import { belegeApi } from './api'

export const texte = kataloge.de.seiten.statistik

const kennzahlen = () =>
  Promise.resolve({
    statistik: {
      karten: 7,
      felder: 2,
      leere: 0,
      belegung: [
        { name: 'name', karten: 7 },
        { name: 'rarity', karten: 5 },
      ],
    },
  })

// seltenheiten: drei gewöhnliche, zwei seltene. **Die Kennungen sind nicht 1
// bis 5**, damit ein Test sieht, dass die des Balkens ankommen und nicht die
// Reihenfolge der Zeilen.
export const seltenheiten: WertzahlDTO[] = [
  { name: 'common', karten: 3, kartenIDs: ['1', '2', '5'] },
  { name: 'rare', karten: 2, kartenIDs: ['3', '7'] },
]

const dimensionen = [
  { name: 'color category', beschreibung: '', numerisch: false },
]

// farben sind zwei Gruppen einer Auszählung, eine davon ohne Karten.
export const farben = [
  { name: 'White', karten: 2, farbe: '', kartenIDs: ['1', '3'] },
  { name: 'Land', karten: 0, farbe: '', kartenIDs: [] },
]

// auszaehlung ist die Antwort auf eine Auszählung mit diesen Gruppen.
export function auszaehlung(gruppen: typeof farben) {
  return {
    dimension: 'color category',
    beschreibung: '',
    numerisch: false,
    gruppen,
    befunde: [],
    fehler: null,
  }
}

// Lesepult zeigt, was die Kartenliste aus dem Raum bekäme.
function Lesepult() {
  const filter = useStatistikfilter()

  return (
    <p data-testid="filter">
      {filter.wahl === undefined
        ? 'keiner'
        : `${filter.wahl.wahl} | ${filter.wahl.quelle} | ${filter.wahl.gruppe} | ${filter.wahl.kartenIDs.join(',')}`}
    </p>
  )
}

// belege belegt die Seite; `teile` überschreibt einzelne Wege.
export function belege(teile = {}) {
  belegeApi({
    Statistik: kennzahlen,
    Verteilung: () => Promise.resolve({ verteilung: seltenheiten }),
    Auszaehldimensionen: () => Promise.resolve({ dimensionen, fehler: null }),
    Auszaehlen: () => Promise.resolve(auszaehlung(farben)),
    ...teile,
  })
}

// seite zeichnet die Statistik in ihrem Raum, neben dem Leser — sichtbar oder
// verborgen, wie die Schublade es täte.
export function seite(wechseln = vi.fn(), sichtbar = true): ReactNode {
  return (
    <Sprachraum sprache="de">
      <Statistikfilterraum wechseln={wechseln}>
        <Activity mode={sichtbar ? 'visible' : 'hidden'}>
          <Statistik setID="1" />
        </Activity>
        <Lesepult />
      </Statistikfilterraum>
    </Sprachraum>
  )
}

export const filter = () => screen.getByTestId('filter')

// waehleBild wählt den Eintrag „Worüber".
export async function waehleBild(wert: string) {
  await userEvent.selectOptions(
    await screen.findByRole('combobox', { name: texte.worueber }),
    wert,
  )
}

// balken ist der Knopf eines Balkens.
export function balken(name: string) {
  return screen.findByRole('button', { name: new RegExp(name) })
}
