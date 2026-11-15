// **Die eigene Stimme zum Abschluss** — TRADE-F1
// (#2026).
//
// # Der Befund
//
// Der Schritt `bestaetigen` führt von *Abgleich offen* nach *beidseitig
// verbucht* erst, wenn **jede** Partei gestimmt hat
// (`internal/domain/handelsmatrix.go`, `Alle`). Die Lage sagte „Du bist dran:
// Abschluss bestätigen", das Fenster bot aber nur „Gegenseite bestätigt" an —
// die Stimme des Gegenübers. Die eigene ging nur über `mse handel bestaetigen`
// und REST; aus dem Fenster allein war ein beidseitiger Abschluss nicht zu
// erreichen (§16).
//
// # Was jetzt gilt
//
// **Wessen Stimme es ist, sagt der Dienst** (`ich` im Handelsvorgang) — das
// Fenster rechnet sie nicht aus den Parteien heraus. Fehlt sie, führt
// „Abschluss bestätigen"; die Freigabe gilt genau dieser Stimme.
import { within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import { zeigeHandel } from './tauschprobe'
import {
  bestaetigt,
  buchung,
  dieVorgaenge,
  einJournal,
  einStand,
  eineFreigabe,
  imAbschluss,
  trageEin,
  mitBelegnachlauf,
} from './buchungsprobe'
import { kataloge } from '../sprache/kataloge'

import type { Handelsvorgang } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

const offen = {
  zustand: 'abgleichOffen',
  naechster: 'bestaetigen',
  moegliche: ['bestaetigen', 'klaeren'],
  beschluss: 'bestaetigen',
}

// knoepfe nennt die Knöpfe im Kasten samt Gewicht — ohne das Fragezeichen.
function knoepfe(kasten: HTMLElement) {
  return within(kasten)
    .getAllByRole('button')
    .filter((knopf) => !knopf.hasAttribute('aria-label'))
    .map((knopf) => ({
      name: knopf.textContent,
      haupt: knopf.classList.contains('bg-marke'),
    }))
}

function mitStand(
  stand: Handelsvorgang,
  weitere: Record<string, unknown> = {},
) {
  belegeApi(
    mitBelegnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: dieVorgaenge(stand),
      HandelFreigabeErteilen: eineFreigabe,
      Handelsjournal: () => Promise.resolve(einJournal()),
      ...weitere,
    }),
  )
}

it('führt mit der eigenen Stimme, solange sie fehlt', async () => {
  mitStand(imAbschluss(offen))

  const kasten = await buchung(await zeigeHandel())

  await within(kasten).findByRole('button', {
    name: eigene.buchungEigeneStimme,
  })
  expect(knoepfe(kasten)).toEqual([
    { name: eigene.buchungEigeneStimme, haupt: true },
    { name: eigene.buchungBestaetigen, haupt: false },
  ])
})

it('setzt die eigene Stimme mit einer Freigabe genau für sie', async () => {
  const erteilt = vi.fn(eineFreigabe)
  const gestimmt = vi.fn(() =>
    Promise.resolve(einStand({ zustand: 'abgleichOffen' })),
  )
  mitStand(imAbschluss(offen), {
    HandelFreigabeErteilen: erteilt,
    Handelsbestaetigung: gestimmt,
  })

  await trageEin(await zeigeHandel(), eigene.buchungEigeneStimme)

  expect(erteilt).toHaveBeenCalledWith('v-1', 'bestaetigen', 'bich')
  expect(gestimmt).toHaveBeenCalledWith('v-1', 'bich', 'f-1')
})

it('fragt nicht mehr, wenn die eigene Stimme da ist', async () => {
  mitStand(imAbschluss({ ...offen, zustimmungen: ['bich'] }))

  const kasten = await buchung(await zeigeHandel())

  await within(kasten).findByRole('button', {
    name: eigene.buchungBestaetigen,
  })
  expect(knoepfe(kasten)).toEqual([
    { name: eigene.buchungBestaetigen, haupt: false },
  ])
})

it('bietet keine eigene Stimme an, wo der Dienst keine Partei nennt', async () => {
  mitStand(imAbschluss({ ...offen, ich: '' }))

  const kasten = await buchung(await zeigeHandel())

  await within(kasten).findByRole('button', {
    name: eigene.buchungBestaetigen,
  })
  expect(
    within(kasten).queryByRole('button', { name: eigene.buchungEigeneStimme }),
  ).toBeNull()
})
