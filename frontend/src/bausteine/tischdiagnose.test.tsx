// **Wie die Leitung steht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
//   - Fehlende Messgrößen erscheinen als **„unbekannt"** und nicht als Null.
//     Das steht so in der Abnahme des Issues.
//   - Die **Bildzeit** entsteht hier und kommt nicht aus Go.
//   - Ein Tisch, der nicht läuft, sagt es — statt eine leere Tabelle zu
//     zeigen (§6).
import { act, render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import { Bildproben } from '../haken/usePartiediagnose'
import { Sprachraum } from '../sprache/kontext'
import { Tischdiagnose } from './Tischdiagnose'
import { belegeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { PartiediagnoseAntwort } from '../apinebenbei'

const eigene = kataloge.de.seiten.spielen

// bilder sammelt die Rückrufe, damit ein Lauf die Uhr selbst dreht.
//
// **Selbst drehen und nicht abwarten**: `requestAnimationFrame` feuert unter
// jsdom nie von allein, und ein Lauf, der darauf wartete, prüfte die Uhr
// dieser Maschine (§3).
let bilder: ((jetzt: number) => void)[] = []

beforeEach(() => {
  bilder = []

  vi.stubGlobal(
    'requestAnimationFrame',
    (rueckruf: (jetzt: number) => void) => {
      bilder.push(rueckruf)

      return bilder.length
    },
  )

  vi.stubGlobal('cancelAnimationFrame', () => undefined)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

// einStand gibt eine Antwort mit den genannten Abweichungen.
function einStand(
  abweichend: Partial<PartiediagnoseAntwort> = {},
): PartiediagnoseAntwort {
  return {
    laeuft: true,
    rolle: 'wirt',
    partie: 'p1',
    pausiert: false,
    stand: 0,
    standBekannt: false,
    rueckkehrMillis: 0,
    rueckkehrBekannt: false,
    leitungen: [],
    fehler: null,
    ...abweichend,
  }
}

// zeige zeichnet die Diagnose und gibt die Hand.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Tischdiagnose />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// schlage dreht die Uhr um so viele Bilder weiter.
function schlage(anzahl: number, abstand = 16) {
  act(() => {
    for (let bild = 1; bild <= anzahl; bild += 1) {
      const naechster = bilder[bilder.length - 1]

      if (naechster === undefined) {
        throw new Error('es wurde kein Bild angefordert')
      }

      naechster(bild * abstand)
    }
  })
}

it('holt den Stand nicht von selbst und sagt, dass die Bildzeit fehlt', () => {
  const geholt = vi.fn()
  belegeApi({ Partiediagnose: geholt })

  zeige()

  expect(geholt).not.toHaveBeenCalled()

  expect(
    screen.getByText(`${eigene.diagnoseBildzeit}: ${eigene.diagnoseUnbekannt}`),
  ).toBeInTheDocument()
})

// **Die Bildzeit entsteht im Fenster** — nach genug Bildern steht sie da.
it('mittelt die Bildzeit über die Proben', async () => {
  belegeApi({ Partiediagnose: vi.fn().mockResolvedValue(einStand()) })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.diagnoseAnsehen }))

  // **Ein Bild mehr als Proben**: das erste setzt nur den Anfangspunkt, ein
  // Abstand entsteht erst mit dem zweiten.
  schlage(Bildproben + 1, 20)

  expect(
    screen.getByText(`${eigene.diagnoseBildzeit}: 20 ms`),
  ).toBeInTheDocument()
})

// **Keine Dauerschleife** — PLAY-06.AK-04
// (#1714).
// Gemessen am 2026-09-25 in der WebKit-Probe: die Messung forderte 60 Bilder
// je Sekunde an, solange der Tisch aufgeschlagen war — auch nach einem
// verlorenen Grafikkontext, als es nichts mehr zu messen gab.
it('fordert von selbst kein Bild an', () => {
  belegeApi({ Partiediagnose: vi.fn() })

  zeige()

  expect(bilder).toHaveLength(0)
})

it('hört nach den Proben auf', async () => {
  belegeApi({ Partiediagnose: vi.fn().mockResolvedValue(einStand()) })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.diagnoseAnsehen }))

  schlage(Bildproben + 1, 20)

  // **So viele Anforderungen wie Rückrufe** — der letzte fordert kein
  // weiteres Bild an, sonst stünde eines offen, und die Messung liefe
  // weiter, bis jemand die Seite verlässt.
  expect(bilder).toHaveLength(Bildproben + 1)
})

it('bricht eine laufende Messung ab, wenn jemand noch einmal nachsieht', async () => {
  const abgebrochen = vi.fn()
  vi.stubGlobal('cancelAnimationFrame', abgebrochen)
  belegeApi({ Partiediagnose: vi.fn().mockResolvedValue(einStand()) })

  const hand = zeige()
  const knopf = screen.getByRole('button', { name: eigene.diagnoseAnsehen })

  await hand.click(knopf)
  schlage(5)

  // **Erst wenn der erste Ruf zurück ist** — solange er läuft, ruht der Knopf.
  await screen.findByText(
    `${eigene.diagnoseRunde}: ${eigene.diagnoseUnbekannt}`,
  )
  await hand.click(knopf)

  expect(abgebrochen).toHaveBeenCalledWith(6)
})

it('sagt, wenn gerade kein Tisch läuft', async () => {
  belegeApi({
    Partiediagnose: vi.fn().mockResolvedValue(einStand({ laeuft: false })),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.diagnoseAnsehen }))

  expect(await screen.findByText(eigene.diagnoseStill)).toBeInTheDocument()
})

// **Fehlende Messgrößen erscheinen unbekannt** — und nicht als Null.
it('nennt eine ungemessene Runde unbekannt', async () => {
  belegeApi({ Partiediagnose: vi.fn().mockResolvedValue(einStand()) })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.diagnoseAnsehen }))

  expect(
    await screen.findByText(
      `${eigene.diagnoseRunde}: ${eigene.diagnoseUnbekannt}`,
    ),
  ).toBeInTheDocument()

  expect(screen.getByText(eigene.diagnoseOhneLeitung)).toBeInTheDocument()

  expect(
    screen.queryByText(new RegExp(eigene.diagnoseStand)),
  ).not.toBeInTheDocument()
})

it('nennt jede Leitung mit ihrem Weg und ihrer Warteschlange', async () => {
  belegeApi({
    Partiediagnose: vi.fn().mockResolvedValue(
      einStand({
        stand: 12,
        standBekannt: true,
        rueckkehrMillis: 42,
        rueckkehrBekannt: true,
        leitungen: [
          {
            sitz: 'bert',
            weg: 'tcp',
            gegenueber: 'AB:CD',
            seitSekunden: 7,
            wartend: 1,
            verworfen: 2,
          },
        ],
      }),
    ),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.diagnoseAnsehen }))

  expect(await screen.findByText('bert')).toBeInTheDocument()

  for (const stueck of [
    eigene.diagnoseAlsWirt,
    `${eigene.diagnoseStand}: 12`,
    `${eigene.diagnoseRunde}: 42 ms`,
    `${eigene.diagnoseWeg}: tcp`,
    `${eigene.diagnoseGegenueber}: AB:CD`,
    `${eigene.diagnoseSeit}: 7 s`,
    `${eigene.diagnoseWartend}: 1`,
    `${eigene.diagnoseVerworfen}: 2`,
  ]) {
    expect(screen.getByText(stueck)).toBeInTheDocument()
  }
})

// **Ein Gast sieht seine eine Leitung — und dass sie abriss.**
it('sagt dem Gast, dass sein Platz bleibt', async () => {
  belegeApi({
    Partiediagnose: vi.fn().mockResolvedValue(
      einStand({
        rolle: 'gast',
        pausiert: true,
        leitungen: [
          {
            sitz: '',
            weg: 'schleife',
            gegenueber: '',
            seitSekunden: 0,
            wartend: 0,
            verworfen: 0,
          },
        ],
      }),
    ),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.diagnoseAnsehen }))

  expect(await screen.findByText(eigene.diagnoseAlsGast)).toBeInTheDocument()
  expect(screen.getByText(eigene.diagnosePausiert)).toBeInTheDocument()

  expect(
    screen.queryByText(new RegExp(eigene.diagnoseGegenueber)),
  ).not.toBeInTheDocument()
})

it('meldet einen Fehler', async () => {
  belegeApi({
    Partiediagnose: vi.fn().mockResolvedValue(
      einStand({
        fehler: {
          code: 'erlaubnis',
          schluessel: 'erlaubnis',
          meldung: 'darfst du nicht',
        },
      }),
    ),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.diagnoseAnsehen }))

  // **Der Satz kommt aus dem Katalog** — was Go schickt, ist ein Schlüssel,
  // und die Oberfläche setzt ihn in die Sprache dieses Fensters (§9).
  expect(await screen.findByRole('alert')).toHaveTextContent(
    kataloge.de.fehler.erlaubnis,
  )
})
