// **Der eine modale Dialog** — Werkstück U2 von Phase 5.1
// (#1413).
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { Dialog } from './Dialog'
import { Sprachraum } from '../sprache/kontext'

function Probe() {
  const [offen, setzeOffen] = useState(false)

  return (
    <>
      <button
        type="button"
        onClick={() => {
          setzeOffen(true)
        }}
      >
        Palette öffnen
      </button>

      {offen && (
        <Dialog
          bezeichnung="Befehlspalette"
          schliessen={() => {
            setzeOffen(false)
          }}
        >
          <div>
            <input type="text" aria-label="Wohin?" />
            <button type="button">Erster</button>
            <button type="button">Letzter</button>
          </div>
        </Dialog>
      )}
    </>
  )
}

describe('Dialog', () => {
  it('ist modal, trägt seinen Namen und setzt den Fokus auf das erste Feld', async () => {
    render(
      <Sprachraum sprache="de">
        <Probe />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: 'Palette öffnen' }),
    )

    const dialog = screen.getByRole('dialog', { name: 'Befehlspalette' })
    expect(dialog).toHaveAttribute('aria-modal', 'true')
    expect(screen.getByRole('textbox', { name: 'Wohin?' })).toHaveFocus()
  })

  // **Er taucht auf, statt zu springen** — AP-4 (#2361), B-11: Deckkraft und
  // vier Punkte Lage in `--dauer-mittel` (`auftauchen`), der Schleier mit.
  // Mantines eigener Übergang spielt hier nicht: der Dialog wird offen
  // gezeichnet, und Mantine beginnt dann im Zustand „eingetreten“.
  it('taucht mit Kasten und Schleier in der mittleren Dauer auf', async () => {
    render(
      <Sprachraum sprache="de">
        <Probe />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: 'Palette öffnen' }),
    )

    const dialog = screen.getByRole('dialog', { name: 'Befehlspalette' })
    expect(dialog).toHaveClass('auftauchen')
    expect(document.querySelector('.mantine-Modal-overlay')).toHaveClass(
      'auftauchen',
    )
  })

  it('schließt mit Escape und gibt den Fokus an den Auslöser zurück', async () => {
    render(
      <Sprachraum sprache="de">
        <Probe />
      </Sprachraum>,
    )

    const ausloeser = screen.getByRole('button', { name: 'Palette öffnen' })
    await userEvent.click(ausloeser)
    await userEvent.keyboard('{Escape}')

    expect(screen.queryByRole('dialog')).toBeNull()

    // **Gewartet wird auf die Rückgabe, nicht auf das Schließen.**
    //
    // Mantine gibt den Fokus in einem Effekt zurück, nicht im selben Zug wie
    // das Schließen — und §3 sagt dazu: gewartet wird auf die **Wirkung**.
    // Die eigene Fassung tat es synchron in einem `useEffect`-Aufräumer; die
    // Zusage ist dieselbe geblieben, ihr Zeitpunkt nicht.
    await waitFor(() => {
      expect(ausloeser).toHaveFocus()
    })
  })

  // **Escape greift auch, wenn der Fokus den Kasten verlassen hat** — ein
  // Klick auf den Schleier setzt ihn auf den Anfang der Seite, und der Dialog
  // muss sich trotzdem schließen lassen. Dieselbe Zusage wie bei der Palette
  // seit #245.
  it('schließt mit Escape, auch wenn der Fokus draußen steht', async () => {
    render(
      <Sprachraum sprache="de">
        <Probe />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: 'Palette öffnen' }),
    )
    ;(document.activeElement as HTMLElement).blur()
    expect(document.activeElement).toBe(document.body)

    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('steht wahlweise in der Mitte', () => {
    render(
      <Sprachraum sprache="de">
        <Dialog
          bezeichnung="Mitte"
          ausrichtung="mitte"
          schliessen={() => undefined}
        >
          <p>Gelesen</p>
        </Dialog>
      </Sprachraum>,
    )

    expect(screen.getByRole('dialog', { name: 'Mitte' })).toBeInTheDocument()
  })

  // **Tab bleibt im Kasten** — vom letzten auf das erste, mit ⇧ vom ersten
  // auf das letzte.
  it('hält den Fokus im Kasten', async () => {
    render(
      <Sprachraum sprache="de">
        <Probe />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: 'Palette öffnen' }),
    )

    await userEvent.tab()
    await userEvent.tab()
    expect(screen.getByRole('button', { name: 'Letzter' })).toHaveFocus()

    await userEvent.tab()
    expect(screen.getByRole('textbox', { name: 'Wohin?' })).toHaveFocus()

    await userEvent.tab({ shift: true })
    expect(screen.getByRole('button', { name: 'Letzter' })).toHaveFocus()
  })

  // **Ein Dialog ohne Bedienelement fängt nichts** — Tab tut, was der
  // Browser tut. Der Zweig steht, damit ein leerer Kasten nicht wirft.
  it('lässt Tab in einem leeren Kasten in Ruhe', () => {
    render(
      <Sprachraum sprache="de">
        <Dialog bezeichnung="Leer" schliessen={() => undefined}>
          <p>Nur Text</p>
        </Dialog>
      </Sprachraum>,
    )

    // **Die Taste trifft den Kasten selbst**: ohne Bedienelement liegt der
    // Fokus nirgends darin, und ein `userEvent.tab()` erreichte den Horcher
    // gar nicht.
    fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Tab' })
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })
})
