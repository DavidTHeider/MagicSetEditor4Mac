// **Der Weg über dem Tisch** — UX-03.AK-02
// (#1675).
//
// # Wogegen hier geprüft wird
//
// **Gegen „Spielen › Spielen".** So stand es über dem Tisch bis zum
// 2026-09-24: der Rahmen kannte nur das offene Set als Bezugsobjekt, und die
// Partie war keines. Das Artboard *Tisch* sagt *Spielen › Partie gegen Bo ›
// Am Tisch*. Geprüft wird durch den Rahmen und die Seite zusammen — die
// Zusage steht an der Naht zwischen beiden, und jeder für sich wäre grün.
import { render, screen, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { Rahmen } from '../rahmen/Rahmen'
import { Spielen } from './Spielen'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, raeumeApi } from '../test/api'
import { element, eineAntwort } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { mitDecks } from './spielenprobe'

const texte = kataloge.de
const eigene = texte.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// weg liest die Stationen der Brotkrume — ohne das Zeichen dazwischen, das
// einem Vorleseprogramm verborgen ist.
function weg(): string[] {
  return within(
    screen.getByRole('navigation', { name: texte.allgemein.navigation.weg }),
  )
    .getAllByRole('listitem')
    .map((station) => station.textContent.replace('›', '').trim())
}

it('nennt am Tisch die Partie und endet bei ihm', async () => {
  mitDecks({ PartieAnlegen: () => Promise.resolve(eineAntwort()) })

  render(
    <Sprachraum sprache="de">
      <Rahmen aktuell="spielen" wechseln={vi.fn()}>
        <Spielen wechseln={() => undefined} />
      </Rahmen>
    </Sprachraum>,
  )

  const hand = nutzerereignis.setup()

  // **Vorher sagt der Weg, was die Seite tut** — *Spielen › Partie
  // anfangen*, wie im Artboard *Eine Partie anfangen*; und er steht da,
  // bevor eine Verneinung etwas sagt (§3).
  expect(weg()).toEqual([texte.allgemein.bereiche.spielen, eigene.anfangen])
  expect(
    screen.getByRole('heading', { level: 1, name: eigene.anfangTitel }),
  ).toBeInTheDocument()

  await hand.type(await screen.findByLabelText(`${eigene.sitzName} 1`), 'Anna')
  await hand.type(screen.getByLabelText(`${eigene.sitzName} 2`), 'Bert')

  const felder = screen.getAllByLabelText(eigene.sitzDeck)

  await hand.selectOptions(element(felder[0]), '1')
  await hand.selectOptions(element(felder[1]), '2')
  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  await screen.findByRole('heading', { level: 1, name: eigene.tischTitel })

  expect(weg()).toEqual([
    texte.allgemein.bereiche.spielen,
    'Partie gegen Bert',
    eigene.tischTitel,
  ])

  // **Und wer die Partie schließt, steht wieder auf der Seite** — der Weg
  // der Partie geht mit ihr.
  await hand.click(screen.getByRole('button', { name: eigene.verlassen }))

  await screen.findByRole('heading', { level: 1, name: eigene.anfangTitel })

  expect(weg()).toEqual([texte.allgemein.bereiche.spielen, eigene.anfangen])

  await antwortAbwarten()
})
