// **Zurück, und was dort stand, steht noch da** — UX-03.AK-02
// (#1675):
// *„Zurück/Vorwärts, Direktaufruf und Wiederaufnahme funktionieren ohne
// Verlust ungespeicherter Eingaben."*
//
// # Der Befund
//
// **Eine Seite, die man verließ, war fort.** Wer auf den Decks einen Namen
// für ein neues Deck tippte, kurz in die Sammlung sah und zurückkam, fand ein
// leeres Feld — jede Seite entstand beim Zurückkommen neu, samt ihrem
// Zustand. Und es gab gar kein Zurück: nur denselben Weg noch einmal.
//
// # Warum durch die ganze Anwendung
//
// **Weil die Zusage über drei Stellen läuft**: der Gang in `Anwendung`, die
// Kürzel im `Rahmen` und das Aufheben der Seiten in `Seiteninhalt`. Jede hat
// ihren eigenen Lauf; keiner sagt, dass der Weg trägt.
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { App } from './App'
import { Sprachraum } from './sprache/kontext'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { kataloge } from './sprache/kataloge'
import { leisteneintrag } from './test/leistenweg'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function zeige() {
  belegeApi({
    ...belegeFenster('de'),
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [], fehler: null }),
    Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
    Sammlungsfenster: () =>
      Promise.resolve({
        exemplare: [],
        gesamt: 0,
        insgesamt: 0,
        stueck: 0,
        fehler: null,
      }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <App />
    </Sprachraum>,
  )
}

// taste drückt ⌘ mit einer Klammertaste — gedeutet über die Taste.
function taste(code: 'BracketLeft' | 'BracketRight') {
  fireEvent.keyDown(window, {
    key: code === 'BracketLeft' ? '[' : ']',
    code,
    metaKey: true,
  })
}

// aufschlagen klappt „Deck anlegen" auf — er beginnt zu.
async function aufschlagen() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.seiten.decks.neu }),
  )
}

// deckname findet das Namensfeld im aufgeklappten Kasten.
async function deckname(): Promise<HTMLElement> {
  const anlage = await screen.findByRole('region', {
    name: texte.seiten.decks.neu,
  })

  return within(anlage).getByRole('textbox', { name: texte.seiten.decks.name })
}

describe('der Seitengang durch die Anwendung', () => {
  it('führt mit ⌘[ zurück und mit ⌘] wieder vor', async () => {
    zeige()

    await userEvent.click(await leisteneintrag('decks'))
    await aufschlagen()
    await deckname()
    await userEvent.click(await leisteneintrag('sammlung'))
    expect(
      await screen.findByRole('heading', {
        level: 1,
        name: texte.seiten.sammlung.titel,
      }),
    ).toBeVisible()

    taste('BracketLeft')
    expect(await deckname()).toBeVisible()

    taste('BracketRight')
    expect(
      await screen.findByRole('heading', {
        level: 1,
        name: texte.seiten.sammlung.titel,
      }),
    ).toBeVisible()
  })

  it('hebt auf, was im Formular stand, als man ging', async () => {
    zeige()

    await userEvent.click(await leisteneintrag('decks'))
    await aufschlagen()
    await userEvent.type(await deckname(), 'Mono-Rot')

    await userEvent.click(await leisteneintrag('sammlung'))
    taste('BracketLeft')

    expect(await deckname()).toHaveValue('Mono-Rot')
  })

  // **Auch der Direktaufruf** — über die Leiste statt über ⌘[.
  it('hebt es auch auf, wenn man über die Leiste zurückkommt', async () => {
    zeige()

    await userEvent.click(await leisteneintrag('decks'))
    await aufschlagen()
    await userEvent.type(await deckname(), 'Mono-Rot')

    await userEvent.click(await leisteneintrag('sammlung'))
    await userEvent.click(await leisteneintrag('decks'))

    expect(await deckname()).toHaveValue('Mono-Rot')
  })
})
