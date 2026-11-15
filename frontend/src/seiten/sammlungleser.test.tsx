// **Ein Leser bekommt auch im Leerzustand kein Aufnehmen** — Befund 4 des
// Zweitblicks über Phase G
// (#2276).
//
// G10 sperrte *+ Aufnehmen* im Kopf; der Knopf im Leerzustand öffnete
// dieselbe Schublade ohne Sperre. Das Einlesen bleibt offen — die Schublade
// gibt auch das Ausgeben her, und das darf ein Leser.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Sammlung } from './Sammlung'

const eigene = kataloge.de.seiten.sammlung

afterEach(raeumeApi)

describe('Die leere Sammlung für einen Leser', () => {
  it('sperrt das Aufnehmen und lässt das Einlesen offen', async () => {
    belegeApi({
      Sammlung: () => Promise.resolve({ exemplare: [], fehler: null }),
      Sammlungsfenster: () =>
        Promise.resolve({
          exemplare: [],
          gesamt: 0,
          insgesamt: 0,
          stueck: 0,
          fehler: null,
        }),
      Portfolio: () =>
        Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
      Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
      MarkenJeObjekt: () => Promise.resolve({ marken: {}, fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Schreibrechtraum nurLesend>
          <Sammlung bestandGeaendert={vi.fn()} />
        </Schreibrechtraum>
      </Sprachraum>,
    )

    const leer = (await screen.findByText(eigene.leerTitel)).closest(
      '[data-leerzustand]',
    )
    expect(leer).not.toBeNull()

    const knoepfe = [...(leer?.querySelectorAll('button') ?? [])]
    const aufnehmen = knoepfe.find((k) =>
      k.textContent.includes(eigene.neuesTitel),
    )
    const einlesen = knoepfe.find((k) =>
      k.textContent.includes(eigene.listeTitel),
    )

    expect(aufnehmen).toHaveAttribute('aria-disabled', 'true')
    expect(einlesen).not.toHaveAttribute('aria-disabled')
  })
})
