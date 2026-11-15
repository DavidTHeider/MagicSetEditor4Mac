// **Ein Ort für „hier raus"** —
// #677.
//
// # Was sich mit UXR-09 geändert hat
//
// Hier standen vier Fälle über einen **Wegweiser**: fünf Flächen mit je
// einem Satz, zwei Knöpfe „Dorthin" auf die Laschen *Bilder* und
// *Druckbogen*, und zwei Ausgänge, die erst mit dem geladenen Set standen.
// Seit UXR-09 (#1917)
// ist die Fläche ein Assistent: die Laschen gibt es nicht mehr, auf die die
// Knöpfe führten, und was wartet, steht in `vorlagenimassistenten.test.tsx`.
// **Die Zusage aus #677 bleibt** — es gibt einen Ort, an dem alle Ausgänge
// stehen, und jeder sagt, wofür er ist —, und sie steht hier.
import { render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { ausgabeziele } from '../ausgabe'
import { ohneAuszeichnung } from '../sprache/kurzform'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Ausgeben } from './Ausgeben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.ausgeben

describe('Die Fläche Ausgeben', () => {
  // **Alle fünf stehen da, jeder mit seinem Satz** — als Beschreibung am
  // Optionsfeld, damit ein Vorleseprogramm ihn mit dem Namen sagt.
  //
  // **Ohne Auszeichnung verglichen** (#1100): die Sätze tragen `` `…` `` und
  // `**…**`, und `bausteine/Satz.tsx` löst beides in eigene Kinder auf.
  it('nennt jeden Ausgang mit seinem Satz', async () => {
    belegeApi({
      // **Der Assistent merkt sich das Ziel** (AP-10, #2367).
      EinstellungSetzen: () => Promise.resolve({}),
      Sets: () => new Promise(() => undefined),
    })

    render(
      <Sprachraum sprache="de">
        <Ausgeben setID="4" seite="ausgeben" wechseln={vi.fn()} />
      </Sprachraum>,
    )

    const knoepfe = within(
      await screen.findByRole('group', { name: texte.wohin }),
    ).getAllByRole('radio')

    expect(knoepfe).toHaveLength(ausgabeziele.length)

    for (const [stelle, ziel] of ausgabeziele.entries()) {
      expect(knoepfe[stelle]).toHaveAccessibleDescription(
        ohneAuszeichnung(texte.ziele[ziel].kurz),
      )
    }
  })
})
