// **Die Kästchen der Kartenliste stehen erst im Auswahlmodus** — UXR-05
// (#1913).
//
// # Was hier gemessen wird
//
// **Dass die Liste ohne Auswahl kein Feld zeichnet.** Bis UXR-05 trug jede
// gezeichnete Zeile ein Kästchen für die Massenänderung. Das Sichtfenster
// zeichnet die Zeilen im Bild und je acht davor und dahinter
// (`useSichtfenster`) — bei 1.000 Karten im Fenster rund zwei Dutzend, unter
// der Vorrichtung dieses Laufs siebzehn —, und jede war ein Feld mehr auf
// der Fläche (R3 aus `docs/UIREVIEW.md`: höchstens zwölf). Gebraucht werden
// sie für einen seltenen Weg: mehrere Karten zugleich ändern.
//
// **Dass der Weg dahin erhalten bleibt** — per Maus und per Tastatur, und
// auch über den Menüeintrag „Alle Karten anhaken".
//
// **Und dass das Abschalten nichts versteckt**: wer den Modus verlässt, hebt
// die Anhakung auf, statt sie unsichtbar stehen zu lassen.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO } from '../api'
import { ereignisAllesAuswaehlen } from '../fenster'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

const texte = kataloge.de.seiten.karten

afterEach(() => {
  raeumeApi()
  delete (globalThis as { runtime?: unknown }).runtime
})

// tausend sind die Karten eines großen Sets — so viele, wie ein Mensch in
// einem Cube oder einer Sammlung von Entwürfen tatsächlich führt.
const tausend: KarteDTO[] = Array.from({ length: 1000 }, (_, stelle) => ({
  id: `k${String(stelle)}`,
  nummer: stelle,
  werte: [{ feld: 'name', wert: `Karte ${String(stelle + 1)}` }],
}))

// horcher hält die Rückrufe, die die Seite bei der Laufzeit anmeldet — der
// Weg, auf dem die Menüleiste „Alle Karten anhaken" schickt.
function laufzeitAttrappe(): Map<string, () => void> {
  const angemeldet = new Map<string, () => void>()

  ;(globalThis as { runtime?: unknown }).runtime = {
    EventsOn: (name: string, rueckruf: () => void) => {
      angemeldet.set(name, rueckruf)

      return () => angemeldet.delete(name)
    },
  }

  return angemeldet
}

function zeige(karten: KarteDTO[] = tausend) {
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    ...ausKarten(() => Promise.resolve({ karten })),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID={undefined}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

// gitter ist die Kartenliste — die Kästchen werden **in ihr** gezählt, damit
// ein Kästchen anderswo auf der Seite (die Spaltenwahl, der Editor) die
// Zählung nicht verfälscht.
async function gitter() {
  return screen.findByRole('grid', { name: texte.titel })
}

// zeilen zählt die gezeichneten Zeilen — erst, wenn mehr als zwölf dastehen:
// ab dann wären es mit Kästchen mehr Felder, als R3 zulässt.
async function zeilen(liste: HTMLElement) {
  await waitFor(() => {
    expect(within(liste).getAllByRole('row').length).toBeGreaterThan(12)
  })

  return within(liste).getAllByRole('row').length
}

describe('Der Auswahlmodus der Kartenliste', () => {
  it('zeichnet bei 1.000 Karten keine Kästchen, solange niemand auswählt', async () => {
    zeige()

    const liste = await gitter()
    await zeilen(liste)

    expect(within(liste).queryAllByRole('checkbox')).toHaveLength(0)
  })

  it('zeigt die Kästchen nach „Mehrere wählen", und angehakt geht in die Massenänderung', async () => {
    zeige()
    const liste = await gitter()

    const schalter = screen.getByRole('button', {
      name: texte.massen.mehrereWaehlen,
    })
    expect(schalter).toHaveAttribute('aria-pressed', 'false')

    const gezeichnet = await zeilen(liste)
    await userEvent.click(schalter)

    expect(schalter).toHaveAttribute('aria-pressed', 'true')
    expect(within(liste).getAllByRole('checkbox')).toHaveLength(gezeichnet)

    await userEvent.click(
      within(liste).getByRole('checkbox', {
        name: `Karte 3 — ${texte.massen.auswaehlen}`,
      }),
    )

    expect(
      await screen.findByRole('button', { name: texte.massen.anwenden }),
    ).toBeInTheDocument()
  })

  it('hebt beim Verlassen des Modus die Anhakung auf, statt sie zu verstecken', async () => {
    zeige()
    const liste = await gitter()
    const schalter = screen.getByRole('button', {
      name: texte.massen.mehrereWaehlen,
    })

    await userEvent.click(schalter)
    await userEvent.click(
      within(liste).getByRole('checkbox', {
        name: `Karte 1 — ${texte.massen.auswaehlen}`,
      }),
    )
    expect(
      await screen.findByRole('button', { name: texte.massen.anwenden }),
    ).toBeInTheDocument()

    await userEvent.click(schalter)

    expect(within(liste).queryAllByRole('checkbox')).toHaveLength(0)
    expect(
      screen.queryByRole('button', { name: texte.massen.anwenden }),
    ).not.toBeInTheDocument()
    expect(schalter).toHaveAttribute('aria-pressed', 'false')
  })

  it('bleibt nach dem letzten Abhaken im Modus — das Kästchen verschwindet nicht unter dem Zeiger', async () => {
    zeige()
    const liste = await gitter()

    await userEvent.click(
      screen.getByRole('button', { name: texte.massen.mehrereWaehlen }),
    )
    const kaestchen = within(liste).getByRole('checkbox', {
      name: `Karte 2 — ${texte.massen.auswaehlen}`,
    })

    await userEvent.click(kaestchen)
    await userEvent.click(kaestchen)

    expect(kaestchen).toBeInTheDocument()
    expect(kaestchen).not.toBeChecked()
  })

  it('schaltet den Modus ein, wenn die Menüleiste alle Karten anhakt', async () => {
    const horcher = laufzeitAttrappe()
    zeige()
    const liste = await gitter()
    await effekteNachholen()

    const gezeichnet = await zeilen(liste)
    horcher.get(ereignisAllesAuswaehlen)?.()

    await waitFor(() => {
      expect(within(liste).getAllByRole('checkbox')).toHaveLength(gezeichnet)
    })
    expect(
      screen.getByRole('button', { name: texte.massen.mehrereWaehlen }),
    ).toHaveAttribute('aria-pressed', 'true')

    // **Und ein einzelnes Abhaken lässt die übrigen stehen** — der Modus
    // hängt nicht daran, dass alle angehakt sind.
    await userEvent.click(
      within(liste).getByRole('checkbox', {
        name: `Karte 1 — ${texte.massen.auswaehlen}`,
      }),
    )
    expect(
      within(liste).getByRole('checkbox', {
        name: `Karte 2 — ${texte.massen.auswaehlen}`,
      }),
    ).toBeChecked()
  })

  it('erreicht das Kästchen im Modus mit der Tastatur — → in die Zeile, Leertaste hakt an', async () => {
    zeige()
    const liste = await gitter()

    await userEvent.click(
      screen.getByRole('button', { name: texte.massen.mehrereWaehlen }),
    )

    const erste = within(liste).getAllByRole('row')[0]
    erste?.focus()
    await userEvent.keyboard('{ArrowRight}')

    const kaestchen = within(liste).getByRole('checkbox', {
      name: `Karte 1 — ${texte.massen.auswaehlen}`,
    })
    expect(kaestchen).toHaveFocus()

    await userEvent.keyboard(' ')

    expect(kaestchen).toBeChecked()
  })
})
