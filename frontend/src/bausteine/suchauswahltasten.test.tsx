// **Eine Auswahl mit Suche, Teil 2: Tippen, Tasten und Maus** — UXR-F1
// (#2007).
//
// **Alles über die rohe Eingabe** — Tippen, Pfeile, Enter, Escape, Klick (§3):
// ein Lauf über fertige Zustände sähe nicht, ob die Taste ankommt. Die Gestalt
// steht in `suchauswahl.test.tsx`.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { merkeEscape } from '../test/escape'
import {
  Probe,
  angeboten,
  ansage,
  feld,
  stile,
  zonen,
} from '../test/suchauswahlprobe'

const worte = kataloge.de.bausteine.suchauswahl

describe('Suchauswahl — Tippen', () => {
  it.each([
    ['über den Namen', 'cairo', ['Africa/Cairo']],
    ['ohne Rücksicht auf Groß und klein', 'PARIS', ['Europe/Paris']],
    ['über ein Teilwort', 'rop', ['Europe/Berlin', 'Europe/Paris']],
    ['mit Leerzeichen statt Unterstrich', 'new york', ['America/New_York']],
  ])('grenzt die Liste %s ein', async (_, suche, soll) => {
    render(<Probe />)

    await userEvent.click(feld())
    await userEvent.keyboard(suche)

    expect(angeboten()).toEqual(soll)
  })

  it('durchsucht auch den Zusatz', async () => {
    render(<Probe eintraege={stile} anfang="magic-m15" />)

    await userEvent.click(feld())
    await userEvent.keyboard('m15-alt')

    expect(angeboten()).toEqual(['Altered Art magic-m15-altered'])
  })

  // **Tippen öffnet**, auch wenn nur der Fokus da war — wer tippt, will sehen,
  // was passt.
  it('öffnet beim Tippen, ohne dass vorher geklickt wurde', async () => {
    render(<Probe />)

    await userEvent.tab()
    await userEvent.keyboard('{Control>}a{/Control}pa')

    expect(feld()).toHaveValue('pa')
    expect(angeboten()).toEqual(['Europe/Paris'])
  })

  it('hebt den ersten Treffer hervor, damit Enter ihn nimmt', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)

    await userEvent.click(feld())
    await userEvent.keyboard('paris{Enter}')

    expect(aendern).toHaveBeenCalledExactlyOnceWith('Europe/Paris')
    expect(feld()).toHaveValue('Europe/Paris')
  })

  // **Eine leere Suche wählt nichts**: wer den Text löscht und Enter drückt,
  // soll nicht die erste Zeile der Liste bekommen.
  it('wählt bei leerer Suche mit Enter nichts', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)

    // **Getippt und wieder gelöscht** — eine leere Suche, nicht die
    // ungetippte Liste, in der die Wahl hervorgehoben steht.
    await userEvent.click(feld())
    await userEvent.keyboard('x{Backspace}{Enter}')

    expect(aendern).not.toHaveBeenCalled()
    expect(feld()).not.toHaveAttribute('aria-activedescendant')
    expect(angeboten()).toHaveLength(5)
  })

  it('meldet, was nicht passt, mit einem Satz aus dem Katalog', async () => {
    render(<Probe />)

    await userEvent.click(feld())
    await userEvent.keyboard('tokio')

    const satz = worte.keiner.replace('{suche}', 'tokio')

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(feld()).toHaveAttribute('aria-expanded', 'false')
    // **Zweimal**: sichtbar in der Liste und ansagend in der Live-Region.
    expect(screen.getAllByText(satz)).toHaveLength(2)
    expect(ansage()).toHaveTextContent(satz)
    expect(feld()).not.toHaveAttribute('aria-controls')
    expect(feld()).not.toHaveAttribute('aria-activedescendant')
  })

  it('bleibt mit den Pfeilen stehen, wenn nichts passt', async () => {
    render(<Probe />)

    await userEvent.click(feld())
    await userEvent.keyboard('tokio{ArrowDown}{ArrowUp}')

    expect(feld()).toHaveValue('tokio')
    expect(feld()).not.toHaveAttribute('aria-activedescendant')
  })

  it('wählt mit Enter nichts, wenn nichts passt', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)

    await userEvent.click(feld())
    await userEvent.keyboard('tokio{Enter}')

    expect(aendern).not.toHaveBeenCalled()
  })

  it('sagt die Zahl der Treffer an', async () => {
    render(<Probe />)

    expect(ansage()).toBeEmptyDOMElement()

    await userEvent.click(feld())

    expect(ansage()).toHaveTextContent(worte.treffer.replace('{anzahl}', '5'))

    await userEvent.keyboard('europe')

    expect(ansage()).toHaveTextContent(worte.treffer.replace('{anzahl}', '2'))
  })
})

describe('Suchauswahl — die Tastatur', () => {
  it('öffnet mit Pfeil nach unten und hebt die Wahl hervor', async () => {
    render(<Probe />)
    await userEvent.tab()

    await userEvent.keyboard('{ArrowDown}')

    expect(screen.getByRole('listbox')).toBeInTheDocument()
    expect(feld()).toHaveAttribute(
      'aria-activedescendant',
      screen.getByRole('option', { name: 'Europe/Berlin' }).id,
    )
  })

  it('öffnet mit Pfeil nach oben', async () => {
    render(<Probe />)
    await userEvent.tab()

    await userEvent.keyboard('{ArrowUp}')

    expect(screen.getByRole('listbox')).toBeInTheDocument()
  })

  // **Ohne Wahl beginnt der Pfeil oben**, nicht in der Luft.
  it('beginnt ohne Wahl bei der ersten Zeile', async () => {
    render(<Probe anfang="" eintraege={zonen.slice(1)} />)
    await userEvent.tab()

    await userEvent.keyboard('{ArrowDown}')

    expect(feld()).toHaveAttribute(
      'aria-activedescendant',
      screen.getByRole('option', { name: 'Africa/Cairo' }).id,
    )
  })

  it('wandert mit den Pfeilen durch die Liste und verfängt sich nicht', async () => {
    render(<Probe />)
    await userEvent.tab()
    await userEvent.keyboard('{ArrowDown}')

    await userEvent.keyboard('{ArrowDown}')

    expect(feld()).toHaveAttribute(
      'aria-activedescendant',
      screen.getByRole('option', { name: 'Europe/Paris' }).id,
    )

    // **Rundherum**: von der letzten Zeile geht es zur ersten.
    await userEvent.keyboard('{ArrowDown}')

    expect(feld()).toHaveAttribute(
      'aria-activedescendant',
      screen.getByRole('option', { name: 'Wie das System' }).id,
    )

    await userEvent.keyboard('{ArrowUp}')

    expect(feld()).toHaveAttribute(
      'aria-activedescendant',
      screen.getByRole('option', { name: 'Europe/Paris' }).id,
    )
  })

  it('springt bei offener Liste ohne Hervorhebung mit Pfeil nach oben zur letzten Zeile', async () => {
    render(<Probe anfang="" eintraege={zonen.slice(1)} />)
    await userEvent.click(feld())

    expect(feld()).not.toHaveAttribute('aria-activedescendant')

    await userEvent.keyboard('{ArrowUp}')

    expect(feld()).toHaveAttribute(
      'aria-activedescendant',
      screen.getByRole('option', { name: 'Europe/Paris' }).id,
    )
  })

  it('wählt mit Enter und schließt die Liste', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)
    await userEvent.tab()

    await userEvent.keyboard('{ArrowDown}{ArrowDown}{Enter}')

    expect(aendern).toHaveBeenCalledExactlyOnceWith('Europe/Paris')
    expect(feld()).toHaveValue('Europe/Paris')
    expect(feld()).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(feld()).toHaveFocus()
  })

  it('wählt mit Enter nichts, wenn die Liste zu ist', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)
    await userEvent.tab()

    await userEvent.keyboard('{Enter}')

    expect(aendern).not.toHaveBeenCalled()
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  // **Die Wahl, die schon gilt, ist keine Änderung**: ein Schreibweg hinter
  // `aendern` liefe sonst für nichts.
  it('meldet nichts, wenn dieselbe Zeile noch einmal gewählt wird', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)

    await userEvent.click(feld())
    await userEvent.keyboard('{Enter}')

    expect(aendern).not.toHaveBeenCalled()
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  // **Ein Enter während einer Eingabe mit Eingabemethode** (Japanisch,
  // Chinesisch) bestätigt die Umwandlung und wählt keine Zeile.
  it('wählt nichts, solange eine Eingabemethode umwandelt', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)
    await userEvent.click(feld())
    await userEvent.keyboard('{ArrowDown}')

    fireEvent.keyDown(feld(), { key: 'Enter', isComposing: true })

    expect(aendern).not.toHaveBeenCalled()
    expect(screen.getByRole('listbox')).toBeInTheDocument()
  })

  it('lässt andere Tasten dem Feld', async () => {
    render(<Probe />)
    await userEvent.click(feld())

    await userEvent.keyboard('{Home}{End}{ArrowLeft}')

    expect(screen.getByRole('listbox')).toBeInTheDocument()
  })
})

describe('Suchauswahl — Escape', () => {
  // **Escape schließt, stellt den vorigen Wert her und nimmt die Taste**:
  // eine Schublade um das Feld bliebe sonst nicht stehen.
  it('schließt die Liste, verwirft die Eingabe und verbraucht die Taste', async () => {
    const aendern = vi.fn()
    render(<Probe aendern={aendern} />)
    const verbraucht = merkeEscape()

    await userEvent.click(feld())
    await userEvent.keyboard('paris')

    expect(feld()).toHaveValue('paris')

    await userEvent.keyboard('{Escape}')

    expect(verbraucht()).toBe(true)
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(feld()).toHaveValue('Europe/Berlin')
    expect(feld()).toHaveAttribute('aria-expanded', 'false')
    expect(aendern).not.toHaveBeenCalled()
    expect(feld()).toHaveFocus()
  })

  // **Ist die Liste zu, gehört Escape den anderen**: sonst ließe sich eine
  // Schublade mit einem Auswahlfeld darin nie schließen.
  it('lässt Escape bei geschlossener Liste durch', async () => {
    render(<Probe />)
    const verbraucht = merkeEscape()
    await userEvent.tab()

    await userEvent.keyboard('{Escape}')

    expect(verbraucht()).toBe(false)
  })

  it('hält die Taste von außen fern, solange die Liste offen ist', async () => {
    // **Am Dokument, in der Bubble-Phase**: dort hören Rahmen, die ein Escape
    // für sich nehmen wollen, nachdem das Feld es gesehen hat.
    const draussen = vi.fn()
    document.addEventListener('keydown', draussen)

    try {
      render(<Probe />)

      await userEvent.click(feld())
      await userEvent.keyboard('{Escape}')

      expect(draussen).not.toHaveBeenCalled()

      await userEvent.keyboard('{Escape}')

      expect(draussen).toHaveBeenCalledTimes(1)
    } finally {
      document.removeEventListener('keydown', draussen)
    }
  })
})
