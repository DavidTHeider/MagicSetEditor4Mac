// **Ein Menü hinter einem Knopf** — Werkstück U2 von Phase 5.1
// (#1413).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Menue, naechsteStelle, type Menueeintrag } from './Menue'
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
    {
      schluessel: 'loeschen',
      name: 'Karte löschen',
      aufWahl: () => {
        gewaehlt('loeschen')
      },
    },
  ]
}

describe('naechsteStelle', () => {
  const liste = eintraege(() => undefined)

  it.each([
    ['ArrowDown', 0, 2],
    ['ArrowDown', 3, 0],
    ['ArrowUp', 0, 3],
    ['ArrowUp', 2, 0],
    ['Home', 2, 0],
    ['End', 0, 3],
    ['Enter', 2, 2],
  ])('%s von %d führt auf %d', (taste, von, ziel) => {
    expect(naechsteStelle(liste, von, taste)).toBe(ziel)
  })

  // **Ein Menü, in dem alles ruht, bleibt stehen** — der Pfeil findet
  // nichts und ändert nichts.
  it('bleibt stehen, wenn kein Eintrag bedienbar ist', () => {
    const ruhend = liste.map((eintrag) => ({
      schluessel: eintrag.schluessel,
      name: eintrag.name,
      angehalten: true as const,
    }))

    expect(naechsteStelle(ruhend, 1, 'ArrowDown')).toBe(1)
  })
})

describe('Menue', () => {
  it('sagt am Knopf, dass ein Menü dahinter ist', () => {
    render(<Menue bezeichnung="Karte" eintraege={eintraege(() => undefined)} />)

    const knopf = screen.getByRole('button', { name: 'Karte' })
    expect(knopf).toHaveAttribute('aria-haspopup', 'menu')
    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByRole('menu')).toBeNull()
  })

  it('öffnet per Klick, setzt den Fokus auf den ersten Eintrag und wählt mit ⏎', async () => {
    const gewaehlt = vi.fn()
    render(<Menue bezeichnung="Karte" eintraege={eintraege(gewaehlt)} />)

    const knopf = screen.getByRole('button', { name: 'Karte' })
    await userEvent.click(knopf)

    const menue = screen.getByRole('menu', { name: 'Karte' })
    expect(knopf).toHaveAttribute('aria-expanded', 'true')
    expect(knopf).toHaveAttribute('aria-controls', menue.id)
    expect(
      screen.getByRole('menuitem', { name: 'Karte duplizieren' }),
    ).toHaveFocus()

    await userEvent.keyboard('{Enter}')

    expect(gewaehlt).toHaveBeenCalledWith('kopieren')
    expect(screen.queryByRole('menu')).toBeNull()
    expect(knopf).toHaveFocus()
  })

  // **↓ am Knopf öffnet und geht auf den ersten, ↑ auf den letzten** — das
  // Muster der WAI-ARIA Authoring Practices für einen Menüknopf.
  it('öffnet mit ↓ auf dem ersten und mit ↑ auf dem letzten Eintrag', async () => {
    render(<Menue bezeichnung="Karte" eintraege={eintraege(() => undefined)} />)

    const knopf = screen.getByRole('button', { name: 'Karte' })
    knopf.focus()
    await userEvent.keyboard('{ArrowUp}')

    expect(
      screen.getByRole('menuitem', { name: 'Karte löschen' }),
    ).toHaveFocus()

    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('menu')).toBeNull()
    expect(knopf).toHaveFocus()

    await userEvent.keyboard('{ArrowDown}')
    expect(
      screen.getByRole('menuitem', { name: 'Karte duplizieren' }),
    ).toHaveFocus()
  })

  // **Das Escape, das das Menü schließt, ist verbraucht** — eine Schublade
  // darum bleibt offen (UX-08, #1680).
  it('verbraucht das Escape, mit dem es schließt', async () => {
    render(<Menue bezeichnung="Karte" eintraege={eintraege(() => undefined)} />)

    screen.getByRole('button', { name: 'Karte' }).focus()
    await userEvent.keyboard('{ArrowDown}')

    const verbraucht = merkeEscape()
    await userEvent.keyboard('{Escape}')

    expect(screen.queryByRole('menu')).toBeNull()
    expect(verbraucht()).toBe(true)
  })

  it('wandert mit den Pfeilen über ruhende Einträge hinweg und zyklisch', async () => {
    render(<Menue bezeichnung="Karte" eintraege={eintraege(() => undefined)} />)

    await userEvent.click(screen.getByRole('button', { name: 'Karte' }))
    await userEvent.keyboard('{ArrowDown}')

    // **„Nach oben" ruht** und wird übersprungen.
    expect(screen.getByRole('menuitem', { name: 'Nach unten' })).toHaveFocus()
    expect(screen.getByRole('menuitem', { name: 'Nach oben' })).toHaveAttribute(
      'aria-disabled',
      'true',
    )

    await userEvent.keyboard('{End}')
    expect(
      screen.getByRole('menuitem', { name: 'Karte löschen' }),
    ).toHaveFocus()

    await userEvent.keyboard('{ArrowDown}')
    expect(
      screen.getByRole('menuitem', { name: 'Karte duplizieren' }),
    ).toHaveFocus()

    await userEvent.keyboard('{Home}')
    expect(
      screen.getByRole('menuitem', { name: 'Karte duplizieren' }),
    ).toHaveFocus()
  })

  it('wählt einen ruhenden Eintrag nicht', async () => {
    const gewaehlt = vi.fn()
    render(<Menue bezeichnung="Karte" eintraege={eintraege(gewaehlt)} />)

    await userEvent.click(screen.getByRole('button', { name: 'Karte' }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Nach oben' }))

    expect(gewaehlt).not.toHaveBeenCalled()
    expect(screen.getByRole('menu')).toBeInTheDocument()
  })

  // **Tab verlässt das Menü und schließt es** — ohne den Fokus
  // zurückzuholen, denn wer Tab drückt, will weiter.
  it('schließt mit Tab, ohne den Fokus zurückzuholen', async () => {
    render(
      <>
        <Menue bezeichnung="Karte" eintraege={eintraege(() => undefined)} />
        <button type="button">Danach</button>
      </>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Karte' }))
    await userEvent.tab()

    // **Weiter, nicht bloß weg** — `not.toHaveFocus()` auf dem Knopf hielt
    // auch, wenn der Fokus ins Dokument fiel. Seit user-event 14.6.5 (#1784)
    // fiel er dann auf den Knopf zurück, so wie ein Browser von vorn beginnt.
    expect(screen.getByRole('button', { name: 'Danach' })).toHaveFocus()
    await waitFor(() => {
      expect(screen.queryByRole('menu')).toBeNull()
    })
  })

  it('schließt bei einem Klick daneben', async () => {
    render(
      <>
        <Menue bezeichnung="Karte" eintraege={eintraege(() => undefined)} />
        <p>Daneben</p>
      </>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Karte' }))
    expect(screen.getByRole('menu')).toBeInTheDocument()

    await userEvent.click(screen.getByText('Daneben'))
    expect(screen.queryByRole('menu')).toBeNull()
  })

  it('schließt mit einem zweiten Klick auf den Knopf', async () => {
    render(<Menue bezeichnung="Karte" eintraege={eintraege(() => undefined)} />)

    const knopf = screen.getByRole('button', { name: 'Karte' })
    await userEvent.click(knopf)
    await userEvent.click(knopf)

    expect(screen.queryByRole('menu')).toBeNull()
    expect(knopf).toHaveFocus()
  })

  // **Ein Zeichen auf dem Knopf, der Name fürs Vorleseprogramm** — die
  // Paketzeile zeigt „…", nennt aber das Paket.
  it('trägt seine Bezeichnung als Namen, wenn etwas anderes darauf steht', () => {
    render(
      <Menue
        bezeichnung="Handlungen für M15"
        eintraege={eintraege(() => undefined)}
      >
        …
      </Menue>,
    )

    expect(
      screen.getByRole('button', { name: 'Handlungen für M15' }),
    ).toHaveTextContent('…')
  })
})

// **Eine andere Taste auf dem Knopf tut nichts** — ↓ und ↑ öffnen, der Rest
// bleibt dem Browser.
it('lässt andere Tasten auf dem Knopf in Ruhe', async () => {
  render(<Menue bezeichnung="Karte" eintraege={eintraege(() => undefined)} />)

  screen.getByRole('button', { name: 'Karte' }).focus()
  await userEvent.keyboard('a')

  expect(screen.queryByRole('menu')).toBeNull()
})

// **Am rechten Rand öffnet es nach links** — UXR-04 (#1912): das Menü einer
// Setkarte steht rechts in der Karte, und ein Menü, das von dort nach rechts
// aufgeht, ragte über die Kachel hinaus und am Fenster vorbei. Ohne Angabe
// bleibt es, wie es war: unter dem Knopf, nach rechts.
// **Seit AP-6** (#2363) steht das Menü fest am Fenster, gemessen am Knopf;
// links, rechts und das Aufklappen über den Knopf prüft
// `menuelage.test.tsx` gegen eine gesetzte Lage des Knopfs.
it('steht fest am Fenster', async () => {
  render(<Menue bezeichnung="Karte" eintraege={eintraege(() => undefined)} />)

  await userEvent.click(screen.getByRole('button', { name: 'Karte' }))

  expect(screen.getByRole('menu')).toHaveStyle({ position: 'fixed' })
})
