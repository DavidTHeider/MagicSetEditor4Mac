// **Der Arbeitsplatz und die Menüleiste** — UXR-F14
// (#2029).
//
// Zwei Wege, und beide laufen hier durch den echten Baustein: der Menüeintrag
// meldet ein Ereignis, und die Spalte klappt; wie sie danach dasteht, geht
// über `DarstellungsstandMelden` an die native Leiste. Ein Test gegen den
// Haken allein prüfte nur, dass er etwas ruft — nicht, dass die Spalte dabei
// zuklappt.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import {
  anfangsstand,
  type Darstellungsstand,
  type Fensterteil,
} from '../darstellungsstand'
import { Darstellungsstandraum } from '../darstellungsstandraum'
import { ereignisInspector, ereignisNavigator } from '../fenster'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Verdeckung } from '../verdeckung'

let raeumeBreite: () => void = () => undefined

beforeEach(() => {
  raeumeBreite = gibJsdomEineBreite(1400)
})

afterEach(() => {
  raeumeBreite()
  raeumeApi()
  delete (globalThis as { runtime?: unknown }).runtime
})

// laufzeitAttrappe sammelt die Rückrufe, die der Baustein anmeldet — dieselbe
// Naht wie in `menueleiste.test.tsx`.
function laufzeitAttrappe() {
  const angemeldet = new Map<string, (...daten: unknown[]) => void>()

  ;(globalThis as { runtime?: unknown }).runtime = {
    EventsOn: (name: string, rueckruf: (...daten: unknown[]) => void) => {
      angemeldet.set(name, rueckruf)

      return () => angemeldet.delete(name)
    },
  }

  return {
    // ausloesen lässt macOS den Eintrag der Leiste wählen.
    ausloesen: (name: string) => {
      act(() => {
        angemeldet.get(name)?.()
      })
    },
    angemeldet: (name: string) => angemeldet.has(name),
  }
}

function Probe({
  navigator = true,
  inspector = true,
  verdeckt = false,
  liste = <p>Die Liste</p>,
}: {
  navigator?: boolean
  inspector?: boolean
  verdeckt?: boolean
  liste?: ReactNode
}) {
  // **Eine fehlende Spalte fehlt ganz**, statt `undefined` zu heißen — die
  // Eigenschaften des Bausteins sind genau (`exactOptionalPropertyTypes`).
  const spalten = {
    ...(navigator
      ? { navigator: { bezeichnung: 'Karten', inhalt: liste } }
      : {}),
    ...(inspector
      ? { inspector: { bezeichnung: 'Felder', inhalt: <p>Das Formular</p> } }
      : {}),
  }

  return (
    <Sprachraum sprache="de">
      <Darstellungsstandraum>
        <Verdeckung value={verdeckt}>
          <Arbeitsplatz
            bezeichnung="Kartenarbeitsplatz"
            zentrum={<p>Die Vorschau</p>}
            {...spalten}
          />
        </Verdeckung>
      </Darstellungsstandraum>
    </Sprachraum>
  )
}

const steht: Fensterteil = { vorhanden: true, offen: true }
const zu: Fensterteil = { vorhanden: true, offen: false }

// stand baut den erwarteten Stand — die Seitenleiste meldet hier niemand.
function stand(
  navigator: Fensterteil,
  inspector: Fensterteil,
): Darstellungsstand {
  return { ...anfangsstand, navigator, inspector }
}

function belege() {
  const DarstellungsstandMelden = vi.fn(() => Promise.resolve({ fehler: null }))
  belegeApi({ DarstellungsstandMelden })

  return DarstellungsstandMelden
}

describe('Arbeitsplatz und Menüleiste', () => {
  it('meldet, dass beide Spalten da sind und stehen', async () => {
    const melden = belege()
    laufzeitAttrappe()

    render(<Probe />)

    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(stand(steht, steht))
    })
  })

  it('klappt den Navigator auf den Eintrag hin und meldet es', async () => {
    const melden = belege()
    const leiste = laufzeitAttrappe()

    render(<Probe />)
    await effekteNachholen()

    leiste.ausloesen(ereignisNavigator)

    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(stand(zu, steht))
    })

    // **Und derselbe Eintrag klappt ihn wieder auf** — er schaltet um, er
    // stellt nichts.
    leiste.ausloesen(ereignisNavigator)

    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toBeInTheDocument()
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(stand(steht, steht))
    })
  })

  it('klappt den Inspector auf den Eintrag hin und meldet es', async () => {
    const melden = belege()
    const leiste = laufzeitAttrappe()

    render(<Probe />)
    await effekteNachholen()

    leiste.ausloesen(ereignisInspector)

    expect(screen.queryByRole('complementary', { name: 'Felder' })).toBeNull()
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(stand(steht, zu))
    })
  })

  // **Knopf und Eintrag sind ein Weg**: was der Knopf zuklappt, steht in der
  // Leiste als „einblenden", ohne dass jemand es ihr sagt.
  it('meldet auch, was der Knopf klappt', async () => {
    const melden = belege()
    laufzeitAttrappe()

    render(<Probe />)
    await effekteNachholen()

    await userEvent.click(
      screen.getByRole('button', { name: 'Navigator ausblenden' }),
    )

    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(stand(zu, steht))
    })
  })

  // **Eine Seite ohne Inspector hat den Eintrag nicht** — er ist gegraut, und
  // käme sein Ereignis trotzdem, geschähe nichts.
  it('meldet einen fehlenden Inspector und hört nicht auf seinen Eintrag', async () => {
    const melden = belege()
    const leiste = laufzeitAttrappe()

    render(<Probe inspector={false} />)
    await effekteNachholen()

    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(
        stand(steht, { vorhanden: false, offen: false }),
      )
    })

    leiste.ausloesen(ereignisInspector)
    await antwortAbwarten()

    // **Und der Navigator hört weiter** — sonst wäre die Verneinung grün, weil
    // gar nichts angemeldet ist.
    expect(leiste.angemeldet(ereignisNavigator)).toBe(true)
    leiste.ausloesen(ereignisNavigator)
    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
  })

  // **Dasselbe für den Navigator**: eine Seite kann nur den Inspector haben.
  it('meldet einen fehlenden Navigator und hört nicht auf seinen Eintrag', async () => {
    const melden = belege()
    const leiste = laufzeitAttrappe()

    render(<Probe navigator={false} />)
    await effekteNachholen()

    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(
        stand({ vorhanden: false, offen: false }, steht),
      )
    })

    leiste.ausloesen(ereignisNavigator)
    await antwortAbwarten()

    // **Und der Inspector hört weiter** — dieselbe Gegenprobe wie darüber.
    leiste.ausloesen(ereignisInspector)
    expect(screen.queryByRole('complementary', { name: 'Felder' })).toBeNull()
  })

  // **Unter einer Schublade ist der Eintrag gegraut** (UX-08): die Tasten
  // gehören dann nicht der Seite darunter. Die Spalte steht weiter da — der
  // graue Eintrag sagt, wie.
  it('meldet unter einer Schublade „nicht vorhanden" und klappt nichts', async () => {
    const melden = belege()
    const leiste = laufzeitAttrappe()

    render(<Probe verdeckt />)
    await effekteNachholen()

    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(
        stand(
          { vorhanden: false, offen: true },
          { vorhanden: false, offen: true },
        ),
      )
    })

    leiste.ausloesen(ereignisNavigator)
    leiste.ausloesen(ereignisInspector)
    await antwortAbwarten()

    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('complementary', { name: 'Felder' }),
    ).toBeInTheDocument()
  })

  // **Bei Enge liegt die Spalte über dem Zentrum** — und der Eintrag öffnet
  // sie dort, wie der Knopf: er meint, was zu sehen ist, nicht den Wunsch
  // dahinter.
  it('öffnet bei Enge den Navigator über dem Zentrum und meldet ihn offen', async () => {
    raeumeBreite()
    raeumeBreite = gibJsdomEineBreite(1000)

    const melden = belege()
    const leiste = laufzeitAttrappe()

    render(<Probe />)
    await effekteNachholen()

    // Er weicht: zu, solange niemand ihn aufmacht.
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(stand(zu, steht))
    })

    leiste.ausloesen(ereignisNavigator)

    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toHaveAttribute('data-ueber', 'zentrum')
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(stand(steht, steht))
    })
  })

  // **Über das Menü geöffnet, nimmt die Spalte über dem Zentrum den Fokus** —
  // gemessen am Fenster (`darstellung`, #2029): der Fokus blieb sonst, wo er
  // war, und Escape schloss die Spalte nicht (`escapeSchliesst` verlangt den
  // Fokus in ihr oder auf `body`). **Im selben Schritt, in dem sie erscheint**,
  // nicht im nächsten Bild: am Fenster stand die Spalte schon da, der Fokus
  // nach einem schweren Navigator aber erst 300 ms später — ein Escape
  // dazwischen ging ins Leere.
  it('gibt einer Spalte über dem Zentrum den Fokus, die das Menü öffnet', async () => {
    raeumeBreite()
    raeumeBreite = gibJsdomEineBreite(1000)

    const leiste = laufzeitAttrappe()
    render(
      <>
        <button type="button">Woanders</button>
        <Probe />
      </>,
    )
    await effekteNachholen()
    screen.getByRole('button', { name: 'Woanders' }).focus()

    leiste.ausloesen(ereignisNavigator)

    const spalte = screen.getByRole('complementary', { name: 'Karten' })
    expect(spalte.contains(document.activeElement)).toBe(true)

    await userEvent.keyboard('{Escape}')
    expect(
      screen.queryByRole('complementary', { name: 'Karten' }),
    ).not.toBeInTheDocument()
  })

  it('gibt den Fokus dem ersten Bedienelement der Spalte', async () => {
    raeumeBreite()
    raeumeBreite = gibJsdomEineBreite(1000)

    const leiste = laufzeitAttrappe()
    render(<Probe liste={<button type="button">Erste Karte</button>} />)
    await effekteNachholen()

    leiste.ausloesen(ereignisNavigator)

    expect(screen.getByRole('button', { name: 'Erste Karte' })).toHaveFocus()
  })

  // **Eine Spalte daneben ist keine Überlagerung** und nimmt keinen Fokus —
  // bei Platz steht der Inspector neben dem Zentrum.
  it('lässt den Fokus stehen, wenn das Menü den Inspector daneben öffnet', async () => {
    const leiste = laufzeitAttrappe()
    render(
      <>
        <button type="button">Woanders</button>
        <Probe />
      </>,
    )
    await effekteNachholen()
    const woanders = screen.getByRole('button', { name: 'Woanders' })
    woanders.focus()

    leiste.ausloesen(ereignisInspector)
    leiste.ausloesen(ereignisInspector)
    await effekteNachholen()

    expect(
      screen.getByRole('complementary', { name: 'Felder' }),
    ).toBeInTheDocument()
    expect(woanders).toHaveFocus()
  })

  it('nimmt beide Spalten zurück, wenn der Arbeitsplatz geht', async () => {
    const melden = belege()
    laufzeitAttrappe()

    const { rerender } = render(<Probe />)
    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(stand(steht, steht))
    })

    rerender(
      <Sprachraum sprache="de">
        <Darstellungsstandraum>{null}</Darstellungsstandraum>
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(melden).toHaveBeenLastCalledWith(anfangsstand)
    })
  })
})
