// **Wo Griffe stehen und wo nicht** —
// #890.
//
// **Eine eigene Datei**, weil `symbolgriffe.test.tsx` über die Zeilengrenze
// wuchs (§2). Der Schnitt folgt der Frage: dort **wie ein Griff bewegt
// wird**, hier **wann es überhaupt einen gibt**.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { leererSetkopf } from '../test/setkopfprobe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Symbol } from './Symbol'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

function punkt(x: number, y: number) {
  return {
    x,
    y,
    griffDavorX: 0,
    griffDavorY: 0,
    griffDanachX: 0,
    griffDanachY: 0,
    kurveDanach: false,
    sperre: 'free',
  }
}

function dreieck() {
  return {
    teile: [
      {
        art: 'Form',
        name: 'Dreieck',
        kombination: 'overlap',
        punkte: [punkt(0.5, 0.1), punkt(0.9, 0.8), punkt(0.1, 0.8)],
        spiegelart: '',
        kopien: 0,
        mitteX: 0,
        mitteY: 0,
        griffX: 0,
        griffY: 0,
        teile: [],
      },
    ],
    name: '',
    fehler: null,
  }
}

// **Die Maße werden gestellt.** In jsdom hat kein Element eine Größe;
// `getBoundingClientRect` gibt überall Nullen, und eine Umrechnung durch null
// gäbe `NaN`. Der Lauf stellt deshalb ein 320 × 320 großes SVG an der linken
// oberen Ecke — genau die Maße, die die Zeichenfläche im Browser hat.

describe('Griffe', () => {
  // **Ein Punkt ohne Kurve hat keine Griffe.**
  //
  // Ein Griff an einer Geraden zeigt auf eine Rundung, die es nicht gibt.
  it('fehlen an einer Geraden', async () => {
    belegeApi({
      Grundform: vi.fn().mockResolvedValue(dreieck()),
      Setangaben: vi.fn().mockResolvedValue({
        felder: [],
        werte: [],
        kopf: leererSetkopf,
        fehler: null,
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Symbol setID="7" />
      </Sprachraum>,
    )

    ;(await screen.findByRole('button', { name: texte.dreieck })).click()
    await screen.findByRole('button', { name: 'Punkt 1 von Dreieck' })

    expect(
      screen.queryByRole('button', { name: /^Griff/ }),
    ).not.toBeInTheDocument()
  })
})
