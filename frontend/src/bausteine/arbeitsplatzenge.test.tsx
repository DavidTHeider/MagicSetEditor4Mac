// **Der Arbeitsplatz bei Enge** — UX-14.AK-03 und S06.NAT-03
// (#1686).
//
// **Aus `arbeitsplatz.test.tsx` herausgezogen**, als die Läufe zur Enge
// dazukamen und die Datei über die Zeilengrenze ging (§2). Der Schnitt folgt
// der Frage: dort, wie der Arbeitsplatz steht, hier, wie er ausweicht.
import { act, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  onTestFinished,
} from 'vitest'

import { Arbeitsplatz } from './Arbeitsplatz'
import { setzeSchriftgroesse } from '../erscheinungsbild'
import { effekteNachholen } from '../test/effekte'
import { gibJsdomEineBreite } from '../test/sichtbreite'
import { Sprachraum } from '../sprache/kontext'

// **jsdom kennt keine Breite** — dieselbe Vorrichtung wie im Nachbarn.
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

describe('Der Arbeitsplatz bei Enge', () => {
  it('klappt den Navigator bei enger Breite von selbst ein', () => {
    raeumeBreite()
    raeumeBreite = gibJsdomEineBreite(1000)

    render(<Probe />)

    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
    expect(
      screen.getByText(
        'Zu schmal für alle drei Spalten — der Navigator geht über dem Zentrum auf',
      ),
    ).toBeInTheDocument()
  })

  // **Nichts ist gesperrt** — UX-14.AK-03 (#1686),
  // Artboard *Enge*: *„Inspector und Navigator können ausweichen statt
  // Inhalte abzuschneiden."* Bis zum 2026-09-25 war der Knopf bei Enge
  // angehalten, und bei 1.024 × 700 war die Kartenliste gar nicht erreichbar.
  it('lässt den Navigator bei Enge über dem Zentrum aufgehen, statt ihn zu sperren', async () => {
    raeumeBreite()
    raeumeBreite = gibJsdomEineBreite(1000)

    render(<Probe />)
    await effekteNachholen()

    const knopf = screen.getByRole('button', { name: 'Navigator einblenden' })
    expect(knopf).toBeEnabled()
    await userEvent.click(knopf)

    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toHaveAttribute('data-ueber', 'zentrum')
    expect(
      screen.getByRole('button', { name: 'Navigator ausblenden' }),
    ).toHaveAttribute('aria-expanded', 'true')

    // **Der Knopf schließt sie wieder**, und ein zweiter Druck öffnet.
    await userEvent.click(
      screen.getByRole('button', { name: 'Navigator ausblenden' }),
    )
    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
    await userEvent.click(
      screen.getByRole('button', { name: 'Navigator einblenden' }),
    )

    // **Escape schließt, und der Fokus steht wieder am Knopf** — auch wenn er
    // inzwischen in der Spalte war. Ohne das Wegnehmen stünde er vom Klick
    // noch am Knopf, und der Lauf bewiese nichts.
    ;(document.activeElement as HTMLElement | null)?.blur()
    fireEvent.keyDown(window, { key: 'Escape' })

    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
    expect(
      screen.getByRole('button', { name: 'Navigator einblenden' }),
    ).toHaveFocus()
  })

  // **Maximieren und zurück erhalten die Spalte** — S06.NAT-03 (#1686): wer
  // sie bei Enge über dem Zentrum geöffnet hat, findet sie nach dem Umweg über
  // ein großes Fenster wieder dort; dazwischen steht sie, wie gewünscht.
  it('hält die geöffnete Spalte über Maximieren und Zurück', async () => {
    const rueckrufe: (() => void)[] = []
    const vorher = globalThis.ResizeObserver
    globalThis.ResizeObserver = class {
      constructor(rueckruf: () => void) {
        rueckrufe.push(rueckruf)
      }

      observe() {
        return undefined
      }

      unobserve() {
        return undefined
      }

      disconnect() {
        return undefined
      }
    } as unknown as typeof ResizeObserver
    onTestFinished(() => {
      globalThis.ResizeObserver = vorher
    })
    const neueBreite = (breite: number) => {
      raeumeBreite()
      raeumeBreite = gibJsdomEineBreite(breite)
      act(() => {
        for (const rueckruf of rueckrufe) {
          rueckruf()
        }
      })
    }

    raeumeBreite()
    raeumeBreite = gibJsdomEineBreite(1000)
    render(<Probe />)
    await effekteNachholen()
    await userEvent.click(
      screen.getByRole('button', { name: 'Navigator einblenden' }),
    )

    neueBreite(1400)
    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).not.toHaveAttribute('data-ueber')

    neueBreite(1000)
    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toHaveAttribute('data-ueber', 'zentrum')
  })

  // **Dieselbe Taste wie sonst** — ⌥⌘[ öffnet ihn bei Enge über dem Zentrum.
  it('öffnet den Navigator bei Enge auch über ⌥⌘[', async () => {
    raeumeBreite()
    raeumeBreite = gibJsdomEineBreite(1000)

    render(<Probe />)
    await effekteNachholen()

    fireEvent.keyDown(window, {
      key: '„',
      code: 'BracketLeft',
      metaKey: true,
      altKey: true,
    })

    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toHaveAttribute('data-ueber', 'zentrum')
  })

  // **Wer weicht, entscheidet die Seite** (U7, #1418): die Vorlagenwerkstatt
  // gibt bei Enge den Inspector her, denn ihr Navigator ist der einzige Weg
  // zu einer Datei — und auch der Inspector geht dann über dem Zentrum auf.
  it('gibt bei enger Breite auf Wunsch den Inspector her und lässt den Navigator stehen', async () => {
    raeumeBreite()
    raeumeBreite = gibJsdomEineBreite(1000)

    render(<Probe beiEnge="inspector" />)
    await effekteNachholen()

    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('complementary', { name: 'Felder' })).toBeNull()
    expect(
      screen.getByText(
        'Zu schmal für alle drei Spalten — der Inspector geht über dem Zentrum auf',
      ),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: 'Inspector einblenden' }),
    )

    expect(
      screen.getByRole('complementary', { name: 'Felder' }),
    ).toHaveAttribute('data-ueber', 'zentrum')
    expect(
      screen.getByRole('button', { name: 'Navigator ausblenden' }),
    ).toBeEnabled()

    // **Escape gibt den Fokus an den Knopf dieser Spalte zurück**, nicht an
    // den des Navigators.
    ;(document.activeElement as HTMLElement | null)?.blur()
    fireEvent.keyDown(window, { key: 'Escape' })
    expect(screen.queryByRole('complementary', { name: 'Felder' })).toBeNull()
    expect(
      screen.getByRole('button', { name: 'Inspector einblenden' }),
    ).toHaveFocus()
  })

  // **Ein Schriftwechsel misst neu** — Nachlauf zum Zweitblick über Phase 6
  // von 0.9.0, zweiter Teil, R5: die Grenze rechnet in der Wurzelschrift,
  // gemessen wurde aber nur bei einer neuen Breite. Bei 1.200 Punkten passt
  // der Navigator mit 16 Punkten, mit 20 nicht mehr.
  it('klappt den Navigator ein, wenn die Schrift zur Laufzeit größer wird', () => {
    raeumeBreite()
    raeumeBreite = gibJsdomEineBreite(1200)

    render(<Probe />)
    expect(
      screen.getByRole('complementary', { name: 'Karten' }),
    ).toBeInTheDocument()

    onTestFinished(() => {
      document.documentElement.style.fontSize = ''
      document.documentElement.style.removeProperty('--schriftgroesse')
    })
    act(() => {
      document.documentElement.style.fontSize = '20px'
      setzeSchriftgroesse(document.documentElement, 'sehrgross')
    })

    expect(screen.queryByRole('complementary', { name: 'Karten' })).toBeNull()
  })
})
