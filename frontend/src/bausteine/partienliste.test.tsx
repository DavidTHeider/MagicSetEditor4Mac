// **Der Weg zurück in die eigene Partie** —
// #1600.
//
// # Wogegen hier geprüft wird
//
//   - **Ohne Partie sagt der Leerzustand, was zu tun ist** (§6) — und nicht
//     „keine Daten“.
//   - **Je Sitz ein Eintrag**, und er nennt Partie **und** Sitz — ein Aufruf
//     allein wäre auch grün, wenn jeder Eintrag denselben Platz meinte (§10).
//     Seit UXR-13 (#1921) stehen die Sitze im Menü „Weiterspielen" der
//     Zeile und nicht mehr als Knopf je Sitz: bei neunzehn Partien waren das
//     über hundert Bedienelemente neben dem Anfang einer neuen.
//   - **Die Liste zeigt keine Karte** — sie ist eine Liste und keine Sicht
//     (G1).
import { render, screen, within } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Partienliste } from './Partienliste'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Partiezeile } from '../apipartie'
import { stilleVerwaltung } from '../test/verwaltung'

const eigene = kataloge.de.seiten.spielen

// eineZeile gibt eine Partie, wie die Liste sie bekommt.
function eineZeile(abweichend: Partial<Partiezeile> = {}): Partiezeile {
  return {
    id: 'p1',
    fassung: '3',
    spiel: 'magic',
    stand: 'laeuft',
    sitze: [
      { kennung: 'sitz-1', name: 'David' },
      { kennung: 'sitz-2', name: 'Gegenüber' },
    ],
    objekte: 12,
    zug: 0,
    eigenerSitz: '',
    dran: false,
    ...abweichend,
  }
}

// zeige zeichnet die Liste und gibt die Hand.
function zeige(
  partien: Partiezeile[],
  einsteigen: (partieID: string, sitz: string) => void = () => undefined,
  fehler?: string,
) {
  render(
    <Sprachraum sprache="de">
      <Partienliste
        geladen
        wiederholen={() => undefined}
        partien={partien}
        fehler={fehler}
        einsteigen={einsteigen}
        verwaltung={stilleVerwaltung()}
      />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// weiterspielen wählt im Menü einer Zeile den Sitz — der Weg, den ein Mensch
// seit UXR-13 nimmt.
async function weiterspielen(
  hand: ReturnType<typeof nutzerereignis.setup>,
  partie: string,
  eintrag: string,
) {
  await hand.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.partienWeiterspielenIn, { name: partie }),
    }),
  )
  await hand.click(screen.getByRole('menuitem', { name: eintrag }))
}

// **Ohne Partie steht da, was zu tun ist.**
it('erklärt den nächsten Schritt, wenn es keine Partie gibt', () => {
  zeige([])

  expect(screen.getByText(eigene.partienLeer)).toBeInTheDocument()
})

// **Je Sitz ein Eintrag, und er meint seinen Sitz.**
it('steigt in den gewählten Sitz ein', async () => {
  const eingestiegen = vi.fn()

  const hand = zeige([eineZeile()], eingestiegen)

  await weiterspielen(
    hand,
    'David · Gegenüber',
    einsetzen(eigene.partienEinsteigenAls, { name: 'Gegenüber' }),
  )

  expect(eingestiegen).toHaveBeenCalledWith('p1', 'sitz-2')

  // **Und der andere Eintrag meint den anderen** — ohne diese Zeile wäre der
  // Lauf auch grün, wenn beide denselben Platz meinten.
  await weiterspielen(
    hand,
    'David · Gegenüber',
    einsetzen(eigene.partienEinsteigenAls, { name: 'David' }),
  )

  expect(eingestiegen).toHaveBeenLastCalledWith('p1', 'sitz-1')
})

// **Zuschauen geht aus der Liste** — mit der leeren Sitzkennung, die
// `Partie.Sicht("")` als Zuschauer liest (PLAY-05.AK-02, #1713). Bis UXR-13
// ging es nur über die Sichtwahl am Tisch, also erst nach dem Hinsetzen.
it('schaut einer Partie ohne Sitz zu', async () => {
  const eingestiegen = vi.fn()

  const hand = zeige([eineZeile()], eingestiegen)

  await weiterspielen(hand, 'David · Gegenüber', eigene.partienZuschauen)

  expect(eingestiegen).toHaveBeenCalledWith('p1', '')
})

// **Die Zeile nennt die Namen und die Zahl der Karten — und keine Karte.**
it('nennt die Sitze und zählt die Objekte', () => {
  zeige([eineZeile()])

  expect(screen.getByText(/David · Gegenüber/)).toBeInTheDocument()
  expect(screen.getByText(`${eigene.partienObjekte}: 12`)).toBeInTheDocument()
})

// **Ein Zug steht da, wo es einen gibt.**
//
// **Und null ist keiner** (G9, #1592): eine Partie aus dem Sandkasten hat
// keinen Zugbegriff, und eine „Zug 0“ wäre eine Auskunft, die es nicht gibt.
it('nennt den Zug nur, wo es zugweise zugeht', () => {
  const { unmount } = render(
    <Sprachraum sprache="de">
      <Partienliste
        geladen
        wiederholen={() => undefined}
        partien={[eineZeile()]}
        fehler={undefined}
        einsteigen={() => undefined}
        verwaltung={stilleVerwaltung()}
      />
    </Sprachraum>,
  )

  expect(
    screen.queryByText(new RegExp(eigene.partienZug)),
  ).not.toBeInTheDocument()

  unmount()

  zeige([eineZeile({ zug: 4 })])

  expect(
    screen.getByText(new RegExp(`${eigene.partienZug} 4`)),
  ).toBeInTheDocument()
})

// **Eine Partie ohne Namen steht mit ihrer Kennung da.**
//
// **Weil sie sonst als leere Zeile dastünde** — ein Tisch, den niemand
// benennen kann, ist immer noch einer (§8).
it('nennt eine namenlose Partie bei ihrer Kennung', async () => {
  const eingestiegen = vi.fn()

  const hand = zeige(
    [eineZeile({ sitze: [{ kennung: 'sitz-1', name: '' }] })],
    eingestiegen,
  )

  expect(screen.getByText(/p1/)).toBeInTheDocument()

  await weiterspielen(
    hand,
    'p1',
    einsetzen(eigene.partienEinsteigenAls, { name: 'sitz-1' }),
  )

  expect(eingestiegen).toHaveBeenCalledWith('p1', 'sitz-1')
})

// **Ein Fehler steht als Meldung da und nicht als leere Liste.**
it('meldet, was beim Lesen schiefging', () => {
  zeige([], () => undefined, 'Der Bestand ist zu.')

  expect(screen.getByRole('alert')).toHaveTextContent('Der Bestand ist zu.')
})

// **„dran" steht an der Partie, auf die gewartet wird — und nur dort.**
// UXR-F16 (#2031).
//
// **Drei Partien, drei Gründe** (die Rechnung selbst steht in `service`, hier
// kommt nur an, was sie sagt): eine, auf die gewartet wird; eine, in der man
// einen Sitz hat und nicht dran ist; und eine, in der man zusieht. Die Marke
// gehört **an ihre Zeile** — wer nur zählte, dass irgendwo eine steht, wäre auch
// grün, wenn sie an der falschen hinge (§10).
it('zeigt die Marke an der Partie, auf die gewartet wird', () => {
  const anna = (id: string, name: string) => [
    { kennung: `${id}-a`, name },
    { kennung: `${id}-b`, name: 'Gegenüber' },
  ]

  zeige([
    eineZeile({
      id: 'p1',
      sitze: anna('p1', 'Anna'),
      eigenerSitz: 'p1-a',
      dran: true,
    }),
    eineZeile({
      id: 'p2',
      sitze: anna('p2', 'Bert'),
      eigenerSitz: 'p2-a',
      dran: false,
    }),
    eineZeile({ id: 'p3', sitze: anna('p3', 'Cora') }),
  ])

  // zeileVon gibt die Zeile, in der dieser Name steht.
  const zeileVon = (name: string): HTMLElement => {
    const zeile = screen.getByText(`${name} · Gegenüber`).closest('li')

    if (zeile === null) {
      throw new Error(`keine Zeile für ${name}`)
    }

    return zeile
  }

  expect(
    within(zeileVon('Anna')).getByText(eigene.partienDran),
  ).toBeInTheDocument()
  expect(
    within(zeileVon('Bert')).queryByText(eigene.partienDran),
  ).not.toBeInTheDocument()
  expect(
    within(zeileVon('Cora')).queryByText(eigene.partienDran),
  ).not.toBeInTheDocument()
})

// **Die Marke ist Text und kein Bedienelement, und sie ist übersetzt.**
//
// **Gemessen an einer zweiten Sprache**, weil ein Wort, das nur im Katalog
// der Ausgangssprache stünde, in der anderen als Schlüssel dastünde.
it('nennt die Marke in der Sprache des Fensters', () => {
  render(
    <Sprachraum sprache="en">
      <Partienliste
        geladen
        wiederholen={() => undefined}
        partien={[eineZeile({ eigenerSitz: 'sitz-1', dran: true })]}
        fehler={undefined}
        einsteigen={() => undefined}
        verwaltung={stilleVerwaltung()}
      />
    </Sprachraum>,
  )

  const englisch = kataloge.en.seiten.spielen.partienDran

  expect(englisch).not.toBe(eigene.partienDran)
  expect(screen.getByText(englisch)).toBeInTheDocument()
  expect(screen.queryByText(eigene.partienDran)).not.toBeInTheDocument()
})
