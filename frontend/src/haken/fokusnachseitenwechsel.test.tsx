// **Nach einem Seitenwechsel steht der Fokus nicht im Leeren** — UX-10.AK-02
// (#1682).
//
// Gemessen am 2026-09-25 in der WebKit-Probe: nach ⌘2 und ⌘6 bewegte Tab den
// Fokus **gar nicht**. Er stand auf `body`, und WebKit suchte von einem Knoten
// der vorigen Seite aus weiter, den der Wechsel abgeräumt hatte. **jsdom hat
// diesen Startpunkt nicht** — geprüft wird hier, wohin der Fokus geht; dass
// Tab danach weitergeht, zeigt das Drehbuch `ux10-tastatur`.
import { act, fireEvent, render, screen } from '@testing-library/react'
import { Activity } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { useSeitengang } from './useSeitengang'

afterEach(() => {
  document.body.replaceChildren()
})

// Probe ist ein Rahmen im Kleinen: eine Leiste, die bleibt, und ein Inhalt,
// der mit der Seite wechselt.
function Probe() {
  const gang = useSeitengang(undefined)

  return (
    <>
      <nav>
        <button
          type="button"
          onClick={() => {
            gang.wechseln('sammlung')
          }}
        >
          Leiste
        </button>
      </nav>
      <main id="inhalt" tabIndex={-1}>
        {gang.aktuell === 'start' ? (
          <button
            type="button"
            onClick={() => {
              gang.wechseln('sets')
            }}
          >
            Zu den Sets
          </button>
        ) : (
          <p>{gang.aktuell}</p>
        )}
      </main>
    </>
  )
}

// Aufbewahrt ist der Rahmen, wie die Anwendung ihn baut: **eine besuchte
// Seite bleibt, verborgen** (`Seiteninhalt.tsx`) — ihr Knopf wird nicht
// entfernt, sondern versteckt, und der Fokus steht noch auf ihm.
function Aufbewahrt() {
  const gang = useSeitengang(undefined)

  return (
    <main id="inhalt" tabIndex={-1}>
      <Activity mode={gang.aktuell === 'start' ? 'visible' : 'hidden'}>
        <button
          type="button"
          onClick={() => {
            gang.wechseln('sets')
          }}
        >
          Zu den Sets
        </button>
      </Activity>
      <Activity mode={gang.aktuell === 'sets' ? 'visible' : 'hidden'}>
        <p>sets</p>
      </Activity>
    </main>
  )
}

describe('der Fokus nach einem Seitenwechsel', () => {
  it('geht in den Inhalt, wenn der gedrückte Knopf mit der Seite ging', () => {
    render(<Probe />)
    const knopf = screen.getByRole('button', { name: 'Zu den Sets' })
    knopf.focus()

    act(() => {
      fireEvent.click(knopf)
    })

    expect(screen.getByRole('main')).toHaveFocus()
  })

  // **Auch, wenn die Seite aufbewahrt und nur verborgen wird** — in WebKit
  // stand der Fokus dann auf einem versteckten Knopf, und Tab suchte von
  // einer Seite aus weiter, die niemand mehr sah.
  it('geht in den Inhalt, wenn der gedrückte Knopf verborgen wird', () => {
    render(<Aufbewahrt />)
    const knopf = screen.getByRole('button', { name: 'Zu den Sets' })
    knopf.focus()

    act(() => {
      fireEvent.click(knopf)
    })

    expect(screen.getByText('sets')).toBeVisible()
    expect(screen.getByRole('main')).toHaveFocus()
  })

  // **Steht eine Schublade offen, bleibt der Fokus in ihr** — angestrichen
  // vom Zweitblick über Phase 6.3 (M2): ein Laschenwechsel mit ⌥⌘4–7 setzte
  // ihn auf den Inhalt **hinter** der Schublade, ohne Ring, und Escape
  // schloss sie danach nicht mehr.
  it('bleibt in einer offenen Schublade', () => {
    render(
      <>
        <Aufbewahrt />
        <div role="dialog" aria-modal="true" aria-label="Schublade">
          <button type="button">In der Schublade</button>
        </div>
      </>,
    )
    const knopf = screen.getByRole('button', { name: 'Zu den Sets' })
    knopf.focus()

    act(() => {
      fireEvent.click(knopf)
    })

    expect(
      screen.getByRole('button', { name: 'In der Schublade' }),
    ).toHaveFocus()
  })

  // **Eine verborgene Schublade fängt ihn nicht** — gefunden an der
  // WebKit-Abnahme `ux10-tastatur` (WK03, #2104): die Handelsschublade blieb
  // nach dem Schließen verborgen im Baum (Activity), der Wechsel zu ⌘6 wollte
  // den Fokus in sie setzen, und er blieb auf `body`.
  it('geht an einer verborgenen Schublade vorbei in den Inhalt', () => {
    render(
      <>
        <Aufbewahrt />
        <Activity mode="hidden">
          <div role="dialog" aria-modal="true" aria-label="Schublade">
            <button type="button">In der Schublade</button>
          </div>
        </Activity>
      </>,
    )
    const knopf = screen.getByRole('button', { name: 'Zu den Sets' })
    knopf.focus()

    act(() => {
      fireEvent.click(knopf)
    })

    expect(screen.getByRole('main')).toHaveFocus()
  })

  // **Wer die Leiste bedient, bleibt dort** — der Knopf überlebt den Wechsel.
  it('lässt ihn stehen, wo er den Wechsel überlebt', () => {
    render(<Probe />)
    const leiste = screen.getByRole('button', { name: 'Leiste' })
    leiste.focus()

    act(() => {
      fireEvent.click(leiste)
    })

    expect(screen.getByText('sammlung')).toBeInTheDocument()
    expect(leiste).toHaveFocus()
  })

  // **Wer die Anwendung öffnet, beginnt mit Tab beim Sprunglink** — beim
  // ersten Zeichnen ist nichts gewechselt.
  it('rührt ihn beim ersten Zeichnen nicht an', () => {
    render(<Probe />)

    expect(document.activeElement).toBe(document.body)
  })
})
