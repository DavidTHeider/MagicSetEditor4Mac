// **Welche Knöpfe der Buchung stehen** — UXR-15
// (#1923),
// nachgeschärft beim Abschluss von Phase 6.8.
//
// # Was vorher war
//
// **Drei Knöpfe in jeder Phase und keiner im Hauptgewicht.** „Verbuchen",
// „Abgleich melden" und „Gegenseite bestätigt" standen auch vor der Übergabe
// da und meldeten dann einen Fehler; und wo der nächste Schritt einer von
// ihnen war, führte keiner (R1 in `docs/UIREVIEW.md`). Nach einer Buchung
// blieben Lage und Schritte darüber auf dem Stand von vorher stehen.
//
// # Was jetzt gilt
//
// **Vor der Übergabe ein Satz statt drei Knöpfen.** Danach steht
// „Verbuchen", bis gebucht ist — bis zur Hüllenprobe `nat04handel` (#1672)
// auch danach, denn ein zweiter Druck war der einzige Weg, den Beleg zu
// sehen; seither liest der Kasten ihn —, „Abgleich melden", wo der Vertrag
// ihn nennt, und „Gegenseite bestätigt", solange ein Abschluss auf die
// Stimme des Gegenübers wartet. **Der nächste Schritt führt**, und nach jedem lädt der
// Raum den Handel nach.
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import { einVorgang } from './handelslisteprobe'
import { zeigeHandel } from './tauschprobe'
import {
  bestaetigt,
  buchung,
  dieVorgaenge,
  einJournal,
  einStand,
  eineFreigabe,
  imAbschluss,
  mitBelegnachlauf,
  trageEin,
} from './buchungsprobe'
import { kataloge } from '../sprache/kataloge'

import type { Handelsvorgang } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

// mitStaenden belegt den Raum mit diesen Ständen des Handels `v-1`.
function mitStaenden(erster: Handelsvorgang, ...weitere: Handelsvorgang[]) {
  belegeApi(
    mitBelegnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: dieVorgaenge(erster, ...weitere),
      HandelFreigabeErteilen: eineFreigabe,
      Handelsverbuchung: () => Promise.resolve(einStand()),
      Handelsjournal: () => Promise.resolve(einJournal()),
    }),
  )
}

// knoepfe nennt die Knöpfe im Kasten — ohne das Fragezeichen der Erklärung,
// das als einziges seinen Namen im `aria-label` trägt.
function knoepfe(kasten: HTMLElement) {
  return within(kasten)
    .getAllByRole('button')
    .filter((knopf) => !knopf.hasAttribute('aria-label'))
    .map((knopf) => ({
      name: knopf.textContent,
      haupt: knopf.classList.contains('bg-marke'),
    }))
}

it('zeigt vor der Übergabe keinen Knopf, sondern sagt warum', async () => {
  mitStaenden(einVorgang({ vorgang: 'v-1' }))

  const kasten = await buchung(await zeigeHandel())

  expect(
    await within(kasten).findByText(eigene.buchungNochNicht),
  ).toBeInTheDocument()
  expect(knoepfe(kasten)).toEqual([])
})

it('führt nach der Übergabe mit „Verbuchen"', async () => {
  mitStaenden(imAbschluss())

  const kasten = await buchung(await zeigeHandel())

  await within(kasten).findByRole('button', { name: eigene.buchungVerbuchen })
  expect(knoepfe(kasten)).toEqual([
    { name: eigene.buchungVerbuchen, haupt: true },
  ])
})

// **Was der Vertrag nicht als nächsten nennt, führt nicht** — auch wenn es
// gehen darf: hier sagt er „klären", und „Verbuchen" steht leise daneben.
it('führt nicht mit „Verbuchen", wo der Vertrag anderes nennt', async () => {
  mitStaenden(imAbschluss({ naechster: 'klaeren' }))

  const kasten = await buchung(await zeigeHandel())

  await within(kasten).findByRole('button', { name: eigene.buchungVerbuchen })
  expect(knoepfe(kasten)).toEqual([
    { name: eigene.buchungVerbuchen, haupt: false },
  ])
})

it('führt nach dem Buchen mit dem Abgleich', async () => {
  mitStaenden(
    imAbschluss({
      zustand: 'lokalVerbucht',
      naechster: 'abgleichen',
      moegliche: ['abgleichen', 'klaeren'],
    }),
  )

  const kasten = await buchung(await zeigeHandel())

  await within(kasten).findByRole('button', { name: eigene.buchungAbgleichen })
  expect(knoepfe(kasten)).toEqual([
    { name: eigene.buchungAbgleichen, haupt: true },
  ])
})

// **Die eigene Stimme ist hier schon da** (`zustimmungen: ['bich']`) — seit
// #2026 steht sonst auch „Abschluss bestätigen"; was der Fall hält, ist die
// Stimme des Gegenübers (`handelseigenestimme.test.tsx` hält die eigene).
it('fragt nach der Stimme des Gegenübers, solange sie fehlt', async () => {
  const offen = {
    zustand: 'abgleichOffen',
    naechster: 'bestaetigen',
    moegliche: ['bestaetigen', 'klaeren'],
    beschluss: 'bestaetigen',
    zustimmungen: ['bich'],
  }
  mitStaenden(imAbschluss(offen))

  const kasten = await buchung(await zeigeHandel())

  await within(kasten).findByRole('button', {
    name: eigene.buchungBestaetigen,
  })
  expect(knoepfe(kasten)).toEqual([
    { name: eigene.buchungBestaetigen, haupt: false },
  ])
})

it('fragt nicht mehr, wenn das Gegenüber gestimmt hat', async () => {
  mitStaenden(
    imAbschluss({
      zustand: 'abgleichOffen',
      naechster: 'bestaetigen',
      moegliche: ['bestaetigen', 'klaeren'],
      beschluss: 'bestaetigen',
      zustimmungen: ['bgegenueber'],
    }),
  )

  const kasten = await buchung(await zeigeHandel())

  // **Übrig bleibt die eigene Stimme, und sie führt** (#2026) — bis dahin
  // stand hier nur „Verbuchen", und der Abschluss war im Fenster nicht zu
  // erreichen.
  await within(kasten).findByRole('button', {
    name: eigene.buchungEigeneStimme,
  })
  expect(knoepfe(kasten)).toEqual([
    { name: eigene.buchungEigeneStimme, haupt: true },
  ])
})

// **Nach dem Buchen steht die neue Lage da** — bis hierher blieb „Du bist
// dran: verbuchen" stehen, bis jemand den Handel neu aufschlug.
it('lädt den Handel nach einer Buchung nach', async () => {
  mitStaenden(
    imAbschluss(),
    imAbschluss({
      zustand: 'lokalVerbucht',
      naechster: 'abgleichen',
      moegliche: ['abgleichen', 'klaeren'],
    }),
  )

  const hand = await zeigeHandel()
  await trageEin(hand, eigene.buchungVerbuchen)

  expect(
    await screen.findByText('Du bist dran: Abgleich melden'),
  ).toBeInTheDocument()
  expect(
    await screen.findByRole('button', { name: eigene.buchungAbgleichen }),
  ).toHaveClass('bg-marke')
})
