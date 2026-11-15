// **Die Liste zählt Stellen, nicht Schlüssel**
// (#666).
import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenzeileDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { Kartenliste } from './Kartenliste'

function karte(id: string, nummer: number, name: string): KartenzeileDTO {
  return { id, nummer, name, felder: 1 }
}

function zeige(karten: readonly KartenzeileDTO[]) {
  render(
    <Sprachraum sprache="de">
      <Kartenliste
        spalten={[]}
        marken={{}}
        setID="7"
        stand={0}
        karten={karten}
        gewaehlt={undefined}
        waehlen={vi.fn()}
        angehakt={[]}
        anhaken={vi.fn()}
        mitKaestchen={false}
      />
    </Sprachraum>,
  )
}

describe('Die Kartenliste', () => {
  it('zählt von eins an', () => {
    zeige([karte('1', 0, 'Erste'), karte('2', 1, 'Zweite')])

    expect(screen.getAllByText(/^[0-9]+$/).map((s) => s.textContent)).toEqual([
      '1',
      '2',
    ])
  })

  // **Eine Lücke im Schlüssel ändert die Zählung nicht.**
  //
  // `cards.reihenfolge` ist ein Sortierschlüssel und keine dichte Zählung: ein
  // Löschen lässt die Lücke stehen, damit die Rücknahme dieselbe Stelle
  // wiederbekommt (`internal/storage/karteloeschen.go`). Angezeigt wird die
  // **Stelle** — in MSE2 gibt es gar keine gespeicherte Nummer, `Set::cards`
  // ist ein `vector` und die Anzeige sein Index.
  //
  // Vor dieser Zusicherung zeigte ein Set mit zwei Karten sie als 2 und 3.
  it('zählt auch dann dicht, wenn im Schlüssel eine Lücke steht', () => {
    zeige([karte('8', 1, 'Feuerball'), karte('9', 2, 'Ohne Namen')])

    expect(screen.getAllByText(/^[0-9]+$/).map((s) => s.textContent)).toEqual([
      '1',
      '2',
    ])
  })
})

// **Sie zeichnet nur, was zu sehen ist**
// (#891).
//
// Unter jsdom hat kein Element eine Höhe; ohne gesetzte Maße zeichnet die
// Liste alles, und das ist der Zweig, den jede andere Prüfung nimmt. Hier
// stehen sie, damit die **Abstandhalter** einen Test bekommen — ohne sie wäre
// die Liste so hoch wie ihr Fenster, und niemand käme über die zwanzigste
// Karte hinaus.
describe('Die Kartenliste unter echten Maßen', () => {
  // 486 Punkte Sicht und 41 je Zeile — gemessen an der laufenden Anwendung am
  // 2026-08-25, bei 1.440 Punkten Fensterbreite.
  const sicht = 486
  const zeilenhoehe = 41

  function messbarMachen() {
    // **`HTMLElement.prototype` und nicht der einzelne Knoten**: die Liste
    // hängt ihre Maßreferenz an die **erste gezeichnete** Zeile, und die
    // wechselt beim Rollen. Ein Test, der einen Knoten von Hand belegt, träfe
    // nach dem ersten Rollen den falschen.
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
      height: zeilenhoehe,
    } as DOMRect)

    vi.spyOn(HTMLElement.prototype, 'clientHeight', 'get').mockReturnValue(
      sicht,
    )
  }

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('zeichnet einen Bruchteil und trägt den Rest als Abstand', () => {
    messbarMachen()

    const viele = Array.from({ length: 3788 }, (_, stelle) =>
      karte(String(stelle), stelle, `Karte ${String(stelle)}`),
    )

    zeige(viele)

    // ceil(486/41) = 12 sichtbar, plus 8 Überhang je Seite: 28 gezeichnet.
    const zeilen = screen.getAllByRole('button')

    expect(zeilen).toHaveLength(28)

    // **Und der Rest steht als Höhe da.** `3788 - 28` Zeilen à 41 Punkte —
    // ohne diesen Abstandhalter hätte die Liste keinen Rollbalken.
    const abstand = document.querySelector('[role="presentation"]')

    expect(abstand).toHaveStyle({ height: `${String((3788 - 28) * 41)}px` })
  })

  // **Und nach dem Rollen steht auch oben einer.** Er hält die gezeichneten
  // Zeilen an ihrer Stelle; ohne ihn sprängen sie beim ersten Rollen an den
  // Anfang der Liste.
  it('trägt nach dem Rollen oben und unten einen Abstand', async () => {
    messbarMachen()

    const viele = Array.from({ length: 3788 }, (_, stelle) =>
      karte(String(stelle), stelle, `Karte ${String(stelle)}`),
    )

    zeige(viele)

    const behaelter = screen.getByRole('grid')

    behaelter.scrollTop = zeilenhoehe * 100
    fireEvent.scroll(behaelter)

    // erste = 100 - 8 = 92, letzte = 92 + 12 + 16 = 120
    await screen.findByRole('button', { name: /^Karte 92 —/ })

    // **Die Abstandhalter tragen `role="presentation"`** (#1290): ein Kind
    // eines Gitters müsste sonst eine Zeile sein, und sie sind keine.
    const abstaende = [
      ...behaelter.querySelectorAll(':scope > [role="presentation"]'),
    ]

    expect(abstaende).toHaveLength(2)
    expect(abstaende[0]).toHaveStyle({ height: `${String(92 * 41)}px` })
    expect(abstaende[1]).toHaveStyle({
      height: `${String((3788 - 120) * 41)}px`,
    })
  })

  // **Die wahre Stelle steht an jeder Zeile, die wahre Zahl am Gitter** —
  // sonst hörte eine Vorlesehilfe „Eintrag 3 von 28", während es 3.788 sind.
  //
  // Seit #1290 ist die Liste ein `grid`: die Stelle heißt `aria-rowindex`, und
  // `aria-rowcount` steht **einmal** oben statt an jeder Zeile.
  it('nennt jeder Zeile ihre Stelle in der ganzen Liste', () => {
    messbarMachen()

    const viele = Array.from({ length: 3788 }, (_, stelle) =>
      karte(String(stelle), stelle, `Karte ${String(stelle)}`),
    )

    zeige(viele)

    const erste = screen
      .getByRole('button', { name: /^Karte 0 —/ })
      .closest('[role="row"]')

    expect(erste).toHaveAttribute('aria-rowindex', '1')
    expect(screen.getByRole('grid')).toHaveAttribute('aria-rowcount', '3788')
  })
})

// **Ein Bild in der Zeile** — Werkstück 2 von Phase 3.5
// (#1166).
//
// # Warum es hier steht und nicht bei `useListenbilder`
//
// Dort steht, **wann** geholt wird; hier, **dass** die Liste zeigt, was
// ankommt. Ein Haken, der Bilder liefert, die niemand einsetzt, wäre in beiden
// Prüfungen grün.
describe('das Bild in der Zeile', () => {
  it('zeigt das Bild, das der Haken liefert', async () => {
    belegeApi({
      Kartenbilder: vi.fn().mockResolvedValue({
        bilder: { '1': 'data:image/png;base64,AAA' },
        fehler: undefined,
      }),
    })

    zeige([karte('1', 1, 'Mirri'), karte('2', 2, 'Rith')])

    // **Über `presentation`**, denn das Bild trägt `alt=""` und liegt unter
    // `aria-hidden`: daneben steht der Name, und zweimal dieselbe Karte
    // vorgelesen zu bekommen hilft niemandem.
    const bild = await screen.findByRole('presentation', { hidden: true })

    expect(bild).toHaveAttribute('src', 'data:image/png;base64,AAA')

    // **Und die Zeile ohne Bild trägt keines** — der Normalfall. Ohne diese
    // Zeile wäre auch eine Liste grün, die überall dasselbe Bild setzt.
    expect(screen.getAllByRole('presentation', { hidden: true })).toHaveLength(
      1,
    )
  })
})
