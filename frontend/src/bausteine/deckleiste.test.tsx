// **Die Kennzahlen des offenen Decks** — AP-9
// (#2366),
// Entwurf `Bereiche09`: Hauptdeck mit Sideboard, Prüfung mit Format, was in
// der Sammlung fehlt — jede Zahl mit Ziel.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import type { HabenstandAntwort } from '../apidecks'
import type { Deckpruefung } from '../haken/useDeckpruefung'
import type { Ladezustand } from '../haken/useLaden'
import { deckprobe, urteilprobe } from '../seiten/deckprobe'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Deckleiste } from './Deckleiste'

const texte = kataloge.de.seiten.decks

function pruefung(teil: Partial<Deckpruefung> = {}): Deckpruefung {
  return {
    urteil: urteilprobe(),
    befunde: [],
    laedt: false,
    fehler: undefined,
    nachladen: vi.fn(),
    ohneFormat: false,
    unzulaessig: () => false,
    ...teil,
  }
}

function abgleich(
  teil: Partial<Ladezustand<HabenstandAntwort | null>> = {},
): Ladezustand<HabenstandAntwort | null> {
  return {
    daten: { zeilen: [], fehlend: 4, unbekannt: 0, fehler: null },
    laedt: false,
    fehler: undefined,
    nachladen: vi.fn(),
    ...teil,
  }
}

function zeige(
  p: Deckpruefung,
  a: Ladezustand<HabenstandAntwort | null>,
  ziele = { zeigePlaetze: vi.fn(), zeigePruefung: vi.fn() },
) {
  render(
    <Sprachraum sprache="de">
      <Deckleiste
        deck={deckprobe({ hauptdeck: 60, sideboard: 15 })}
        pruefung={p}
        abgleich={a}
        {...ziele}
      />
    </Sprachraum>,
  )

  return ziele
}

const leiste = () =>
  within(screen.getByRole('toolbar', { name: texte.leiste.bezeichnung }))

describe('die Leiste des Decks', () => {
  it('nennt jede Zahl mit ihrem Ziel', async () => {
    const ziele = zeige(pruefung(), abgleich())

    await userEvent.click(
      leiste().getByRole('button', { name: 'Hauptdeck 60 Sideboard 15' }),
    )
    expect(ziele.zeigePlaetze).toHaveBeenCalledTimes(1)

    const urteil = urteilprobe()
    await userEvent.click(
      leiste().getByRole('button', {
        name: `${texte.leiste.pruefung} ${texte.legal} ${urteil.formatname}`,
      }),
    )
    await userEvent.click(
      leiste().getByRole('button', { name: `${texte.leiste.fehlt} 4` }),
    )
    expect(ziele.zeigePruefung).toHaveBeenCalledTimes(2)
  })

  it('sagt „nicht zulässig“, und unbekannte Karten zählen für sich', () => {
    zeige(
      pruefung({ urteil: { ...urteilprobe(), zulaessig: false } }),
      abgleich({
        daten: { zeilen: [], fehlend: 0, unbekannt: 3, fehler: null },
      }),
    )

    expect(leiste().getByText(texte.postenUnzulaessig)).toBeInTheDocument()
    expect(
      leiste().getByRole('button', {
        name: `${texte.leiste.fehlt} 0 3 unbekannt`,
      }),
    ).toBeInTheDocument()
  })

  // **Ungeprüft ist kein „nicht zulässig“** (Vermerk 0061): ohne Format und
  // nach einem Fehler steht ein Satz, keine rote Behauptung.
  it('sagt ohne Format und nach einem Fehler, dass nicht geprüft ist', () => {
    zeige(
      pruefung({ urteil: null, ohneFormat: true }),
      abgleich({ daten: null, fehler: 'kaputt' }),
    )
    expect(leiste().getByText(texte.leiste.ohneFormat)).toBeInTheDocument()
    expect(
      leiste().getByText(texte.leiste.nichtAbgeglichen),
    ).toBeInTheDocument()
  })

  it('sagt nach einem Fehler der Prüfung „nicht geprüft“', () => {
    zeige(pruefung({ urteil: null, fehler: 'kaputt' }), abgleich())
    expect(leiste().getByText(texte.leiste.nichtGeprueft)).toBeInTheDocument()
  })

  it('steht beim Laden in Endhöhe', () => {
    zeige(
      pruefung({ urteil: null, laedt: true }),
      abgleich({ daten: undefined, laedt: true }),
    )
    expect(
      screen.getByRole('toolbar', { name: texte.leiste.bezeichnung }),
    ).toHaveAttribute('aria-busy', 'true')
  })
})
