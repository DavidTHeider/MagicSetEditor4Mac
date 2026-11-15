// **Was ein anderer Prozess abschaltet, verschwindet auch im offenen
// Fenster** — Befund 2 des Gesamtreviews über 0.6.0 (#1660).
//
// Die Einstellungen las das Fenster bei einem ersetzten Bestand und beim
// Profilwechsel; den fremden Stand (`useBestandsstand`) kannte nur die Ebene
// darunter. Schaltete die Kommandozeile `spielen` aus, standen Seitenleiste
// und Spielseite weiter da, bis ein anderer Anlass die Einstellungen neu
// las — und der erste Schritt dort wurde erst an der Go-Wache abgewiesen.
import { act, render, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import { App } from './App'
import { abstand } from './haken/useBestandsstand'
import { belegeApi, raeumeApi } from './test/api'

beforeEach(() => {
  vi.useFakeTimers({ shouldAdvanceTime: true })
})

afterEach(() => {
  vi.useRealTimers()
  raeumeApi()
})

// spielenSchalter ist der Eintrag, wie Go ihn schickt.
const spielenSchalter = (wert: string) => ({
  schluessel: 'schalter.spielen',
  wert,
  wirksam: wert,
  werte: ['an', 'aus'],
  bereich: 'anwendung',
})

// deutsch hält das Fenster in der Sprache dieses Laufs.
const deutsch = {
  schluessel: 'sprache',
  wert: 'de',
  wirksam: 'de',
  werte: ['de', 'en'],
  bereich: 'anwendung',
}

it('verbirgt einen Bereich, den ein anderer Prozess abschaltet', async () => {
  const bestand = { stand: 7, spielen: 'an' }

  belegeApi({
    Datenstand: () => Promise.resolve({ stand: bestand.stand, fehler: null }),
    Einstellungen: () =>
      Promise.resolve({
        einstellungen: [deutsch, spielenSchalter(bestand.spielen)],
      }),
  })

  render(<App />)

  const leiste = () =>
    within(document.getElementById('seitenleiste') ?? document.body)

  expect(
    await leiste().findByRole('button', { name: /Spielen/ }),
  ).toBeInTheDocument()

  // **Drüben schaltet jemand ab** — der Datenstand bewegt sich.
  bestand.spielen = 'aus'
  bestand.stand = 8

  await act(async () => {
    await vi.advanceTimersByTimeAsync(abstand * 2)
  })

  await waitFor(() => {
    expect(leiste().queryByRole('button', { name: /Spielen/ })).toBeNull()
  })

  // **Dass die Leiste noch zuhört** (§3): die anderen Bereiche stehen.
  expect(leiste().getByRole('button', { name: /Handeln/ })).toBeInTheDocument()
})
