import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { raeumeApi, belegeApi } from '../test/api'
import {
  feldwerte,
  fuelleAus,
  mitVorschau,
  texte,
  vorschauAntwort,
  vorschauen,
  waehleFeld,
  zeige,
} from './massenprobe'

// **Die Massenänderung mit Vorschau** — Werkstück U6 von Phase 5.1 (#1417).
describe('Massenänderung', () => {
  afterEach(() => {
    raeumeApi()
  })

  it('erklärt bei leerer Auswahl den nächsten Schritt', () => {
    zeige([])

    expect(screen.getByText(texte.keineAuswahl)).toBeInTheDocument()
  })

  it('holt die Vorschau für die ganze Auswahl und zeigt je Karte Bisher und Danach', async () => {
    const FelderVorschau = vi.fn(() => Promise.resolve(vorschauAntwort()))
    belegeApi({ ...feldwerte, FelderVorschau })
    zeige()

    await vorschauen()

    expect(FelderVorschau).toHaveBeenCalledExactlyOnceWith(
      '1',
      ['1', '2'],
      'typ',
      'Artefakt',
    )

    const zeilen = within(screen.getByRole('table')).getAllByRole('row')
    // Kopfzeile und zwei Karten
    expect(zeilen).toHaveLength(3)
    expect(zeilen[1]).toHaveTextContent('Feuerball')
    expect(zeilen[1]).toHaveTextContent(texte.keinWertBisher)
    expect(zeilen[2]).toHaveTextContent('Zauber')
    expect(zeilen[2]).toHaveTextContent('Artefakt')
    expect(
      screen.getByText('2 Karte(n) werden geändert, 0 bleiben.'),
    ).toBeInTheDocument()
  })

  it('nennt eine Karte ohne Namen bei ihrer Kennung', async () => {
    const antwort = vorschauAntwort()
    antwort.zeilen = antwort.zeilen.map((zeile, nummer) =>
      nummer === 0 ? { ...zeile, name: '' } : zeile,
    )
    belegeApi({ ...feldwerte, FelderVorschau: () => Promise.resolve(antwort) })
    zeige()

    await vorschauen()

    const zeilen = within(screen.getByRole('table')).getAllByRole('row')
    expect(zeilen[1]).toHaveTextContent('1')
    expect(zeilen[1]).not.toHaveTextContent('Feuerball')
  })

  it('nennt, was schon so steht, und hat dann nichts anzuwenden', async () => {
    belegeApi({
      ...feldwerte,
      FelderVorschau: () =>
        Promise.resolve({
          feld: 'typ',
          wert: 'Zauber',
          zeilen: [
            {
              karteID: '1',
              name: 'Feuerball',
              alt: 'Zauber',
              gabAlt: true,
              neu: 'Zauber',
              ausschluss: 'gleich',
            },
            {
              karteID: '2',
              name: 'Eisblitz',
              alt: 'Zauber',
              gabAlt: true,
              neu: 'Zauber',
              ausschluss: 'gleich',
            },
          ],
        }),
    })
    zeige()

    await vorschauen('typ', 'Zauber')

    expect(screen.getAllByText(texte.bleibtGleich)).toHaveLength(2)
    expect(screen.getByText(texte.nichtsZuTun)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: texte.anwenden })).toBeDisabled()
  })

  it('wendet die Vorschau mit ihrer Erwartung an und lädt nach', async () => {
    const FelderSetzenMitErwartung = vi.fn(() =>
      Promise.resolve({ betroffen: [], vorgang: 'v1' }),
    )
    belegeApi({ ...mitVorschau, FelderSetzenMitErwartung })
    const nachladen = zeige()

    await fuelleAus()

    await waitFor(() => {
      expect(nachladen).toHaveBeenCalledOnce()
    })
    expect(FelderSetzenMitErwartung).toHaveBeenCalledExactlyOnceWith(
      '1',
      [
        { karteID: '1', wert: '', gabEs: false },
        { karteID: '2', wert: 'Zauber', gabEs: true },
      ],
      'typ',
      'Artefakt',
    )
    expect(screen.getByRole('status')).toHaveTextContent('2 Karte(n) geändert.')
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('lässt beim Anwenden die Karten weg, die schon so stehen', async () => {
    const FelderSetzenMitErwartung = vi.fn(() =>
      Promise.resolve({ betroffen: [], vorgang: 'v1' }),
    )
    belegeApi({
      ...feldwerte,
      FelderVorschau: () =>
        Promise.resolve({
          feld: 'typ',
          wert: 'Artefakt',
          zeilen: [
            {
              karteID: '1',
              name: 'Feuerball',
              alt: '',
              gabAlt: false,
              neu: 'Artefakt',
              ausschluss: '',
            },
            {
              karteID: '2',
              name: 'Eisblitz',
              alt: 'Artefakt',
              gabAlt: true,
              neu: 'Artefakt',
              ausschluss: 'gleich',
            },
          ],
        }),
      FelderSetzenMitErwartung,
    })
    zeige()

    await fuelleAus()

    await waitFor(() => {
      expect(FelderSetzenMitErwartung).toHaveBeenCalledExactlyOnceWith(
        '1',
        [{ karteID: '1', wert: '', gabEs: false }],
        'typ',
        'Artefakt',
      )
    })
  })

  it('meldet einen Konflikt, hebt die Karten hervor und lädt nicht nach', async () => {
    belegeApi({
      ...mitVorschau,
      FelderSetzenMitErwartung: () =>
        Promise.resolve({
          betroffen: ['2'],
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'konflikt',
            meldung:
              'Das Feld steht nicht mehr auf dem Wert, den die Vorschau gezeigt hat.',
          },
        }),
    })
    const nachladen = zeige()

    await fuelleAus()

    expect(await screen.findByText(texte.konflikt)).toBeInTheDocument()
    expect(screen.getByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )

    // **Die Karte im Konflikt steht seit UXR-06 vorn** (#1914): die Tafel
    // zeigt zehn Zeilen, und dahinter sähe niemand, welche es ist
    // (`massenfalte.test.tsx`). Bis dahin stand sie an ihrer Stelle.
    const zeilen = within(screen.getByRole('table')).getAllByRole('row')
    expect(zeilen[1]).toHaveTextContent('Eisblitz')
    expect(zeilen[1]).toHaveTextContent(texte.inzwischenGeaendert)
    expect(zeilen[2]).not.toHaveTextContent(texte.inzwischenGeaendert)
    expect(screen.getByRole('button', { name: texte.anwenden })).toBeDisabled()
    expect(nachladen).not.toHaveBeenCalled()
  })

  it('räumt die Vorschau, wenn Feld oder Wert wechseln', async () => {
    belegeApi(mitVorschau)
    zeige()

    await vorschauen()
    await userEvent.type(screen.getByLabelText(texte.wert), '!')

    expect(screen.queryByRole('table')).not.toBeInTheDocument()
    expect(screen.getAllByRole('listitem')).toHaveLength(2)

    await userEvent.click(screen.getByRole('button', { name: texte.vorschau }))
    await screen.findByRole('table')
    await waehleFeld('')

    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  // **Ein Wert für viele Karten wird nicht von selbst berichtigt** — G03
  // (#2250), dieselbe Regel wie im Feld einer Karte (`Feldeingabe`).
  it('berichtigt den Wert nicht von selbst', async () => {
    belegeApi(feldwerte)
    zeige()

    await waehleFeld('typ')

    expect(screen.getByLabelText(texte.wert)).toHaveAttribute(
      'autocorrect',
      'off',
    )
    expect(screen.getByLabelText(texte.wert)).toHaveAttribute(
      'autocapitalize',
      'off',
    )
  })

  it('meldet eine abgelehnte Vorschau', async () => {
    belegeApi({
      ...feldwerte,
      FelderVorschau: () => Promise.reject(new Error('abgebrochen')),
    })
    zeige()

    await waehleFeld('typ')
    await userEvent.type(screen.getByLabelText(texte.wert), 'Artefakt')
    await userEvent.click(screen.getByRole('button', { name: texte.vorschau }))

    expect(await screen.findByRole('alert')).toHaveTextContent('abgebrochen')
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('meldet ein abgelehntes Anwenden ohne nachzuladen', async () => {
    belegeApi({
      ...mitVorschau,
      FelderSetzenMitErwartung: () => Promise.reject(new Error('abgebrochen')),
    })
    const nachladen = zeige()

    await fuelleAus()

    expect(await screen.findByRole('alert')).toHaveTextContent('abgebrochen')
    expect(nachladen).not.toHaveBeenCalled()
    // Die Vorschau bleibt stehen — sie ist nicht falsch, nur nicht geschrieben.
    expect(screen.getByRole('table')).toBeInTheDocument()
  })
})
