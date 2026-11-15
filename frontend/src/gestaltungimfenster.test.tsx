// **Die Gestaltung an der ganzen Anwendung** — UX-09
// (#1681).
//
// **Gegen `App` und nicht gegen den Raum**, weil die Zusagen Nähte sind: das
// Zurücksetzen aus der Menüleiste endet erst, wenn die **Anwendung** neu
// liest, und die eigenen Farben gehen über die Einstellungsseite in den
// Bestand. Die Läufe am Raum und am Beiwerk prüfen je eine Hälfte.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, raeumeApi } from './test/api'

const texte = kataloge.de
const worte = texte.seiten.einstellungen.gestaltung

afterEach(() => {
  raeumeApi()
  delete (globalThis as { runtime?: unknown }).runtime
  for (const name of ['data-farbwelt', 'data-hintergrund', 'style']) {
    document.documentElement.removeAttribute(name)
  }
})

function eintrag(schluessel: string, wert: string, werte: string[] = []) {
  return { schluessel, wert, wirksam: wert, werte, bereich: 'anwendung' }
}

function einstellungen(farbwelt: string) {
  return () =>
    Promise.resolve({
      einstellungen: [
        eintrag('sprache', 'de', ['de', 'en']),
        eintrag('farbwelt', farbwelt, [
          'pergament',
          'klar',
          'schiefer',
          'moos',
        ]),
        eintrag('hintergrund', 'aus', [
          'aus',
          'fuellen',
          'einpassen',
          'kacheln',
        ]),
        eintrag('gestaltung', ''),
      ],
    })
}

describe('Die Gestaltung an der Anwendung', () => {
  // **Das Zurücksetzen endet mit einem neuen Lesen** — sonst stünde im
  // Fenster weiter, was der Bestand schon vergessen hat.
  it('liest nach »Darstellung zurücksetzen« aus der Leiste neu', async () => {
    const horcher = new Map<string, () => void>()
    ;(globalThis as { runtime?: unknown }).runtime = {
      EventsOn: (name: string, rueckruf: () => void) => {
        horcher.set(name, rueckruf)

        return () => {
          horcher.delete(name)
        }
      },
    }
    belegeApi({
      Einstellungen: vi
        .fn()
        .mockImplementationOnce(einstellungen('moos'))
        .mockImplementation(einstellungen('pergament')),
      Ruecksetzungsvorschau: vi.fn(() =>
        Promise.resolve({
          revision: 'r1',
          art: 'layout',
          betroffen: [],
          hindernisse: [],
          bereit: true,
          nameNoetig: false,
          nachweisNoetig: false,
          sicherungEmpfohlen: false,
          fehler: null,
        }),
      ),
      Zuruecksetzen: vi.fn(() =>
        Promise.resolve({
          art: 'layout',
          zurueckgesetzt: 1,
          beiseite: '',
          fehler: null,
        }),
      ),
    })

    render(<App />)
    await waitFor(() => {
      expect(document.documentElement.getAttribute('data-farbwelt')).toBe(
        'moos',
      )
    })

    await waitFor(() => {
      expect(horcher.has('menue:darstellung-zuruecksetzen')).toBe(true)
    })
    act(() => {
      horcher.get('menue:darstellung-zuruecksetzen')?.()
    })

    await waitFor(() => {
      expect(document.documentElement.hasAttribute('data-farbwelt')).toBe(false)
    })
  })

  // **Die eigenen Farben gehen über die Seite in den Bestand** — mit
  // demselben Schlüssel, den MCP und Kommandozeile schreiben.
  it('schreibt eigene Farben über die Einstellungsseite', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({}))
    belegeApi({ Einstellungen: einstellungen('klar'), EinstellungSetzen })

    render(<App />)
    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.einstellungen.titel,
      }),
    )
    await userEvent.click(
      await screen.findByRole('tab', {
        name: texte.seiten.einstellungen.gruppen.erscheinungsbild,
      }),
    )
    // **Seit #2465 in einem Abschnitt** — erst aufklappen.
    await userEvent.click(
      await screen.findByRole('button', { name: worte.merkmale }),
    )

    await userEvent.type(
      await screen.findByLabelText(worte.merkmale),
      '--mse-marke: #4e2f16;',
    )
    await userEvent.click(
      screen.getByRole('button', { name: worte.uebernehmen }),
    )

    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledWith(
        'gestaltung',
        '--mse-marke: #4e2f16;',
      )
    })
  })
})
