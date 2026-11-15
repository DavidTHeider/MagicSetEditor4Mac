// **Wer am Zug ist, steht am Tisch** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
//   - **Ohne Zug keine Leiste** — eine Partie aus G1 bis G8 ist ein
//     Sandkasten, und „Zug 0“ wäre eine Auskunft, die es nicht gibt.
//   - **Wer dran ist, liest es** — und wer nicht, erfährt, auf wen gewartet
//     wird (§6).
//   - **Passen geht nur, wer Priorität hat** (CR 117.3d).
//   - **Wo niemand Priorität hat, steht das da** (CR 502.4, CR 514.3).
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Zugleiste } from './Zugleiste'
import { einZug, einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'

import type { Partie } from '../apipartie'

const eigene = kataloge.de.seiten.spielen

// zeige zeichnet die Leiste und gibt die Hand.
function zeige(
  partie: Partie,
  passen = () => undefined,
  laeuft = false,
  abkuerzen = () => undefined,
  zugende = () => undefined,
) {
  render(
    <Sprachraum sprache="de">
      <Zugleiste
        partie={partie}
        erlaubt={['zug.passen', 'zug.abkuerzen', 'zug.beenden']}
        laeuft={laeuft}
        passen={passen}
        abkuerzen={abkuerzen}
        zugende={zugende}
      />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// **Ohne laufenden Zug steht keine Leiste da.**
it('zeichnet nichts, solange es nicht zugweise zugeht', () => {
  zeige(einePartie())

  expect(screen.queryByLabelText(eigene.zugLeiste)).not.toBeInTheDocument()
})

// **Wer dran ist, liest es — und passt.**
it('nennt den Zug und lässt den Dranseienden passen', async () => {
  const gepasst = vi.fn()

  const hand = zeige(
    einePartie({
      fuer: 'sitz1',
      zug: einZug({
        nummer: 3,
        aktiv: 'sitz1',
        phase: 'anfang',
        schritt: 'versorgung',
        prioritaet: 'sitz1',
      }),
    }),
    gepasst,
  )

  expect(screen.getByText(`${eigene.zug} 3`)).toBeInTheDocument()
  expect(
    screen.getByText(`${eigene.phaseAnfang} · ${eigene.schrittVersorgung}`),
  ).toBeInTheDocument()
  expect(screen.getByText(eigene.zugDuBistDran)).toBeInTheDocument()

  await hand.click(screen.getByRole('button', { name: eigene.zugPassen }))

  expect(gepasst).toHaveBeenCalledTimes(1)
})

// **Wer nicht dran ist, erfährt, auf wen gewartet wird — und passt nicht.**
it('nennt den Wartegrund und hält den Knopf an', () => {
  zeige(
    einePartie({
      fuer: 'sitz2',
      zug: einZug({
        nummer: 1,
        aktiv: 'sitz1',
        phase: 'hauptVorKampf',
        prioritaet: 'sitz1',
      }),
    }),
  )

  expect(
    screen.getByText(new RegExp(`${eigene.zugWartetAuf} Anna`)),
  ).toBeInTheDocument()

  expect(screen.getByRole('button', { name: eigene.zugPassen })).toBeDisabled()

  // **Eine Hauptphase hat keinen Schritt** — CR 500.1 zerlegt nur Anfang,
  // Kampf und Ende.
  expect(screen.getByText(eigene.phaseHauptVorKampf)).toBeInTheDocument()
})

// **Wo niemand Priorität hat, steht das da** — CR 502.4, CR 514.3.
it('sagt, wenn niemand Priorität hat', () => {
  zeige(
    einePartie({
      fuer: 'sitz1',
      zug: einZug({
        nummer: 1,
        aktiv: 'sitz1',
        phase: 'anfang',
        schritt: 'enttappen',
      }),
    }),
  )

  expect(screen.getByText(eigene.zugNiemandDran)).toBeInTheDocument()
})

// **Eine offene Abwurfpflicht steht am Tisch** — CR 514.1.
it('nennt die offene Abwurfpflicht', () => {
  zeige(
    einePartie({
      fuer: 'sitz1',
      zug: einZug({
        nummer: 1,
        aktiv: 'sitz1',
        phase: 'ende',
        schritt: 'aufraeumen',
        abwurf: 2,
      }),
    }),
  )

  expect(screen.getByText(`${eigene.zugAbwurf}: 2`)).toBeInTheDocument()
})

// **Ein Sitz, den die Sicht nicht führt, steht bei seiner Kennung.**
//
// **Und nicht als leere Stelle** (§8): fremde Daten sind der Normalfall, und
// ein Name, den niemand kennt, ist kein Grund, nichts zu sagen.
it('nennt einen unbekannten Sitz bei seiner Kennung', () => {
  zeige(
    einePartie({
      fuer: 'sitz1',
      zug: einZug({ nummer: 1, aktiv: 'sitz9', prioritaet: 'sitz1' }),
    }),
  )

  expect(screen.getByText(`${eigene.zugAktiv}: sitz9`)).toBeInTheDocument()
})

// **Eine Phase, die der Katalog nicht kennt, steht bei ihrem Namen.**
//
// **Und nicht als leere Stelle** (§8): ein Zustand kann aus einem Bestand
// kommen, den jemand mit einem anderen Werkzeug geöffnet hat — eine Leiste,
// die daran nichts mehr sagte, wäre unbrauchbar.
it('nennt eine unbekannte Phase und einen unbekannten Schritt', () => {
  zeige(
    einePartie({
      fuer: 'sitz1',
      zug: einZug({
        nummer: 1,
        aktiv: 'sitz1',
        phase: 'erfunden',
        schritt: 'auch',
        prioritaet: 'sitz1',
      }),
    }),
  )

  expect(screen.getByText('erfunden · auch')).toBeInTheDocument()
})

// **Die Abkürzung steht neben dem Passen** — CR 732 (M1, #1603).
//
// **Zehn Abschnitte je Zug vergeben Priorität**; bei zwei Sitzen sind das
// zwanzig Passen, in denen nichts geschieht. Dieser Knopf ist die Abkürzung.
it('reicht auf Zuruf durch', async () => {
  const abgekuerzt = vi.fn()

  const hand = zeige(
    einePartie({
      zug: einZug({ nummer: 1, aktiv: 'sitz-1', prioritaet: 'sitz-1' }),
      fuer: 'sitz-1',
    }),
    () => undefined,
    false,
    abgekuerzt,
  )

  await hand.click(screen.getByRole('button', { name: eigene.zugAbkuerzen }))

  expect(abgekuerzt).toHaveBeenCalledOnce()
})

// **Wer nicht dran ist, kürzt nicht ab** — CR 732.2a: vorschlagen darf, wer
// Priorität hat.
it('lässt den, der nicht dran ist, nicht abkürzen', () => {
  zeige(
    einePartie({
      zug: einZug({ nummer: 1, aktiv: 'sitz-1', prioritaet: 'sitz-2' }),
      fuer: 'sitz-1',
    }),
  )

  expect(
    screen.getByRole('button', { name: eigene.zugAbkuerzen }),
  ).toBeDisabled()
})

// **Und der Zug lässt sich vorzeitig beenden** — CR 724 (M1, #1603).
it('beendet den Zug auf Zuruf', async () => {
  const beendet = vi.fn()

  const hand = zeige(
    einePartie({
      zug: einZug({ nummer: 1, aktiv: 'sitz-1', prioritaet: 'sitz-1' }),
      fuer: 'sitz-1',
    }),
    () => undefined,
    false,
    () => undefined,
    beendet,
  )

  await hand.click(screen.getByRole('button', { name: eigene.zugBeenden }))

  expect(beendet).toHaveBeenCalledOnce()
})
