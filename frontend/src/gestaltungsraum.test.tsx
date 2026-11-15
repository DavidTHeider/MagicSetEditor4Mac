// **Der Gestaltungsraum legt alles aus UX-09 ans Wurzelelement** — und gibt
// der Einstellungsseite, was sie zeigt
// (#1681).
import { act, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Gestaltungsraum, useGestaltung } from './gestaltungsraum'
import { Nutzereinstellungsraum } from './nutzereinstellungsraum'
import { Sprachraum } from './sprache/kontext'
import { belegeApi, raeumeApi } from './test/api'
import { probeprofil } from './test/profil'

const wurzel = document.documentElement

afterEach(() => {
  raeumeApi()
  for (const name of [
    'data-farbwelt',
    'data-hintergrund',
    'data-kontrast',
    'data-bewegung',
    'data-erscheinungsbild',
    'style',
  ]) {
    wurzel.removeAttribute(name)
  }
  delete (globalThis as { runtime?: unknown }).runtime
})

// Anzeige zeigt, was der Raum der Einstellungsseite gibt.
function Anzeige() {
  const stand = useGestaltung()

  return (
    <output>
      <span>{`bild:${stand?.fensterbild.adresse ?? 'kein Raum'}`}</span>
      <span>{`befunde:${String(stand?.befunde.length)}`}</span>
      <span>{`schmuck:${String(stand?.schmuckNurOben)}`}</span>
      <span>{`fehler:${stand?.ruecksetzfehler ?? 'keiner'}`}</span>
    </output>
  )
}

function zeige(werte: Record<string, string>, neuLesen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Nutzereinstellungsraum>
        <Gestaltungsraum
          einstellungen={{ wert: (schluessel) => werte[schluessel] ?? '' }}
          neuLesen={neuLesen}
        >
          <Anzeige />
        </Gestaltungsraum>
      </Nutzereinstellungsraum>
    </Sprachraum>,
  )

  return neuLesen
}

function belege(extra: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    GezeigtesProfil: vi.fn(() =>
      Promise.resolve({
        profil: probeprofil({
          id: '10',
          nutzerID: '7',
          name: 'Mein Profil',
          fassung: '1',
        }),
      }),
    ),
    Nutzereinstellungen: vi.fn(() =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: 'barrierefreiheit.kontrast',
            wert: 'system',
            werte: ['system', 'hoch'],
          },
        ],
      }),
    ),
    ...extra,
  })
}

describe('Der Gestaltungsraum', () => {
  it('legt Farbwelt und Fensterbild ans Wurzelelement — das Bild mit Stand', async () => {
    belege({
      Fensterbild: vi.fn(() =>
        Promise.resolve({ adresse: '/fensterbild/7.png', fehler: null }),
      ),
    })
    zeige({ farbwelt: 'moos', hintergrund: 'kacheln' })

    await screen.findByText('bild:/fensterbild/7.png')
    await waitFor(() => {
      expect(wurzel).toHaveAttribute('data-hintergrund', 'kacheln')
    })
    expect(wurzel).toHaveAttribute('data-farbwelt', 'moos')
    expect(wurzel.style.getPropertyValue('--mse-fensterbild')).toBe(
      'url("/fensterbild/7.png?stand=0")',
    )
  })

  // **Hoher Kontrast sticht alles hier** — auch eine Farbwelt, die gewählt
  // ist, und ein Bild, das daliegt.
  it('setzt bei hohem Kontrast weder Farbwelt noch Bild', async () => {
    belege({
      Nutzereinstellungen: vi.fn(() =>
        Promise.resolve({
          einstellungen: [
            {
              schluessel: 'barrierefreiheit.kontrast',
              wert: 'hoch',
              werte: ['system', 'hoch'],
            },
          ],
        }),
      ),
      Fensterbild: vi.fn(() =>
        Promise.resolve({ adresse: '/fensterbild/7.png', fehler: null }),
      ),
    })
    zeige({ farbwelt: 'klar', hintergrund: 'fuellen' })

    await screen.findByText('bild:/fensterbild/7.png')
    await waitFor(() => {
      expect(wurzel).toHaveAttribute('data-kontrast', 'hoch')
    })
    expect(wurzel).not.toHaveAttribute('data-farbwelt')
    expect(wurzel).not.toHaveAttribute('data-hintergrund')
  })

  it('nennt eine Schmuckschrift, die an den Überschriften bleibt', async () => {
    belege()
    zeige({ schriftbereich: 'ganz', schriftart: 'kalligrafie' })

    expect(await screen.findByText('schmuck:true')).toBeInTheDocument()
    expect(wurzel.style.getPropertyValue('--font-text')).toBe('')
  })

  // **Die Sperre** — unter jsdom kennt der Stil nur, was am Element steht;
  // eine eigene Textfarbe gleich dem eigenen Grund ist unlesbar.
  it('gibt die Befunde unlesbarer eigener Farben weiter und setzt sie nicht', async () => {
    belege()
    zeige({ gestaltung: '--mse-grund: #ffffff;\n--mse-text: #fefefe;' })

    expect(await screen.findByText('befunde:1')).toBeInTheDocument()
    expect(wurzel.style.getPropertyValue('--mse-text')).toBe('')
  })

  // **Mantine folgt dem Erscheinungsbild** — Gegenprobe A03 im Gast unter
  // macOS 15.2 (#2217): Mantine stand im dunklen Erscheinungsbild auf
  // »light«, und was nicht über die Merkmale aus `mantinemerkmale.css` lief —
  // Symbolfarben, Hover —, kam aus der hellen Palette.
  it.each([
    ['dunkel', 'dark'],
    ['hell', 'light'],
    ['', 'light'],
  ])('stellt Mantine bei »%s« auf %s', async (wahl, schema) => {
    belege()
    zeige({ erscheinungsbild: wahl })

    await waitFor(() => {
      expect(wurzel).toHaveAttribute('data-mantine-color-scheme', schema)
    })
  })

  it('stellt Mantine dunkel, wenn das System dunkel ist', async () => {
    const frage = vi
      .spyOn(window, 'matchMedia')
      .mockImplementation((medien: string) => ({
        matches: medien === '(prefers-color-scheme: dark)',
        media: medien,
        onchange: null,
        addListener: () => undefined,
        removeListener: () => undefined,
        addEventListener: () => undefined,
        removeEventListener: () => undefined,
        dispatchEvent: () => false,
      }))

    try {
      belege()
      zeige({ erscheinungsbild: 'system' })

      await waitFor(() => {
        expect(wurzel).toHaveAttribute('data-mantine-color-scheme', 'dark')
      })
    } finally {
      frage.mockRestore()
    }
  })

  it('zeigt ohne Raum keinen Stand', () => {
    render(<Anzeige />)

    expect(screen.getByText('bild:kein Raum')).toBeInTheDocument()
  })
})

describe('Darstellung zurücksetzen aus der Menüleiste', () => {
  // horcher fängt, was der Raum an der Laufzeit anmeldet.
  function mitLaufzeit() {
    const horcher = new Map<string, (...daten: unknown[]) => void>()
    ;(globalThis as { runtime?: unknown }).runtime = {
      EventsOn: (name: string, rueckruf: (...daten: unknown[]) => void) => {
        horcher.set(name, rueckruf)

        return () => {
          horcher.delete(name)
        }
      },
    }

    return horcher
  }

  it('führt die Layoutrücksetzung aus und lässt neu lesen', async () => {
    const horcher = mitLaufzeit()
    const zuruecksetzen = vi.fn(() =>
      Promise.resolve({
        art: 'layout',
        zurueckgesetzt: 4,
        beiseite: '',
        fehler: null,
      }),
    )
    belege({
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
      Zuruecksetzen: zuruecksetzen,
    })
    const neuLesen = zeige({ farbwelt: 'klar' })

    await waitFor(() => {
      expect(horcher.has('menue:darstellung-zuruecksetzen')).toBe(true)
    })
    act(() => {
      horcher.get('menue:darstellung-zuruecksetzen')?.()
    })

    await waitFor(() => {
      expect(neuLesen).toHaveBeenCalledTimes(1)
    })
    expect(zuruecksetzen).toHaveBeenCalledWith('layout', '', '', '', '', 'r1')
  })

  it('meldet, woran das Zurücksetzen scheiterte, und liest nicht neu', async () => {
    const horcher = mitLaufzeit()
    belege({
      Ruecksetzungsvorschau: vi.fn(() =>
        Promise.resolve({
          revision: '',
          art: 'layout',
          betroffen: [],
          hindernisse: [],
          bereit: false,
          nameNoetig: false,
          nachweisNoetig: false,
          sicherungEmpfohlen: false,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Der Bestand antwortet nicht.',
          },
        }),
      ),
    })
    const neuLesen = zeige({})

    await waitFor(() => {
      expect(horcher.has('menue:darstellung-zuruecksetzen')).toBe(true)
    })
    act(() => {
      horcher.get('menue:darstellung-zuruecksetzen')?.()
    })

    expect(await screen.findByText(/^fehler:(?!keiner)/)).toBeInTheDocument()
    expect(neuLesen).not.toHaveBeenCalled()
  })
})
