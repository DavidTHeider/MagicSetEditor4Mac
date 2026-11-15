// **Aufräumen, von der Seite aus** — Werkstück O8 von Phase 5.2
// (#1432).
//
// **Über die Einstellungsseite und nicht über den Haken unmittelbar**: geprüft
// wird der Weg, den jemand geht — nachsehen, wählen, aufräumen. Was das
// Bauteil zeigt, steht in `seiten/datenpflege.test.tsx`.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { zuAbschnitt } from './test/gebiete'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.datenpflege

const einBefund = {
  art: 'kartenwerteOhneKarte',
  klasse: 'sicher',
  anzahl: 3,
  revision: 'abc',
  beispiele: ['card_values:1:name'],
}

async function zumBestandsgebiet() {
  const nutzer = userEvent.setup()

  render(<App />)

  await nutzer.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.einstellungen.titel,
    }),
  )

  await zuAbschnitt(
    kataloge.de.seiten.einstellungen.gruppen.daten,
    kataloge.de.seiten.einstellungen.abschnitte.eingriffe,
  )
  await screen.findByRole('heading', { name: worte.titel })

  return nutzer
}

describe('Datenpflege', () => {
  it('sieht nach, räumt auf und zeigt die Bilanz', async () => {
    const Pflegen = vi.fn(() =>
      Promise.resolve({
        gesichert: '/wo/es/liegt.db',
        zeilen: [
          {
            art: 'kartenwerteOhneKarte',
            ausgang: 'erledigt',
            anzahl: 3,
            grund: '',
          },
        ],
        vorher: [einBefund],
        nachher: [],
        fehler: null,
      }),
    )

    belegeApi({
      ...belegeFenster('de'),
      Pflegevorschau: () =>
        Promise.resolve({
          erstellt: '2026-09-10T12:00:00Z',
          gueltigbis: '2026-09-10T12:15:00Z',
          befunde: [einBefund],
          fehler: null,
        }),
      Pflegen,
    })

    const nutzer = await zumBestandsgebiet()

    await nutzer.click(screen.getByRole('button', { name: worte.nachsehen }))

    await nutzer.click(
      await screen.findByRole('button', { name: worte.aufraeumen }),
    )

    await waitFor(() => {
      // **Der Zeitpunkt der Vorschau reist mit** — er ist die Frist.
      expect(Pflegen).toHaveBeenCalledWith('2026-09-10T12:00:00Z', [
        { art: 'kartenwerteOhneKarte', revision: 'abc' },
      ])
    })

    expect(await screen.findByText(/Weggeräumt \(3\)/)).toBeInTheDocument()

    // **Die Vorschau ist danach weg** — sie beschreibt einen Stand, den es
    // nicht mehr gibt.
    expect(
      screen.queryByRole('button', { name: worte.aufraeumen }),
    ).not.toBeInTheDocument()
  })

  // **Ein Fehler beim Pflegen lässt die Bilanz stehen** (§7, Regel 3):
  // gepflegt ist gepflegt, und ein Fehler beim Nachsehen danach darf die
  // Zeilen nicht verschlucken.
  it('meldet einen Fehler und zeigt trotzdem, was geschah', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Pflegevorschau: () =>
        Promise.resolve({
          erstellt: '2026-09-10T12:00:00Z',
          gueltigbis: '2026-09-10T12:15:00Z',
          befunde: [einBefund],
          fehler: null,
        }),
      Pflegen: () =>
        Promise.resolve({
          gesichert: '/wo/es/liegt.db',
          zeilen: [
            {
              art: 'kartenwerteOhneKarte',
              ausgang: 'erledigt',
              anzahl: 3,
              grund: '',
            },
          ],
          vorher: [einBefund],
          nachher: [einBefund],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'sicherungUnbrauchbar',
            meldung: 'roh',
          },
        }),
    })

    const nutzer = await zumBestandsgebiet()

    await nutzer.click(screen.getByRole('button', { name: worte.nachsehen }))
    await nutzer.click(
      await screen.findByRole('button', { name: worte.aufraeumen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.sicherungUnbrauchbar,
    )

    expect(screen.getByText(/Weggeräumt \(3\)/)).toBeInTheDocument()
  })

  it('meldet einen Fehler beim Nachsehen', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Pflegevorschau: () =>
        Promise.resolve({
          erstellt: '',
          gueltigbis: '',
          befunde: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'sicherungUnbrauchbar',
            meldung: 'roh',
          },
        }),
    })

    const nutzer = await zumBestandsgebiet()

    await nutzer.click(screen.getByRole('button', { name: worte.nachsehen }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.sicherungUnbrauchbar,
    )
  })
})
