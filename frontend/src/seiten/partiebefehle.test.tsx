// **Die Handlungen am Tisch** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Wogegen hier geprüft wird
//
// **Dass beim Klick das Richtige hinausgeht** — Art und Nutzlast, nicht nur
// „irgendein Ruf". Ein Knopf, der `objekt.tappen` ohne Objektkennung
// schickte, riefe genauso oft.
//
// **Anhalten und Beenden stehen nicht mehr hier**, sondern unter dem Tisch
// (`partiefuehrung.test.tsx`).
import { render, screen, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { Partiebefehle } from './Partiebefehle'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { oeffne, stelle } from '../test/handlungsprobe'

import {
  element,
  einLauf,
  einObjekt,
  einePartie,
  ruhenderLauf,
  volleAuskunft,
} from './partieprobe'

import type { Partie, Spielobjekt } from '../apipartie'
import type { Partielauf } from '../haken/usePartie'

const eigene = kataloge.de.seiten.spielen
const handlungen = kataloge.de.seiten.handlungen

// zeige zeichnet die Handlungen und gibt die Hand.
// **`objekt` steht in einem Verbund und nicht als Vorgabewert.** Ein
// `undefined` löst einen Vorgabewert aus — der Lauf „ohne gewählte Karte"
// bekäme damit still eine, und seine Aussage wäre das Gegenteil dessen, was
// sein Name sagt.
function zeige(
  lauf: Partielauf,
  was: { objekt?: Spielobjekt | undefined; partie?: Partie } = {},
) {
  const objekt =
    'objekt' in was
      ? was.objekt
      : einObjekt({ zone: 'tisch', name: 'Wald', geheim: false })

  const partie = was.partie ?? einePartie()

  const entwaehlt = vi.fn()

  // **Die volle Auskunft an der gewählten Karte** — hier geht es darum,
  // was ein Knopf stellt; was der Kern anbietet, prüft
  // `partiebefehleauskunft.test.tsx` (P4, #2372).
  const mitAuskunft =
    objekt === undefined || lauf.handlungen.length > 0
      ? lauf
      : { ...lauf, handlungen: [volleAuskunft(objekt.kennung)] }

  render(
    <Sprachraum sprache="de">
      <Partiebefehle
        lauf={mitAuskunft}
        partie={partie}
        objekt={objekt}
        entwaehle={entwaehlt}
        texturen={ruhenderLauf()}
      />
    </Sprachraum>,
  )

  return { hand: nutzerereignis.setup(), entwaehlt }
}

// **Ohne gewählte Karte steht der Satz, der sagt, was zu tun ist.**
it('sagt, dass eine Karte zu wählen ist', () => {
  zeige(einLauf(), { objekt: undefined })

  expect(screen.getByText(handlungen.handlungenKeineKarte)).toBeInTheDocument()
})

// **Tappen schickt die Kennung mit.**
it('tappt die gewählte Karte', async () => {
  const befohlen = vi.fn()

  const { hand } = zeige(einLauf({ befehlen: befohlen }))

  await hand.click(
    screen.getByRole('button', {
      name: handlungen.handlungsname['objekt.tappen'],
    }),
  )

  expect(befohlen).toHaveBeenCalledWith('objekt.tappen', {
    objekt: 'sitz1/k1/0',
  })
})

// **Und eine getappte Karte bekommt den umgekehrten Knopf.**
//
// **Ein Knopf und nicht zwei**: zwei nebeneinander ließen einen davon immer
// ins Leere gehen, und niemand sähe, welcher gilt.
it('bietet einer getappten Karte das Enttappen', async () => {
  const befohlen = vi.fn()

  const { hand } = zeige(einLauf({ befehlen: befohlen }), {
    objekt: einObjekt({
      zone: 'tisch',
      name: 'Wald',
      geheim: false,
      getappt: true,
    }),
  })

  expect(
    screen.queryByRole('button', {
      name: handlungen.handlungsname['objekt.tappen'],
    }),
  ).not.toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', {
      name: handlungen.handlungsname['objekt.enttappen'],
    }),
  )

  expect(befohlen).toHaveBeenCalledWith('objekt.enttappen', {
    objekt: 'sitz1/k1/0',
  })
})

// **Umdrehen ebenso.**
it('dreht die gewählte Karte um', async () => {
  const befohlen = vi.fn()

  const { hand } = zeige(einLauf({ befehlen: befohlen }))

  await hand.click(
    screen.getByRole('button', {
      name: handlungen.handlungsname['objekt.umdrehen'],
    }),
  )

  expect(befohlen).toHaveBeenCalledWith('objekt.umdrehen', {
    objekt: 'sitz1/k1/0',
  })
})

// **In eine ungeordnete Zone geht keine Stelle mit.**
//
// **Und das ist kein Weglassen, sondern eine Aussage**: „oben auf den
// Friedhof" weist der Kern ab, und ein Feld, das es anböte, lüde zu einem
// Fehler ein.
it('schiebt ohne Stelle in eine ungeordnete Zone', async () => {
  const befohlen = vi.fn()
  const { hand, entwaehlt } = zeige(einLauf({ befehlen: befohlen }))

  await hand.selectOptions(await schiebefeld(hand, 'zone'), 'friedhof')
  await hand.selectOptions(await schiebefeld(hand, 'stelle'), 'unten')

  await stelle(hand, 'objekt.schieben')

  // **Die Stelle geht mit, auch in eine ungeordnete Zone** — und der Kern
  // entscheidet, ob sie zählt. Hier stand bis zum 2026-09-17 das Gegenteil:
  // das Fenster ließ das Feld weg, wenn die Zone ungeordnet war, und hatte
  // damit dieselbe Regel ein zweites Mal (§2,
  // #1623).
  expect(befohlen).toHaveBeenCalledWith('objekt.schieben', {
    objekt: 'sitz1/k1/0',
    zone: 'friedhof',
    stelle: 'unten',
  })

  // **Und die Wahl ist aufgehoben** (§7, Regel 1): die Karte liegt danach
  // woanders.
  expect(entwaehlt).toHaveBeenCalled()
})

// **In eine geordnete Zone geht die Stelle mit.**
it('schiebt mit Stelle in die Bibliothek', async () => {
  const befohlen = vi.fn()
  const { hand } = zeige(einLauf({ befehlen: befohlen }))

  await hand.selectOptions(await schiebefeld(hand, 'zone'), 'bibliothek')
  await hand.selectOptions(await schiebefeld(hand, 'stelle'), 'zufaellig')

  await stelle(hand, 'objekt.schieben')

  expect(befohlen).toHaveBeenCalledWith('objekt.schieben', {
    objekt: 'sitz1/k1/0',
    zone: 'bibliothek',
    stelle: 'zufaellig',
  })
})

// **Durchsehen und Mischen gelten der Zone und brauchen keine Karte.**
it('sieht die Bibliothek durch und mischt sie', async () => {
  const befohlen = vi.fn()
  const { hand } = zeige(einLauf({ befehlen: befohlen }), { objekt: undefined })

  await stelle(hand, 'zone.blick')
  await stelle(hand, 'zone.mischen')

  expect(befohlen).toHaveBeenNthCalledWith(1, 'zone.blick', {
    zone: 'bibliothek',
  })
  expect(befohlen).toHaveBeenNthCalledWith(2, 'zone.mischen', {
    zone: 'bibliothek',
  })
})

// schiebefeld gibt eines der zwei Felder des Schiebens.
//
// **Die Felder stehen erst auf Zuruf da** (UX-08,
// #1680);
// aufgeklappt wird einmal, und das zweite Feld steht dann schon.
async function schiebefeld(
  hand: ReturnType<typeof nutzerereignis.setup>,
  name: 'zone' | 'stelle',
): Promise<HTMLElement> {
  const offen =
    screen.queryByRole('group', {
      name: handlungen.handlungsname['objekt.schieben'],
    }) ?? (await oeffne(hand, 'objekt.schieben'))

  return within(offen).getByLabelText(handlungen.handlungsfeld[name])
}

// **An einer pausierten Partie geht kein Befehl hinaus.**
//
// **Aufklappen geht trotzdem** — es schickt nichts; gesperrt ist, was
// schickt.
it('lässt an einer pausierten Partie keine Befehle zu', async () => {
  const { hand } = zeige(einLauf(), {
    partie: einePartie({ stand: 'pausiert' }),
  })

  expect(
    screen.getByRole('button', {
      name: handlungen.handlungsname['objekt.tappen'],
    }),
  ).toBeDisabled()

  const blick = await oeffne(hand, 'zone.blick')

  expect(
    within(blick).getByRole('button', { name: handlungen.handlungStellen }),
  ).toBeDisabled()
})

// **Während ein Befehl läuft, ruht der nächste** (§7, Regel 4).
it('sperrt die Knöpfe, solange ein Befehl läuft', async () => {
  const { hand } = zeige(einLauf({ laeuft: (was) => was === 'partie-p1' }))

  // **Jeder Knopf jeder Gruppe** — einer allein sagte nichts über die
  // anderen (§10). Gesperrt ist, was sofort schickt; wer aufklappt, schickt
  // nichts und bleibt bedienbar.
  for (const titel of [
    handlungen.handlungenKarte,
    handlungen.handlungenZone,
    handlungen.handlungenSitz,
  ]) {
    const gruppe = element(screen.getByText(titel).closest('section'))

    for (const knopf of within(gruppe).getAllByRole('button')) {
      if (knopf.hasAttribute('aria-expanded')) {
        expect(knopf).toBeEnabled()
      } else {
        expect(knopf).toBeDisabled()
      }
    }
  }

  // **Und was aufgeklappt ist, stellt nicht.**
  const blick = await oeffne(hand, 'zone.blick')

  expect(
    within(blick).getByRole('button', { name: handlungen.handlungStellen }),
  ).toBeDisabled()
})

// **Eine unbekannte Zielzone fällt auf ihr Wort zurück.**
//
// **Derselbe Rückfall wie im Tisch** — und er wird hier gemessen, weil die
// beiden Übersetzer getrennt dastehen und getrennt brechen können.
it('zeigt eine unbekannte Zone unverändert', async () => {
  const { hand } = zeige(einLauf())

  // **Die sieben Zonen stehen im Auswahlfeld** — jede mit ihrem Wort aus
  // dem Katalog und keine mit ihrer Kennung.
  expect(await schiebefeld(hand, 'zone')).toHaveTextContent(eigene.zoneKommando)
})

// **Und die Vorrichtung selbst hält, was sie zusagt.**
//
// **Ein Lauf über einen Helfer der Läufe** — er sieht überflüssig aus und ist
// es nicht: `element` ist die Stelle, an der jeder andere Lauf einen Knoten
// entgegennimmt. Wäre sie still, spräche ein Fehler drei Zeilen weiter von
// etwas anderem.
it('scheitert laut, wo kein Element steht', () => {
  expect(() => element(null)).toThrow()
  expect(() => element(undefined)).toThrow()

  const knopf = document.createElement('button')

  expect(element(knopf)).toBe(knopf)
})
