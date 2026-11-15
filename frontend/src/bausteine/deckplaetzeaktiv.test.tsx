// **Die Handlungen stehen an einer Zeile, nicht an jeder** — UXR-11
// (#1919).
//
// Ein Commanderdeck hat 76 Plätze, und bis hierher trug jeder vier Knöpfe:
// 304 Bedienelemente in der Mitte, und wer mit der Tastatur an das Ende
// wollte, drückte dreihundertmal Tabulator (R2 und R6 in `docs/UIREVIEW.md`).
// Jetzt trägt sie die Zeile mit der Tabstelle und die unter dem Zeiger —
// dieselben vier Handlungen, für jede Zeile erreichbar.
//
// **Seit AP-6** (#2363)
// **hinter einem Knopf „…“** und nicht als vier: Menge ±, Bereich und
// Herausnehmen liegen im Menü der Zeile und im Rechtsklick (`Bereiche09`) —
// drei Bedienelemente weniger im Bild.
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Deckplaetze } from './Deckplaetze'
import { Sprachraum } from '../sprache/kontext'
import { commanderdeck, standarddeck } from '../test/deckbestand'
import { kataloge } from '../sprache/kataloge'
import { an } from '../test/erstes'
import { verstoesse } from '../test/axe'

import type { DeckDTO } from '../apidecks'

const texte = kataloge.de.seiten.decks

function zeige(deck: DeckDTO) {
  const aufMenge = vi.fn()
  const aufBereich = vi.fn()
  const aufNehmen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Deckplaetze
        deck={deck}
        angehalten={false}
        unzulaessig={() => false}
        aufMenge={aufMenge}
        aufBereich={aufBereich}
        aufNehmen={aufNehmen}
      />
    </Sprachraum>,
  )

  const liste = screen.getByRole('list', { name: texte.plaetze })

  return {
    aufMenge,
    aufBereich,
    aufNehmen,
    liste,
    zeilen: within(liste).getAllByRole('listitem'),
  }
}

// menueVon nennt den Knopf „…“ des Platzes an einer Stelle.
function menueVon(stelle: number, deck: DeckDTO = commanderdeck): string {
  return `${texte.plaetze}: ${name(deck, stelle)}`
}

// name gibt den Namen des Platzes an einer Stelle.
function name(deck: DeckDTO, stelle: number): string {
  return deck.plaetze[stelle]?.bezeichnung ?? ''
}

describe('die Handlungen der Plätze', () => {
  it('stehen an einer Zeile hinter »…« — ein Commanderdeck trägt einen Knopf und nicht 304', async () => {
    const { aufMenge, liste, zeilen } = zeige(commanderdeck)

    expect(zeilen).toHaveLength(76)
    expect(within(liste).getAllByRole('button')).toHaveLength(1)

    await userEvent.click(
      within(an(zeilen, 0)).getByRole('button', { name: menueVon(0) }),
    )
    await userEvent.click(screen.getByRole('menuitem', { name: texte.mehrVon }))
    expect(aufMenge).toHaveBeenCalledWith(
      commanderdeck.plaetze[0],
      (commanderdeck.plaetze[0]?.menge ?? 0) + 1,
    )
  })

  it('geben genau einer Zeile die Tabstelle', () => {
    const { zeilen } = zeige(commanderdeck)

    expect(zeilen.filter((zeile) => zeile.tabIndex === 0)).toHaveLength(1)
    expect(zeilen[0]).toHaveAttribute('tabindex', '0')
  })

  it('wandern mit den Pfeiltasten, Pos1 und Ende', async () => {
    const { liste, zeilen } = zeige(commanderdeck)

    an(zeilen, 0).focus()
    await userEvent.keyboard('{ArrowDown}')

    expect(zeilen[1]).toHaveFocus()
    expect(
      within(liste).getByRole('button', { name: menueVon(1) }),
    ).toBeInTheDocument()
    expect(within(liste).getAllByRole('button')).toHaveLength(1)

    await userEvent.keyboard('{End}')
    expect(zeilen[75]).toHaveFocus()

    await userEvent.keyboard('{ArrowUp}')
    expect(zeilen[74]).toHaveFocus()

    await userEvent.keyboard('{Home}')
    expect(zeilen[0]).toHaveFocus()
  })

  it('führen die Tastatur aus der Zeile zu ihrem Menü', async () => {
    const { zeilen } = zeige(standarddeck)

    an(zeilen, 0).focus()
    await userEvent.tab()

    expect(
      screen.getByRole('button', { name: menueVon(0, standarddeck) }),
    ).toHaveFocus()
  })

  it('erscheinen unter dem Zeiger, und ein Klick darauf trifft diesen Platz', async () => {
    const { aufMenge, liste, zeilen } = zeige(standarddeck)
    const dritte = an(zeilen, 2)

    await userEvent.hover(dritte)
    expect(within(liste).getAllByRole('button')).toHaveLength(2)

    await userEvent.unhover(dritte)
    expect(within(dritte).queryAllByRole('button')).toHaveLength(0)

    await userEvent.hover(dritte)
    await userEvent.click(
      within(dritte).getByRole('button', {
        name: menueVon(2, standarddeck),
      }),
    )
    await userEvent.click(screen.getByRole('menuitem', { name: texte.mehrVon }))
    expect(aufMenge).toHaveBeenCalledWith(standarddeck.plaetze[2], 5)

    // **Wer hineingeklickt hat, hat die Zeile gewählt** — sie behält ihren
    // Knopf, auch wenn der Zeiger weiterzieht.
    await userEvent.unhover(dritte)
    expect(within(dritte).getAllByRole('button')).toHaveLength(1)
  })

  it('nehmen die Tabstelle mit, wenn jemand in eine Zeile klickt', async () => {
    const { zeilen } = zeige(standarddeck)

    await userEvent.click(
      within(an(zeilen, 3)).getByText(name(standarddeck, 3)),
    )

    expect(zeilen[3]).toHaveFocus()
    expect(zeilen[3]).toHaveAttribute('tabindex', '0')
    expect(zeilen[0]).toHaveAttribute('tabindex', '-1')
  })

  it('sagen an einer ruhenden Zeile, dass sie im Sideboard liegt', () => {
    const { zeilen } = zeige(standarddeck)
    const imSideboard = an(zeilen, standarddeck.plaetze.length - 1)

    expect(within(imSideboard).getByText(texte.sideboard)).toBeInTheDocument()
    expect(within(an(zeilen, 1)).queryByText(texte.hauptdeck)).toBeNull()
  })
})

describe('das Kontextmenü einer ruhenden Zeile', () => {
  it('bietet dieselben vier Handlungen an', async () => {
    const { aufNehmen, zeilen } = zeige(commanderdeck)

    fireEvent.contextMenu(an(zeilen, 40))
    await userEvent.click(
      await screen.findByRole('menuitem', { name: texte.nehmen }),
    )

    expect(aufNehmen).toHaveBeenCalledWith(commanderdeck.plaetze[40]?.id)
  })

  it('lässt die Zeile stehen, während die Pfeile im Menü wandern', async () => {
    const { zeilen } = zeige(commanderdeck)
    const erste = an(zeilen, 0)

    erste.focus()
    await userEvent.keyboard('{Shift>}{F10}{/Shift}')
    expect(await screen.findByRole('menu')).toBeInTheDocument()

    await userEvent.keyboard('{ArrowDown}')

    expect(erste).toHaveAttribute('tabindex', '0')
    expect(zeilen[1]).toHaveAttribute('tabindex', '-1')
  })
})

describe('eine Sondertaste', () => {
  it.each(['shiftKey', 'metaKey', 'altKey', 'ctrlKey'])(
    'lässt die Zeile mit %s stehen — die Taste gehört dann jemand anderem',
    (taste) => {
      const { zeilen } = zeige(commanderdeck)
      const erste = an(zeilen, 0)

      erste.focus()
      fireEvent.keyDown(erste, { key: 'ArrowDown', [taste]: true })

      expect(erste).toHaveFocus()
      expect(zeilen[1]).toHaveAttribute('tabindex', '-1')
    },
  )

  it('und eine fremde Taste ebenso', () => {
    const { zeilen } = zeige(commanderdeck)
    const erste = an(zeilen, 0)

    erste.focus()
    fireEvent.keyDown(erste, { key: 'a' })

    expect(erste).toHaveFocus()
  })
})

describe('die Zeilen als Fokusstellen', () => {
  // **Eine fokussierbare Zeile ohne eigene Rolle** — axe gegen WCAG 2.2 AA,
  // mit einer Zeile unter dem Zeiger, damit zwei Zeilen Knöpfe tragen.
  it('bestehen axe, auch mit einer Zeile unter dem Zeiger', async () => {
    const { liste, zeilen } = zeige(standarddeck)

    await userEvent.hover(an(zeilen, 5))

    expect(await verstoesse(liste)).toEqual([])
  })
})
