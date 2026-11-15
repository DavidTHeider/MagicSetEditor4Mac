// **Ein Leistungsprofil, gezielt** — OPS-05.AK-03
// (#1722):
// eine Dauer wählen, aufzeichnen, und danach steht da, wo es liegt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Leistungsprofil } from './Leistungsprofil'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { LeistungsprofilAntwort } from '../apimetriken'

const texte = kataloge.de.seiten.technik.leistungsprofil

afterEach(raeumeApi)

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Leistungsprofil />
    </Sprachraum>,
  )
}

describe('Das Leistungsprofil', () => {
  it('zeichnet über die gewählte Dauer auf und nennt die Datei', async () => {
    const aufzeichnen = vi.fn(() =>
      Promise.resolve({
        datei: '/Bestand/Leistungsprofile/mse-20260926-120000.000.pprof',
        sekunden: 5,
        bytes: 2048,
        fehler: null,
      }),
    )
    belegeApi({ Leistungsprofil: aufzeichnen })
    zeige()

    const dauer = screen.getByRole('spinbutton', { name: texte.sekunden })
    await userEvent.clear(dauer)
    await userEvent.type(dauer, '5')
    await userEvent.click(
      screen.getByRole('button', { name: texte.aufzeichnen }),
    )

    expect(
      await screen.findByText(/mse-20260926-120000\.000\.pprof/),
    ).toBeInTheDocument()
    expect(aufzeichnen).toHaveBeenCalledWith(5)
    expect(screen.getByText(/2,0 KB/)).toBeInTheDocument()
  })

  it('zeigt einen Fehler am Knopf', async () => {
    belegeApi({
      Leistungsprofil: vi.fn(() =>
        Promise.resolve({
          datei: '',
          sekunden: 0,
          bytes: 0,
          fehler: {
            code: 'konflikt' as const,
            schluessel: '',
            meldung:
              'Es läuft schon ein Leistungsprofil — warte, bis es fertig ist.',
          },
        }),
      ),
    })
    zeige()

    await userEvent.click(
      screen.getByRole('button', { name: texte.aufzeichnen }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  it('sperrt den Knopf, solange es läuft', async () => {
    let fertig: (wert: LeistungsprofilAntwort) => void = () => undefined
    belegeApi({
      Leistungsprofil: vi.fn(
        () =>
          new Promise<LeistungsprofilAntwort>((erledigt) => {
            fertig = erledigt
          }),
      ),
    })
    zeige()

    const knopf = screen.getByRole('button', { name: texte.aufzeichnen })
    await userEvent.click(knopf)

    await waitFor(() => {
      expect(knopf).toBeDisabled()
    })
    expect(screen.getByText(texte.laeuft)).toBeInTheDocument()

    fertig({ datei: '/x.pprof', sekunden: 10, bytes: 1, fehler: null })

    await waitFor(() => {
      expect(knopf).toBeEnabled()
    })
  })

  // **Die Grenze steht am Feld** — eine Null wird angezeigt und nicht
  // abgeschickt.
  it('hält eine Dauer außerhalb der Grenze am Feld an', async () => {
    const aufzeichnen = vi.fn()
    belegeApi({ Leistungsprofil: aufzeichnen })
    zeige()

    const dauer = screen.getByRole('spinbutton', { name: texte.sekunden })
    await userEvent.clear(dauer)
    await userEvent.type(dauer, '0')

    expect(dauer).toHaveAttribute('aria-invalid', 'true')
    expect(dauer).toHaveAccessibleDescription(texte.grenze)
    expect(
      screen.getByRole('button', { name: texte.aufzeichnen }),
    ).toBeDisabled()
    expect(aufzeichnen).not.toHaveBeenCalled()
  })
})
