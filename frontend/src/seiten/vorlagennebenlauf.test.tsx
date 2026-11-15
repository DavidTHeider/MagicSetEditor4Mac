// **Die Paketwerkstatt: Wechsel, Fehler, zwei Vorgänge** — Werkstück 2 von
// Phase 3.3 (#1164).
//
// # Warum das eine eigene Datei ist
//
// `vorlagen.test.tsx` fährt den **Weg**: Paket wählen, Datei aufschlagen,
// ändern, speichern. Hier stehen die Fälle daneben — was beim Wechsel des
// Bezugsobjekts geräumt wird (§7, Regel 1), was ein Fehler auf jedem der drei
// Wege anrichtet, und was ein zweiter Klick tut, während der erste liest
// (Regel 4).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { paketDTO } from '../test/vorlagenprobe'
import { belegeWerkstatt } from '../test/werkstattprobe'

import { Vorlagen } from './Vorlagen'

const texte = kataloge.de.seiten.vorlagen

describe('Die Paketwerkstatt', () => {
  afterEach(raeumeApi)

  // **Ein Paketwechsel räumt die offene Datei** (§7, Regel 1).
  it('räumt die offene Datei beim Paketwechsel', async () => {
    belegeWerkstatt({
      Pakete: () =>
        Promise.resolve({
          pakete: [paketDTO('probe.mse-style'), paketDTO('zweit.mse-style')],
          befunde: [],
          fehler: null,
        }),
      Paketdateien: (paket: string) =>
        Promise.resolve({
          dateien: paket === '/pakete/probe.mse-style' ? ['style'] : ['game'],
          beiwerk: [],
          fehler: null,
        }),
      PaketdateiLesen: (paket: string) =>
        Promise.resolve({
          inhalt: `aus ${paket}\n`,
          stand: 'alt',
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagen />
      </Sprachraum>,
    )

    await userEvent.click(await screen.findByRole('button', { name: 'style' }))

    expect(await screen.findByLabelText(texte.inhalt)).toHaveValue(
      'aus /pakete/probe.mse-style\n',
    )

    // **Der Wechsel räumt den Puffer**: das Feld zeigt die andere Datei, nicht
    // den vorigen Text unter neuem Namen (§7, Regel 1).
    await userEvent.click(screen.getByRole('button', { name: 'game' }))

    await waitFor(() => {
      expect(screen.getByLabelText(texte.inhalt)).toHaveValue(
        'aus /pakete/zweit.mse-style\n',
      )
    })
  })

  // **Ein Paket ohne bearbeitbare Datei sagt es** (§6: leere Zustände
  // erklären den nächsten Schritt).
  it('sagt es, wenn ein Paket nichts zu bearbeiten hat', async () => {
    belegeWerkstatt({
      Pakete: () =>
        Promise.resolve({
          pakete: [paketDTO('probe.mse-style')],
          befunde: [],
          fehler: null,
        }),
      Paketdateien: () =>
        Promise.resolve({ dateien: [], beiwerk: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagen />
      </Sprachraum>,
    )

    expect(await screen.findByText(texte.keineDateien)).toBeInTheDocument()
  })

  // **Ein Fehler beim Aufzählen kommt an — von jeder der beiden Listen.**
  //
  // `useLaden` hält für jede einen eigenen Stand; die Seite zeigt den, der da
  // ist. Ohne den zweiten Fall sagte der Lauf auch dann ja, wenn nur der
  // Paketstand angesehen würde.
  it('meldet einen Fehler beider Listen', async () => {
    for (const belegung of [
      {
        Pakete: () =>
          Promise.resolve({
            pakete: [],
            befunde: [],
            fehler: {
              code: 'unbekannt' as const,
              schluessel: 'unbekannt',
              meldung: 'kein Paketverzeichnis',
            },
          }),
      },
      {
        Pakete: () =>
          Promise.resolve({
            pakete: [paketDTO('probe.mse-style')],
            befunde: [],
            fehler: null,
          }),
        Paketdateien: () =>
          Promise.resolve({
            dateien: [],
            beiwerk: [],
            fehler: {
              code: 'nichtGefunden' as const,
              schluessel: 'nichtGefunden',
              meldung: 'weg',
            },
          }),
      },
    ]) {
      raeumeApi()
      belegeWerkstatt(belegung)

      const { unmount } = render(
        <Sprachraum sprache="de">
          <Vorlagen />
        </Sprachraum>,
      )

      expect(await screen.findByRole('alert')).toBeInTheDocument()

      unmount()
    }
  })

  // **Ein zweiter Klick, während der erste liest, tut nichts** (§7, Regel 4).
  //
  // Zwei Lesevorgänge auf dieselbe Datei gingen gegen denselben Stand, und der
  // zweite trüge seine Antwort über die des ersten — mit demselben Inhalt, aber
  // ohne dass jemand sich darauf verlassen dürfte.
  it('liest eine Datei nicht zweimal zugleich', async () => {
    let weiter: (() => void) | undefined
    let gelesen = 0

    belegeWerkstatt({
      Pakete: () =>
        Promise.resolve({
          pakete: [paketDTO('probe.mse-style')],
          befunde: [],
          fehler: null,
        }),
      Paketdateien: () =>
        Promise.resolve({ dateien: ['style'], beiwerk: [], fehler: null }),
      PaketdateiLesen: async () => {
        gelesen++

        await new Promise<void>((los) => {
          weiter = los
        })

        return { inhalt: 'mse version: 2.0.0\n', stand: 'alt', fehler: null }
      },
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagen />
      </Sprachraum>,
    )

    const knopf = await screen.findByRole('button', { name: 'style' })

    await userEvent.click(knopf)
    await userEvent.click(knopf)

    expect(gelesen).toBe(1)

    weiter?.()

    expect(await screen.findByLabelText(texte.inhalt)).toHaveValue(
      'mse version: 2.0.0\n',
    )
  })

  // **Und ein Fehler beim Lesen kommt an.**
  it('meldet einen Fehler beim Aufschlagen', async () => {
    belegeWerkstatt({
      Pakete: () =>
        Promise.resolve({
          pakete: [paketDTO('probe.mse-style')],
          befunde: [],
          fehler: null,
        }),
      Paketdateien: () =>
        Promise.resolve({ dateien: ['style'], beiwerk: [], fehler: null }),
      PaketdateiLesen: () =>
        Promise.resolve({
          inhalt: '',
          stand: '',
          fehler: {
            code: 'nichtGefunden',
            schluessel: 'nichtGefunden',
            meldung: 'weg',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagen />
      </Sprachraum>,
    )
    await userEvent.click(await screen.findByRole('button', { name: 'style' }))

    expect(await screen.findByRole('alert')).toBeInTheDocument()

    // **Kein halbes Feld**: was nicht zu lesen war, wird nicht aufgeschlagen.
    expect(screen.queryByLabelText(texte.inhalt)).not.toBeInTheDocument()
  })
})
