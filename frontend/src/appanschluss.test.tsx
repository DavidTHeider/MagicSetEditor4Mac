// **Von der Sammlung ins Deck — durch die ganze Anwendung** — UX-03.AK-01
// (#1675).
//
// # Warum dieser Lauf durch die ganze Anwendung geht
//
// **Weil die Zusage über fünf Bauteile läuft**: der Sammlungsinspector gibt
// ab, der Anschlussraum hält fest und wechselt die Seite, `Seiteninhalt` und
// `Anwendungsinhalt` reichen durch, und die Kartenwahl des Decks nimmt auf.
// Jedes Stück hat seinen eigenen Lauf; **keiner von ihnen sagt, dass der Weg
// trägt** — und ein Raum, den `Anwendung` nicht aufspannt, schluckt jede
// Übergabe still, weil `useAnschluss` außerhalb taub ist (§10: ein Bauteil
// ohne Aufrufer ist keine Anwendungsfunktion).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { App } from './App'
import { Sprachraum } from './sprache/kontext'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { einsetzen } from './sprache/einsetzen'
import { kataloge } from './sprache/kataloge'
import { keineHerkunft } from './apiherkunft'
import { deckprobe } from './seiten/deckprobe'
import { leisteneintrag } from './test/leistenweg'
import { platzhalter } from './haken/sammlungszeilen'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

const dasExemplar = {
  ...platzhalter(0),
  id: 7,
  fassung: 1,
  printing: 'p-4',
  menge: 2,
  zustand: 'nearMint',
  sprache: 'de',
  name: 'Aschenwächter',
  bezeichnung: 'Aschenwächter',
  herkunft: keineHerkunft,
}

it('trägt ein Exemplar der Sammlung in die Kartenwahl des Decks', async () => {
  belegeApi({
    ...belegeFenster('de'),
    Sammlungsfenster: () =>
      Promise.resolve({
        exemplare: [dasExemplar],
        gesamt: 1,
        insgesamt: 1,
        stueck: 1,
        fehler: null,
      }),
    ExemplarDecks: () => Promise.resolve({ decks: [], fehler: null }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
    Formate: () => Promise.resolve({ formate: [], fehler: null }),
    Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <App />
    </Sprachraum>,
  )

  await userEvent.click(await leisteneintrag('sammlung'))

  // **Erst das Exemplar wählen** — der Inspector zeigt nur, was gewählt ist.
  // Die Liste ist ein Gitter (`Fensterliste`), also wird die **Zeile**
  // angeklickt und kein Knopf.
  await userEvent.click(
    await screen.findByRole('row', { name: /Aschenwächter/ }),
  )

  await userEvent.click(
    await screen.findByRole('button', {
      name: texte.seiten.sammlung.insDeck,
    }),
  )

  // **Die Fläche ist gewechselt, und sie sagt, worauf sie wartet.** Die
  // Kartenwahl steht erst bei offenem Deck da — der Satz führt dorthin.
  expect(
    await screen.findByText(
      einsetzen(texte.seiten.decks.anschlussOhneDeck, {
        name: 'Aschenwächter',
      }),
    ),
  ).toBeInTheDocument()

  await userEvent.click(
    await screen.findByRole('button', {
      name: `${texte.seiten.decks.waehlen}: Aggro`,
    }),
  )

  // **Und dort steht der Druck schon im Feld.**
  expect(await screen.findByLabelText(texte.seiten.decks.druck)).toHaveValue(
    'p-4',
  )
  expect(
    screen.getByText(
      einsetzen(texte.seiten.decks.anschlussWartet, {
        name: 'Aschenwächter',
      }),
    ),
  ).toBeInTheDocument()
})
