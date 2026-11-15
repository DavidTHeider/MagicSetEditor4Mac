// **Die Hilfe führt zur Frage, statt alles zu zeigen** — UXR-22
// (#1930).
//
// Gemessen am nativen Fenster stand die Hilfe bei einer Rolltiefe von 4,1:
// das erste Gebiet zählte alle fünfundzwanzig Seiten mit je einem Absatz auf.
// Wer die Hilfe aufschlägt, hat eine Frage — sie bekommt ein Suchfeld, die
// Seite, von der man kam, steht obenan, und das kürzeste Gebiet ist offen.
import userEvent from '@testing-library/user-event'
import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import type { Seitenschluessel } from '../navigation'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Hilfe } from './Hilfe'

const { hilfe, seiten } = kataloge.de

function zeige(herkunft?: Seitenschluessel) {
  render(
    <Sprachraum sprache="de">
      <Hilfe
        fassung="0.0.0-test"
        zuUeber={false}
        angekommen={vi.fn()}
        einfuehrungZeigen={vi.fn()}
        herkunft={herkunft}
      />
    </Sprachraum>,
  )
}

function suchfeld() {
  return screen.getByRole('searchbox', { name: hilfe.suche.bezeichnung })
}

describe('Hilfe, Suche', () => {
  it('findet einen Begriff über seinen Namen, ohne Rücksicht auf Groß und Klein', async () => {
    zeige()

    await userEvent.type(suchfeld(), 'DRUCKBOGEN')

    const treffer = screen.getByRole('region', { name: hilfe.suche.gefunden })
    expect(
      within(treffer)
        .getAllByRole('term')
        .map((knoten) => knoten.textContent),
    ).toContain(hilfe.begriffe.druckbogen.name)
    // **Die Gebiete weichen**, solange gesucht wird — sonst stünde der
    // Treffer über einer Leiste, die alles andere weiter anbietet.
    expect(screen.queryByRole('tablist')).toBeNull()
  })

  it('findet eine Seite über ihren Satz', async () => {
    zeige()

    // Ein Wort, das nur im Satz der Statistikseite steht, nicht im Namen.
    const satz = hilfe.seiten.statistik.text
    const wort = satz.split(/\s+/).find((w) => w.length > 9) ?? satz
    await userEvent.type(suchfeld(), wort.replace(/[^\p{L}]/gu, ''))

    const treffer = screen.getByRole('region', { name: hilfe.suche.gefunden })
    expect(
      within(treffer)
        .getAllByRole('term')
        .map((knoten) => knoten.textContent),
    ).toContain(hilfe.seiten.statistik.ueberschrift)
  })

  it('nennt den nächsten Schritt, wenn nichts passt', async () => {
    zeige()

    await userEvent.type(suchfeld(), 'qqqxz')

    expect(screen.queryByRole('term')).toBeNull()
    expect(
      screen.getByText(hilfe.suche.nichts.replace('{suche}', 'qqqxz')),
    ).toBeInTheDocument()

    // **Und tut ihn** (G02, #2249): der Satz riet, das Feld zu leeren — der
    // Knopf tut es.
    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.listenansicht.sucheLeeren,
      }),
    )

    expect(suchfeld()).toHaveValue('')
    expect(screen.getByRole('tablist')).toBeInTheDocument()
  })

  it('bringt die Gebiete zurück, wenn das Feld leer wird', async () => {
    zeige()

    await userEvent.type(suchfeld(), 'set')
    expect(screen.queryByRole('tablist')).toBeNull()

    await userEvent.clear(suchfeld())

    expect(screen.getByRole('tablist')).toBeInTheDocument()
    expect(
      screen.queryByRole('region', { name: hilfe.suche.gefunden }),
    ).toBeNull()
  })
})

describe('Hilfe, was zuerst steht', () => {
  it('schlägt die Begriffe auf, wie das Artboard', () => {
    zeige()

    expect(
      screen.getByRole('tab', { name: seiten.hilfe.glossar }),
    ).toHaveAttribute('aria-selected', 'true')
  })

  it('zeigt die Hilfe der Seite, von der man kam', () => {
    zeige('statistik')

    const bereich = screen.getByRole('region', {
      name: hilfe.suche.herkunftUeberschrift,
    })
    expect(bereich).toHaveTextContent(hilfe.seiten.statistik.ueberschrift)
  })

  // **Ein Satz, der Rest hinter dem Fragezeichen** — der ganze Absatz hob
  // die Rolltiefe im engen Fenster auf 2,8 (R4 erlaubt 2,5).
  it('zeigt von der Herkunft den ersten Satz und den Rest auf Nachfrage', () => {
    zeige('statistik')

    const bereich = screen.getByRole('region', {
      name: hilfe.suche.herkunftUeberschrift,
    })
    // Der sichtbare Satz steht vor dem Fragezeichen; der Rest im Tooltip.
    const sichtbar = bereich.querySelector('dd p > span')?.textContent ?? ''
    expect(sichtbar).toContain('Was im offenen Set steckt')
    expect(sichtbar).not.toContain('Zum Nachsehen')
    expect(within(bereich).getAllByRole('button')).toHaveLength(2)
  })

  it('führt von der Herkunft zur Übersicht aller Seiten', async () => {
    zeige('statistik')

    const bereich = screen.getByRole('region', {
      name: hilfe.suche.herkunftUeberschrift,
    })
    await userEvent.click(
      within(bereich).getByRole('button', {
        name: seiten.hilfe.seitenuebersicht,
      }),
    )

    expect(
      await screen.findByRole('tab', {
        name: seiten.hilfe.seitenuebersicht,
        selected: true,
      }),
    ).toBeInTheDocument()
  })

  it('zeigt keinen Herkunftsbereich ohne Herkunft', () => {
    zeige()

    expect(
      screen.queryByRole('region', { name: hilfe.suche.herkunftUeberschrift }),
    ).toBeNull()
  })
})
