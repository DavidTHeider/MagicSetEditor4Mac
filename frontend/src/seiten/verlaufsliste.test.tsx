import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenzeileDTO, VerlaufsschrittDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { Verlaufsliste } from './Verlaufsliste'

const texte = kataloge.de.seiten.karten.verlauf

const karten: KartenzeileDTO[] = [
  { id: '7', nummer: 0, name: 'Feuerball', felder: 1 },
  { id: '8', nummer: 1, name: 'Eisblitz', felder: 2 },
]

function schritt(teil: Partial<VerlaufsschrittDTO>): VerlaufsschrittDTO {
  return {
    id: '1',
    zeitpunkt: '2026-09-09T02:01:00.123456789Z',
    handlung: 'feld',
    karteID: '7',
    feld: 'name',
    vorher: 'Alt',
    gabVorher: true,
    nachher: 'Neu',
    gabNachher: true,
    zurueckgenommen: false,
    urheber: 'fenster',
    herkunft: '',
    vorgang: '',
    umfang: 1,
    ...teil,
  }
}

// zeile gibt die n-te Zeile — oder scheitert benannt, statt `undefined`
// durchzureichen.
function zeile(zeilen: readonly HTMLElement[], n: number): HTMLElement {
  const gefunden = zeilen[n]
  if (gefunden === undefined) {
    throw new Error(`keine Zeile ${String(n)}`)
  }

  return gefunden
}

function zeige(zuruecknehmen?: (vorgang: string) => void, verlaufsstand = 0) {
  return render(
    <Sprachraum sprache="de">
      <Verlaufsliste
        setID="1"
        karten={karten}
        verlaufsstand={verlaufsstand}
        zuruecknehmen={zuruecknehmen}
      />
    </Sprachraum>,
  )
}

// **Die Verlaufsliste** — Werkstück U6 von Phase 5.1 (#1417).
describe('Verlaufsliste', () => {
  afterEach(() => {
    raeumeApi()
  })

  it('nennt je Schritt Karte, Feld, Bisher, Danach und Wer — und die Zahl der Karten', async () => {
    belegeApi({
      Verlauf: () =>
        Promise.resolve({
          schritte: [
            schritt({
              id: '3',
              karteID: '8',
              feld: 'typ',
              vorher: 'Zauber',
              nachher: 'Land',
              vorgang: 'g1',
              umfang: 2,
              urheber: 'server',
            }),
            schritt({ id: '1', gabVorher: false, vorher: '' }),
          ],
        }),
    })
    zeige()

    const zeilen = within(await screen.findByRole('table')).getAllByRole('row')
    expect(zeilen).toHaveLength(3)
    expect(zeilen[1]).toHaveTextContent('2026-09-09 02:01:00')
    expect(zeilen[1]).toHaveTextContent('Eisblitz')
    expect(zeilen[1]).toHaveTextContent('2 Karten')
    expect(zeilen[1]).toHaveTextContent('typ')
    expect(zeilen[1]).toHaveTextContent('Zauber')
    expect(zeilen[1]).toHaveTextContent('Land')
    expect(zeilen[1]).toHaveTextContent(texte.urheberServer)
    expect(zeilen[2]).toHaveTextContent('Feuerball')
    expect(zeilen[2]).toHaveTextContent(texte.keinWert)
    expect(zeilen[2]).toHaveTextContent(texte.urheberFenster)
  })

  // **Die Tabelle rollt in sich, nicht die Seite** — UX-14.AK-03 (#1686):
  // sechs Spalten sprengten bei 1.024 × 700 die mittlere Spalte, und die
  // Knöpfe „Bilder vergleichen" lagen hinter dem rechten Rand (gefunden von
  // `ux14-groessen`). Ein Rollbereich ist per Tastatur erreichbar und
  // benannt, sonst rollte ihn nur, wer eine Maus hat.
  it('legt die Tabelle in einen eigenen, erreichbaren Rollbereich', async () => {
    belegeApi({
      Verlauf: () => Promise.resolve({ schritte: [schritt({ id: '1' })] }),
    })
    zeige()

    const tabelle = await screen.findByRole('table')
    const bereich = tabelle.parentElement
    expect(bereich).toHaveClass('overflow-x-auto')
    expect(bereich).toHaveAttribute('tabindex', '0')
    expect(screen.getByRole('region', { name: texte.ueberschrift })).toBe(
      bereich,
    )
  })

  it('nennt einen fehlenden Wert danach und einen unbekannten Urheber beim Namen', async () => {
    belegeApi({
      Verlauf: () =>
        Promise.resolve({
          schritte: [
            schritt({
              id: '2',
              gabNachher: false,
              nachher: '',
              urheber: 'fremd',
            }),
          ],
        }),
    })
    zeige()

    const zeilen = within(await screen.findByRole('table')).getAllByRole('row')
    expect(zeilen[1]).toHaveTextContent(texte.keinWert)
    expect(zeilen[1]).toHaveTextContent('fremd')
  })

  it('klappt den Vergleich über denselben Knopf wieder zu', async () => {
    belegeApi({
      Verlauf: () => Promise.resolve({ schritte: [schritt({ id: '1' })] }),
      Probevorschau: () => Promise.resolve({ vorschau: null }),
    })
    zeige()

    const knopf = await screen.findByRole('button', { name: texte.vergleichen })
    await userEvent.click(knopf)
    expect(knopf).toHaveAttribute('aria-pressed', 'true')
    expect(
      await screen.findByRole('button', { name: texte.vergleichSchliessen }),
    ).toBeInTheDocument()

    await userEvent.click(knopf)
    expect(knopf).toHaveAttribute('aria-pressed', 'false')
    expect(
      screen.queryByRole('button', { name: texte.vergleichSchliessen }),
    ).not.toBeInTheDocument()
  })

  it('nennt eine unbekannte Karte bei ihrer Kennung und eine andere Handlung beim Wort', async () => {
    belegeApi({
      Verlauf: () =>
        Promise.resolve({
          schritte: [
            schritt({ id: '4', karteID: '99', handlung: 'karte', feld: '' }),
          ],
        }),
    })
    zeige()

    const zeilen = within(await screen.findByRole('table')).getAllByRole('row')
    expect(zeilen[1]).toHaveTextContent('99')
    expect(zeilen[1]).toHaveTextContent(texte.handlungKarte)
  })

  it('erklärt einen leeren Verlauf', async () => {
    belegeApi({ Verlauf: () => Promise.resolve({ schritte: [] }) })
    zeige()

    expect(await screen.findByText(texte.leer)).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('nimmt genau den genannten Schritt zurück — und nur, was noch gilt und einen Vorgang hat', async () => {
    const zuruecknehmen = vi.fn()
    belegeApi({
      Verlauf: () =>
        Promise.resolve({
          schritte: [
            schritt({ id: '3', vorgang: 'g1', umfang: 2 }),
            schritt({ id: '2', vorgang: 'g0', zurueckgenommen: true }),
            schritt({ id: '1' }),
          ],
        }),
    })
    zeige(zuruecknehmen)

    const zeilen = within(await screen.findByRole('table')).getAllByRole('row')
    expect(
      within(zeile(zeilen, 2)).queryByRole('button', {
        name: texte.zuruecknehmen,
      }),
    ).toBeNull()
    expect(zeilen[2]).toHaveTextContent(texte.zurueckgenommen)
    expect(
      within(zeile(zeilen, 3)).queryByRole('button', {
        name: texte.zuruecknehmen,
      }),
    ).toBeNull()

    await userEvent.click(
      within(zeile(zeilen, 1)).getByRole('button', {
        name: texte.zuruecknehmen,
      }),
    )
    expect(zuruecknehmen).toHaveBeenCalledExactlyOnceWith('g1')
  })

  it('zeigt auf Wunsch beide Karten nebeneinander und schließt den Vergleich wieder', async () => {
    const Probevorschau = vi.fn((_: string, __: string, wert: string) =>
      Promise.resolve({
        vorschau: {
          stil: '',
          gezeichnet: '',
          bild: `/vorschau/${wert}.png`,
          breite: 10,
          hoehe: 14,
          befunde: [],
          flaechen: [],
          vorgaben: [],
          nennungen: [],
          dpi: 0,
        },
      }),
    )
    belegeApi({
      Verlauf: () => Promise.resolve({ schritte: [schritt({ id: '1' })] }),
      Probevorschau,
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.vergleichen }),
    )

    expect(
      await screen.findByRole('img', {
        name: 'Karte Feuerball mit dem Wert von vorher',
      }),
    ).toHaveAttribute('src', '/vorschau/Alt.png')
    expect(
      await screen.findByRole('img', {
        name: 'Karte Feuerball mit dem Wert von danach',
      }),
    ).toHaveAttribute('src', '/vorschau/Neu.png')
    expect(Probevorschau).toHaveBeenCalledWith('7', 'name', 'Alt', 1)

    await userEvent.click(
      screen.getByRole('button', { name: texte.vergleichSchliessen }),
    )
    expect(screen.queryByRole('img')).not.toBeInTheDocument()
  })

  // **Nacheinander, nicht nebeneinander.** `App.Probevorschau` teilt sich den
  // Lauf mit der Kartenvorschau, und ein neuer Ruf bricht den vorigen ab
  // (§7): zwei Bilder zugleich angefordert, kam eines nie an — gefunden in
  // der Sichtprüfung von U6, die „danach“-Seite blieb leer.
  it('holt die zweite Seite erst, wenn die erste da ist', async () => {
    let erstesFreigeben: (() => void) | undefined
    const Probevorschau = vi.fn((_: string, __: string, wert: string) =>
      wert === 'Alt'
        ? new Promise<{ vorschau: null }>((los) => {
            erstesFreigeben = () => {
              los({ vorschau: null })
            }
          })
        : Promise.resolve({ vorschau: null }),
    )
    belegeApi({
      Verlauf: () => Promise.resolve({ schritte: [schritt({ id: '1' })] }),
      Probevorschau,
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.vergleichen }),
    )

    await waitFor(() => {
      expect(Probevorschau).toHaveBeenCalledTimes(1)
    })
    await antwortAbwarten()
    expect(Probevorschau).toHaveBeenCalledTimes(1)

    erstesFreigeben?.()

    await waitFor(() => {
      expect(Probevorschau).toHaveBeenCalledTimes(2)
    })
    expect(Probevorschau).toHaveBeenLastCalledWith('7', 'name', 'Neu', 1)
  })

  it('vergleicht mit der leeren Zeichenkette, wo es keinen Wert gab', async () => {
    const Probevorschau = vi.fn(() => Promise.resolve({ vorschau: null }))
    belegeApi({
      Verlauf: () =>
        Promise.resolve({
          schritte: [
            schritt({
              id: '1',
              gabVorher: false,
              vorher: 'egal',
              gabNachher: false,
              nachher: 'egal',
            }),
          ],
        }),
      Probevorschau,
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.vergleichen }),
    )

    await waitFor(() => {
      expect(Probevorschau).toHaveBeenCalledTimes(2)
    })
    expect(Probevorschau).toHaveBeenNthCalledWith(1, '7', 'name', '', 1)
    expect(Probevorschau).toHaveBeenNthCalledWith(2, '7', 'name', '', 1)
  })

  it('lädt neu, wenn der Verlaufsstand steigt', async () => {
    const Verlauf = vi.fn(() => Promise.resolve({ schritte: [] }))
    belegeApi({ Verlauf })
    const { rerender } = zeige(undefined, 0)
    await screen.findByText(texte.leer)

    rerender(
      <Sprachraum sprache="de">
        <Verlaufsliste
          setID="1"
          karten={karten}
          verlaufsstand={1}
          zuruecknehmen={undefined}
        />
      </Sprachraum>,
    )

    await screen.findByText(texte.leer)
    expect(Verlauf).toHaveBeenCalledTimes(2)
  })
})
