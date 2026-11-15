// **Der Arbeitsplatz: Navigator, Zentrum, Inspector** — Werkstück U2 von
// Phase 5.1 (#1413).
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import { engeGrenze, navigatorPasst } from './arbeitsplatzmasse'
import { Verdeckung } from '../verdeckung'
import { effekteNachholen } from '../test/effekte'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Sprachraum } from '../sprache/kontext'

// **jsdom kennt keine Breite** — ohne diese Zeilen wäre jeder Arbeitsplatz
// eng, und der Fall „breit genug" ließe sich nie betreten.
let raeumeBreite: () => void = () => undefined

beforeEach(() => {
  raeumeBreite = gibJsdomEineBreite(1400)
})

afterEach(() => {
  raeumeBreite()
})

function Probe({ beiEnge }: { beiEnge?: 'navigator' | 'inspector' }) {
  return (
    <Sprachraum sprache="de">
      <Arbeitsplatz
        bezeichnung="Kartenarbeitsplatz"
        navigator={{ bezeichnung: 'Karten', inhalt: <p>Die Liste</p> }}
        zentrum={<p>Die Vorschau</p>}
        inspector={{ bezeichnung: 'Felder', inhalt: <p>Das Formular</p> }}
        beiEnge={beiEnge}
      />
    </Sprachraum>
  )
}

describe('navigatorPasst', () => {
  it('lässt den Navigator ab der Grenze stehen', () => {
    expect(navigatorPasst(engeGrenze)).toBe(true)
    expect(navigatorPasst(engeGrenze - 1)).toBe(false)
  })

  // **Die Grenze wächst mit der Schrift** — G20 aus dem Zweitblick über
  // Phase 6.3 (#1664), verlagert in Phase 6.11 (#1672). Die Spalten stehen in
  // `rem` (`w-60`, `w-80`), die Grenze stand in Punkten: bei „Sehr groß"
  // (125 %, 20 Punkte Wurzelschrift) blieb der Navigator ab 1.072 stehen,
  // und dem Zentrum blieben 332 statt 480 · 1,25.
  it('rechnet in der Wurzelschrift — bei „Sehr groß" wächst sie mit', () => {
    expect(navigatorPasst(engeGrenze, 20)).toBe(false)
    expect(navigatorPasst(engeGrenze * 1.25 - 1, 20)).toBe(false)
    expect(navigatorPasst(engeGrenze * 1.25, 20)).toBe(true)
  })

  // **Die Grenze ist am Arbeitsplatz gemessen, nicht am Inhalt** — gefunden
  // in der Sichtprüfung von U5 (#1416): bei 1.440 Punkten Fenster misst der
  // Arbeitsplatz **1.086**, weil Rahmen und Setkopf 114 Punkte Rand nehmen,
  // und mit einer Grenze von 1.200 stand der Navigator auf keinem Fenster,
  // das es gibt. Das Messfenster (1.280, Arbeitsplatz 926) bleibt eng.
  it('steht beim 1.440er-Fenster und nicht beim Messfenster', () => {
    expect(navigatorPasst(1086)).toBe(true)
    expect(navigatorPasst(926)).toBe(false)
  })
})

describe('Arbeitsplatz', () => {
  it('zeigt drei Regionen, jede mit ihrem Namen', () => {
    render(<Probe />)

    expect(
      screen.getByRole('region', { name: 'Kartenarbeitsplatz' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toHaveTextContent('Die Liste')
    expect(
      screen.getByRole('complementary', { name: 'Felder' }),
    ).toHaveTextContent('Das Formular')
    expect(screen.getByText('Die Vorschau')).toBeInTheDocument()
  })

  it('klappt den Inspector über seinen Knopf ein und aus', async () => {
    render(<Probe />)

    const knopf = screen.getByRole('button', { name: 'Inspector ausblenden' })
    expect(knopf).toHaveAttribute('aria-expanded', 'true')

    await userEvent.click(knopf)

    expect(screen.queryByRole('complementary', { name: 'Felder' })).toBeNull()
    const wieder = screen.getByRole('button', { name: 'Inspector einblenden' })
    expect(wieder).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(wieder)
    expect(
      screen.getByRole('complementary', { name: 'Felder' }),
    ).toBeInTheDocument()
  })

  // **⌥⌘[ und ⌥⌘]** — die Tasten aus `tastatur.ts`, gedeutet über `code`,
  // weil ⌥ das Zeichen umschreibt.
  it('klappt Navigator und Inspector über die Tasten', async () => {
    render(<Probe />)
    await effekteNachholen()

    fireEvent.keyDown(window, {
      key: '„',
      code: 'BracketLeft',
      metaKey: true,
      altKey: true,
    })
    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()

    fireEvent.keyDown(window, {
      key: '‘',
      code: 'BracketRight',
      metaKey: true,
      altKey: true,
    })
    expect(screen.queryByRole('complementary', { name: 'Felder' })).toBeNull()

    fireEvent.keyDown(window, {
      key: '„',
      code: 'BracketLeft',
      metaKey: true,
      altKey: true,
    })
    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toBeInTheDocument()
  })

  // **Unter einer Schublade klappt nichts** (UX-08, #1680): ⌥⌘[ gehört dann
  // niemandem, der es sieht.
  it('lässt die Tasten los, solange er verdeckt ist', async () => {
    const { rerender } = render(
      <Verdeckung value>
        <Probe />
      </Verdeckung>,
    )
    await effekteNachholen()

    const links = { key: '„', code: 'BracketLeft', metaKey: true, altKey: true }

    fireEvent.keyDown(window, links)
    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toBeInTheDocument()

    // **Der Beleg, dass jemand zuhört**: aufgedeckt klappt dieselbe Taste.
    rerender(
      <Verdeckung value={false}>
        <Probe />
      </Verdeckung>,
    )
    await effekteNachholen()

    fireEvent.keyDown(window, links)
    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
  })

  // **Unter 1.200 Punkten klappt der Navigator von selbst ein** — und der
  // Knopf sagt es: er bietet das Einblenden an, das erst wieder geht, wenn
  // Platz ist.
  // **Das Zentrum erzwingt keine Mindestbreite.** Im Mindestfenster
  // (1.024 × 700) misst der Arbeitsplatz 670 Punkte; Zentrum 480 und
  // Inspector 320 passten nicht hinein, und die Seite rollte waagerecht —
  // gefunden in der Sichtprüfung von U5. Die 480 aus 0076 hält die Grenze
  // zum Einklappen, nicht ein CSS-Maß.
  it('lässt das Zentrum schmaler als 480 werden, statt die Seite rollen zu lassen', () => {
    render(<Probe />)

    const zentrum = screen.getByText('Die Vorschau').parentElement
    expect(zentrum).toHaveClass('min-w-0')
    expect(zentrum).not.toHaveClass('min-w-[30rem]')
  })

  it('kommt ohne Navigator und Inspector aus', () => {
    render(
      <Sprachraum sprache="de">
        <Arbeitsplatz bezeichnung="Nur Mitte" zentrum={<p>Allein</p>} />
      </Sprachraum>,
    )

    expect(screen.queryByRole('button')).toBeNull()
    expect(screen.getByText('Allein')).toBeInTheDocument()

    // **Die Tasten tun ohne Region nichts** — und werfen nichts.
    fireEvent.keyDown(window, {
      key: '„',
      code: 'BracketLeft',
      metaKey: true,
      altKey: true,
    })
    expect(screen.getByText('Allein')).toBeInTheDocument()
  })
})

it('klappt den Navigator über seinen Knopf ein und aus', async () => {
  raeumeBreite()
  raeumeBreite = gibJsdomEineBreite(1400)

  render(<Probe />)

  await userEvent.click(
    screen.getByRole('button', { name: 'Navigator ausblenden' }),
  )
  expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()

  await userEvent.click(
    screen.getByRole('button', { name: 'Navigator einblenden' }),
  )
  expect(
    screen.getByRole('complementary', { name: 'Karten' }),
  ).toBeInTheDocument()
})

// **Mit Gedächtnis** — Werkstück U4 von Phase 5.1
// (#1415):
// die Seite reicht `useArbeitsplatzgedaechtnis` herein, und der Baustein
// zeigt, was dort steht, und meldet, was jemand klappt.
describe('Arbeitsplatz mit Gedächtnis', () => {
  function gedaechtnis(
    navigator: boolean,
    inspector: boolean,
    merke = vi.fn(),
    laeuft: (region: string) => boolean = () => false,
  ) {
    return {
      geladen: true,
      bereit: true,
      navigator,
      inspector,
      navigatorbreite: 300,
      inspectorbreite: 400,
      fehler: undefined,
      laeuft,
      merke,
      breiteWuenschen: vi.fn(),
    }
  }

  function ProbeMitGedaechtnis({
    stand,
  }: {
    stand: ReturnType<typeof gedaechtnis>
  }) {
    return (
      <Sprachraum sprache="de">
        <Arbeitsplatz
          bezeichnung="Kartenarbeitsplatz"
          navigator={{ bezeichnung: 'Karten', inhalt: <p>Die Liste</p> }}
          zentrum={<p>Die Vorschau</p>}
          inspector={{ bezeichnung: 'Felder', inhalt: <p>Das Formular</p> }}
          gedaechtnis={stand}
        />
      </Sprachraum>
    )
  }

  it('zeigt die Regionen so, wie das Gedächtnis sie kennt, mit ihren Breiten', () => {
    render(<ProbeMitGedaechtnis stand={gedaechtnis(false, true)} />)

    expect(
      screen.queryByRole('complementary', { name: 'Karten' }),
    ).not.toBeInTheDocument()

    const inspector = screen.getByRole('complementary', { name: 'Felder' })
    expect(inspector).toBeInTheDocument()
    expect(inspector).toHaveStyle({ width: '400px' })
  })

  it('gibt dem Navigator seine Breite aus dem Gedächtnis', () => {
    render(<ProbeMitGedaechtnis stand={gedaechtnis(true, false)} />)

    expect(screen.getByRole('complementary', { name: 'Karten' })).toHaveStyle({
      width: '300px',
    })
    expect(
      screen.queryByRole('complementary', { name: 'Felder' }),
    ).not.toBeInTheDocument()
  })

  it('meldet einen Klick auf den Knopf an das Gedächtnis, statt selbst zu klappen', async () => {
    const merke = vi.fn()
    render(<ProbeMitGedaechtnis stand={gedaechtnis(true, true, merke)} />)

    await userEvent.click(
      screen.getByRole('button', { name: 'Navigator ausblenden' }),
    )

    expect(merke).toHaveBeenCalledWith('navigator', false)
    // **Der Baustein wartet auf das Gedächtnis**: solange es nichts anderes
    // sagt, steht der Navigator.
    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: 'Inspector ausblenden' }),
    )

    expect(merke).toHaveBeenCalledWith('inspector', false)
  })

  it('meldet auch die Tasten an das Gedächtnis', async () => {
    const merke = vi.fn()
    render(<ProbeMitGedaechtnis stand={gedaechtnis(false, false, merke)} />)
    await effekteNachholen()

    fireEvent.keyDown(window, {
      key: '[',
      code: 'BracketLeft',
      metaKey: true,
      altKey: true,
    })
    fireEvent.keyDown(window, {
      key: ']',
      code: 'BracketRight',
      metaKey: true,
      altKey: true,
    })

    expect(merke).toHaveBeenCalledWith('navigator', true)
    expect(merke).toHaveBeenCalledWith('inspector', true)
  })

  // **Ein Vorgang je Datensatz** (§7): während das Gedächtnis schreibt, ist
  // der Knopf angehalten.
  it('hält den Knopf an, solange das Gedächtnis schreibt', () => {
    render(
      <ProbeMitGedaechtnis
        stand={gedaechtnis(
          true,
          true,
          vi.fn(),
          (region) => region === 'navigator',
        )}
      />,
    )

    expect(
      screen.getByRole('button', { name: 'Navigator ausblenden' }),
    ).toBeDisabled()
    expect(
      screen.getByRole('button', { name: 'Inspector ausblenden' }),
    ).toBeEnabled()
  })
})
