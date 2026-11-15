// **Lange Balkenlisten stehen gefaltet da** — UXR-08
// (#1916).
//
// # Der Befund
//
// Am Probebestand misst die Statistik eine Rolltiefe von 1,2 — er hat wenige
// Felder und wenige Werte. Ein echtes Set nicht: an `White.mse-set` (3.761
// Karten, aus der Setdatei gezählt am 2026-10-02) tragen gut vierzig Felder
// einen Wert, und die Verteilung über `illustrator` hat **586** Balken, über
// `name` **3.756**. Untereinander in einer Schublade bricht das R4 aus
// `docs/UIREVIEW.md` um das Vielfache.
//
// # Was gilt
//
// Zehn Balken vorn, der Rest stückweise hinter „Weitere … zeigen", und daneben
// die Zahl, wie viele es sind. Die Mengen hier sind die gemessenen, die Namen
// eigene (§11).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import type { WertzahlDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'
import { waehleZeile } from '../test/suchwahl'
import { Statistik } from './Statistik'

const texte = kataloge.de.seiten.statistik
const worte = kataloge.de.bausteine.listenansicht

// absteigend sind `anzahl` Einträge mit fallender Kartenzahl, wie die
// Abfragen sie liefern.
function absteigend(vorsatz: string, anzahl: number): WertzahlDTO[] {
  return Array.from({ length: anzahl }, (_, stelle) => ({
    name: [vorsatz, String(stelle + 1)].join(' '),
    karten: anzahl - stelle,
    kartenIDs: [],
  }))
}

// einSet ist ein Set mit 3.761 Karten und 42 belegten Feldern.
const einSet = {
  karten: 3761,
  felder: 42,
  leere: 0,
  belegung: absteigend('Feld', 42),
}

function zeige() {
  belegeApi({
    Statistik: () => Promise.resolve({ statistik: einSet }),
    Verteilung: () =>
      Promise.resolve({ verteilung: absteigend('Zeichner', 586) }),
    Auszaehldimensionen: () =>
      Promise.resolve({ dimensionen: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Statistik setID="7" />
    </Sprachraum>,
  )
}

function stand(gezeigt: number, gesamt: number) {
  return screen.getByText(
    einsetzen(worte.zahl, {
      gezeigt: String(gezeigt),
      gesamt: String(gesamt),
    }),
  )
}

afterEach(() => {
  raeumeApi()
})

describe('Eine lange Statistik', () => {
  it('zeigt von der Belegung zehn Felder und den Rest auf Verlangen', async () => {
    zeige()

    await screen.findByRole('region', { name: texte.belegung })

    expect(screen.getAllByRole('listitem')).toHaveLength(10)
    expect(stand(10, 42)).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(worte.weitere, { anzahl: '32' }),
      }),
    )

    expect(screen.getAllByRole('listitem')).toHaveLength(42)
  })

  it('zeigt von 586 Werten eines Felds zehn und dann hundert', async () => {
    zeige()

    await screen.findByRole('region', { name: texte.belegung })
    await waehleZeile(
      screen.getByRole('combobox', { name: texte.worueber }),
      /^Feld 3 /,
    )

    await waitFor(() => {
      expect(screen.getByText('Zeichner 1')).toBeInTheDocument()
    })
    expect(screen.getAllByRole('listitem')).toHaveLength(10)
    expect(stand(10, 586)).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(worte.weitere, { anzahl: '100' }),
      }),
    )

    expect(screen.getAllByRole('listitem')).toHaveLength(110)

    // **Ein anderes Feld beginnt wieder gefaltet** — die hundert gehörten
    // zum vorigen (§7, Regel 1).
    await waehleZeile(
      screen.getByRole('combobox', { name: texte.worueber }),
      /^Feld 4 /,
    )

    await waitFor(() => {
      expect(screen.getAllByRole('listitem')).toHaveLength(10)
    })
  })
})
