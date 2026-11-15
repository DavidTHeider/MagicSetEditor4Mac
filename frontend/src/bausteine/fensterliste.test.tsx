// **Die Fensterliste** — Werkstück U8 von Phase 5.1 (#1419, Vermerk 0085):
// bei fester Zeilenhöhe stehen nur die Zeilen im Sichtfenster im Baum, die
// Auskunft für das Vorleseprogramm nennt trotzdem alle, und die Tastatur
// erreicht jede Zeile — auch die zehntausendste.
import { fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Fensterliste } from './Fensterliste'

interface Eintrag {
  id: number
  name: string
}

const viele: Eintrag[] = Array.from({ length: 10000 }, (_, i) => ({
  id: i + 1,
  name: `Karte ${String(i + 1)}`,
}))

// **jsdom rechnet kein Layout**: die Höhe des Fensters kommt aus
// `clientHeight`, und die stellen wir — 400 Punkte, zehn Zeilen à 40.
let hoehe = 400
beforeEach(() => {
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get: () => hoehe,
  })
})
afterEach(() => {
  hoehe = 400
})

// zeile gibt die Zeile zu einem Text — oder scheitert laut, statt still.
function zeile(text: string): HTMLElement {
  const gefunden = screen.getByText(text).closest('[role="row"]')
  if (!(gefunden instanceof HTMLElement)) {
    throw new Error(`keine Zeile zu ${text}`)
  }

  return gefunden
}

function zeige(eintraege: readonly Eintrag[] = viele) {
  return render(
    <Fensterliste
      bezeichnung="Exemplare"
      eintraege={eintraege}
      zeilenhoehe={40}
      schluessel={(e) => e.id}
      spalten={['Name']}
      zeichne={(e) => <div role="gridcell">{e.name}</div>}
    />,
  )
}

describe('Die Fensterliste', () => {
  it('zeichnet nur das Sichtfenster samt Überhang, nennt aber alle Zeilen', () => {
    zeige()

    const gitter = screen.getByRole('grid', { name: 'Exemplare' })
    // Die Kopfzeile zählt mit — ARIA zählt Kopfzeilen (Zeile 1) wie Daten.
    expect(gitter).toHaveAttribute('aria-rowcount', '10001')

    const zeilen = within(gitter).getAllByRole('row')
    // zehn sichtbare plus Überhang — und weit weniger als zehntausend.
    expect(zeilen.length).toBeGreaterThanOrEqual(11)
    expect(zeilen.length).toBeLessThan(40)
    expect(zeilen[0]).toHaveAttribute('aria-rowindex', '1')
    expect(zeilen[1]).toHaveAttribute('aria-rowindex', '2')
    expect(screen.getByText('Karte 1')).toBeInTheDocument()
    expect(screen.queryByText('Karte 5000')).not.toBeInTheDocument()
  })

  it('folgt dem Rollen und zeichnet die Zeilen dort', () => {
    zeige()

    const fenster = screen.getByTestId('fenster')
    fenster.scrollTop = 4000
    fireEvent.scroll(fenster)

    expect(screen.getByText('Karte 101')).toBeInTheDocument()
    expect(screen.queryByText('Karte 1')).not.toBeInTheDocument()
  })

  // **Die Tastatur erreicht jede Zeile**: ↓ wandert, End springt ans Ende —
  // und die Liste rollt dorthin, damit die Zeile im Baum steht.
  it('wandert mit den Pfeiltasten und springt mit End ans Ende', () => {
    zeige()

    const erste = zeile('Karte 1')
    erste.focus()
    fireEvent.keyDown(erste, { key: 'ArrowDown' })

    const zweite = zeile('Karte 2')
    expect(zweite).toHaveFocus()
    expect(zweite.tabIndex).toBe(0)
    expect(erste.tabIndex).toBe(-1)

    fireEvent.keyDown(zweite, { key: 'End' })
    const letzte = zeile('Karte 10000')
    expect(letzte).toHaveFocus()
    expect(letzte).toHaveAttribute('aria-rowindex', '10001')

    fireEvent.keyDown(letzte, { key: 'Home' })
    expect(zeile('Karte 1')).toHaveFocus()
  })

  it('hält den Halt am Datensatz, wenn die Liste sich ändert', () => {
    const { rerender } = zeige()

    const dritte = zeile('Karte 3')
    dritte.focus()
    fireEvent.keyDown(dritte, { key: 'ArrowDown' })
    expect(zeile('Karte 4')).toHaveFocus()

    // Ein Filter nimmt die ersten drei weg — der Halt bleibt bei Karte 4.
    rerender(
      <Fensterliste
        bezeichnung="Exemplare"
        eintraege={viele.slice(3)}
        zeilenhoehe={40}
        schluessel={(e) => e.id}
        spalten={['Name']}
        zeichne={(e) => <div role="gridcell">{e.name}</div>}
      />,
    )

    const halte = screen
      .getAllByRole('row')
      .filter((zeile) => zeile.getAttribute('tabindex') === '0')
    expect(halte).toHaveLength(1)
    expect(halte[0]).toHaveTextContent('Karte 4')
  })

  // **Eine Taste mit ⌘ und ein Buchstabe sind kein Wandern.**
  it('lässt andere Tasten in Ruhe', () => {
    zeige()

    const erste = zeile('Karte 1')
    erste.focus()
    fireEvent.keyDown(erste, { key: 'ArrowDown', metaKey: true })
    fireEvent.keyDown(erste, { key: 'a' })

    expect(erste).toHaveFocus()
  })

  // **Wer den Halt hält, erfährt es** — der Inspector hängt daran.
  it('meldet den gehaltenen Eintrag beim Wandern und beim Fokus', () => {
    const aufHalt = vi.fn()
    render(
      <Fensterliste
        bezeichnung="Exemplare"
        eintraege={viele}
        zeilenhoehe={40}
        schluessel={(e) => e.id}
        spalten={['Name']}
        zeichne={(e) => <div role="gridcell">{e.name}</div>}
        aufHalt={aufHalt}
      />,
    )

    const zweite = zeile('Karte 2')
    zweite.focus()
    expect(aufHalt).toHaveBeenLastCalledWith(viele[1])

    fireEvent.keyDown(zweite, { key: 'ArrowDown' })
    expect(aufHalt).toHaveBeenLastCalledWith(viele[2])
  })

  it('sagt es, wenn nichts da ist, und zeichnet keine Zeile', () => {
    zeige([])

    expect(screen.getByRole('grid', { name: 'Exemplare' })).toHaveAttribute(
      'aria-rowcount',
      '1',
    )
    // Nur die Kopfzeile.
    expect(screen.queryAllByRole('row')).toHaveLength(1)
  })
})

// **→ und ← führen in die Zeile und zurück** (#2249) — wie in der Kartenliste
// (`inDieZeile`). Ohne das stand jedes Häkchen der Sammlung einzeln im
// Tabweg: dreihundert Exemplare, dreihundert Halte, statt einem für das
// Gitter. Gefunden vom Wächter über die Barrierefreiheit, als er auf die
// geladene Seite wartete.
describe('Die Tasten in der Zeile', () => {
  it('führen mit → in die Zeile und mit ← zurück', () => {
    render(
      <Fensterliste
        bezeichnung="Exemplare"
        eintraege={viele.slice(0, 3)}
        zeilenhoehe={40}
        schluessel={(e) => e.id}
        spalten={['Wahl', 'Name']}
        zeichne={(e) => (
          <>
            <div role="gridcell">
              <input type="checkbox" tabIndex={-1} aria-label={e.name} />
            </div>
            <div role="gridcell">{e.name}</div>
          </>
        )}
      />,
    )

    const erste = zeile('Karte 1')
    erste.focus()

    fireEvent.keyDown(erste, { key: 'ArrowRight' })
    expect(within(erste).getByRole('checkbox')).toHaveFocus()

    fireEvent.keyDown(within(erste).getByRole('checkbox'), {
      key: 'ArrowLeft',
    })
    expect(erste).toHaveFocus()
  })
})

// **Eine Zeile antwortet dem Zeiger** — AP-4
// (#2361),
// Befund B-11: bis dahin stand die Sammlungszeile ohne Hover da, und ein
// Klick hatte keine Antwort, bis sich der Inspector füllte.
describe('Rückmeldung der Zeile', () => {
  it('hebt den Grund unter dem Zeiger, in der kurzen Dauer', () => {
    zeige(viele.slice(0, 3))

    expect(zeile(viele[0]?.name ?? '')).toHaveClass(
      'hover:bg-betont',
      'transition-colors',
      'duration-(--dauer-kurz)',
    )
  })

  it('zeigt den Zeiger einer Hand nur, wenn die Zeile selbst wählt', () => {
    zeige(viele.slice(0, 3))
    expect(zeile(viele[0]?.name ?? '')).not.toHaveClass('cursor-pointer')
  })

  it('zeigt ihn, wenn ein Klick die Zeile wählt', () => {
    render(
      <Fensterliste
        bezeichnung="Exemplare"
        eintraege={viele.slice(0, 3)}
        zeilenhoehe={40}
        schluessel={(e) => e.id}
        spalten={['Name']}
        zeichne={(e) => <div role="gridcell">{e.name}</div>}
        aufKlick={vi.fn()}
      />,
    )

    expect(zeile(viele[0]?.name ?? '')).toHaveClass('cursor-pointer')
  })
})

// **Die Köpfe so breit wie ihre Zellen** — AP-12 (#2369): im Handbuchbild
// der Sammlung stand „Menge“ über der Bezeichnung, weil jeder Kopf gleich
// breit war und die Zellen nicht. Und die Liste ist ein Behälter, damit eine
// Zeile nach ihrer eigenen Breite fragen kann statt nach der des Fensters.
describe('Die Kopfzeile', () => {
  it('gibt einem Kopf die Klasse seiner Spalte, sonst teilt sie gleich', () => {
    render(
      <Fensterliste
        bezeichnung="Exemplare"
        eintraege={viele.slice(0, 1)}
        zeilenhoehe={40}
        schluessel={(e) => e.id}
        spalten={[{ titel: 'Wahl', klasse: 'w-8 shrink-0' }, 'Name']}
        zeichne={(e) => <div role="gridcell">{e.name}</div>}
      />,
    )

    expect(screen.getByRole('columnheader', { name: 'Wahl' })).toHaveClass(
      'w-8',
      'shrink-0',
      'truncate',
    )
    expect(screen.getByRole('columnheader', { name: 'Name' })).toHaveClass(
      'flex-1',
      'truncate',
    )
    expect(screen.getByTestId('fenster')).toHaveClass('@container')
  })
})
