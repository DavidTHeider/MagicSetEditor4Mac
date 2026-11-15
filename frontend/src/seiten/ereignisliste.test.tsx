// **Was zuletzt geschah, lesbar** —
// #1836.
//
// # Wogegen hier geprüft wird
//
// **Dass „Hineinsehen" die Karten nennt, der Reihe nach, und nicht als rohes
// JSON.** Es ist die einzige Stelle, an der jemand erfährt, was in seiner
// Bibliothek liegt; bis zum 2026-09-24 stand dort
// `Nur für dich: {"karten":[{"karte":"p0-1","name":"",…}]}`.
import { render, screen, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { Ereignisliste } from './Ereignisliste'
import { Sprachraum } from '../sprache/kontext'
import { einLauf, einObjekt, einePartie } from './partieprobe'
import { kataloge } from '../sprache/kataloge'

import type { Spielereignis } from '../apipartie'
import type { Karteneigenschaften } from '../apikarte'

// eigenschaftenMit gibt die Eigenschaften einer Karte mit diesem Namen.
function eigenschaftenMit(name: string): Karteneigenschaften {
  const ohne = { steht: false, stern: false, zahl: 0 }

  return {
    name,
    farben: [],
    manawert: 0,
    supertypen: [],
    typen: [],
    untertypen: [],
    staerke: ohne,
    widerstandskraft: ohne,
    loyalitaet: ohne,
    verteidigung: ohne,
    unbekannteTypen: [],
  }
}

const eigene = kataloge.de.seiten.spielen
const handlungen = kataloge.de.seiten.handlungen

// zeige zeichnet die Liste mit den genannten Ereignissen.
function zeige(ereignisse: Spielereignis[]) {
  render(
    <Sprachraum sprache="de">
      <Ereignisliste lauf={einLauf({ ereignisse })} />
    </Sprachraum>,
  )
}

// einEreignis gibt ein Ereignis mit leerem öffentlichem Teil.
function einEreignis(teile: Partial<Spielereignis>): Spielereignis {
  return {
    folge: '1',
    art: 'zone.blick',
    oeffentlich: {},
    privat: {},
    kette: 'abcd',
    ...teile,
  }
}

it('sagt, wenn noch nichts geschehen ist', () => {
  zeige([])

  expect(screen.getByText(eigene.ereignisseLeer)).toBeInTheDocument()
})

// **Mit dem Wort der Handlung** — „Hineinsehen" und nicht `zone.blick`.
it('nennt die Handlung beim Wort', () => {
  zeige([einEreignis({ art: 'zone.blick' })])

  expect(
    screen.getByText(handlungen.handlungsname['zone.blick']),
  ).toBeInTheDocument()
  expect(screen.queryByText('zone.blick')).not.toBeInTheDocument()
})

// **Eine Art ohne Wort bleibt, wie sie heißt** (§8): der Kern darf eine
// neue einführen, und eine leere Zeile sähe aus wie ein Fehler der Seite.
it('zeigt eine Art ohne Wort unverändert', () => {
  zeige([einEreignis({ art: 'partie.neuartig' })])

  expect(screen.getByText('partie.neuartig')).toBeInTheDocument()
})

it('nennt die Karten eines Blicks der Reihe nach', () => {
  zeige([
    einEreignis({
      privat: {
        karten: [
          { objekt: 'ada/p1/0', karte: 'p1', name: 'Berg', ordnung: 2 },
          {
            objekt: 'ada/p0/0',
            karte: 'p0',
            name: 'A-Canopy Tactician',
            ordnung: 1,
          },
        ],
      },
    }),
  ])

  const liste = screen.getByRole('list', { name: eigene.ereignisPrivat })

  expect(
    within(liste)
      .getAllByRole('listitem')
      .map((eintrag) => eintrag.textContent),
  ).toEqual(['A-Canopy Tactician', 'Berg'])

  // **Kein rohes JSON mehr.**
  expect(screen.queryByText(/"karten"/)).not.toBeInTheDocument()
})

// **Eine Ergänzung ohne Karten steht, wie sie kam** — legt der Kern morgen
// eine zweite Art an, soll sie nicht still verschwinden.
it('zeigt eine Ergänzung ohne Karten, wie sie kam', () => {
  zeige([einEreignis({ art: 'spieler.wuerfeln', privat: { wurf: 3 } })])

  expect(
    screen.getByText(`${eigene.ereignisPrivat}: {"wurf":3}`),
  ).toBeInTheDocument()
})

// **Das Ergebnis eines Blicks kommt in den Blick.** Die Liste steht am Ende
// der Spalte, unter allen Handlungen; wer „Hineinsehen…" stellt, sah bis
// dahin nur, dass eine Mischpflicht dazukam (gesehen in WebKit am
// 2026-09-24). **Nur beim Erscheinen**: wer danach liest und scrollt, wird
// nicht zurückgeholt.
it('holt die Karten eines Blicks in den Blick, wenn sie erscheinen', () => {
  const gerollt = vi.spyOn(Element.prototype, 'scrollIntoView')

  const blick = einEreignis({
    privat: { karten: [{ objekt: 'a/1', name: 'Wald', ordnung: 1 }] },
  })

  const { rerender } = render(
    <Sprachraum sprache="de">
      <Ereignisliste lauf={einLauf({ ereignisse: [blick] })} />
    </Sprachraum>,
  )

  expect(gerollt).toHaveBeenCalledTimes(1)
  expect(gerollt.mock.contexts[0]).toBe(
    screen.getByRole('list', { name: eigene.ereignisPrivat }),
  )

  // **Dasselbe Ergebnis noch einmal gezeichnet rollt nicht noch einmal.**
  rerender(
    <Sprachraum sprache="de">
      <Ereignisliste lauf={einLauf({ ereignisse: [blick] })} />
    </Sprachraum>,
  )

  expect(gerollt).toHaveBeenCalledTimes(1)

  gerollt.mockRestore()
})

// **Wer und womit** —
// #1992.
// Gemessen am 2026-10-01 in der Abnahme-VM mit VoiceOver: der Eintrag hieß
// „Tap" — ohne Karte und ohne Sitz, also ohne Bild nicht zu verstehen.
it('nennt den Sitz und die Karte, soweit sie zu sehen ist', () => {
  const partie = einePartie({
    objekte: [
      einObjekt({ kennung: 'sitz1/k1/0', name: 'Glutwind', geheim: false }),
    ],
  })

  render(
    <Sprachraum sprache="de">
      <Ereignisliste
        lauf={einLauf({
          partie,
          ereignisse: [
            einEreignis({
              art: 'objekt.tappen',
              oeffentlich: { objekt: 'sitz1/k1/0', sitz: 'sitz1' },
            }),
          ],
        })}
      />
    </Sprachraum>,
  )

  const eintrag = screen.getByRole('listitem')

  expect(eintrag).toHaveTextContent('Anna')
  expect(eintrag).toHaveTextContent(handlungen.handlungsname['objekt.tappen'])
  expect(eintrag).toHaveTextContent('Glutwind')
})

// **Was verdeckt ist, bleibt es** — ein Name, den diese Sicht nicht kennt,
// steht auch hier nicht; der Sitz schon.
it('nennt keine Karte, die diese Sicht nicht kennt', () => {
  render(
    <Sprachraum sprache="de">
      <Ereignisliste
        lauf={einLauf({
          ereignisse: [
            einEreignis({
              art: 'objekt.tappen',
              oeffentlich: { objekt: 'sitz2/k9/0', sitz: 'sitz2' },
            }),
          ],
        })}
      />
    </Sprachraum>,
  )

  expect(screen.getByRole('listitem')).toHaveTextContent(
    `Bert: ${handlungen.handlungsname['objekt.tappen']}`,
  )
})

// **Ohne Partie und ohne Namen nur die Handlung** — eine Sicht, die noch
// lädt, und eine Karte, deren Name diese Sicht nicht kennt.
it('nennt nur die Handlung, solange es keine Partie gibt', () => {
  render(
    <Sprachraum sprache="de">
      <Ereignisliste
        lauf={einLauf({
          partie: undefined,
          ereignisse: [
            einEreignis({
              art: 'objekt.tappen',
              oeffentlich: { objekt: 'sitz1/k1/0', sitz: 'sitz1' },
            }),
          ],
        })}
      />
    </Sprachraum>,
  )

  expect(screen.getByRole('listitem')).toHaveTextContent(
    handlungen.handlungsname['objekt.tappen'],
  )
  expect(screen.getByRole('listitem')).not.toHaveTextContent('Anna')
})

it('nennt keinen leeren Namen', () => {
  render(
    <Sprachraum sprache="de">
      <Ereignisliste
        lauf={einLauf({
          ereignisse: [
            einEreignis({
              art: 'objekt.tappen',
              oeffentlich: { objekt: 'sitz1/k1/0', sitz: 'sitz1' },
            }),
          ],
        })}
      />
    </Sprachraum>,
  )

  expect(screen.getByRole('listitem')).toHaveTextContent(
    `Anna: ${handlungen.handlungsname['objekt.tappen']}`,
  )
  expect(screen.getByRole('listitem')).not.toHaveTextContent('—')
})

// **Angesagt wird, was dazukommt** — #1992. Gemessen am 2026-10-01 in der
// Abnahme-VM: eine Liste mit `role="log"` in einem beschrifteten Abschnitt
// sagt VoiceOver an, sobald es nicht gerade spricht. **Sie steht schon da,
// wenn noch nichts geschehen ist** — eine Live-Region, die erst mit ihrem
// ersten Eintrag entsteht, meldet ihn nicht.
it('ist ein Protokoll, das auch leer schon dasteht', () => {
  zeige([])

  const protokoll = screen.getByRole('log', { name: eigene.ereignisse })

  expect(protokoll).toBeEmptyDOMElement()
  expect(screen.getByText(eigene.ereignisseLeer)).toBeInTheDocument()
})

// **Der Name einer Karte steht in ihren Eigenschaften** — gemessen am
// 2026-10-01 in der Abnahme-VM: nach dem ersten Schnitt hieß der Eintrag
// „Ich: Untap", ohne Karte. Der Kern schickt `name` am Objekt nur bei einem
// Spielstein; bei einer Karte steht er in `eigenschaften` (`kartenwort`).
// **Und was geheim ist, bleibt ungenannt.**
it('nennt eine Karte mit dem Namen aus ihren Eigenschaften, eine geheime nicht', () => {
  const partie = einePartie({
    objekte: [
      einObjekt({
        kennung: 'sitz1/k1/0',
        geheim: false,
        eigenschaften: eigenschaftenMit('Glutwind'),
      }),
      einObjekt({
        kennung: 'sitz1/k2/0',
        geheim: true,
        eigenschaften: eigenschaftenMit('Verdeckt'),
      }),
    ],
  })

  render(
    <Sprachraum sprache="de">
      <Ereignisliste
        lauf={einLauf({
          partie,
          ereignisse: [
            einEreignis({
              folge: '1',
              art: 'objekt.tappen',
              oeffentlich: { objekt: 'sitz1/k1/0', sitz: 'sitz1' },
            }),
            einEreignis({
              folge: '2',
              art: 'objekt.tappen',
              oeffentlich: { objekt: 'sitz1/k2/0', sitz: 'sitz1' },
            }),
          ],
        })}
      />
    </Sprachraum>,
  )

  const [erster, zweiter] = screen.getAllByRole('listitem')

  expect(erster).toHaveTextContent('Glutwind')
  expect(zweiter).not.toHaveTextContent('Verdeckt')
})
