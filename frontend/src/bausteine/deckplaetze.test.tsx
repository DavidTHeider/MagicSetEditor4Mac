// **Die Plätze mit +/−, Bereichswechsel und Nehmen** — Werkstück U10 von
// Phase 5.1 (#1421).
//
// **Jede Geste hat einen Tastenweg**, und das ist hier zu sehen: alles sind
// Knöpfe, und der Lauf drückt sie über die Tastatur.
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Deckplaetze } from './Deckplaetze'
import { Sprachraum } from '../sprache/kontext'
import { deckprobe, platzprobe } from '../seiten/deckprobe'
import { kataloge } from '../sprache/kataloge'

import type { DeckplatzDTO } from '../apidecks'
import { platzeintrag } from '../test/deckhandlung'

const texte = kataloge.de.seiten.decks

function zeige(plaetze: DeckplatzDTO[], angehalten = false) {
  const aufMenge = vi.fn()
  const aufBereich = vi.fn()
  const aufNehmen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Deckplaetze
        deck={deckprobe({ plaetze })}
        angehalten={angehalten}
        unzulaessig={(platzID) => platzID === 'verboten'}
        aufMenge={aufMenge}
        aufBereich={aufBereich}
        aufNehmen={aufNehmen}
      />
    </Sprachraum>,
  )

  return { aufMenge, aufBereich, aufNehmen }
}

describe('die Plätze eines Decks', () => {
  it('sagt es, wenn keiner da ist', () => {
    zeige([])

    expect(screen.getByText(texte.plaetzeLeer)).toBeInTheDocument()
  })

  it('zählt herauf und herunter', async () => {
    const platz = platzprobe({ bezeichnung: 'Blitz', menge: 4 })
    const { aufMenge } = zeige([platz])

    await userEvent.click(await platzeintrag('Blitz', texte.mehrVon))
    expect(aufMenge).toHaveBeenCalledWith(platz, 5)

    await userEvent.click(await platzeintrag('Blitz', texte.wenigerVon))
    expect(aufMenge).toHaveBeenCalledWith(platz, 3)
  })

  // **Eine Mengenänderung nimmt nichts heraus** — DOM-03.AK-03 (#1698).
  // Bis dahin nahm „eine weniger" bei eins den Platz still aus dem Deck; das
  // war eine Löschung in der Gestalt einer Zählung. Herausnehmen ist der
  // eigene Knopf daneben.
  it('zählt nicht unter eins — herausnehmen ist ein eigener Eintrag', async () => {
    const platz = platzprobe({ bezeichnung: 'Blitz', menge: 1 })
    const { aufMenge, aufNehmen } = zeige([platz])

    expect(await platzeintrag('Blitz', texte.wenigerVon)).toHaveAttribute(
      'aria-disabled',
      'true',
    )
    expect(aufMenge).not.toHaveBeenCalled()
    expect(aufNehmen).not.toHaveBeenCalled()
  })

  it('wechselt den Bereich in beide Richtungen', async () => {
    const haupt = platzprobe({ id: '1', bezeichnung: 'Blitz' })
    const side = platzprobe({
      id: '2',
      bezeichnung: 'Gegenzauber',
      bereich: 'sideboard',
    })
    const { aufBereich } = zeige([haupt, side])

    await userEvent.click(await platzeintrag('Blitz', texte.insSideboard))
    expect(aufBereich).toHaveBeenCalledWith(haupt, 'sideboard')

    // **Die zweite Zeile trägt ihre Knöpfe erst unter dem Zeiger** — seit
    // UXR-11 (#1919) stehen sie an der Zeile mit der Tabstelle und der unter
    // dem Zeiger, nicht an jeder (`deckplaetzeaktiv.test.tsx`).
    await userEvent.hover(screen.getByText('Gegenzauber'))
    await userEvent.click(await platzeintrag('Gegenzauber', texte.insHauptdeck))
    expect(aufBereich).toHaveBeenCalledWith(side, 'hauptdeck')
  })

  it('nimmt einen Platz heraus', async () => {
    const platz = platzprobe({ bezeichnung: 'Blitz' })
    const { aufNehmen } = zeige([platz])

    await userEvent.click(await platzeintrag('Blitz', texte.nehmen))

    expect(aufNehmen).toHaveBeenCalledWith(platz.id)
  })

  it('nennt einen Platz ohne Namen bei seiner Kennung — in dieser Reihenfolge', () => {
    zeige([
      platzprobe({ id: '1', bezeichnung: '', printing: 'abc' }),
      platzprobe({ id: '2', bezeichnung: '', printing: '', druckID: '9' }),
      platzprobe({
        id: '3',
        bezeichnung: '',
        printing: '',
        druckID: '',
        karteID: '7',
      }),
    ])

    expect(screen.getByText('abc')).toBeInTheDocument()
    expect(screen.getByText('9')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
  })

  it('hält jeden Eintrag an, solange ein Vorgang läuft', async () => {
    zeige([platzprobe({ bezeichnung: 'Blitz' })], true)

    expect(await platzeintrag('Blitz', texte.mehrVon)).toHaveAttribute(
      'aria-disabled',
      'true',
    )
  })
})

// **Die Gesten, die das Issue nennt** — Befund 3 des Zweitblicks über Phase
// 5.1 (#1405):
// #1421
// verlangt neben `+/-` ein **Kontextmenü**, und produktiv gab es keines.
describe('das Kontextmenü einer Platzzeile', () => {
  it('bietet dieselben vier Handlungen wie die Knöpfe', async () => {
    zeige([platzprobe({ id: '3', bezeichnung: 'Blitz' })])

    fireEvent.contextMenu(
      within(screen.getByRole('list', { name: texte.plaetze })).getByRole(
        'listitem',
      ),
    )

    const menue = await screen.findByRole('menu')
    const eintraege = within(menue).getAllByRole('menuitem')

    expect(eintraege.map((eintrag) => eintrag.textContent)).toEqual([
      texte.mehrVon,
      texte.wenigerVon,
      texte.insSideboard,
      texte.nehmen,
    ])
  })

  it('führt aus, was darin gewählt wird', async () => {
    const { aufBereich } = zeige([
      platzprobe({ id: '3', bezeichnung: 'Blitz' }),
    ])

    fireEvent.contextMenu(
      within(screen.getByRole('list', { name: texte.plaetze })).getByRole(
        'listitem',
      ),
    )

    await userEvent.click(
      await screen.findByRole('menuitem', { name: texte.insSideboard }),
    )

    expect(aufBereich).toHaveBeenCalledWith(
      expect.objectContaining({ id: '3' }),
      'sideboard',
    )
  })
})

// **Die zwei Zeichenknöpfe trugen einen Tooltip** — Befund 1 des Zweitblicks
// über Phase 5.1 ([Vermerk 0077](../../../docs/entscheidungen/0077-eigene-primitive-statt-radix.md)).
// **Seit AP-6** (#2363) stehen − und + als Einträge im Menü „…“, und ein
// Eintrag sagt in Worten, was er tut; der Knopf „…“ nennt den Platz.
describe('der Knopf »…« einer Zeile', () => {
  it('nennt den Platz und führt zu den Mengeneinträgen', async () => {
    zeige([platzprobe({ id: '3', bezeichnung: 'Blitz' })])

    expect(
      screen.getByRole('button', { name: `${texte.plaetze}: Blitz` }),
    ).toHaveTextContent(texte.handlungenKurz)
    expect(await platzeintrag('Blitz', texte.mehrVon)).toHaveTextContent(
      texte.mehrVon,
    )
  })

  it('führt die drei übrigen Handlungen ebenso aus', async () => {
    const { aufMenge, aufNehmen } = zeige([
      platzprobe({ id: '3', bezeichnung: 'Blitz', menge: 4 }),
    ])

    const menue = async (name: string) => {
      fireEvent.contextMenu(
        within(screen.getByRole('list', { name: texte.plaetze })).getByRole(
          'listitem',
        ),
      )

      await userEvent.click(await screen.findByRole('menuitem', { name }))
    }

    await menue(texte.mehrVon)
    expect(aufMenge).toHaveBeenCalledWith(
      expect.objectContaining({ id: '3' }),
      5,
    )

    await menue(texte.wenigerVon)
    expect(aufMenge).toHaveBeenCalledWith(
      expect.objectContaining({ id: '3' }),
      3,
    )

    await menue(texte.nehmen)
    expect(aufNehmen).toHaveBeenCalledWith('3')
  })

  // **Bei eins ruht „eine weniger"** — im Menü wie am Knopf (DOM-03.AK-03).
  it('bietet bei eins im Menü kein „eine weniger" an', async () => {
    const { aufMenge, aufNehmen } = zeige([
      platzprobe({ id: '3', bezeichnung: 'Blitz', menge: 1 }),
    ])

    fireEvent.contextMenu(
      within(screen.getByRole('list', { name: texte.plaetze })).getByRole(
        'listitem',
      ),
    )

    expect(
      await screen.findByRole('menuitem', { name: texte.wenigerVon }),
    ).toHaveAttribute('aria-disabled', 'true')
    expect(aufMenge).not.toHaveBeenCalled()
    expect(aufNehmen).not.toHaveBeenCalled()
  })

  // **Und aus dem Sideboard zurück** — der Eintrag heißt dann anders.
  it('bietet aus dem Sideboard den Weg ins Hauptdeck', async () => {
    const { aufBereich } = zeige([
      platzprobe({ id: '3', bezeichnung: 'Blitz', bereich: 'sideboard' }),
    ])

    fireEvent.contextMenu(
      within(screen.getByRole('list', { name: texte.plaetze })).getByRole(
        'listitem',
      ),
    )

    await userEvent.click(
      await screen.findByRole('menuitem', { name: texte.insHauptdeck }),
    )

    expect(aufBereich).toHaveBeenCalledWith(
      expect.objectContaining({ id: '3' }),
      'hauptdeck',
    )
  })
})

// **Die Liste bleibt eine Liste** — Befund 3 des Nachlaufs zum Zweitblick über
// Phase 5.1: das Kontextmenü stand als `div` zwischen `ul` und `li`, und damit
// war der Baum für ein Vorleseprogramm keine Liste mehr (axe, Regel `list`).
describe('der Baum der Platzliste', () => {
  it('hat unter dem ul nur li und sonst nichts', () => {
    zeige([
      platzprobe({ id: '3', bezeichnung: 'Blitz' }),
      platzprobe({ id: '4', bezeichnung: 'Wald' }),
    ])

    const liste = screen.getByRole('list', { name: texte.plaetze })

    expect([...liste.children].map((kind) => kind.tagName)).toEqual([
      'LI',
      'LI',
    ])

    // **Und das Gitter greift**: `grid-cols-subgrid` gilt nur für ein
    // direktes Kind des Gitters.
    expect(
      [...liste.children].every((kind) =>
        kind.className.includes('grid-cols-subgrid'),
      ),
    ).toBe(true)
  })
})
