// **Die Schlüsselwörter des Spiels als Text** — UXR-07
// (#1915).
//
// # Was vorher war
//
// Aufgeklappt zeichnete die Seite je Schlüsselwort des Spiels eine ganze
// Bearbeitungszeile — vier Textfelder und eine Auswahl, alle abgeschaltet.
// Bei `magic.mse-game` sind es **368** (`keyword:` in `keywords_en`, gezählt
// am 2026-10-02): **1.840 Felder** für etwas, das niemand ändern kann.
// R6 aus `docs/UIREVIEW.md` faltet ab 200, und R7 verlangt das passende
// Bedienelement — für Unveränderliches ist das **Text**.
//
// # Was MSE tut
//
// Es zeigt sie **in der Liste**, grau (`KeywordList::OnGetItemAttr`,
// `src/gui/control/keyword_list.cpp`, Zeilen 225–235), und erst wer **eines**
// wählt, sieht dessen Felder — nicht beschreibbar, mit einem Hinweis
// darüber (`KeywordsPanel::onKeywordSelect`, `src/gui/set/keywords_panel.cpp`,
// Zeilen 355–370, Fassung f9d9356 vom 2021-07-06). **Eines zur Zeit, nicht
// alle**: hier stehen sie als Text, ohne Wahl — **adapted**, weil eine Zeile
// Text dieselbe Auskunft gibt wie ein abgeschaltetes Feld, und ohne die
// Bedienelemente.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SchluesselwortDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { Schluesselwoerter } from './Schluesselwoerter'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.schluesselwoerter

const fliegend: SchluesselwortDTO = {
  id: '',
  wort: 'Fliegend',
  muster: 'Fliegend',
  erinnerung: 'Kann nur von Fliegenden geblockt werden.',
  regeln: '',
  modus: 'core',
  fest: true,
  gueltig: true,
}

function zeige(woerter: SchluesselwortDTO[]) {
  const Parameterhilfen = vi.fn()
  belegeApi({
    Schluesselwoerter: () =>
      Promise.resolve({ woerter, modi: ['core'], fehler: null }),
    SchluesselwortSetzen: vi.fn(),
    Parameterhilfen,
  })

  render(
    <Sprachraum sprache="de">
      <Schluesselwoerter setID="3" />
    </Sprachraum>,
  )

  return Parameterhilfen
}

describe('Die Schlüsselwörter des Spiels', () => {
  // **Der Knopf heißt, wie er aussieht**, und sagt, ob er offen ist — vorher
  // trug er „Die des Spiels zeigen“ als Namen und „1 aus dem Spiel“ als
  // Aufschrift (WCAG 2.5.3, Label in Name).
  it('stehen hinter einem Knopf, der sagt, was er ist', async () => {
    zeige([fliegend])

    const knopf = await screen.findByRole('button', {
      name: '1 aus dem Spiel',
    })
    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByText('Fliegend')).not.toBeInTheDocument()
  })

  // **Und sie fragen nach keinen Parameterhilfen** — ein Schlüsselwort des
  // Spiels gehört dem Spielpaket. Das hielt bis dahin ein Wächter im Haken
  // gegen eine leere Wortkennung; jetzt ruft ihn diese Liste gar nicht.
  it('zeigen sich als Text und nicht als abgeschaltete Felder', async () => {
    const Parameterhilfen = zeige([fliegend])

    await userEvent.click(
      await screen.findByRole('button', { name: '1 aus dem Spiel' }),
    )

    const liste = screen.getByRole('list', { name: texte.ausDemSpiel })
    const eintrag = within(liste).getByRole('listitem')
    expect(eintrag).toHaveTextContent('Fliegend')
    expect(eintrag).toHaveTextContent('core')
    expect(eintrag).toHaveTextContent(
      'Kann nur von Fliegenden geblockt werden.',
    )

    expect(screen.queryAllByRole('textbox')).toHaveLength(0)
    expect(screen.queryAllByRole('combobox')).toHaveLength(0)
    expect(within(liste).queryAllByRole('button')).toHaveLength(0)
    expect(Parameterhilfen).not.toHaveBeenCalled()
  })

  it('hält die Regeln der Barrierefreiheit', async () => {
    zeige([fliegend])

    await userEvent.click(
      await screen.findByRole('button', { name: '1 aus dem Spiel' }),
    )
    await screen.findByRole('list', { name: texte.ausDemSpiel })

    expect(await verstoesse(document.body)).toEqual([])
  })
})
