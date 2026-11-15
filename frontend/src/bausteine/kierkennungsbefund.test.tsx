// **Was die Anzeige über ein Foto sagt — und was sie nicht sagt** — Werkstück
// A13.3 von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Die Unsicherheit.** Das Issue verlangt „mögliche Identitäten **mit
// Unsicherheit**"; sie ist hier eine gezählte Zahl und kein Prozentwert, und
// sie muss auch dann dastehen, wenn die Liste kürzer ist als sie.
//
// **Und die Trennung der Evidenz**: „OCR/Layout/Face/Set/Kunstwerk als
// **getrennte Evidenz** behandeln". Eine abgeschnittene Fußzeile und ein
// leeres Feld sind zwei Auskünfte.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { KIErkennungsbefund } from './KIErkennungsbefund'
import { Sprachraum } from '../sprache/kontext'
import {
  einKandidat,
  eineBeobachtung,
  eineErkennung,
  erkennungstexte as eigene,
} from '../test/kierkennung'

import type { KIErkennungAntwort } from '../antwortenerkennung'

// zeige stellt die Anzeige allein hin — ohne den Kasten darum.
function zeige(antwort: KIErkennungAntwort) {
  render(
    <Sprachraum sprache="de">
      <KIErkennungsbefund antwort={antwort} />
    </Sprachraum>,
  )
}

// **Eine unlesbare Stelle steht als unlesbar da — nicht als leeres Feld.**
//
// **Das ist der Satz, für den es die Unterscheidung gibt.** „Die Fußzeile ist
// abgeschnitten" erklärt, warum ein Kandidat offenbleibt; „dort steht nichts"
// erklärt es nicht.
it('unterscheidet unlesbar, leer gelesen und gelesen', () => {
  zeige(
    eineErkennung({
      beobachtungen: [
        eineBeobachtung(),
        eineBeobachtung({ art: 'nummer', wert: '', lesbar: false }),
        eineBeobachtung({ art: 'typzeile', wert: '', lesbar: true }),
      ],
    }),
  )

  expect(screen.getByText(eigene.kiErkennungUnlesbar)).toBeInTheDocument()
  expect(screen.getByText(eigene.kiErkennungOhneWert)).toBeInTheDocument()

  // **Zweimal**: einmal als gelesene Stelle, einmal als Name des Kandidaten,
  // den die Suche dazu gefunden hat. Genau diese Wiederholung ist die
  // Auskunft — sie zeigt, dass der Treffer zum Gelesenen passt.
  expect(screen.getAllByText('Blitzschlag')).toHaveLength(2)
})

// **Dieselbe Art darf zweimal vorkommen.**
//
// **Weil eine geteilte Karte zwei Namen hat** — ein Modell, das beide liest,
// hat richtig gelesen. Der Lauf hält, dass beide dastehen; ohne ihn wäre ein
// Schlüssel aus der Art allein doppelt, und React zeigte einen davon.
it('zeigt zwei Namen auf einer geteilten Karte', () => {
  zeige(
    eineErkennung({
      beobachtungen: [
        eineBeobachtung({ wert: 'Feuer' }),
        eineBeobachtung({ wert: 'Eis' }),
      ],
    }),
  )

  expect(screen.getByText('Feuer')).toBeInTheDocument()
  expect(screen.getByText('Eis')).toBeInTheDocument()
})

// **Eine Art, die der Katalog nicht kennt, steht roh da.**
//
// **Statt zu verschwinden.** Ein Modell, das eine siebte Art erfindet, ist
// ein Befund — und der Befund nützt nichts, wenn die Anzeige die Zeile
// stillschweigend wegließe (§8).
it('zeigt eine unbekannte Evidenzart unter ihrem rohen Namen', () => {
  zeige(
    eineErkennung({
      beobachtungen: [eineBeobachtung({ art: 'wasserzeichen', wert: 'X' })],
      nichtGelesen: ['wasserzeichen'],
      befunde: [
        {
          art: 'erfundeneevidenz',
          stelle: 'wasserzeichen',
          meldung: '"wasserzeichen" ist keine Evidenzart.',
        },
      ],
    }),
  )

  expect(screen.getAllByText(/wasserzeichen/)).not.toHaveLength(0)

  expect(screen.getByRole('alert')).toHaveTextContent(
    '"wasserzeichen" ist keine Evidenzart.',
  )
})

// **Die Zahl der passenden Drucke steht da, auch wenn die Liste kürzer ist.**
//
// **Das ist die Unsicherheit** — gezählt statt geschätzt. Eine Liste von
// zehn Zeilen über 72 Drucken sähe ohne sie nach zehn aus.
it('nennt die Zahl der passenden Drucke und dass es weitergeht', () => {
  zeige(
    eineErkennung({
      kandidaten: [
        einKandidat(),
        einKandidat({ kennung: 'p2', gestuetzt: [] }),
      ],
      drucke: 72,
      weitere: true,
      eindeutig: false,
    }),
  )

  expect(screen.getByText(/72/)).toBeInTheDocument()
  expect(
    screen.getByText(new RegExp(eigene.kiErkennungWeitere)),
  ).toBeInTheDocument()

  // **Und je Kandidat steht, was ihn stützt** — der eine über den Namen, der
  // andere über gar nichts.
  expect(screen.getAllByText(/Kartenname/)).toHaveLength(2)
  expect(screen.getByText(eigene.kiErkennungOhneStuetze)).toBeInTheDocument()
})

// **Ohne Kandidaten steht da, warum.**
//
// **Und nicht »0 Treffer«** (§6): ein leerer Zustand erklärt den nächsten
// Schritt.
it('erklärt, wenn kein Druck passt', () => {
  zeige(
    eineErkennung({
      nichtGelesen: [],
      kandidaten: [],
      drucke: 0,
      eindeutig: false,
      alter: '',
    }),
  )

  expect(screen.getByText(eigene.kiErkennungOhneKandidat)).toBeInTheDocument()

  // **Ohne Stand kein Standsatz** — eine Suche, die nicht lief, hat kein
  // Alter, und „Stand: " wäre eine halbe Angabe.
  const standanfang = eigene.kiErkennungStand.slice(
    0,
    eigene.kiErkennungStand.indexOf('{'),
  )

  expect(screen.queryByText(new RegExp(standanfang))).not.toBeInTheDocument()

  // **Und keine Überschrift über eine leere Liste.**
  expect(
    screen.queryByText(eigene.kiErkennungNichtGelesen),
  ).not.toBeInTheDocument()
})
