// **Der Navigator der Sammlung trägt die Sicht** — UXR-10
// (#1918),
// nach dem Artboard `Sammlung`: Suche, gespeicherte Filter, eigene Marken —
// und was man mit einer Auswahl tut, solange es eine gibt.
//
// # Der Befund
//
// **Mit einem echten Bestand gezählt**, nicht am leeren Probebestand, an dem
// das Fenster gemessen wurde: vierhundert Exemplare, acht Marken, eine
// davon gewählt, drei Zeilen angehakt und eine im Inspector. Dann standen
// **fünfzehn Felder** auf der Fläche (R3 lässt zwölf zu) — die Pflege der
// Marken mit drei Feldern offen, der Bulkbezug mit seiner Auswahl, die
// Ansicht als Auswahlliste im Navigator. Und ohne Auswahl standen
// *Ausgewählte ändern* und *Ausgewählte löschen* da und sagten auf Klick
// nur »Nichts gewählt.«.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  bestand,
  belegeSammlung,
  felder,
  hauptknoepfeJeRegion,
  marken,
  texte,
  zeigeSammlung,
} from '../test/sammlungsseite'

const markentexte = kataloge.de.bausteine.objektmarken

beforeEach(() => {
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get: () => 600,
  })
})

afterEach(() => {
  raeumeApi()
})

const vierhundert = bestand(400)

// belegeMitMarken gibt dem Bestand acht Marken und einen Filter, der die
// ersten zwanzig trifft — **und einen Katalog, der sie führt**: mit einem
// leeren sagte die Pflege nur „keine Marken" und zeichnete keines ihrer drei
// Felder, und der Lauf zählte zwölf statt fünfzehn.
function belegeMitMarken() {
  const katalog = vi.fn(() =>
    Promise.resolve({
      marken: [...new Set(Object.values(marken(vierhundert)).flat())].map(
        (marke) => ({ marke, anzahl: 5 }),
      ),
      fehler: null,
    }),
  )

  belegeSammlung(vierhundert, {
    MarkenJeObjekt: () =>
      Promise.resolve({ marken: marken(vierhundert), fehler: null }),
    MarkenFiltern: () =>
      Promise.resolve({
        kennungen: vierhundert.slice(0, 20).map((eines) => String(eines.id)),
        fehler: null,
      }),
    Markenkatalog: katalog,
  })

  return katalog
}

describe('Der Navigator der Sammlung', () => {
  it('zeichnet mit Marke, Auswahl und gehaltenem Exemplar höchstens zwölf Felder', async () => {
    belegeMitMarken()
    zeigeSammlung()

    await userEvent.click(
      await screen.findByRole('button', {
        name: markentexte.filterWaehlen.replace('{marke}', 'Handelsstapel'),
      }),
    )
    await waitFor(() => {
      expect(screen.queryByText('Karte 0399')).toBeNull()
    })

    for (const name of ['Karte 0001', 'Karte 0002', 'Karte 0003']) {
      await userEvent.click(screen.getByRole('checkbox', { name }))
    }
    await userEvent.click(screen.getByText('Karte 0004'))

    expect(felder()).toBeLessThanOrEqual(12)
    expect(hauptknoepfeJeRegion()).toBeLessThanOrEqual(1)
  })

  // **Selten heißt: eingeklappt** — und erst dann gefragt. Die Pflege
  // fragte den Markenkatalog bei jedem Aufschlagen der Seite.
  it('klappt die Pflege der Marken und den Bulkbezug ein', async () => {
    const katalog = belegeMitMarken()
    zeigeSammlung()

    await screen.findByText('Karte 0001')

    for (const titel of [markentexte.pflegeUeberschrift, texte.bulkTitel]) {
      expect(screen.getByRole('button', { name: titel })).toHaveAttribute(
        'aria-expanded',
        'false',
      )
    }
    expect(screen.queryByLabelText(texte.bulkUmfang)).toBeNull()
    expect(katalog).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', { name: markentexte.pflegeUeberschrift }),
    )
    await waitFor(() => {
      expect(katalog).toHaveBeenCalled()
    })
  })

  it('zeigt die Massenänderung erst mit einer Auswahl', async () => {
    belegeMitMarken()
    zeigeSammlung()

    await screen.findByText('Karte 0001')

    expect(
      screen.queryByRole('button', { name: texte.masseAendern }),
    ).toBeNull()
    expect(screen.queryByLabelText(texte.masseLagerortSetzen)).toBeNull()

    // **„Preise für alle holen" braucht keine Auswahl** — es bleibt.
    expect(
      screen.getByRole('button', { name: texte.massePreise }),
    ).toBeInTheDocument()

    await userEvent.click(screen.getByRole('checkbox', { name: 'Karte 0001' }))

    const masse = within(screen.getByRole('region', { name: texte.masseTitel }))
    expect(
      masse.getByRole('button', { name: texte.masseAendern }),
    ).toBeInTheDocument()
    expect(masse.getByLabelText(texte.masseLagerortSetzen)).toBeInTheDocument()
  })
})
