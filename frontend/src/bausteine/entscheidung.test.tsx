// **Eine Rückfrage an Ort und Stelle** — Werkstück U2 von Phase 5.1
// (#1413).
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { Entscheidung } from './Entscheidung'
import { merkeEscape } from '../test/escape'

// Probe stellt eine Rückfrage hinter einen Auslöser — so, wie jede Seite sie
// benutzt: der Knopf öffnet sie, die Antwort schließt sie.
function Probe({
  aufBestaetigen = () => undefined,
}: {
  aufBestaetigen?: () => void
}) {
  const [gefragt, setzeGefragt] = useState(false)

  return (
    <>
      <button
        type="button"
        onClick={() => {
          setzeGefragt(true)
        }}
      >
        Löschen …
      </button>

      {gefragt && (
        <Entscheidung
          frage="Diese Karte löschen?"
          erklaerung="Mit ⌘Z lässt es sich zurücknehmen."
          bestaetigen="Löschen"
          abbrechen="Abbrechen"
          aufBestaetigen={() => {
            aufBestaetigen()
            setzeGefragt(false)
          }}
          aufAbbrechen={() => {
            setzeGefragt(false)
          }}
        />
      )}
    </>
  )
}

describe('Entscheidung', () => {
  // **Auch wenn der Klick keinen Fokus gab** — UX-08.AK-03 (#1680). WebKit
  // gibt einem Knopf beim Klick keinen Fokus, die Gewohnheit von macOS; die
  // Frage merkte sich dann `body` und gab den Fokus beim Schließen dorthin.
  // Gefunden von der WebKit-Probe `ux10-aktionen` am 2026-09-25. Nachgestellt
  // mit `fireEvent`, das anders als `userEvent` keinen Fokus setzt.
  //
  // **Und der Fokus bleibt dabei, wo er war** — das zeigte erst die zweite
  // Probe: WebKit lässt ihn beim Klick auf dem zuvor fokussierten Element, hier
  // „Navigator ausblenden", und genau dorthin kehrte er zurück.
  it('gibt den Fokus dem gedrückten Knopf zurück, der keinen bekam', () => {
    render(
      <>
        <input aria-label="Vorher" />
        <Probe />
      </>,
    )

    screen.getByRole('textbox', { name: 'Vorher' }).focus()
    const knopf = screen.getByRole('button', { name: 'Löschen …' })
    fireEvent.pointerDown(knopf)
    fireEvent.click(knopf)
    expect(knopf).not.toHaveFocus()
    expect(screen.getByRole('button', { name: 'Abbrechen' })).toHaveFocus()

    fireEvent.keyDown(screen.getByRole('button', { name: 'Abbrechen' }), {
      key: 'Escape',
    })

    expect(knopf).toHaveFocus()
  })

  it('ist eine Gruppe, die ihre Frage als Namen trägt', async () => {
    render(<Probe />)

    await userEvent.click(screen.getByRole('button', { name: 'Löschen …' }))

    const gruppe = screen.getByRole('group', { name: 'Diese Karte löschen?' })
    expect(gruppe).toHaveAccessibleDescription(
      'Mit ⌘Z lässt es sich zurücknehmen.',
    )
  })

  // **Der Fokus springt auf das Abbrechen** — die ungefährliche Antwort. Wer
  // zweimal ⏎ drückt, hat damit nichts gelöscht.
  it('setzt den Fokus auf das Abbrechen', async () => {
    render(<Probe />)

    await userEvent.click(screen.getByRole('button', { name: 'Löschen …' }))

    expect(screen.getByRole('button', { name: 'Abbrechen' })).toHaveFocus()
  })

  it('bestätigt und gibt den Fokus an den Auslöser zurück', async () => {
    const bestaetigt = vi.fn()
    render(<Probe aufBestaetigen={bestaetigt} />)

    const ausloeser = screen.getByRole('button', { name: 'Löschen …' })
    await userEvent.click(ausloeser)
    await userEvent.click(screen.getByRole('button', { name: 'Löschen' }))

    expect(bestaetigt).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('group')).toBeNull()
    expect(ausloeser).toHaveFocus()
  })

  // **Escape ist Abbrechen** — und nur hier: das Ereignis steigt nicht zum
  // Fenster auf, wo die Palette darauf horcht.
  it('bricht mit Escape ab und gibt den Fokus zurück', async () => {
    const bestaetigt = vi.fn()
    const fenster = vi.fn()
    window.addEventListener('keydown', fenster)

    render(<Probe aufBestaetigen={bestaetigt} />)

    const ausloeser = screen.getByRole('button', { name: 'Löschen …' })
    await userEvent.click(ausloeser)
    await userEvent.keyboard('{Escape}')

    expect(bestaetigt).not.toHaveBeenCalled()
    expect(screen.queryByRole('group')).toBeNull()
    expect(ausloeser).toHaveFocus()
    expect(fenster).not.toHaveBeenCalled()

    window.removeEventListener('keydown', fenster)
  })

  // **Und es sagt, dass es die Taste verbraucht hat** — sonst schlösse eine
  // Schublade um die Rückfrage mit demselben Escape gleich mit (UX-08,
  // #1680). `stopPropagation` hält nur, wer **nach** ihm horcht; die
  // Schublade horcht in der Capture-Phase und damit davor.
  it('verbraucht das Escape, mit dem es abbricht', async () => {
    render(<Probe aufBestaetigen={vi.fn()} />)

    await userEvent.click(screen.getByRole('button', { name: 'Löschen …' }))

    const verbraucht = merkeEscape()
    await userEvent.keyboard('{Escape}')

    expect(verbraucht()).toBe(true)
  })

  // **Solange es läuft, ruht beides** (§7, Regel 4).
  it('hält beide Knöpfe an, solange der Vorgang läuft', () => {
    render(
      <Entscheidung
        frage="Dieses Set löschen?"
        bestaetigen="Endgültig löschen"
        abbrechen="Abbrechen"
        laeuft
        aufBestaetigen={() => undefined}
        aufAbbrechen={() => undefined}
      />,
    )

    expect(
      screen.getByRole('button', { name: 'Endgültig löschen' }),
    ).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Abbrechen' })).toBeDisabled()
  })

  // **Ohne Erklärung keine Beschreibung** — ein `aria-describedby`, das ins
  // Leere zeigt, ist schlechter als keines.
  it('trägt ohne Erklärung keine Beschreibung', () => {
    render(
      <Entscheidung
        frage="Dieses Deck löschen?"
        bestaetigen="Löschen"
        abbrechen="Abbrechen"
        aufBestaetigen={() => undefined}
        aufAbbrechen={() => undefined}
      >
        <p>Die Karten selbst bleiben.</p>
      </Entscheidung>,
    )

    const gruppe = screen.getByRole('group', { name: 'Dieses Deck löschen?' })
    expect(gruppe).not.toHaveAttribute('aria-describedby')
    expect(screen.getByText('Die Karten selbst bleiben.')).toBeInTheDocument()
  })

  // **Ein Auslöser, den es nicht mehr gibt, bekommt keinen Fokus** — nach
  // einem Löschen ist die Zeile weg, und der Fokus fällt auf den Anfang der
  // Seite statt ins Nirgendwo.
  it('lässt den Fokus, wo er ist, wenn der Auslöser verschwunden ist', async () => {
    function Zeile() {
      const [gefragt, setzeGefragt] = useState(false)
      const [weg, setzeWeg] = useState(false)

      if (weg) {
        return <p>Die Zeile ist weg.</p>
      }

      return (
        <>
          <button
            type="button"
            onClick={() => {
              setzeGefragt(true)
            }}
          >
            Entfernen …
          </button>
          {gefragt && (
            <Entscheidung
              frage="Dieses Paket entfernen?"
              bestaetigen="Entfernen"
              abbrechen="Abbrechen"
              aufBestaetigen={() => {
                setzeWeg(true)
              }}
              aufAbbrechen={() => {
                setzeGefragt(false)
              }}
            />
          )}
        </>
      )
    }

    render(<Zeile />)

    await userEvent.click(screen.getByRole('button', { name: 'Entfernen …' }))
    await userEvent.click(screen.getByRole('button', { name: 'Entfernen' }))

    expect(screen.getByText('Die Zeile ist weg.')).toBeInTheDocument()
    expect(document.activeElement).toBe(document.body)
  })
})

// **Eine andere Taste tut nichts** — nur Escape ist Abbrechen.
it('lässt andere Tasten in der Gruppe in Ruhe', async () => {
  render(
    <Entscheidung
      frage="Dieses Set löschen?"
      bestaetigen="Löschen"
      abbrechen="Abbrechen"
      aufBestaetigen={() => undefined}
      aufAbbrechen={() => undefined}
    />,
  )

  await userEvent.keyboard('a')

  expect(
    screen.getByRole('group', { name: 'Dieses Set löschen?' }),
  ).toBeInTheDocument()
})
