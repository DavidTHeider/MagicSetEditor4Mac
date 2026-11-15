// **Die Hilfe führt ein Inhaltsverzeichnis statt elf Laschen** — AP-11
// (#2368),
// Befund B-29, Entwurf `Ruhe09`: „Ein Inhaltsverzeichnis links statt der
// Laschenleiste. Die Suche bleibt oben.“ Elf Laschen standen in zwei Reihen.
import userEvent from '@testing-library/user-event'
import { render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Hilfe } from './Hilfe'

const texte = kataloge.de
const hilfe = texte.hilfe

// **Die elf Ziele, die vorher eine Lasche hatten** — keines geht verloren.
const ziele = [
  texte.seiten.hilfe.glossar,
  texte.seiten.hilfe.seitenuebersicht,
  hilfe.fuehrungen.ueberschrift,
  texte.seiten.einstellungen.kuerzelUeberschrift,
  hilfe.sprache.ueberschrift,
  hilfe.schnittstellen.ueberschrift,
  hilfe.datenschutz.ueberschrift,
  hilfe.betrieb.ueberschrift,
  hilfe.erststart.ueberschrift,
  hilfe.weiterlesen.ueberschrift,
  hilfe.ueber.ueberschrift,
]

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Hilfe
        fassung="0.0.0-test"
        zuUeber={false}
        angekommen={vi.fn()}
        einfuehrungZeigen={vi.fn()}
      />
    </Sprachraum>,
  )
}

describe('das Inhaltsverzeichnis der Hilfe', () => {
  it('steht senkrecht und führt alle elf Ziele', () => {
    zeige()

    const verzeichnis = screen.getByRole('tablist', {
      name: texte.seiten.hilfe.aufDieserSeite,
    })
    expect(verzeichnis).toHaveAttribute('aria-orientation', 'vertical')
    expect(
      within(verzeichnis)
        .getAllByRole('tab')
        .map((ziel) => ziel.textContent),
    ).toEqual(ziele)
  })

  it('geht mit ↓ zum nächsten Ziel und schlägt es auf', async () => {
    zeige()

    const [erstesZiel = '', zweitesZiel = ''] = ziele
    screen.getByRole('tab', { name: erstesZiel }).focus()
    await userEvent.keyboard('{ArrowDown}')

    const zweites = screen.getByRole('tab', { name: zweitesZiel })
    expect(zweites).toHaveFocus()
    expect(zweites).toHaveAttribute('aria-selected', 'true')
    expect(
      screen.getByRole('heading', { name: zweitesZiel }),
    ).toBeInTheDocument()
  })

  it('hat die Suche über sich', () => {
    zeige()

    const suche = screen.getByRole('searchbox', {
      name: hilfe.suche.bezeichnung,
    })
    const verzeichnis = screen.getByRole('tablist', {
      name: texte.seiten.hilfe.aufDieserSeite,
    })
    expect(
      suche.compareDocumentPosition(verzeichnis) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  })
})
