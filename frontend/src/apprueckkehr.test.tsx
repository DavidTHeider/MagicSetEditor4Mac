// **Eine Rückkehr quittieren, von der Seite aus** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// **Über die Einstellungsseite und nicht über den Haken unmittelbar**: geprüft
// wird der Weg, den jemand geht — hinschauen, nachsehen, quittieren. Was das
// Bauteil zeigt, steht in `seiten/ruecksprung.test.tsx`.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { zuAbschnitt } from './test/gebiete'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'

import type { RueckkehrDTO } from './apirueckkehr'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.ruecksprung

const offene: RueckkehrDTO = {
  id: 7,
  zurueckgespielt: '2026-09-10T12:00:00Z',
  datenstand: '2026-09-01T08:30:00Z',
  quelle: 'bestand.msesicherung',
  klaerung: 'offen',
  geklaertAm: '',
  mengen: [
    {
      schluessel: 'scryfall:abc',
      zustand: 'nm',
      sprache: 'en',
      finish: 'nonfoil',
      menge: 3,
      name: 'Lightning Bolt',
      setcode: '',
      nummer: '',
    },
  ],
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

describe('Rückkehren', () => {
  it('quittiert eine offene Rückkehr und lädt danach neu', async () => {
    const RueckkehrQuittieren = vi.fn(() => Promise.resolve({ fehler: null }))

    // **Beim zweiten Lesen ist sie quittiert** — so verhält sich der Bestand
    // nach einem geglückten Schreiben, und nur so lässt sich sehen, dass
    // wirklich nachgeladen wurde (§7, Regel 3).
    let gelesen = 0

    belegeApi({
      ...belegeFenster('de'),
      Rueckkehren: () => {
        gelesen += 1

        return Promise.resolve({
          rueckkehren: [
            gelesen === 1
              ? offene
              : {
                  ...offene,
                  klaerung: 'quittiert',
                  geklaertAm: '2026-09-10T13:00:00Z',
                },
          ],
          fehler: null,
        })
      },
      RueckkehrQuittieren,
    })

    const nutzer = await zumBestandsgebiet()

    await nutzer.click(
      await screen.findByRole('button', { name: worte.quittieren }),
    )

    await waitFor(() => {
      expect(RueckkehrQuittieren).toHaveBeenCalledWith(7)
    })

    // **Gewartet wird auf die Wirkung** und nicht auf den Ruf: der Knopf
    // verschwindet erst, wenn der zweite Lauf angekommen ist.
    await waitFor(() => {
      expect(
        screen.queryByRole('button', { name: worte.quittieren }),
      ).not.toBeInTheDocument()
    })

    expect(gelesen).toBe(2)
  })

  // **Nach einem gescheiterten Schreiben wird nicht nachgeladen** (§7,
  // Regel 3): sonst überschriebe ein Lesefehler die Meldung, die sagt, was zu
  // tun ist — und die Rückkehr bliebe offen, ohne dass jemand erführe, warum.
  it('meldet einen Fehler und lässt die Rückkehr stehen', async () => {
    let gelesen = 0

    belegeApi({
      ...belegeFenster('de'),
      Rueckkehren: () => {
        gelesen += 1

        return Promise.resolve({ rueckkehren: [offene], fehler: null })
      },
      RueckkehrQuittieren: () =>
        Promise.resolve({
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'sicherungUnbrauchbar',
            meldung: 'roh',
          },
        }),
    })

    const nutzer = await zumBestandsgebiet()

    await nutzer.click(
      await screen.findByRole('button', { name: worte.quittieren }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.sicherungUnbrauchbar,
    )

    expect(
      screen.getByRole('button', { name: worte.quittieren }),
    ).toBeInTheDocument()

    expect(gelesen).toBe(1)
  })
})
