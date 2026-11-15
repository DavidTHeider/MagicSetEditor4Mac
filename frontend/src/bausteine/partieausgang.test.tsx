// **Wie die Partie ausgegangen ist** — Werkstück M17 der Regelmaschine
// (#1618).
//
// # Wogegen hier geprüft wird
//
//   - **Solange sie läuft, steht nichts da** — ein Ergebnis, das vor dem Ende
//     erschiene, wäre eine Behauptung.
//   - **Der Sieger steht mit Namen da**, und wer gegangen ist, mit **Grund**:
//     `beendet` sagt nicht, warum (CR 104.3).
//   - **Unentschieden ist ein eigener Ausgang** (CR 104.4a) und kein
//     „niemand hat gewonnen“.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Partieausgang } from './Partieausgang'
import { Sprachraum } from '../sprache/kontext'
import { einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'

import type { Partie } from '../apipartie'

const eigene = kataloge.de.seiten.spielen

function zeige(partie: Partie) {
  render(
    <Sprachraum sprache="de">
      <Partieausgang partie={partie} />
    </Sprachraum>,
  )
}

// **Eine laufende Partie hat kein Ergebnis.**
it('zeichnet nichts, solange die Partie läuft', () => {
  zeige(einePartie())

  expect(screen.queryByLabelText(eigene.ausgangTitel)).not.toBeInTheDocument()
})

// **Der Sieger steht mit Namen da, der Verlierer mit Grund.**
it('nennt den Sieger und den Grund des Ausscheidens', () => {
  zeige(
    einePartie({
      stand: 'beendet',
      ausgang: {
        sieger: 'sitz1',
        unentschieden: false,
        ausgeschieden: [{ sitz: 'sitz2', grund: 'aufgabe' }],
      },
    }),
  )

  expect(screen.getByText(`${eigene.ausgangSieger}: Anna`)).toBeInTheDocument()

  expect(screen.getByText(`Bert — ${eigene.grundAufgabe}`)).toBeInTheDocument()
})

// **Unentschieden ist ein eigener Ausgang** — CR 104.4a.
it('nennt ein Unentschieden als solches', () => {
  zeige(
    einePartie({
      stand: 'beendet',
      ausgang: {
        sieger: '',
        unentschieden: true,
        ausgeschieden: [
          { sitz: 'sitz1', grund: 'leben' },
          { sitz: 'sitz2', grund: 'leben' },
        ],
      },
    }),
  )

  expect(screen.getByText(eigene.ausgangUnentschieden)).toBeInTheDocument()
  expect(
    screen.queryByText(new RegExp(eigene.ausgangSieger)),
  ).not.toBeInTheDocument()

  // **Und beide Gründe stehen da** — zweimal derselbe ist zweimal wahr.
  expect(screen.getAllByText(new RegExp(eigene.grundLeben))).toHaveLength(2)
})

// **Ein Grund, den der Katalog nicht kennt, steht bei seinem Namen.**
//
// **Und nicht als leere Stelle** (§8): ein Zustand kann aus einem Bestand
// kommen, den jemand mit einem anderen Werkzeug geöffnet hat.
it('nennt einen unbekannten Grund bei seinem Namen', () => {
  zeige(
    einePartie({
      stand: 'beendet',
      ausgang: {
        sieger: 'sitz1',
        unentschieden: false,
        ausgeschieden: [{ sitz: 'sitz9', grund: 'erfunden' }],
      },
    }),
  )

  expect(screen.getByText('sitz9 — erfunden')).toBeInTheDocument()
})
