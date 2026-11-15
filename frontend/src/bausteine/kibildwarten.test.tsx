// **Was geschieht, während ein Bild entsteht** — Werkstück A17 von Phase 5.3
// (#1532).
//
// # Die Frage, die kein Lauf gestellt hat
//
// **Was, wenn die Antwort nach einem Kartenwechsel zurückkommt?** Ein
// Bildauftrag dauert Sekunden bis Minuten; in dieser Zeit blättert jemand
// weiter. Kommt der Vorschlag dann an, gehört er zu einer Karte, die niemand
// mehr ansieht — §7, Regel 2.
//
// **Gelöst ist das über den `key` am [KIBereich]** (`Karteneditor.tsx`): ein
// Kartenwechsel baut das Bauteil ab, und mit ihm den offenen Auftrag. Hier
// steht der Lauf, der es festhält — **am langsamen Fall und nicht am
// schnellen**, denn am schnellen gibt es kein Dazwischen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { KIBild } from './KIBild'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import {
  bildtexte as eigene,
  einKatalog,
  einVorschlag,
  gibFrei,
  mitFreigabe,
  ohneNachweis,
} from '../test/kibild'

import type { KIBildvorschlagAntwort } from '../antworten'

afterEach(() => {
  raeumeApi()
})

// zeigeFuer rendert den Kasten für eine Karte — **mit `key` wie im Editor**.
//
// **Der `key` ist der Punkt dieses Laufs.** Ohne ihn bliebe das Bauteil beim
// Kartenwechsel stehen und trüge die Antwort der alten Karte in die neue.
//
// # Die Gegenprobe, und wie sie beim ersten Mal nichts maß
//
// **`key` an **beiden** Stellen entfernen, dann wird der zweite Lauf rot** —
// „expected document not to contain element, found <figure".
//
// **Beim ersten Anlauf stand er nur hier unten weg**, und der Lauf blieb
// grün: React sieht ein Element, das erst einen `key` trägt und dann keinen,
// als ein **anderes** — es baut ab und neu auf, genau wie mit `key`. Die
// Entschärfung war wirkungslos, und das Grün sagte nichts über den Test
// (§10: eine Gegenprobe, die nichts ändert, ist keine).
function zeigeFuer(karteID: string) {
  const gezeichnet = render(
    <Sprachraum sprache="de">
      <KIBild
        key={karteID}
        karteID={karteID}
        bildfelder={['image']}
        zuschneiden={vi.fn(() => Promise.resolve(true))}
        stand={0}
      />
    </Sprachraum>,
  )

  return (neue: string) => {
    gezeichnet.rerender(
      <Sprachraum sprache="de">
        <KIBild
          key={neue}
          karteID={neue}
          bildfelder={['image']}
          zuschneiden={vi.fn(() => Promise.resolve(true))}
          stand={0}
        />
      </Sprachraum>,
    )
  }
}

// starteAuftrag klappt auf, füllt aus, gibt frei — und hält die Antwort an.
//
// **Sie kommt erst, wenn der Lauf sie ruft.** Genau das fehlte bisher: jedes
// Doppel antwortete sofort, und zwischen Absenden und Antwort lag nichts, in
// dem etwas geschehen könnte.
async function starteAuftrag() {
  let antworten: ((antwort: KIBildvorschlagAntwort) => void) | undefined

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildVorschlagen: () =>
        new Promise<KIBildvorschlagAntwort>((fertig) => {
          antworten = fertig
        }),
    }),
  )

  const wechsle = zeigeFuer('4')
  const hand = userEvent.setup()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  await hand.selectOptions(
    await screen.findByLabelText(eigene.kiVorschlagAnbieter),
    'openai/gpt-image-2.5-flare',
  )

  await hand.type(screen.getByLabelText(eigene.kiBildAufgabe), 'ein Blitz')
  await hand.click(screen.getByRole('button', { name: eigene.kiBildErzeugen }))

  await gibFrei(hand)

  // **Der Beleg, dass der Auftrag wirklich läuft** — ohne ihn sagte alles
  // Weitere nichts (§3): eine Verneinung nach einem Ereignis, das niemand
  // gehört hat, trifft immer zu.
  expect(await screen.findByText(eigene.kiBildAbbrechen)).toBeInTheDocument()

  return {
    wechsle,
    antworte: () => {
      antworten?.(einVorschlag())
    },
  }
}

// **Der Steuerlauf: ohne Kartenwechsel steht der Vorschlag da.**
//
// # Warum er dasteht
//
// **Weil die Verneinung darunter sonst nichts misst** (§10). Ein Lauf, in dem
// der Vorschlag **nie** ankommt — weil die Vorrichtung zu kurz wartet oder
// die Antwort gar nicht durchgereicht wird —, wäre grün und sagte nichts über
// den Kartenwechsel. Dieser hier zeigt, dass derselbe Weg ohne den Wechsel
// wirklich etwas hinstellt.
it('stellt einen Vorschlag hin, solange die Karte dieselbe bleibt', async () => {
  const auftrag = await starteAuftrag()

  auftrag.antworte()

  expect(await screen.findByRole('figure')).toBeInTheDocument()
})

// **Und mit Kartenwechsel landet er nirgends.**
it('trägt einen Vorschlag nach einem Kartenwechsel nirgends ein', async () => {
  const auftrag = await starteAuftrag()

  // **Jetzt blättert jemand weiter** — mitten im laufenden Auftrag.
  auftrag.wechsle('5')
  auftrag.antworte()

  // **Und die Antwort wird abgewartet, bevor verneint wird** (§3): „kein
  // Vorschlag da" ist erst dann eine Aussage, wenn die Antwort durch war.
  await antwortAbwarten()

  expect(screen.queryByRole('figure')).not.toBeInTheDocument()
  expect(screen.queryByText(eigene.kiBildAbbrechen)).not.toBeInTheDocument()
})
