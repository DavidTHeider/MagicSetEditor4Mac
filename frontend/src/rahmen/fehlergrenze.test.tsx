// **Ein Fehler beim Zeichnen nimmt nicht das ganze Fenster mit** — F01
// (#2340),
// Befund A-01.
//
// Bis hierher gab es im Fenster keine Fehlergrenze: ein Wurf beim Zeichnen
// eines Bereichs hängte den ganzen Baum ab, und das Fenster stand leer — ohne
// Satz, ohne Rückweg, ohne Diagnose. Geprüft wird der Bereich (Rahmen und
// Seitenleiste bleiben), die Wurzel (ein Satz statt eines leeren Fensters),
// der Fokus auf dem Rückweg, die Meldung ans Protokoll und die Naht der
// Hüllenprobe.
import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { seitenschluessel, type Seitenschluessel } from '../navigation'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { Rahmen } from './Rahmen'
import { Wurzelgrenze } from './Wurzelgrenze'

const eigene = kataloge.de.allgemein.fehlergrenze

afterEach(() => {
  raeumeApi()
  delete (globalThis as { __mseProbewurf?: string }).__mseProbewurf
  sessionStorage.clear()
  vi.restoreAllMocks()
})

// wirft ist eine Seite, die beim Zeichnen wirft, solange `zustand.wirft` gilt.
const zustand = { wirft: true }
function Wirft() {
  if (zustand.wirft) {
    throw new Error('Cannot read properties of undefined')
  }

  return <p>{kataloge.de.seiten.start.was}</p>
}

function zeige(aktuell: Seitenschluessel, seite = <Wirft />) {
  const melden = vi.fn(() => Promise.resolve({ fehler: null }))
  belegeApi({ FensterfehlerMelden: melden })
  const ansicht = render(
    <Sprachraum sprache="de">
      <Rahmen aktuell={aktuell} wechseln={vi.fn()}>
        {seite}
      </Rahmen>
    </Sprachraum>,
  )

  return { melden, ansicht }
}

describe('Die Grenze eines Bereichs', () => {
  it('zeigt Satz, Kennung und Rückweg — und der Rahmen bleibt', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = true
    const { melden } = zeige('sets')

    const flaeche = await screen.findByRole('alert', {
      name: eigene.bereichTitel,
    })
    expect(flaeche).toHaveTextContent(/[0-9a-f]{4}-[0-9a-f]{4}/)
    expect(screen.getByRole('button', { name: eigene.neuLaden })).toHaveFocus()
    expect(
      screen.getByRole('button', { name: eigene.diagnoseOeffnen }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('navigation', {
        name: kataloge.de.allgemein.navigation.bezeichnung,
      }),
    ).toBeInTheDocument()

    const kennung = /([0-9a-f]{4}-[0-9a-f]{4})/.exec(flaeche.textContent)?.[1]
    expect(melden).toHaveBeenCalledWith(
      expect.objectContaining({ kennung, ebene: 'bereich', bereich: 'sets' }),
    )
  })

  it('zeichnet den Bereich neu, wenn der Fehler vorbei ist', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = true
    zeige('sets')
    await screen.findByRole('alert', { name: eigene.bereichTitel })

    zustand.wirft = false
    fireEvent.click(screen.getByRole('button', { name: eigene.neuLaden }))

    expect(
      await screen.findByText(kataloge.de.seiten.start.was),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('alert', { name: eigene.bereichTitel }),
    ).toBeNull()
  })

  it('öffnet die Diagnose an Ort und Stelle', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = true
    zeige('sets')
    await screen.findByRole('alert', { name: eigene.bereichTitel })

    fireEvent.click(
      screen.getByRole('button', { name: eigene.diagnoseOeffnen }),
    )

    expect(
      await screen.findByRole('heading', {
        name: kataloge.de.seiten.technik.paket.ueberschrift,
      }),
    ).toBeInTheDocument()
  })

  it('räumt den Fehler beim Wechsel der Seite', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = true
    const { ansicht } = zeige('sets')
    await screen.findByRole('alert', { name: eigene.bereichTitel })

    zustand.wirft = false
    ansicht.rerender(
      <Sprachraum sprache="de">
        <Rahmen aktuell="start" wechseln={vi.fn()}>
          <Wirft />
        </Rahmen>
      </Sprachraum>,
    )

    expect(
      await screen.findByText(kataloge.de.seiten.start.was),
    ).toBeInTheDocument()
  })

  // **Wechsel und Wurf in derselben Aktualisierung** — so kommt es am
  // Fenster: man geht auf eine Seite, und sie wirft. Ein Räumen beim
  // Wechsel, das nach dem Fangen lief, nahm die Fläche sofort wieder weg
  // (Hüllenprobe `fehlergrenze09`, 2026-10-14).
  it('zeigt die Fläche, wenn die Seite beim Hinwechseln wirft', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = false
    const { ansicht } = zeige('start')
    await screen.findByText(kataloge.de.seiten.start.was)

    // **Die Naht wirft einmal** — wie am Fenster. Ein Wurf, der bei jedem
    // Zeichnen wiederkehrt, verdeckte das Räumen: die Fläche kam mit einer
    // neuen Kennung zurück, und der Test war grün.
    ;(globalThis as { __mseProbewurf?: string }).__mseProbewurf = 'sets'
    ansicht.rerender(
      <Sprachraum sprache="de">
        <Rahmen aktuell="sets" wechseln={vi.fn()}>
          <Wirft />
        </Rahmen>
      </Sprachraum>,
    )

    const flaeche = await screen.findByRole('alert', {
      name: eigene.bereichTitel,
    })
    const kennung = /([0-9a-f]{4}-[0-9a-f]{4})/.exec(flaeche.textContent)?.[1]
    await effekteNachholen()
    expect(
      screen.getByRole('alert', { name: eigene.bereichTitel }),
    ).toHaveTextContent(kennung ?? 'fehlt')
  })

  it('wirft für die Hüllenprobe, wenn sie den Bereich nennt — und nur dann', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = false
    ;(globalThis as { __mseProbewurf?: string }).__mseProbewurf = 'sets'
    zeige('sets')

    await screen.findByRole('alert', { name: eigene.bereichTitel })
    fireEvent.click(screen.getByRole('button', { name: eigene.neuLaden }))

    expect(
      await screen.findByText(kataloge.de.seiten.start.was),
    ).toBeInTheDocument()
  })

  it('meldet einen Fehler der Meldung nicht weiter', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = true
    belegeApi({
      FensterfehlerMelden: () =>
        Promise.reject(new Error('die Leitung ist weg')),
    })
    render(
      <Sprachraum sprache="de">
        <Rahmen aktuell="sets" wechseln={vi.fn()}>
          <Wirft />
        </Rahmen>
      </Sprachraum>,
    )

    expect(
      await screen.findByRole('alert', { name: eigene.bereichTitel }),
    ).toBeInTheDocument()
  })
})

// **In jedem Arbeitsbereich** — die Abnahme von F01 verlangt es, und die
// Liste kommt aus `navigation.ts`, nicht aus einer Kopie daneben: eine zu
// kurze Aufzählung wäre grün, und niemand sähe es.
describe('Die Grenze in jedem Arbeitsbereich', () => {
  it.each(seitenschluessel)(
    '%s: Satz, Kennung, Rückweg und Diagnose',
    async (seite) => {
      vi.spyOn(console, 'error').mockImplementation(() => undefined)
      zustand.wirft = true
      const { melden } = zeige(seite)

      const flaeche = await screen.findByRole('alert', {
        name: eigene.bereichTitel,
      })
      expect(flaeche).toHaveTextContent(/[0-9a-f]{4}-[0-9a-f]{4}/)
      expect(
        screen.getByRole('button', { name: eigene.neuLaden }),
      ).toHaveFocus()
      expect(
        screen.getByRole('button', { name: eigene.diagnoseOeffnen }),
      ).toBeInTheDocument()
      expect(melden).toHaveBeenCalledWith(
        expect.objectContaining({ ebene: 'bereich', bereich: seite }),
      )
    },
  )
})

describe('Die Grenze an der Wurzel', () => {
  // **Die Vorgaben sind der Betrieb**: im Fenster setzt niemand `neuLaden`
  // oder die Sprache ein — die Wurzel lädt das Fenster und spricht die
  // Sprache des Rechners.
  it('lädt ohne eingesetzten Weg das Fenster neu und spricht die Sprache des Rechners', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = true
    belegeApi({
      FensterfehlerMelden: vi.fn(() => Promise.resolve({ fehler: null })),
    })
    const reload = vi.fn()
    vi.stubGlobal('location', { reload })
    vi.spyOn(navigator, 'language', 'get').mockReturnValue('en-GB')

    render(
      <Wurzelgrenze>
        <Wirft />
      </Wurzelgrenze>,
    )

    const knopf = await screen.findByRole('button', {
      name: kataloge.en.allgemein.fehlergrenze.fensterNeuLaden,
    })
    fireEvent.click(knopf)

    expect(reload).toHaveBeenCalledTimes(1)
    vi.unstubAllGlobals()
  })

  it('zeigt einen Satz statt eines leeren Fensters und lädt auf Wunsch neu', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = true
    const melden = vi.fn(() => Promise.resolve({ fehler: null }))
    belegeApi({ FensterfehlerMelden: melden })
    const neuLaden = vi.fn()

    render(
      <Wurzelgrenze neuLaden={neuLaden} systemsprache="en-US">
        <Wirft />
      </Wurzelgrenze>,
    )

    const flaeche = await screen.findByRole('alert', {
      name: kataloge.en.allgemein.fehlergrenze.wurzelTitel,
    })
    expect(flaeche).toHaveTextContent(/[0-9a-f]{4}-[0-9a-f]{4}/)
    const knopf = screen.getByRole('button', {
      name: kataloge.en.allgemein.fehlergrenze.fensterNeuLaden,
    })
    expect(knopf).toHaveFocus()

    await act(async () => {
      fireEvent.click(knopf)
      await Promise.resolve()
    })
    expect(neuLaden).toHaveBeenCalledTimes(1)
    expect(melden).toHaveBeenCalledWith(
      expect.objectContaining({ ebene: 'wurzel', bereich: 'fenster' }),
    )
  })

  it('wirft für die Hüllenprobe auch über ein Neuladen hinweg', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = false
    sessionStorage.setItem('mse.probewurf', 'fenster')
    belegeApi({ FensterfehlerMelden: () => Promise.resolve({ fehler: null }) })

    render(
      <Wurzelgrenze neuLaden={vi.fn()} systemsprache="de-DE">
        <Wirft />
      </Wurzelgrenze>,
    )

    expect(
      await screen.findByRole('alert', { name: eigene.wurzelTitel }),
    ).toBeInTheDocument()
    expect(sessionStorage.getItem('mse.probewurf')).toBeNull()
  })

  it('nimmt einen gesperrten Speicher als leere Naht', () => {
    zustand.wirft = false
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('gesperrt')
    })
    belegeApi({})

    render(
      <Wurzelgrenze>
        <Wirft />
      </Wurzelgrenze>,
    )

    expect(screen.getByText(kataloge.de.seiten.start.was)).toBeInTheDocument()
  })

  it('räumt die Naht auch, wenn das Räumen des Speichers scheitert', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    zustand.wirft = false
    sessionStorage.setItem('mse.probewurf', 'fenster')
    vi.spyOn(Storage.prototype, 'removeItem').mockImplementation(() => {
      throw new Error('gesperrt')
    })
    belegeApi({ FensterfehlerMelden: () => Promise.resolve({ fehler: null }) })

    render(
      <Wurzelgrenze neuLaden={vi.fn()} systemsprache="de-DE">
        <Wirft />
      </Wurzelgrenze>,
    )

    expect(
      await screen.findByRole('alert', { name: eigene.wurzelTitel }),
    ).toBeInTheDocument()
  })

  it('lässt ein gesundes Fenster unberührt', () => {
    zustand.wirft = false
    belegeApi({})
    render(
      <Wurzelgrenze>
        <Wirft />
      </Wurzelgrenze>,
    )

    expect(screen.getByText(kataloge.de.seiten.start.was)).toBeInTheDocument()
  })
})
