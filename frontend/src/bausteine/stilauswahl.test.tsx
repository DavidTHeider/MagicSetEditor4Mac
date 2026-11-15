// **Ein Stil aus vielen, als Auswahl mit Suche** — UXR-07
// (#1915),
// seit UXR-F1 (#2007)
// auf der `Suchauswahl`.
//
// Die Schublade *Eigenschaften* zeichnete in der Lasche *Stil* je
// installiertem Stil ein Optionsfeld: **382 Felder**, zehn davon im Bild
// (`docs/UIREVIEW.md`, R3). Danach kamen ein Suchfeld und ein natives
// Auswahlfeld (UXR-07); jetzt ist es **ein** Kombinationsfeld, das beides
// kann. Geprüft wird, was die Auswahl an ihre Stelle setzt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { verstoesse } from '../test/axe'
import { Stilauswahl } from './Stilauswahl'
import type { Stileintrag } from './Stilauswahl'

const worte = kataloge.de.bausteine

const stile: Stileintrag[] = [
  { kennung: 'magic-m15', anzeigename: 'M15 Normal' },
  { kennung: 'magic-m15-altered', anzeigename: 'Altered Art' },
  { kennung: 'probe-schlicht', anzeigename: 'Schlicht', zusatz: 'probe' },
]

function Probe({
  waehlen = vi.fn(),
  anfang = 'magic-m15',
}: {
  waehlen?: (kennung: string) => void
  anfang?: string
}) {
  const [gewaehlt, setzeGewaehlt] = useState(anfang)

  return (
    <Sprachraum sprache="de">
      <Stilauswahl
        beschriftung="Stil dieses Sets"
        eintraege={stile}
        gewaehlt={gewaehlt}
        waehlen={(kennung) => {
          waehlen(kennung)
          setzeGewaehlt(kennung)
        }}
      />
    </Sprachraum>
  )
}

const feld = () => screen.getByRole('combobox', { name: 'Stil dieses Sets' })

// angeboten nennt, was die Auswahl gerade anbietet.
const angeboten = () =>
  screen.queryAllByRole('option').map((zeile) => zeile.textContent)

describe('Stilauswahl', () => {
  // **Ein Feld statt eines je Stil** — das ist der Befund aus R3.
  it('zeichnet ein einziges Auswahlfeld und kein Optionsfeld je Stil', () => {
    render(<Probe />)

    expect(screen.queryAllByRole('radio')).toHaveLength(0)
    expect(screen.queryAllByRole('combobox')).toHaveLength(1)
    expect(feld()).toHaveValue('M15 Normal')
  })

  // **Die Beschriftung steht sichtbar da** und gehört zum Feld.
  it('beschriftet das Feld sichtbar', () => {
    render(<Probe />)

    expect(screen.getByText('Stil dieses Sets')).toBeVisible()
    expect(feld()).toHaveAccessibleName('Stil dieses Sets')
  })

  // **Name und Kennung stehen in jedem Eintrag**, ein fremdes Spiel hinter
  // der Kennung — dieselbe Auskunft wie in der Liste des neuen Sets.
  it('nennt je Eintrag Namen und Kennung', async () => {
    render(<Probe />)

    await userEvent.click(feld())

    expect(angeboten()).toEqual([
      'M15 Normal magic-m15',
      'Altered Art magic-m15-altered',
      'Schlicht probe-schlicht · probe',
    ])
  })

  it.each([
    ['über den Namen', 'altered', ['Altered Art magic-m15-altered']],
    ['über die Kennung', 'm15-alt', ['Altered Art magic-m15-altered']],
    [
      'ohne Rücksicht auf Groß und klein',
      'ALTERED ART',
      ['Altered Art magic-m15-altered'],
    ],
    [
      'über das fremde Spiel hinter der Kennung',
      'probe',
      ['Schlicht probe-schlicht · probe'],
    ],
  ])('grenzt die Auswahl %s ein', async (_, suche, soll) => {
    render(<Probe />)

    await userEvent.click(feld())
    await userEvent.keyboard(suche)

    expect(angeboten()).toEqual(soll)
  })

  it('sagt, wenn die Suche nichts findet, und behält die Wahl', async () => {
    render(<Probe />)

    await userEvent.click(feld())
    await userEvent.keyboard('planeswalker')

    expect(angeboten()).toEqual([])
    expect(
      screen.getAllByText(
        worte.suchauswahl.keiner.replace('{suche}', 'planeswalker'),
      ),
    ).not.toHaveLength(0)

    // **Die Wahl gilt weiter**, bis eine Zeile gewählt ist — Escape nimmt die
    // Eingabe zurück.
    await userEvent.keyboard('{Escape}')

    expect(feld()).toHaveValue('M15 Normal')
  })

  it('meldet die Wahl', async () => {
    const waehlen = vi.fn()
    render(<Probe waehlen={waehlen} />)

    await userEvent.click(feld())
    await userEvent.click(screen.getByRole('option', { name: /Altered Art/ }))

    expect(waehlen).toHaveBeenCalledExactlyOnceWith('magic-m15-altered')
    expect(feld()).toHaveValue('Altered Art')
  })

  it('wählt mit der Tastatur', async () => {
    const waehlen = vi.fn()
    render(<Probe waehlen={waehlen} />)
    await userEvent.tab()

    await userEvent.keyboard('{ArrowDown}{ArrowDown}{Enter}')

    expect(waehlen).toHaveBeenCalledExactlyOnceWith('magic-m15-altered')
  })

  // **Führt das Set keinen installierten Stil**, steht das Feld leer da und
  // nennt seinen Platzhalter — nicht den ersten Stil. Eine Auswahl ohne
  // passenden Wert fiele sonst still auf ihn.
  it('steht ohne Wahl leer mit einem Platzhalter', () => {
    render(<Probe anfang="" />)

    expect(feld()).toHaveValue('')
    expect(feld()).toHaveAttribute(
      'placeholder',
      worte.stilliste.keinerGewaehlt,
    )
  })

  // **WCAG 2.2 AA** — zu, offen und ohne Treffer.
  it('hält die Regeln der Barrierefreiheit', async () => {
    const { container } = render(<Probe />)

    expect(await verstoesse(container)).toEqual([])

    await userEvent.click(feld())

    expect(await verstoesse(document.body)).toEqual([])

    await userEvent.keyboard('planeswalker')

    expect(await verstoesse(document.body)).toEqual([])
  })
})
