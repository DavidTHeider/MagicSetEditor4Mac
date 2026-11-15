// **Notiz und Marken stehen eingeklappt beieinander**
// (#1271).
//
// # Warum das eine Zusage ist und nicht nur eine Anordnung
//
// Gemessen am 2026-08-28 bei 1280 × 800: die Listenspalte war auf 70vh
// gedeckelt — 560 px —, und darunter standen Notizkasten (150), Marken (150)
// und Spaltenwahl (42) **offen**. Für die Liste blieben rund 180 px, also
// **vier Zeilen** von 103 Karten.
//
// Die Liste ist das, womit jemand auf dieser Lasche anfängt. Vier Zeilen sind
// kein Anfang.
//
// # Was dieser Lauf nicht sieht
//
// **Dass der Inhalt verborgen ist.** jsdom setzt `<details>` nicht um: der
// Text steht im DOM, ob der Ausklapper offen ist oder nicht. Geprüft wird
// deshalb das `open`-Merkmal — und die Höhe, die daran hängt, sieht nur die
// Sichtprüfung (§3).
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'

const texte = kataloge.de.seiten.karten

afterEach(() => {
  raeumeApi()
})

function zeige() {
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Kartenzeilen: () =>
      Promise.resolve({
        zeilen: [{ id: '5', nummer: 0, name: 'Feuerball', felder: 1 }],
      }),
    Karte: () =>
      Promise.resolve({
        karte: {
          id: '5',
          nummer: 0,
          werte: [{ feld: 'name', wert: 'Feuerball' }],
        },
        fehler: null,
      }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null, fehler: null }),
    Kartenbilder: () => Promise.resolve({ bilder: {}, fehler: null }),
    KarteVerweis: () =>
      Promise.resolve({ verweis: null, befunde: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID="5"
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

describe('der Arbeitsstand einer Karte', () => {
  it('steht eingeklappt unter der Liste', async () => {
    zeige()

    const ausklapper = (await screen.findByText(texte.arbeitsstand)).closest(
      'details',
    )

    expect(ausklapper).not.toBeNull()
    expect(ausklapper).not.toHaveAttribute('open')
  })

  // **Und er trägt beides.** Ohne diese Hälfte wäre der Lauf auch grün, wenn
  // der Ausklapper leer wäre — und die Marken wären aus der Anwendung
  // verschwunden, ohne dass jemand es merkt.
  it('trägt Notiz und Marken', async () => {
    zeige()

    const ausklapper = (await screen.findByText(texte.arbeitsstand)).closest(
      'details',
    )

    expect(ausklapper).not.toBeNull()

    const notizen = await screen.findByLabelText(texte.notizen.ueberschrift)
    expect(ausklapper).toContainElement(notizen)

    const marken = await screen.findByRole('region', {
      name: texte.marken.ueberschrift,
    })
    expect(ausklapper).toContainElement(marken)
  })
})
