// **Wie `Wahl` ihr Bedienelement zeichnet und was daran hängt** — UXR-F2
// (#2008).
//
// `wahlkatalog.test.tsx` hält, **welches** Element jede Einstellung des
// Katalogs bekommt; hier steht, was um das Element herum gilt: der Name, die
// Marke für einen Treffer der Suche, der Knopf „Zurücksetzen", der Erbwert und
// das Sperren während des Schreibens.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Wahl, type Wahlbarer } from './Wahl'

const eigene = kataloge.de.seiten.einstellungen

function eintrag(
  schluessel: string,
  werte: string[],
  wert: string,
  vorgabe: string,
): Wahlbarer {
  return {
    schluessel,
    wert,
    wirksam: wert,
    werte,
    ebene: 'profil',
    herkunft: 'profil',
    wirkung: 'sofort',
    vorgabe,
  }
}

function zeige(
  einer: Wahlbarer,
  teile: Partial<Parameters<typeof Wahl>[0]> = {},
) {
  const aendern = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Wahl eintrag={einer} angehalten={false} aendern={aendern} {...teile} />
    </Sprachraum>,
  )

  return aendern
}

const ersteSchritte = eintrag('ersteschritte', ['an', 'aus'], 'aus', 'an')
const aktualisierung = eintrag(
  'aktualisierung',
  ['aus', 'melden', 'holen'],
  'holen',
  'aus',
)
const thema = eintrag(
  'erscheinungsbild',
  ['erben', 'system', 'hell', 'dunkel'],
  'erben',
  'erben',
)

describe('ein Schalter', () => {
  it('trägt den Namen der Einstellung und den Stand', () => {
    zeige(ersteSchritte)

    expect(
      screen.getByRole('switch', { name: eigene.namen.ersteschritte }),
    ).not.toBeChecked()
  })

  it('legt von aus auf an um', async () => {
    const aendern = zeige(ersteSchritte)

    await userEvent.click(screen.getByRole('switch'))
    expect(aendern).toHaveBeenLastCalledWith('an')
  })

  it('legt von an auf aus um', async () => {
    const aendern = zeige({ ...ersteSchritte, wert: 'an', wirksam: 'an' })

    await userEvent.click(screen.getByRole('switch'))
    expect(aendern).toHaveBeenLastCalledWith('aus')
  })

  // **Sichtbar ist an** — die Seitenleiste hat kein „an", und dasselbe Wort
  // der Seite („Verborgen") steht unter dem Schalter.
  it('versteht sichtbar und verborgen als Paar', async () => {
    const aendern = zeige(
      eintrag(
        'seitenleiste',
        ['sichtbar', 'verborgen'],
        'sichtbar',
        'sichtbar',
      ),
    )

    expect(screen.getByRole('switch')).toBeChecked()

    await userEvent.click(screen.getByRole('switch'))
    expect(aendern).toHaveBeenLastCalledWith('verborgen')
  })

  // **Ein Wert, den das Paar nicht kennt, behauptet nichts**: der Schalter
  // steht aus, statt an zu sagen.
  it('steht aus, wenn der Wert nicht im Paar ist', () => {
    zeige(eintrag('ersteschritte', ['an', 'aus'], 'unbekannt', 'an'))

    expect(screen.getByRole('switch')).not.toBeChecked()
  })

  // **Umlegen ist zurücksetzen** — wie bei den Bereichsschaltern (UXR-19).
  it('hat keinen Knopf „Zurücksetzen", auch nicht abseits der Vorgabe', () => {
    zeige(ersteSchritte)

    expect(screen.queryByRole('button', { name: /zurücksetzen/i })).toBeNull()
  })

  it('nimmt nichts an, solange geschrieben wird', () => {
    zeige(ersteSchritte, { angehalten: true })

    expect(screen.getByRole('switch')).toBeDisabled()
  })

  it('trägt die Marke eines Treffers der Suche', () => {
    zeige(ersteSchritte, { gesucht: true })

    expect(screen.getByText(eigene.gesuchtMarke)).toBeVisible()
    expect(
      screen.getByRole('switch', {
        name: new RegExp(eigene.namen.ersteschritte),
      }),
    ).toBeVisible()
  })

  // **Die Zeile trägt ihren Schlüssel** (AP-11, #2368) — Name, Feld, Hinweis
  // und Herkunft zusammen; `data-einstellung` steht nur am Feld.
  it('trägt ihren Schlüssel an der ganzen Zeile', () => {
    zeige(ersteSchritte)

    const zeile = document.querySelector(
      '[data-einstellungszeile="ersteschritte"]',
    )
    expect(zeile).toContainElement(
      document.querySelector('[data-herkunftsmarke]'),
    )
    expect(zeile).toContainElement(
      screen.getByRole('switch', {
        name: new RegExp(eigene.namen.ersteschritte),
      }),
    )
  })

  it('zeigt die Herkunft unter dem Namen', () => {
    zeige(ersteSchritte)

    expect(
      screen.getByText(
        einsetzen(eigene.herkunftMarke, {
          herkunft: eigene.herkunftNamen.profil,
        }),
      ),
    ).toBeVisible()
  })

  it('nimmt den Namen, den der Aufrufer gibt', () => {
    zeige(eintrag('dienst.scryfall', ['aus', 'an'], 'an', 'an'), {
      beschriftung: 'Scryfall',
    })

    expect(screen.getByRole('switch', { name: 'Scryfall' })).toBeChecked()
  })
})

describe('Segmente', () => {
  it('sind eine Gruppe, die den Namen der Einstellung trägt', () => {
    zeige(aktualisierung)

    expect(
      screen.getByRole('radiogroup', { name: eigene.namen.aktualisierung }),
    ).toBeVisible()
    expect(screen.getAllByRole('radio')).toHaveLength(3)
  })

  it('nennen jeden Wert beim Namen aus dem Katalog', () => {
    zeige(aktualisierung)

    expect(
      screen.getByRole('radio', {
        name: eigene.werte['aktualisierung.melden'],
      }),
    ).not.toBeChecked()
    expect(
      screen.getByRole('radio', { name: eigene.werte['aktualisierung.holen'] }),
    ).toBeChecked()
  })

  it('melden den gewählten Wert', async () => {
    const aendern = zeige(aktualisierung)

    await userEvent.click(
      screen.getByRole('radio', {
        name: eigene.werte['aktualisierung.melden'],
      }),
    )

    expect(aendern).toHaveBeenCalledExactlyOnceWith('melden')
  })

  // **Der Erbwert ist ein Segment unter anderen** — „Wie beim Nutzer" steht
  // vorn, ist gewählt und lässt sich von jedem anderen aus wieder wählen.
  it('führen den Erbwert als eigenes Segment', async () => {
    const aendern = zeige({
      ...thema,
      wert: 'hell',
      wirksam: 'hell',
    })

    await userEvent.click(
      screen.getByRole('radio', {
        name: eigene.werte['erscheinungsbild.erben'],
      }),
    )

    expect(aendern).toHaveBeenCalledExactlyOnceWith('erben')
  })

  it('zeigen den Erbwert gewählt und sagen, was dann gilt', () => {
    zeige({ ...thema, wirksam: 'dunkel', herkunft: 'nutzer' })

    expect(
      screen.getByRole('radio', {
        name: eigene.werte['erscheinungsbild.erben'],
      }),
    ).toBeChecked()
    expect(screen.getByText('Es gilt „Dunkel“.')).toBeVisible()
    expect(screen.getByText('von dir')).toBeVisible()
  })

  it('behalten den Knopf „Zurücksetzen" abseits der Vorgabe', async () => {
    const aendern = zeige(aktualisierung)

    await userEvent.click(
      screen.getByRole('button', {
        name: `${eigene.namen.aktualisierung} auf die Vorgabe zurücksetzen`,
      }),
    )

    expect(aendern).toHaveBeenCalledExactlyOnceWith('aus')
  })

  it('haben den Knopf nicht, solange die Vorgabe gilt', () => {
    zeige({ ...aktualisierung, wert: 'aus', wirksam: 'aus' })

    expect(screen.queryByRole('button', { name: /zurücksetzen/i })).toBeNull()
  })

  it('nehmen nichts an, solange geschrieben wird', () => {
    zeige(aktualisierung, { angehalten: true })

    for (const feld of screen.getAllByRole('radio')) {
      // **Gesperrt mit `aria-disabled`**, damit der Fokus bleibt (`uxrf2`).
      expect(feld).toHaveAttribute('aria-disabled', 'true')
    }
  })

  it('tragen die Marke eines Treffers der Suche neben dem Namen', () => {
    zeige(aktualisierung, { gesucht: true })

    expect(screen.getByText(eigene.gesuchtMarke)).toBeVisible()
    expect(
      screen.getByRole('radiogroup', {
        name: new RegExp(eigene.namen.aktualisierung),
      }),
    ).toBeVisible()
  })

  // **Ein Klick auf den Namen wählt nichts** — ein Beschriftungsfeld um
  // mehrere Optionsfelder gehörte dem ersten.
  it('wählen nichts, wenn man auf den Namen klickt', async () => {
    const aendern = zeige(aktualisierung)

    await userEvent.click(screen.getByText(eigene.namen.aktualisierung))

    expect(aendern).not.toHaveBeenCalled()
  })

  it('nehmen die Namen der Werte vom Aufrufer', () => {
    zeige(
      eintrag('frage.suche', ['scryfall', 'mtgapi'], 'scryfall', 'scryfall'),
      {
        beschriftung: 'Kartensuche',
        werteNamen: { scryfall: 'Scryfall', mtgapi: 'magicthegathering.io' },
      },
    )

    expect(
      screen.getByRole('radiogroup', { name: 'Kartensuche' }),
    ).toBeVisible()
    expect(screen.getByRole('radio', { name: 'Scryfall' })).toBeChecked()
    expect(
      screen.getByRole('radio', { name: 'magicthegathering.io' }),
    ).not.toBeChecked()
  })
})

describe('eine Auswahlliste', () => {
  // **Ab fünf Werten, und mit einem Wert auch** — R7: Segmente bis vier.
  it('bleibt es ab fünf Werten', () => {
    zeige(
      eintrag(
        'schriftgroesse',
        ['erben', 'klein', 'normal', 'gross', 'sehrgross'],
        'normal',
        'erben',
      ),
    )

    expect(
      screen.getByRole('combobox', { name: eigene.namen.schriftgroesse }),
    ).toBeVisible()
    expect(screen.queryByRole('radiogroup')).toBeNull()
  })

  it('bleibt es bei einem einzigen Wert', () => {
    zeige(eintrag('frage.suche', ['scryfall'], 'scryfall', 'scryfall'), {
      beschriftung: 'Kartensuche',
    })

    expect(screen.getByRole('combobox', { name: 'Kartensuche' })).toBeVisible()
  })
})
