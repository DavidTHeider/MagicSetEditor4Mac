// **Eine Balkengruppe filtert die Kartenliste** — UXR-F6
// (#2013),
// ein Kompatibilitätsrest aus UXR-08 (§15).
//
// # Was MSE2 tut (nachgelesen am 2026-10-03)
//
// Ein Klick ins Diagramm wählt die Gruppe darunter, und die Liste daneben
// zeigt nur ihre Karten (`GraphControl::onMouseDown` → `StatsPanel::
// onGraphSelect` → `filterCards`, `src/gui/control/graph.cpp` Zeilen
// 1116–1124 und `src/gui/set/stats_panel.cpp` Zeilen 526–547, Fassung f9d9356
// vom 2021-07-06). Gewählt wird **eine** Gruppe je Achse; ein Klick ersetzt
// die Wahl. Wechselt das Diagramm seine Daten, ist die Wahl weg.
//
// # Was hier steht
//
// Die Seite *Statistik* in ihrem Raum, mit einem Leser daneben, der zeigt, was
// die Kartenliste bekäme. Dass die Liste es dann zeigt und dass beide
// denselben Raum teilen, steht in `kartenfilterstatistik.test.tsx`; was nach
// dem Neuladen aus der Wahl wird, in `gruppenabgleich.test.tsx`. Die
// Vorrichtung ist `test/gruppenproben.tsx`.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { raeumeApi } from '../test/api'
import {
  balken,
  belege,
  filter,
  seite,
  texte,
  waehleBild,
} from '../test/gruppenproben'

afterEach(() => {
  raeumeApi()
})

describe('Die Wahl eines Balkens in der Verteilung', () => {
  it('nimmt die Karten des Balkens in den Raum', async () => {
    belege()
    render(seite())

    await waehleBild('feld:rarity')

    const rare = await balken('rare')
    expect(rare).toHaveAttribute('aria-pressed', 'false')
    expect(filter()).toHaveTextContent('keiner')

    await userEvent.click(rare)

    expect(filter()).toHaveTextContent('feld:rarity | rarity | rare | 3,7')
    expect(rare).toHaveAttribute('aria-pressed', 'true')
    expect(await balken('common')).toHaveAttribute('aria-pressed', 'false')
  })

  // **MSE2 ersetzt die Wahl** — es gibt nur eine Gruppe je Achse.
  it('ersetzt die Wahl, wenn ein anderer Balken folgt', async () => {
    belege()
    render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))
    await userEvent.click(await balken('common'))

    expect(filter()).toHaveTextContent('common | 1,2,5')
    expect(await balken('rare')).toHaveAttribute('aria-pressed', 'false')
  })

  // **Ein Klick auf den gewählten Balken hebt die Wahl auf** — angepasst:
  // MSE2 kennt das nicht, dort kommt man mit der Pfeiltaste zurück.
  it('hebt die Wahl auf, wenn derselbe Balken noch einmal gewählt wird', async () => {
    belege()
    render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))
    await userEvent.click(await balken('rare'))

    expect(filter()).toHaveTextContent('keiner')
    expect(await balken('rare')).toHaveAttribute('aria-pressed', 'false')
  })

  // **Ohne Maus**: Tab bis zum Balken, Enter wählt.
  it('wählt mit der Tastatur', async () => {
    belege()
    render(seite())

    await waehleBild('feld:rarity')
    const common = await balken('common')

    common.focus()
    await userEvent.keyboard('{Enter}')

    expect(filter()).toHaveTextContent('common | 1,2,5')
  })

  it('sagt, was ein Klick tut, solange nichts gewählt ist', async () => {
    belege()
    render(seite())

    await waehleBild('feld:rarity')
    await balken('rare')

    expect(screen.getByText(texte.gruppeHinweis)).toBeInTheDocument()
  })

  it('nennt den gewählten Balken und führt zur Kartenliste', async () => {
    const wechseln = vi.fn()

    belege()
    render(seite(wechseln))

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))

    expect(screen.queryByText(texte.gruppeHinweis)).not.toBeInTheDocument()
    expect(
      screen.getByText(
        einsetzen(texte.gruppeGewaehlt, { gruppe: 'rare', karten: '2' }),
      ),
    ).toBeInTheDocument()

    // **Die Wahl steht schon im Raum**, bevor jemand zur Liste geht: der
    // Knopf ist der Weg und nicht das Absenden.
    await userEvent.click(
      screen.getByRole('button', { name: texte.gruppeZeigen }),
    )

    expect(wechseln).toHaveBeenLastCalledWith('karten')
    expect(filter()).toHaveTextContent('rare | 3,7')
  })

  it('hebt die Wahl auf Verlangen auf', async () => {
    belege()
    render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))
    await userEvent.click(
      screen.getByRole('button', { name: texte.gruppeAufheben }),
    )

    expect(filter()).toHaveTextContent('keiner')
    expect(screen.getByText(texte.gruppeHinweis)).toBeInTheDocument()
  })
})

describe('Die Wahl eines Balkens in der Auszählung', () => {
  it('nimmt die Karten der Gruppe in den Raum', async () => {
    belege()
    render(seite())

    await waehleBild('dimension:color category')
    await userEvent.click(await balken('White'))

    expect(filter()).toHaveTextContent(
      'dimension:color category | color category | White | 1,3',
    )
  })

  // **MSE2 wählt auch die leere Gruppe und zeigt eine leere Liste.** Hier ist
  // sie angehalten, und der Titel sagt warum.
  it('hält eine Gruppe ohne Karten an', async () => {
    belege()
    render(seite())

    await waehleBild('dimension:color category')

    const land = await balken('Land')
    expect(land).toBeDisabled()
    expect(land).toHaveAttribute('title', texte.gruppeLeer)

    await userEvent.click(land)
    expect(filter()).toHaveTextContent('keiner')
  })
})

describe('Was keine Gruppe ist', () => {
  // **Die Belegung zählt Felder, und eine Karte steht in vielen.** Ihre Balken
  // sind Zahlen zum Lesen.
  it('lässt die Belegung ohne Wahl', async () => {
    belege()
    render(seite())

    expect(await screen.findByText(/^7 \(100\s%\)$/)).toBeInTheDocument()

    const bild = screen.getByRole('region', { name: texte.belegung })
    expect(within(bild).queryByRole('button')).not.toBeInTheDocument()
    expect(screen.queryByText(texte.gruppeHinweis)).not.toBeInTheDocument()
  })
})

describe('Eine Wahl gilt für das Bild, aus dem sie stammt', () => {
  // **In MSE2 leert `setData` die Wahl**, sobald die Daten wechseln.
  it('endet, wenn ein anderes Bild gewählt wird', async () => {
    belege()
    render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))
    expect(filter()).toHaveTextContent('rare')

    await waehleBild('dimension:color category')

    expect(filter()).toHaveTextContent('keiner')

    // Und auch das Zurück zum ersten Bild bringt sie nicht wieder.
    await waehleBild('feld:rarity')
    expect(await balken('rare')).toHaveAttribute('aria-pressed', 'false')
  })

  it('endet auch beim Wechsel zur Belegung', async () => {
    belege()
    render(seite())

    await waehleBild('feld:rarity')
    await userEvent.click(await balken('rare'))
    await waehleBild('belegung')

    expect(filter()).toHaveTextContent('keiner')
  })
})
