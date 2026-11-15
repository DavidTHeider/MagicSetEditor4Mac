// **Kontakte aus Apple Kontakte** — ID-04
// (#1703).
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KontaktDTO } from '../apikontakte'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { belege, bo, karten, kontakt, texte, zeige } from '../test/kontakte'

describe('Kontakte aus Apple Kontakte', () => {
  afterEach(raeumeApi)

  it('erklärt, was gelesen wird, und dass nichts geprüft ist', async () => {
    belege()
    await zeige()

    expect(
      screen.getByText(/Ablage › Exportieren › vCard exportieren/),
    ).toBeInTheDocument()
    expect(screen.getByText(/nicht geprüft/)).toBeInTheDocument()
    expect(await screen.findByText('Bo')).toBeInTheDocument()
  })

  it('übernimmt nur die Auswahl — Karten und Wege', async () => {
    const uebernommen = vi.fn<
      (auswahl: KontaktDTO[]) => Promise<{
        angelegt: number
        ersetzt: number
        fehler: null
      }>
    >(() => Promise.resolve({ angelegt: 1, ersetzt: 1, fehler: null }))
    belege({ KontakteUebernehmen: uebernommen })
    await zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.waehlen }))
    const vorschau = within(
      await screen.findByRole('region', { name: texte.vorschauTitel }),
    )

    expect(screen.getByText(/keinen Doppelpunkt/)).toBeInTheDocument()

    // **Neu und geändert sind gewählt, eine Vermutung erst auf Wunsch.**
    // **Unverändert steht nicht als Kästchen da** — seit UXR-16 (#1924) als
    // eine Zeile; wählbar war es nie, und mit einer ganzen Ausgabe aus
    // Kontakte waren das 150 abgeschaltete Kästchen samt ihren Wegen.
    const kaesten = vorschau.getAllByRole('checkbox')
    expect(kaesten).toHaveLength(3)
    expect(kaesten[0]).toBeChecked()
    expect(kaesten[1]).toBeChecked()
    expect(kaesten[2]).not.toBeChecked()
    expect(
      vorschau.getByRole('button', {
        name: einsetzen(texte.unveraendert, { anzahl: '1' }),
      }),
    ).toBeInTheDocument()
    expect(
      vorschau.getByText(
        einsetzen(texte.abgleich.vermutlich, { name: 'Ada Lovelace' }),
      ),
    ).toBeInTheDocument()

    // **Nur die Adresse, nicht die Nummer** (AK-01) — die Wege einer Karte
    // stehen seit UXR-16 hinter „Wege wählen".
    await userEvent.click(
      vorschau.getByRole('button', {
        name: `${texte.wegeWaehlen}: Ada Lovelace`,
      }),
    )
    await userEvent.click(
      vorschau.getByRole('checkbox', {
        name: einsetzen(texte.wegZeile, {
          art: texte.weg.telefon,
          wert: '+49 170 0',
        }),
      }),
    )
    await userEvent.click(
      vorschau.getByRole('button', {
        name: einsetzen(texte.uebernehmen, { anzahl: '2' }),
      }),
    )

    await waitFor(() => {
      expect(uebernommen).toHaveBeenCalled()
    })

    const auswahl = uebernommen.mock.calls[0]?.[0] ?? []
    expect(auswahl).toHaveLength(2)
    expect(auswahl[0]?.wege).toEqual([
      { art: 'email', wert: 'ada@example.org' },
    ])
    expect(auswahl[0]?.id).toBe(0)

    // **Ersetzt wird gegen die gelesene Fassung** (§7, Regel 2).
    expect(auswahl[1]).toMatchObject({ id: 3, fassung: 2, name: 'Bo Neu' })

    expect(
      await screen.findByText(
        einsetzen(texte.uebernommen, { angelegt: '1', ersetzt: '1' }),
      ),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('heading', { name: texte.vorschauTitel }),
    ).toBeNull()
  })

  it('lässt nach einem Fehler die Auswahl stehen', async () => {
    belege({
      KontakteUebernehmen: () =>
        Promise.resolve({
          angelegt: 0,
          ersetzt: 0,
          fehler: {
            code: 'konflikt',
            schluessel: 'doppelt',
            meldung: 'doppelt',
          },
        }),
    })
    await zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.waehlen }))
    await userEvent.click(
      await screen.findByRole('button', {
        name: einsetzen(texte.uebernehmen, { anzahl: '2' }),
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.doppelt,
    )
    expect(
      screen.getByRole('heading', { name: texte.vorschauTitel }),
    ).toBeInTheDocument()
  })

  it('ändert beim Abbruch des Dialogs nichts und verwirft auf Wunsch', async () => {
    let abbrechen = false
    belege({
      KontakteEinlesen: () =>
        Promise.resolve({
          karten: abbrechen ? [] : karten(),
          befunde: [],
          abgebrochen: abbrechen,
          fehler: null,
        }),
    })
    await zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.waehlen }))
    await screen.findByRole('heading', { name: texte.vorschauTitel })

    abbrechen = true
    await userEvent.click(screen.getByRole('button', { name: texte.waehlen }))
    await waitFor(() => {
      expect(screen.getByRole('button', { name: texte.waehlen })).toBeEnabled()
    })
    expect(
      screen.getByRole('heading', { name: texte.vorschauTitel }),
    ).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.verwerfen }))
    expect(
      screen.queryByRole('heading', { name: texte.vorschauTitel }),
    ).toBeNull()
  })

  it('meldet einen Fehler beim Einlesen', async () => {
    belege({
      KontakteEinlesen: () =>
        Promise.resolve({
          karten: [],
          befunde: [],
          abgebrochen: false,
          fehler: {
            code: 'nichtGefunden',
            schluessel: 'nichtGefunden',
            meldung: 'weg',
          },
        }),
    })
    await zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.waehlen }))
    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  it('zeigt ein Bild und entfernt nach Rückfrage', async () => {
    const geloescht = vi.fn(() => Promise.resolve({ fehler: null }))
    belege({
      Kontakte: () =>
        Promise.resolve({
          kontakte: [
            bo,
            kontakt({
              id: 6,
              fassung: 1,
              name: 'Di',
              bild: 'iVBORw0KGgo=',
              bildart: 'image/png',
            }),
          ],
          fehler: null,
        }),
      KontaktLoeschen: geloescht,
    })
    await zeige()

    expect(
      await screen.findByRole('img', {
        name: einsetzen(texte.bild, { name: 'Di' }),
      }),
    ).toBeInTheDocument()
    expect(
      screen.getByText(
        einsetzen(texte.wegZeile, {
          art: texte.weg.telefon,
          wert: '+49 170 0',
        }),
      ),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(texte.entfernenKontakt, { name: 'Bo' }),
      }),
    )
    await userEvent.click(
      within(
        screen.getByRole('group', {
          name: einsetzen(texte.entfernenFrage, { name: 'Bo' }),
        }),
      ).getByRole('button', {
        name: texte.abbrechen,
      }),
    )
    expect(geloescht).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(texte.entfernenKontakt, { name: 'Bo' }),
      }),
    )
    await userEvent.click(
      within(
        screen.getByRole('group', {
          name: einsetzen(texte.entfernenFrage, { name: 'Bo' }),
        }),
      ).getByRole('button', {
        name: texte.entfernen,
      }),
    )
    await waitFor(() => {
      expect(geloescht).toHaveBeenCalledWith(3, 2)
    })
  })

  it('meldet einen Fehler beim Entfernen und erklärt eine leere Liste', async () => {
    belege({
      Kontakte: () => Promise.resolve({ kontakte: [bo], fehler: null }),
      KontaktLoeschen: () =>
        Promise.resolve({
          fehler: { code: 'konflikt', schluessel: 'konflikt', meldung: 'x' },
        }),
    })
    await zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: einsetzen(texte.entfernenKontakt, { name: 'Bo' }),
      }),
    )
    await userEvent.click(
      within(
        screen.getByRole('group', {
          name: einsetzen(texte.entfernenFrage, { name: 'Bo' }),
        }),
      ).getByRole('button', {
        name: texte.entfernen,
      }),
    )
    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  it('erklärt eine leere Liste', async () => {
    belege({ Kontakte: () => Promise.resolve({ kontakte: [], fehler: null }) })
    await zeige()

    expect(await screen.findByText(texte.leer)).toBeInTheDocument()
  })
})
