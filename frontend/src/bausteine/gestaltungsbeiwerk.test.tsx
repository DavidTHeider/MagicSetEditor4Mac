// **Fensterbild und eigene Farben auf der Einstellungsseite** — UX-09
// (#1681).
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Eigenfarben, Gestaltungsbeiwerk } from './Gestaltungsbeiwerk'
import { Gestaltungsraum } from '../gestaltungsraum'
import { Nutzereinstellungsraum } from '../nutzereinstellungsraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { probeprofil } from '../test/profil'

import type { Api } from '../api'

const worte = kataloge.de.seiten.einstellungen.gestaltung

afterEach(() => {
  raeumeApi()
  document.documentElement.removeAttribute('style')
  document.documentElement.removeAttribute('data-hintergrund')
})

function zeige({
  adresse = '',
  modus = 'fuellen',
  gestaltung = '',
  ohneGestaltung = false,
  mehr = {},
  api = {},
}: {
  adresse?: string
  modus?: string
  gestaltung?: string
  ohneGestaltung?: boolean
  mehr?: Record<string, string>
  api?: Partial<Api>
}) {
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
    Nutzereinstellungen: vi.fn(() => Promise.resolve({ einstellungen: [] })),
    Fensterbild: vi.fn(() => Promise.resolve({ adresse, fehler: null })),
    ...api,
  })
  const setze = vi.fn()

  const { container } = render(
    <Sprachraum sprache="de">
      <Nutzereinstellungsraum>
        <Gestaltungsraum
          einstellungen={{
            wert: (schluessel) =>
              ({ hintergrund: modus, gestaltung, ...mehr })[schluessel] ?? '',
          }}
          neuLesen={vi.fn()}
        >
          <Gestaltungsbeiwerk hintergrund={modus} />
          <Eigenfarben
            gestaltung={
              ohneGestaltung
                ? undefined
                : {
                    schluessel: 'gestaltung',
                    wert: gestaltung,
                    wirksam: gestaltung,
                    werte: [],
                    bereich: 'anwendung',
                  }
            }
            angehalten={false}
            setze={setze}
          />
        </Gestaltungsraum>
      </Nutzereinstellungsraum>
    </Sprachraum>,
  )

  return { setze, container }
}

// bereiterKnopf wartet, bis »Bild wählen« bedienbar ist — **vorher weiß er
// nicht, für wen er wählt**, und hält deshalb an.
async function bereiterKnopf(): Promise<HTMLElement> {
  const knopf = await screen.findByRole('button', { name: worte.waehlen })
  await waitFor(() => {
    expect(knopf).toBeEnabled()
  })

  return knopf
}

describe('Das Fensterbild', () => {
  // **Ein Knopf, der nichts tut, ist ein Befund** (§6): solange der Mensch
  // nicht bekannt ist, ginge ein Klick ins Leere — also hält der Knopf an,
  // statt zu schweigen. Unter Last kam ein Klick genau dorthin.
  it('hält »Bild wählen« an, bis der Mensch bekannt ist', async () => {
    let geben: (wert: {
      profil: ReturnType<typeof probeprofil>
    }) => void = () => undefined
    const waehlen = vi.fn(() => Promise.resolve({ adresse: '', fehler: null }))
    zeige({
      api: {
        GezeigtesProfil: vi.fn(
          () =>
            new Promise<{ profil: ReturnType<typeof probeprofil> }>((los) => {
              geben = los
            }),
        ),
        FensterbildWaehlen: waehlen,
      },
    })

    const knopf = await screen.findByRole('button', { name: worte.waehlen })
    expect(knopf).toBeDisabled()

    geben({
      profil: probeprofil({
        id: '10',
        nutzerID: '7',
        name: 'Mein Profil',
        fassung: '1',
      }),
    })

    await userEvent.click(await bereiterKnopf())
    await waitFor(() => {
      expect(waehlen).toHaveBeenCalledWith('7')
    })
  })

  it('erklärt, dass es noch keines gibt, und öffnet den Dialog', async () => {
    const waehlen = vi.fn(() =>
      Promise.resolve({ adresse: '/fensterbild/7.png', fehler: null }),
    )
    zeige({ api: { FensterbildWaehlen: waehlen } })

    expect(await screen.findByText(worte.keinBild)).toBeInTheDocument()
    await userEvent.click(await bereiterKnopf())

    await waitFor(() => {
      expect(waehlen).toHaveBeenCalledWith('7')
    })
    expect(
      await screen.findByRole('img', { name: worte.vorschau }),
    ).toHaveStyle({ backgroundRepeat: 'no-repeat' })
  })

  // **Die Vorschau zeigt die gewählte Skalierung** — und „Bild entfernen"
  // ist die Rücksetzung am Bild selbst.
  it('zeigt die Vorschau mit ihrer Skalierung und nimmt das Bild wieder weg', async () => {
    const entfernen = vi.fn(() =>
      Promise.resolve({ adresse: '', fehler: null }),
    )
    zeige({
      adresse: '/fensterbild/7.png',
      modus: 'kacheln',
      api: { FensterbildEntfernen: entfernen },
    })

    const vorschau = await screen.findByRole('img', { name: worte.vorschau })
    expect(vorschau.style.backgroundImage).toContain('/fensterbild/7.png')
    expect(vorschau).toHaveStyle({ backgroundRepeat: 'repeat' })

    await userEvent.click(screen.getByRole('button', { name: worte.entfernen }))

    await waitFor(() => {
      expect(entfernen).toHaveBeenCalledWith('7')
    })
    expect(await screen.findByText(worte.keinBild)).toBeInTheDocument()
  })

  it('sagt, wenn das Bild daliegt, das Fensterbild aber aus ist', async () => {
    zeige({ adresse: '/fensterbild/7.png', modus: 'aus' })

    expect(
      await screen.findByText(/das Fensterbild steht aber auf/),
    ).toBeInTheDocument()
  })

  it('nennt einen Fehler beim Wählen am Bild', async () => {
    zeige({
      api: {
        FensterbildWaehlen: vi.fn(() =>
          Promise.resolve({
            adresse: '',
            fehler: {
              code: 'validierung' as const,
              schluessel: 'validierung',
              meldung: 'Das ist kein Bild.',
            },
          }),
        ),
      },
    })

    await userEvent.click(await bereiterKnopf())

    // **Übersetzt über den Katalog** (§9): die Seite zeigt den Satz zum
    // Schlüssel, nicht den deutschen Satz aus Go.
    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.validierung,
    )
  })
})

describe('Die eigenen Farben', () => {
  // **Der Befund steht am Feld und nennt die Zeile** (§6, §8) — geschrieben
  // wird nichts.
  it('nennt eine Zeile, die die Form verfehlt, und schreibt nicht', async () => {
    const { setze } = zeige({})

    const feld = await screen.findByLabelText(worte.merkmale)
    await userEvent.type(feld, '--mse-grund: #111;{Enter}color: red;')
    await userEvent.click(
      screen.getByRole('button', { name: worte.uebernehmen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('Zeile 2:')
    expect(feld).toHaveAttribute('aria-invalid', 'true')
    expect(setze).not.toHaveBeenCalled()

    await userEvent.type(feld, 'x')
    expect(feld).toHaveAttribute('aria-invalid', 'false')
  })

  it('schreibt eine Form, die passt', async () => {
    const { setze } = zeige({})

    const feld = await screen.findByLabelText(worte.merkmale)
    await userEvent.type(feld, '--mse-marke: #4e2f16;')
    await userEvent.click(
      screen.getByRole('button', { name: worte.uebernehmen }),
    )

    expect(setze).toHaveBeenCalledWith('--mse-marke: #4e2f16;')
  })

  // **Unlesbar gilt nicht** — und die Seite sagt, welches Paar es war, mit
  // Zahl und Schwelle in der Gegend des Fensters.
  it('nennt das Paar, an dem eigene Farben scheitern', async () => {
    zeige({ gestaltung: '--mse-grund: #ffffff;\n--mse-text: #fefefe;' })

    expect(await screen.findByText(worte.nichtLesbar)).toBeInTheDocument()
    expect(
      screen.getByText(
        '--mse-text auf --mse-grund: 1,01 : 1, verlangt 4,50 : 1',
      ),
    ).toBeInTheDocument()
  })

  // **Bei hohem Kontrast gilt keine eigene Farbe** — und die Seite sagt es,
  // statt „Angewandt." zu melden (Zweitblick über Phase 6.3, G1).
  it('sagt bei hohem Kontrast, dass eigene Farben ruhen', async () => {
    zeige({
      gestaltung: '--mse-grund: #ffffff;',
      api: {
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
      },
    })

    expect(await screen.findByText(worte.ruhtBeiKontrast)).toBeInTheDocument()
    expect(screen.queryByText(worte.angewandt)).toBeNull()
  })

  it('sagt, dass lesbare eigene Farben gelten', async () => {
    zeige({ gestaltung: '--mse-grund: #ffffff;' })

    expect(await screen.findByText(worte.angewandt)).toBeInTheDocument()
  })

  it('zeigt kein Feld, wenn es die Einstellung nicht gibt', async () => {
    zeige({ ohneGestaltung: true })

    await screen.findByText(worte.keinBild)
    expect(screen.queryByLabelText(worte.merkmale)).toBeNull()
  })

  it('hält die Regeln der Barrierefreiheit', async () => {
    const { container } = zeige({
      adresse: '/fensterbild/7.png',
      gestaltung: '--mse-grund: #ffffff;',
    })

    await screen.findByRole('img', { name: worte.vorschau })
    expect(await verstoesse(container)).toEqual([])
  })
})

describe('Was der Raum sonst weitergibt', () => {
  afterEach(() => {
    delete (globalThis as { runtime?: unknown }).runtime
  })

  // **Kalligrafie über einem Absatz ist keine Gestaltung** (AK-03) — die
  // Seite sagt, warum die Wahl nur oben gilt.
  it('sagt, dass eine Schmuckschrift an den Überschriften bleibt', async () => {
    zeige({ mehr: { schriftbereich: 'ganz', schriftart: 'kalligrafie' } })

    expect(await screen.findByText(worte.schmuckNurOben)).toBeInTheDocument()
  })

  // **Scheitert das Zurücksetzen aus der Menüleiste**, steht der Grund hier —
  // die Leiste selbst hat keinen Ort dafür.
  it('nennt, woran das Zurücksetzen aus der Leiste scheiterte', async () => {
    const horcher = new Map<string, () => void>()
    ;(globalThis as { runtime?: unknown }).runtime = {
      EventsOn: (name: string, rueckruf: () => void) => {
        horcher.set(name, rueckruf)

        return () => {
          horcher.delete(name)
        }
      },
    }
    zeige({
      api: {
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
      },
    })

    await waitFor(() => {
      expect(horcher.has('menue:darstellung-zuruecksetzen')).toBe(true)
    })
    act(() => {
      horcher.get('menue:darstellung-zuruecksetzen')?.()
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      worte.zuruecksetzenGescheitert,
    )
  })
})

describe('Ohne Raum', () => {
  it('zeichnet nichts', () => {
    render(
      <Sprachraum sprache="de">
        <Gestaltungsbeiwerk hintergrund="aus" />
        <Eigenfarben
          gestaltung={{
            schluessel: 'gestaltung',
            wert: '',
            wirksam: '',
            werte: [],
            bereich: 'anwendung',
          }}
          angehalten={false}
          setze={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(screen.queryByText(worte.fensterbild)).toBeNull()
    expect(screen.queryByLabelText(worte.merkmale)).toBeNull()
  })
})
