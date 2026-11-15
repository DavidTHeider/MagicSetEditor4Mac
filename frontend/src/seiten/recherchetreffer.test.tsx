// **Ohne Einbettungsmodell führt ein Knopf zur Einrichtung** — AI-01.AK-04
// (#1742):
// das Issue nennt genau diesen Fall, „etwa fehlendes Embeddingmodell". Der
// Satz allein sagte, was fehlt; wo es einzurichten ist, musste man suchen.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Recherchetreffer } from './Recherchetreffer'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { RechercheAntwort } from '../apirecherche'

// leer ist eine Antwort ohne Treffer mit diesem Grund.
function leer(leere: string): RechercheAntwort {
  return {
    treffer: [],
    modus: 'lexikalisch',
    leere,
    grund: 'Nichts gefunden.',
    modell: '',
    modellrevision: '',
    suchkonfiguration: '',
    indexfingerabdruck: '',
    raumabdruck: '',
    ausgangsstand: '',
    korpusstaende: [],
    einschraenkungen: [],
    fortsetzung: '',
    fehler: null,
  }
}

function zeige(ergebnis: RechercheAntwort) {
  render(
    <Sprachraum sprache="de">
      <Recherchetreffer ergebnis={ergebnis} oeffnen={undefined} />
    </Sprachraum>,
  )
}

const knopf = { name: kataloge.de.bausteine.kieinrichtung.oeffnen }

it('bietet ohne Einbettungsmodell »KI einrichten« an', () => {
  zeige(leer('ohne-modell'))

  expect(
    screen.getByText(
      `Nichts gefunden. ${kataloge.de.seiten.recherche.rechercheLeereOhneModell}`,
    ),
  ).toBeInTheDocument()
  expect(screen.getByRole('button', knopf)).toBeInTheDocument()
})

// **Und nur dort**: ein fehlender Index wird unter *Diagnose* gebaut, und ein
// Wort, das nichts findet, braucht keine Einrichtung.
it.each(['ohne-index', 'nichts-gefunden', 'gesperrt'])(
  'bietet bei %s keine Einrichtung an',
  (grund) => {
    zeige(leer(grund))

    expect(screen.queryByRole('button', knopf)).not.toBeInTheDocument()
  },
)
