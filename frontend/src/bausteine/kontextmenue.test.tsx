// **Ein Kontextmenü, das auch ohne rechte Maustaste aufgeht** — Werkstück
// U2 von Phase 5.1 (#1413).
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Kontextmenue } from './Kontextmenue'
import { type Menueeintrag } from './Menue'
import { merkeEscape } from '../test/escape'

function eintraege(gewaehlt: (schluessel: string) => void): Menueeintrag[] {
  return [
    {
      schluessel: 'kopieren',
      name: 'Karte duplizieren',
      aufWahl: () => {
        gewaehlt('kopieren')
      },
    },
    {
      schluessel: 'loeschen',
      name: 'Karte löschen',
      aufWahl: () => {
        gewaehlt('loeschen')
      },
    },
  ]
}

function Probe({
  gewaehlt = () => undefined,
}: {
  gewaehlt?: (s: string) => void
}) {
  return (
    <Kontextmenue
      bezeichnung="Handlungen für Blitzschlag"
      eintraege={eintraege(gewaehlt)}
    >
      <button type="button">Blitzschlag</button>
    </Kontextmenue>
  )
}

describe('Kontextmenue', () => {
  it('steht zu, bis jemand es aufruft', () => {
    render(<Probe />)

    expect(screen.queryByRole('menu')).toBeNull()
    expect(
      screen.getByRole('button', { name: 'Blitzschlag' }),
    ).toBeInTheDocument()
  })

  it('geht mit der rechten Maustaste auf und wählt mit ⏎', async () => {
    const gewaehlt = vi.fn()
    render(<Probe gewaehlt={gewaehlt} />)

    const zeile = screen.getByRole('button', { name: 'Blitzschlag' })
    zeile.focus()
    fireEvent.contextMenu(zeile, { clientX: 40, clientY: 20 })

    const menue = screen.getByRole('menu', {
      name: 'Handlungen für Blitzschlag',
    })
    expect(menue).toBeInTheDocument()
    expect(
      screen.getByRole('menuitem', { name: 'Karte duplizieren' }),
    ).toHaveFocus()

    await userEvent.keyboard('{ArrowDown}{Enter}')

    expect(gewaehlt).toHaveBeenCalledWith('loeschen')
    expect(screen.queryByRole('menu')).toBeNull()
    expect(zeile).toHaveFocus()
  })

  // **Eine Ebene über der Seite, und sie taucht auf** — AP-4 (#2361),
  // Vermerk 0065 Festlegungen 10 und 11.
  it('taucht mit Kante und einer Schattenstufe auf', () => {
    render(<Probe />)

    const zeile = screen.getByRole('button', { name: 'Blitzschlag' })
    fireEvent.contextMenu(zeile, { clientX: 40, clientY: 20 })

    expect(screen.getByRole('menu')).toHaveClass(
      'auftauchen',
      'border-2',
      'shadow-ueberlage',
      'contrast-more:shadow-none',
    )
  })

  // **⇧F10 ist der Weg ohne Maus** — und die Menütaste dazu.
  it('geht mit ⇧F10 und der Menütaste auf', async () => {
    render(<Probe />)

    const zeile = screen.getByRole('button', { name: 'Blitzschlag' })
    zeile.focus()
    await userEvent.keyboard('{Shift>}{F10}{/Shift}')
    expect(screen.getByRole('menu')).toBeInTheDocument()

    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('menu')).toBeNull()
    expect(zeile).toHaveFocus()

    await userEvent.keyboard('{ContextMenu}')
    expect(screen.getByRole('menu')).toBeInTheDocument()
  })

  // **Das Escape, das das Menü schließt, ist verbraucht** — eine Schublade
  // darum bleibt offen (UX-08, #1680).
  it('verbraucht das Escape, mit dem es schließt', async () => {
    render(<Probe />)

    screen.getByRole('button', { name: 'Blitzschlag' }).focus()
    await userEvent.keyboard('{Shift>}{F10}{/Shift}')

    const verbraucht = merkeEscape()
    await userEvent.keyboard('{Escape}')

    expect(screen.queryByRole('menu')).toBeNull()
    expect(verbraucht()).toBe(true)
  })

  it('schließt bei einem Klick daneben und mit Tab', async () => {
    render(
      <>
        <Probe />
        <p>Daneben</p>
      </>,
    )

    const zeile = screen.getByRole('button', { name: 'Blitzschlag' })
    fireEvent.contextMenu(zeile)
    expect(screen.getByRole('menu')).toBeInTheDocument()

    await userEvent.click(screen.getByText('Daneben'))
    expect(screen.queryByRole('menu')).toBeNull()

    fireEvent.contextMenu(zeile)
    await userEvent.tab()
    expect(screen.queryByRole('menu')).toBeNull()
  })

  // **Und der Fokus geht weiter, nicht zurück** — wie beim Menü hinter einem
  // Knopf. Seit user-event 14.6.5 (#1784) bestimmt die Vorrichtung das
  // nächste Ziel erst nach dem `keydown`, so wie ein Browser: wer das Menü
  // darin schon abbaute, ließ den Fokus ins Dokument fallen, und der Tab
  // begann wieder vorn.
  it('gibt den Fokus mit Tab an das nächste Element weiter', async () => {
    render(
      <>
        <Probe />
        <button type="button">Danach</button>
      </>,
    )

    fireEvent.contextMenu(screen.getByRole('button', { name: 'Blitzschlag' }))
    await userEvent.tab()

    expect(screen.getByRole('button', { name: 'Danach' })).toHaveFocus()
    await waitFor(() => {
      expect(screen.queryByRole('menu')).toBeNull()
    })
  })

  it('wandert mit Home und End', async () => {
    render(<Probe />)

    fireEvent.contextMenu(screen.getByRole('button', { name: 'Blitzschlag' }))
    await userEvent.keyboard('{End}')
    expect(
      screen.getByRole('menuitem', { name: 'Karte löschen' }),
    ).toHaveFocus()

    await userEvent.keyboard('{Home}')
    expect(
      screen.getByRole('menuitem', { name: 'Karte duplizieren' }),
    ).toHaveFocus()
  })

  it('wählt einen ruhenden Eintrag nicht', async () => {
    const gewaehlt = vi.fn()
    render(
      <Kontextmenue
        bezeichnung="Handlungen"
        eintraege={[
          {
            schluessel: 'oben',
            name: 'Nach oben',
            aufWahl: () => {
              gewaehlt('oben')
            },
            angehalten: true,
          },
          {
            schluessel: 'unten',
            name: 'Nach unten',
            aufWahl: () => {
              gewaehlt('unten')
            },
          },
        ]}
      >
        <button type="button">Zeile</button>
      </Kontextmenue>,
    )

    fireEvent.contextMenu(screen.getByRole('button', { name: 'Zeile' }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Nach oben' }))

    expect(gewaehlt).not.toHaveBeenCalled()
    expect(screen.getByRole('menu')).toBeInTheDocument()
  })

  // **Ohne Einträge kein Menü** — und der Rechtsklick bleibt, was er ist.
  // **Ein Rechtsklick in ein Feld gehört macOS** (GESTE, #2358): dort stehen
  // Kopieren, Einsetzen und die Rechtschreibung. Liegt das Menü um eine
  // ganze Karte samt ihren Feldern, ginge das sonst verloren.
  it.each([
    ['ein Eingabefeld', <input key="e" aria-label="Name" />, 'Name'],
    [
      'ein Textbereich',
      <textarea key="t" aria-label="Regeltext" />,
      'Regeltext',
    ],
    [
      'Bearbeitbares',
      <div key="b" contentEditable aria-label="Freitext" role="textbox" />,
      'Freitext',
    ],
  ])('lässt den Rechtsklick in %s durch', (_, feld, name) => {
    render(
      <Kontextmenue bezeichnung="Handlungen" eintraege={eintraege(vi.fn())}>
        <div>{feld}</div>
      </Kontextmenue>,
    )

    const ereignis = fireEvent.contextMenu(
      screen.getByRole('textbox', { name }),
    )

    expect(ereignis).toBe(true)
    expect(screen.queryByRole('menu')).toBeNull()
  })

  it('tut ohne Einträge nichts', () => {
    render(
      <Kontextmenue bezeichnung="Nichts" eintraege={[]}>
        <button type="button">Zeile</button>
      </Kontextmenue>,
    )

    const zeile = screen.getByRole('button', { name: 'Zeile' })
    fireEvent.contextMenu(zeile)
    fireEvent.keyDown(zeile, { key: 'F10', shiftKey: true })

    expect(screen.queryByRole('menu')).toBeNull()
  })
})
