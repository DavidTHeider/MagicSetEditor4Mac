// **Die Kommandozeile einrichten** — OPS-03 (#1720).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Befehlszeile } from './Befehlszeile'
import type { BefehlsvorschauAntwort } from '../antworten'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.technik

function vorschau(
  stand: string,
  abweichend: Partial<BefehlsvorschauAntwort> = {},
): BefehlsvorschauAntwort {
  return {
    verzeichnis: '/Users/ich/.local/bin',
    imPfad: false,
    pfadzeile: 'export PATH="/Users/ich/.local/bin":"$PATH"',
    pruefbefehl: 'mse meta',
    eintraege: [
      {
        name: 'mse',
        ziel: '/Users/ich/.local/bin/mse',
        quelle: '/A/mse',
        stand,
        verdeckt: '/opt/bin/mse',
      },
      {
        name: 'mse-mcp',
        ziel: '/Users/ich/.local/bin/mse-mcp',
        quelle: '/A/mse-mcp',
        stand: 'unbekannt',
        verdeckt: '',
      },
    ],
    fehler: null,
    ...abweichend,
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Befehlszeile />
    </Sprachraum>,
  )
}

it('nennt Verzeichnis, Stand, Verdeckung, die Pfadzeile und den Prüfbefehl', async () => {
  belegeApi({ Befehlszeile: () => Promise.resolve(vorschau('frei')) })
  zeige()

  expect(
    await screen.findByText(eigene.befehlszeileStaende.frei ?? '', {
      exact: false,
    }),
  ).toBeInTheDocument()
  expect(screen.getByText('unbekannt', { exact: false })).toBeInTheDocument()
  expect(screen.getByText(/opt\/bin\/mse/)).toBeInTheDocument()
  expect(
    screen.getByText('export PATH="/Users/ich/.local/bin":"$PATH"'),
  ).toBeInTheDocument()
  expect(screen.getByText('mse meta')).toBeInTheDocument()
})

it('richtet ein und entfernt — und zeigt den neuen Stand', async () => {
  const befehlszeile = vi.fn((handlung: string) =>
    Promise.resolve(
      vorschau(handlung === 'installieren' ? 'eigen' : 'frei', {
        imPfad: true,
        pfadzeile: '',
      }),
    ),
  )
  belegeApi({ Befehlszeile: befehlszeile })
  zeige()
  await screen.findByText(eigene.befehlszeileStaende.frei ?? '', {
    exact: false,
  })

  await userEvent.click(
    screen.getByRole('button', { name: eigene.befehlszeileEinrichten }),
  )
  expect(
    await screen.findByText(eigene.befehlszeileStaende.eigen ?? '', {
      exact: false,
    }),
  ).toBeInTheDocument()
  expect(befehlszeile).toHaveBeenCalledWith('installieren')
  expect(
    screen.queryByText(eigene.befehlszeileNichtImPfad, { exact: false }),
  ).not.toBeInTheDocument()

  await userEvent.click(
    screen.getByRole('button', { name: eigene.befehlszeileHerausnehmen }),
  )
  await waitFor(() => {
    expect(befehlszeile).toHaveBeenCalledWith('entfernen')
  })
})

it('zeigt einen Fehler beim Einrichten neben dem Stand', async () => {
  belegeApi({
    Befehlszeile: (handlung: string) =>
      handlung === ''
        ? Promise.resolve(vorschau('frei'))
        : Promise.resolve({
            ...vorschau('frei'),
            fehler: {
              code: 'validierung',
              schluessel: 'fehler.validierung',
              meldung: 'das Verzeichnis nicht',
              feld: '',
            },
          }),
  })
  zeige()
  await screen.findByText(eigene.befehlszeileStaende.frei ?? '', {
    exact: false,
  })

  await userEvent.click(
    screen.getByRole('button', { name: eigene.befehlszeileEinrichten }),
  )

  expect(await screen.findByRole('alert')).toBeInTheDocument()
})

// **Der angezeigte Aufruf lässt sich kopieren** — OPS-03.AK-04 (#1720): die
// Zeile fürs Shellprofil und der Prüfbefehl je mit eigenem Knopf, und
// »Kopiert.« steht an dem, der gedrückt wurde.
it('kopiert die Pfadzeile und den Prüfbefehl je für sich', async () => {
  belegeApi({ Befehlszeile: () => Promise.resolve(vorschau('frei')) })
  zeige()
  await screen.findByText('mse meta')

  const hand = userEvent.setup()
  const schreiben = vi.fn(() => Promise.resolve())
  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: { writeText: schreiben },
  })

  await hand.click(
    screen.getByRole('button', { name: eigene.befehlszeileKopierenPfad }),
  )
  expect(schreiben).toHaveBeenLastCalledWith(
    'export PATH="/Users/ich/.local/bin":"$PATH"',
  )
  expect(await screen.findAllByText(eigene.kopiert)).toHaveLength(1)

  await hand.click(
    screen.getByRole('button', { name: eigene.befehlszeileKopierenPruefung }),
  )
  expect(schreiben).toHaveBeenLastCalledWith('mse meta')
  expect(screen.getAllByText(eigene.kopiert)).toHaveLength(1)
})

// **Auch das Nein steht da** — ohne Erlaubnis für die Zwischenablage.
it('sagt, wenn Kopieren nicht ging', async () => {
  belegeApi({ Befehlszeile: () => Promise.resolve(vorschau('frei')) })
  zeige()
  await screen.findByText('mse meta')

  const hand = userEvent.setup()
  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: { writeText: () => Promise.reject(new Error('verweigert')) },
  })

  await hand.click(
    screen.getByRole('button', { name: eigene.befehlszeileKopierenPruefung }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    eigene.kopierfehler,
  )
})
