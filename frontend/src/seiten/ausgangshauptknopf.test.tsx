// **Eine führende Handlung je Ausgang** — UXR-09
// (#1917),
// R1 aus `docs/UIREVIEW.md`.
//
// # Der Befund
//
// Gemessen am 2026-10-02 am nativen Fenster: *Ausgeben*, *Bilder* und
// *Druckbogen* trugen **keinen** Hauptknopf. Der Knopf, der schreibt, sah aus
// wie „Ordner wählen …" daneben — wer die Fläche ansah, fand die Handlung,
// derentwegen er gekommen war, nicht an ihrem Gewicht.
//
// **Geprüft wird die Klasse und nicht ein Bild**: `bg-marke` ist das, woran
// auch das Drehbuch `uirevue` einen Hauptknopf zählt.
import { render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { useVorgang } from '../haken/useVorgang'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Bilder } from './Bilder'
import { Druckbogen } from './Druckbogen'
import { MTGJSONSichern } from './MTGJSONSichern'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function zeige(inhalt: ReactNode) {
  belegeApi({
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
  })

  render(<Sprachraum sprache="de">{inhalt}</Sprachraum>)
}

// MitVorgang gibt dem MTG-JSON-Ausgang den Vorgang, den sonst die Seite hält.
function MitVorgang() {
  const vorgang = useVorgang()

  return (
    <MTGJSONSichern
      satz={{
        id: '4',
        name: 'Mein Probeset',
        spiel: 'magic',
        stil: 'm15',
        herkunft: '/pfad/probeset.mse-set',
        karten: 2,
        herkunftFehlt: false,
      }}
      vorgang={vorgang}
    />
  )
}

describe('Der Knopf, der schreibt, führt', () => {
  it('bei den Bildern', async () => {
    zeige(<Bilder setID="4" />)

    expect(
      await screen.findByRole('button', { name: texte.seiten.bilder.ausgeben }),
    ).toHaveClass('bg-marke')

    // **Und nur er** — „Ordner wählen …" bleibt daneben (R1: einer je
    // Region).
    expect(
      screen.getByRole('button', { name: texte.seiten.bilder.ordnerWaehlen }),
    ).not.toHaveClass('bg-marke')
  })

  it('beim Druckbogen', async () => {
    zeige(<Druckbogen setID="4" />)

    expect(
      await screen.findByRole('button', {
        name: texte.seiten.druckbogen.legen,
      }),
    ).toHaveClass('bg-marke')
  })

  it('bei MTG JSON', () => {
    zeige(<MitVorgang />)

    expect(
      screen.getByRole('button', {
        name: `${texte.seiten.sets.mtgjsonSichern} — Mein Probeset`,
      }),
    ).toHaveClass('bg-marke')
  })
})
