// **Die Leiste über der Partienliste** — PLAY-03.AK-05
// (#1711).
//
// # Wogegen hier geprüft wird
//
// **Dass die Auswahl ankommt**, nicht was sie rechnet — das steht in
// `partienauswahl.test.ts`. Hier geht es um den Weg dazwischen: ein Feld,
// zwei Listen, und was danach dasteht.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it } from 'vitest'

import { Partienliste } from './Partienliste'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Partiezeile } from '../apipartie'
import { stilleVerwaltung } from '../test/verwaltung'

const eigene = kataloge.de.seiten.spielen

function eineZeile(abweichend: Partial<Partiezeile> = {}): Partiezeile {
  return {
    id: 'p1',
    fassung: '3',
    spiel: 'magic',
    stand: 'laeuft',
    sitze: [{ kennung: 'sitz-1', name: 'David' }],
    objekte: 12,
    zug: 0,
    eigenerSitz: '',
    dran: false,
    ...abweichend,
  }
}

// drei Partien, die sich in allem unterscheiden, wonach gesucht wird.
const laeuft = eineZeile({
  id: 'p1',
  sitze: [{ kennung: 's', name: 'Cora' }],
  stand: 'laeuft',
  zug: 4,
})

const pausiert = eineZeile({
  id: 'p2',
  sitze: [{ kennung: 's', name: 'Anna' }],
  stand: 'pausiert',
  zug: 9,
  spiel: 'vs',
})

const drei = [
  laeuft,
  pausiert,
  eineZeile({
    id: 'p3',
    sitze: [{ kennung: 's', name: 'Bert' }],
    stand: 'beendet',
    zug: 1,
  }),
]

function zeige(partien: Partiezeile[]) {
  render(
    <Sprachraum sprache="de">
      <Partienliste
        geladen
        wiederholen={() => undefined}
        partien={partien}
        fehler={undefined}
        einsteigen={() => undefined}
        verwaltung={stilleVerwaltung()}
      />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// namen nennt die Partien, die gerade dastehen — in ihrer Reihenfolge.
//
// **Aus der Überschrift der Zeile** — seit UXR-13 (#1921) steht der Name in
// einer eigenen Zeile über Spiel und Stand, und der Text der ganzen Zeile
// begänne mit „Coramagic".
function namen(): string[] {
  return screen
    .getAllByRole('listitem')
    .map((zeile) => zeile.querySelector('strong')?.textContent ?? '')
}

// **Eine Leiste über einer einzigen Zeile kann nur etwas wegnehmen.**
it('zeigt keine Auswahl, wo es nichts auszuwählen gibt', () => {
  zeige([eineZeile()])

  expect(
    screen.queryByRole('searchbox', { name: eigene.partienSuche }),
  ).not.toBeInTheDocument()
  expect(
    screen.queryByRole('combobox', { name: eigene.partienStand }),
  ).not.toBeInTheDocument()
})

it('sucht in Namen und Spiel', async () => {
  const hand = zeige(drei)

  await hand.type(
    screen.getByRole('searchbox', { name: eigene.partienSuche }),
    'anna',
  )

  expect(namen()).toEqual(['Anna'])

  // **Und das Spiel trifft auch** — es steht in der Zeile, seit man danach
  // suchen kann.
  await hand.clear(screen.getByRole('searchbox', { name: eigene.partienSuche }))
  await hand.type(
    screen.getByRole('searchbox', { name: eigene.partienSuche }),
    'vs',
  )

  expect(namen()).toEqual(['Anna'])
})

// **Der Filter bietet an, was dasteht** — und nichts weiter.
it('filtert nach dem Stand und bietet nur vorkommende an', async () => {
  const hand = zeige([laeuft, pausiert])

  const wahl = screen.getByRole('combobox', { name: eigene.partienStand })

  expect(
    [...wahl.querySelectorAll('option')].map((eines) => eines.textContent),
  ).toEqual([eigene.partienStandAlle, eigene.standLaeuft, eigene.standPausiert])

  await hand.selectOptions(wahl, eigene.standPausiert)

  expect(namen()).toEqual(['Anna'])
})

it('ordnet nach Name und nach Zug', async () => {
  const hand = zeige(drei)

  expect(namen()).toEqual(['Cora', 'Anna', 'Bert'])

  const wahl = screen.getByRole('combobox', { name: eigene.partienOrdnung })

  await hand.selectOptions(wahl, eigene.partienOrdnungName)
  expect(namen()).toEqual(['Anna', 'Bert', 'Cora'])

  await hand.selectOptions(wahl, eigene.partienOrdnungZug)
  expect(namen()).toEqual(['Anna', 'Cora', 'Bert'])
})

// **Die Zahl sagt, wie viel von wie viel dasteht** — sonst sieht eine
// gefilterte Liste aus wie eine kurze.
it('zählt, was dasteht, und nennt das Ganze beim Filtern', async () => {
  const hand = zeige(drei)

  expect(
    screen.getByText(einsetzen(eigene.partienZahl, { zahl: '3' })),
  ).toBeInTheDocument()

  await hand.type(
    screen.getByRole('searchbox', { name: eigene.partienSuche }),
    'anna',
  )

  expect(
    screen.getByText(
      einsetzen(eigene.partienZahlGefiltert, { gezeigt: '1', alle: '3' }),
    ),
  ).toBeInTheDocument()
})

// **Ein Leerzustand erklärt den nächsten Schritt** (§6) — und er ist ein
// anderer als „du hast keine Partie".
it('sagt, wenn nichts passt, und nimmt die Auswahl zurück', async () => {
  const hand = zeige(drei)

  await hand.type(
    screen.getByRole('searchbox', { name: eigene.partienSuche }),
    'zzz',
  )

  expect(screen.getByText(eigene.partienNichtsGefunden)).toBeInTheDocument()
  expect(screen.queryByText(eigene.partienLeer)).not.toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', { name: eigene.partienZuruecksetzen }),
  )

  expect(namen()).toEqual(['Cora', 'Anna', 'Bert'])
})

// **Der Knopf zum Zurücknehmen steht nur da, wo es etwas zurückzunehmen
// gibt.** Ein Knopf ohne Wirkung ist eine Zusage, die niemand einlöst (§10).
it('bietet das Zurücknehmen nur bei einer Auswahl an', () => {
  zeige(drei)

  expect(
    screen.queryByRole('button', { name: eigene.partienZuruecksetzen }),
  ).not.toBeInTheDocument()
})
