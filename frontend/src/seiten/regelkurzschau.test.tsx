// **Was der Tisch tut, auf einen Blick** — das Artboard *Eine Partie
// anfangen*: „setzt durch", „hilft mit", „von Hand", je mit ihren Bereichen
// und der Zahl von allen.
//
// # Wogegen hier geprüft wird
//
// **Dass die Zusammenfassung aus der Matrix kommt und nichts dazuerfindet.**
// Ein Bereich ohne Eintrag ist `unbekannt` — dieselbe Vorgabe wie in der
// Regelschau und im Kern — und steht als „offen" da, nicht als Zusage.
import { render, screen, within } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Regelkurzschau } from './Regelkurzschau'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { regelbereichwort } from './regelwoerter'

import type { RegelfreigabeDTO } from '../apiregeln'

const eigene = kataloge.de.seiten.spielen

// eineFreigabe gibt eine Auskunft über fünf Bereiche in zwei Ästen.
function eineFreigabe(): RegelfreigabeDTO {
  return {
    erlaubt: true,
    regelwerk: 'sandkasten',
    fassung: 1,
    wirksam: '2026-08-19',
    hoechste: 'begleitet',
    matrix: [
      { bereich: 'zonen', grad: 'automatisch', quelle: 'CR 400', grenzen: [] },
      { bereich: 'zug', grad: 'automatisch', quelle: 'CR 500', grenzen: [] },
      { bereich: 'marken', grad: 'begleitet', quelle: '', grenzen: [] },
      { bereich: 'kampf', grad: 'manuell', quelle: '', grenzen: [] },
    ],
    luecken: [],
    hindernisse: [],
    modus: 'anfaenger',
    baum: [
      { kennung: 'ablauf', bereiche: ['zonen', 'zug'] },
      { kennung: 'mechanik', bereiche: ['marken', 'kampf', 'kommander'] },
    ],
    widersprueche: [],
  }
}

// zeile gibt die Zeile eines Grades.
function zeile(wort: string): HTMLElement {
  const zelle = screen.getByText(wort)
  const reihe = zelle.closest('tr')

  if (reihe === null) {
    throw new Error(`keine Zeile zu ${wort}`)
  }

  return reihe
}

it('fasst die Bereiche nach Grad zusammen, mit der Zahl von allen', () => {
  render(
    <Sprachraum sprache="de">
      <Regelkurzschau freigabe={eineFreigabe()} />
    </Sprachraum>,
  )

  const woerter = (bereiche: string[]) =>
    bereiche.map((bereich) => regelbereichwort(eigene, bereich)).join(', ')

  const erwartet = [
    [eigene.gradKurzAutomatisch, woerter(['zonen', 'zug']), '2'],
    [eigene.gradKurzBegleitet, woerter(['marken']), '1'],
    [eigene.gradKurzManuell, woerter(['kampf']), '1'],

    // **Ohne Eintrag heißt: offen** — und nicht verschwiegen.
    [eigene.gradKurzUnbekannt, woerter(['kommander']), '1'],
  ] as const

  for (const [grad, bereiche, zahl] of erwartet) {
    const reihe = zeile(grad)

    expect(within(reihe).getByText(bereiche)).toBeInTheDocument()
    expect(
      within(reihe).getByText(
        einsetzen(eigene.kurzschauAnteil, { zahl, gesamt: '5' }),
      ),
    ).toBeInTheDocument()
  }
})

// **Ein Grad ohne Bereich steht gar nicht da** — eine Zeile „0 von 5"
// sagte nichts, was die anderen nicht schon sagen.
it('lässt einen Grad ohne Bereich weg', () => {
  render(
    <Sprachraum sprache="de">
      <Regelkurzschau
        freigabe={{
          ...eineFreigabe(),
          matrix: [],
          baum: [{ kennung: 'ablauf', bereiche: ['zonen'] }],
        }}
      />
    </Sprachraum>,
  )

  expect(screen.queryByText(eigene.gradKurzAutomatisch)).toBeNull()
  expect(zeile(eigene.gradKurzUnbekannt)).toBeInTheDocument()
})
