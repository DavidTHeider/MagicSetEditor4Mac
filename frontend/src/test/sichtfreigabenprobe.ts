// **Die Vorrichtung der Läufe über die Angaben einer Sicht** — UXR-F18
// (#2033).
//
// **Ein Gegenüber, das handeln darf, zwei Drucke in der Sammlung und der Kasten
// „Was dieses Gegenüber sieht" offen.** Zwei Läufe brauchen sie
// (`sichtfreigabenumfang.test.tsx`, `sichtfreigabenbefund.test.tsx`), und beide
// gingen allein über die Zeilengrenze aus §2.
import { screen } from '@testing-library/react'

import { kataloge } from '../sprache/kataloge'
import { einKontakt, eineListe, zeige } from '../seiten/peerkontaktprobe'
import { belegeApi } from './api'
import { exemplar } from './sammlungsseite'

import type { Sichtfreigabe, Sichtfreigabenantwort } from '../apisichtfreigabe'

export const eigene = kataloge.de.seiten.technik
export const worte = kataloge.de.bausteine.druckwahl

const aufklapper = eigene.freigabenTitel

const dieUmfaenge = [
  'auswahl',
  'dubletten',
  'wuensche',
  'sammlung',
  'digitales',
]

// besitz sind zwei Drucke, die jemand hat.
export const besitz = [
  exemplar({
    id: 1,
    printing: 'bolt',
    bezeichnung: 'Blitzschlag',
    menge: 3,
  }),
  exemplar({
    id: 2,
    printing: 'giant',
    bezeichnung: 'Riesenwuchs',
    menge: 1,
  }),
]

export function freigabenliste(
  freigaben: Sichtfreigabe[] = [],
  fehler: Sichtfreigabenantwort['fehler'] = null,
): Sichtfreigabenantwort {
  return { freigaben, umfaenge: dieUmfaenge, fehler }
}

export function eineFreigabe(
  anders: Partial<Sichtfreigabe> = {},
): Sichtfreigabe {
  return {
    empfaenger: 'bgegenueber',
    umfang: 'sammlung',
    zweck: 'handel',
    auswahl: [],
    behalten: 0,
    ablauf: '',
    erteilt: '2026-09-18T12:00:00Z',
    ...anders,
  }
}

// aufstellen belegt die Bindungen und öffnet den Kasten.
export async function aufstellen(
  mehr: Parameters<typeof belegeApi>[0] = {},
  freigaben: Sichtfreigabenantwort = freigabenliste(),
) {
  belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })]),
      ),
    Sammlungsfreigaben: () => Promise.resolve(freigaben),
    Sammlung: () => Promise.resolve({ exemplare: besitz, fehler: null }),
    ...mehr,
  })

  const hand = await zeige(aufklapper)

  // **Auf die Wirkung warten, nicht auf den Ruf** (§3): erst wenn die Drucke
  // dastehen, ist die Sammlung gelesen und der Kasten fertig gezeichnet. Die
  // Gruppe allein steht schon vorher da, und ein Lauf, der an ihr zugriffe,
  // fände sie leer.
  await screen.findAllByRole('checkbox')

  return hand
}

export type Hand = Awaited<ReturnType<typeof aufstellen>>

export async function umfangWaehlen(hand: Hand, umfang: string) {
  await hand.selectOptions(
    screen.getByLabelText(eigene.freigabenUmfang),
    umfang,
  )
}

export const erteilen = () =>
  screen.getByRole('button', { name: eigene.freigabenErteilen })
