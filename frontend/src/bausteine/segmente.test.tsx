// **Segmente für eine Einstellung mit drei oder vier Werten** — UXR-F2
// (#2008),
// Regel R7 aus `docs/UIREVIEW.md`.
//
// # Warum Optionsfelder und keine Knöpfe
//
// **Ein gedrückter `Knopf` trug damals das Gewicht des Hauptknopfs** (`Knopf`,
// `gedrueckt`; seit AP-3, #2360, ist er ein Segment wie diese), und die Seite
// hat in jeder Region höchstens einen (R1). Dazu sind Optionsfelder, was der
// Fall ist: eine Wahl unter mehreren, mit **einer** Tabstelle und den
// Pfeiltasten darin — ohne dass jemand sie baut.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Segmente } from './Segmente'

const optionen = [
  { wert: 'aus', name: 'Aus' },
  { wert: 'melden', name: 'Nur melden' },
  { wert: 'holen', name: 'Melden und herunterladen' },
]

function Probe({
  wert,
  angehalten = false,
  aendern,
}: {
  wert: string
  angehalten?: boolean
  aendern: (wert: string) => void
}) {
  return (
    <>
      <input aria-label="davor" />
      <span id="name">Nach neuen Fassungen sehen</span>
      <Segmente
        beschriftetVon="name"
        schluessel="aktualisierung"
        optionen={optionen}
        wert={wert}
        angehalten={angehalten}
        aendern={aendern}
      />
      <input aria-label="danach" />
    </>
  )
}

describe('Segmente einer Einstellung', () => {
  it('sind eine Gruppe mit dem Namen der Einstellung und ihrem Schlüssel', () => {
    render(<Probe wert="melden" aendern={vi.fn()} />)

    const gruppe = screen.getByRole('radiogroup', {
      name: 'Nach neuen Fassungen sehen',
    })

    expect(gruppe).toHaveAttribute('data-einstellung', 'aktualisierung')
    expect(screen.getAllByRole('radio')).toHaveLength(3)
  })

  it('zeigen jeden Wert unter seinem Namen und den gewählten gewählt', () => {
    render(<Probe wert="melden" aendern={vi.fn()} />)

    expect(screen.getByRole('radio', { name: 'Aus' })).not.toBeChecked()
    expect(screen.getByRole('radio', { name: 'Nur melden' })).toBeChecked()
    expect(
      screen.getByRole('radio', { name: 'Melden und herunterladen' }),
    ).not.toBeChecked()
  })

  it('melden den Wert, auf den geklickt wird', async () => {
    const aendern = vi.fn()
    render(<Probe wert="aus" aendern={aendern} />)

    await userEvent.click(
      screen.getByRole('radio', { name: 'Melden und herunterladen' }),
    )

    expect(aendern).toHaveBeenCalledExactlyOnceWith('holen')
  })

  // **Ein Klick auf den gewählten Wert ändert nichts** — und schreibt auch
  // nichts: ein Optionsfeld, das schon gewählt ist, meldet keine Änderung.
  it('melden nichts, wenn der gewählte Wert gewählt wird', async () => {
    const aendern = vi.fn()
    render(<Probe wert="aus" aendern={aendern} />)

    await userEvent.click(screen.getByRole('radio', { name: 'Aus' }))

    expect(aendern).not.toHaveBeenCalled()
  })

  // **Eine Tabstelle und die Pfeile darin** — §6: jede Aktion, die mit der
  // Maus geht, geht mit der Tastatur.
  it('haben eine Tabstelle, und die Pfeiltasten wählen darin', async () => {
    const aendern = vi.fn()
    render(<Probe wert="aus" aendern={aendern} />)

    screen.getByRole('textbox', { name: 'davor' }).focus()
    await userEvent.tab()

    expect(screen.getByRole('radio', { name: 'Aus' })).toHaveFocus()

    await userEvent.keyboard('{ArrowRight}')
    expect(aendern).toHaveBeenLastCalledWith('melden')

    await userEvent.keyboard('{ArrowRight}')
    expect(aendern).toHaveBeenLastCalledWith('holen')

    // Das dritte Tab verlässt die Gruppe: es ist **eine** Tabstelle.
    await userEvent.tab()
    expect(screen.getByRole('textbox', { name: 'danach' })).toHaveFocus()
  })

  it('wählen mit ← rückwärts', async () => {
    const aendern = vi.fn()
    render(<Probe wert="melden" aendern={aendern} />)

    screen.getByRole('radio', { name: 'Nur melden' }).focus()
    await userEvent.keyboard('{ArrowLeft}')

    expect(aendern).toHaveBeenLastCalledWith('aus')
  })

  it('nehmen nichts an, solange geschrieben wird', () => {
    render(<Probe wert="aus" angehalten aendern={vi.fn()} />)

    for (const feld of screen.getAllByRole('radio')) {
      expect(feld).toHaveAttribute('aria-disabled', 'true')
    }
  })

  // **Der Fokus bleibt in der Gruppe** — #1990: ein fokussiertes Element, das
  // `disabled` wird, verliert ihn an den Seitenanfang (WebKit; jsdom tut das
  // nicht, der Test stellt es mit `blur()` nach). Ohne das wählte der erste
  // Pfeil, und der zweite ginge ins Leere.
  it('behalten den Fokus über das Schreiben hinweg', async () => {
    const { rerender } = render(<Probe wert="aus" aendern={vi.fn()} />)
    const aus = screen.getByRole('radio', { name: 'Aus' })
    const melden = screen.getByRole('radio', { name: 'Nur melden' })

    aus.focus()
    await userEvent.keyboard('{ArrowRight}')
    expect(melden).toHaveFocus()

    melden.blur()
    rerender(<Probe wert="aus" angehalten aendern={vi.fn()} />)
    expect(document.body).toHaveFocus()

    const fokus = vi.spyOn(melden, 'focus')
    rerender(<Probe wert="melden" aendern={vi.fn()} />)

    expect(melden).toHaveFocus()
    // **Und sichtbar** — gemessen am Fenster (`uxrf2`): WebKit gibt einem per
    // Skript zurückgesetzten Fokus kein `:focus-visible`, und der Ring um das
    // Segment blieb aus, bis die nächste Taste kam (WCAG 2.4.7).
    expect(fokus).toHaveBeenCalledWith({ focusVisible: true })
  })

  // **Gesperrt heißt `aria-disabled` und nicht `disabled`** — gemessen am
  // Fenster (`uxrf2`, #2008): ein `disabled` Optionsfeld verliert in WebKit den
  // Fokus, und ein per Skript zurückgegebener Fokus trägt dort kein
  // `:focus-visible` — der Ring blieb aus. So bleibt der Fokus stehen, und eine
  // Wahl während des Schreibens wird verworfen (§7, Regel 4).
  it('bleiben während des Schreibens fokussierbar und nehmen keine Wahl an', async () => {
    const aendern = vi.fn()
    render(<Probe wert="aus" angehalten aendern={aendern} />)
    const aus = screen.getByRole('radio', { name: 'Aus' })

    expect(aus).not.toBeDisabled()
    expect(aus).toHaveAttribute('aria-disabled', 'true')

    aus.focus()
    await userEvent.keyboard('{ArrowRight}')

    expect(aendern).not.toHaveBeenCalled()
  })

  // **Der Ring hängt an `:focus` und nicht an `:focus-visible`** — gemessen
  // am Fenster (`uxrf2`, #2008): ein Optionsfeld, das die Pfeiltasten
  // fokussiert haben, hat in WKWebView `:focus`, aber kein `:focus-visible`,
  // und der Ring blieb aus (WCAG 2.4.7). Ein Mausklick fokussiert dort kein
  // Optionsfeld, der Ring kommt also nur von der Tastatur.
  it('tragen den Ring über focus-within am Segment', () => {
    render(<Probe wert="aus" aendern={vi.fn()} />)

    const segment = screen.getByRole('radio', { name: 'Aus' }).closest('label')
    expect(segment?.className).toContain('focus-within:outline')
    expect(segment?.className).not.toContain('focus-visible')
  })

  it('nehmen keinen Fokus, den sie beim Wählen nicht hatten', () => {
    const { rerender } = render(<Probe wert="aus" aendern={vi.fn()} />)
    const melden = screen.getByRole('radio', { name: 'Nur melden' })

    fireEvent.click(melden)

    rerender(<Probe wert="aus" angehalten aendern={vi.fn()} />)
    rerender(<Probe wert="melden" aendern={vi.fn()} />)

    expect(melden).not.toHaveFocus()
  })

  // **Ein Wert, den es nicht gibt, wählt keines** — ein Bestand von morgen
  // kann einen tragen; behauptet wird dann nichts.
  it('wählen keines, wenn der Wert nicht darunter ist', () => {
    render(<Probe wert="unbekannt" aendern={vi.fn()} />)

    for (const feld of screen.getAllByRole('radio')) {
      expect(feld).not.toBeChecked()
    }
  })
})

// **Ein Mausdruck nimmt den Fokus nicht, wo er bleiben soll** — QA01
// (#2382),
// gefunden in der WebKit-Abnahme `ux11-ansichten`: in WebKit nahm der Druck
// auf die Beschriftung den Fokus aus der Liste, bevor die Liste ihn in ihren
// Eintrag zurückholen konnte, und er landete auf `main`. Die zwei Knöpfe, die
// das Segment ersetzt hat, nahmen ihn in WebKit nicht.
describe('ein Segment, an dem der Fokus bleibt', () => {
  function zeige(fokusBleibt: boolean, aendern = vi.fn()) {
    render(
      <>
        <span id="name">Ansicht</span>
        <Segmente
          beschriftetVon="name"
          optionen={optionen}
          wert="aus"
          angehalten={false}
          aendern={aendern}
          fokusBleibt={fokusBleibt}
        />
      </>,
    )

    return aendern
  }

  it('hält den Fokus beim Mausdruck fest', () => {
    zeige(true)

    expect(fireEvent.mouseDown(screen.getByText('Nur melden'))).toBe(false)
  })

  it('lässt ihn ohne die Angabe gewähren', () => {
    zeige(false)

    expect(fireEvent.mouseDown(screen.getByText('Nur melden'))).toBe(true)
  })

  it('wählt trotzdem', async () => {
    const aendern = zeige(true)

    await userEvent.click(screen.getByText('Nur melden'))

    expect(aendern).toHaveBeenCalledWith('melden')
  })
})
