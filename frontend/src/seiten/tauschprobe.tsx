// **Die Vorrichtung der Tauschläufe** — Werkstück T4 von Phase 5.6
// (#1469).
//
// **Eine eigene Datei**, weil zwei Läufe sie brauchen und einer sonst über
// die Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `peerkontaktprobe.tsx`.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect } from 'vitest'

import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { einKontakt } from './peerkontaktprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type {
  Tauschabgleichantwort,
  Tauschkandidat,
  Tauschseite,
} from '../apitauschabgleich'

const eigene = kataloge.de.seiten.handeln

// einStand ist ein lesbarer fremder Stand.
export const einStand = JSON.stringify({
  angebot: [{ printing: 'bolt-m10', menge: 4 }],
  wuensche: [{ printing: 'shivan-lea', menge: 1 }],
})

// einKandidat baut einen Treffer.
export function einKandidat(
  anders: Partial<Tauschkandidat> = {},
): Tauschkandidat {
  return {
    wunsch: 'Shivan Dragon',
    angebot: 'Shivan Dragon · LEA #232',
    angebotDruck: 'shivan-lea',
    angebotZustand: 'nearMint',
    angebotSprache: 'de',
    angebotFinish: 'nonfoil',
    menge: 1,
    fehlend: 0,
    rang: 'normal',
    art: 'genau',
    ...anders,
  }
}

// eineSeite baut eine Seite des Vorschlags.
export function eineSeite(stueck: number): Tauschseite {
  return {
    zeilen: [],
    stueck,
    gedeckt: stueck,
    offen: 0,
    wert: 0,
    waehrung: '',
    bepreist: false,
  }
}

// einAbgleich baut die Antwort.
export function einAbgleich(anders: Partial<Tauschabgleichantwort> = {}) {
  return {
    kennung: 'bgegenueber',
    gebe: [einKandidat()],
    bekomme: [einKandidat({ wunsch: 'Lightning Bolt', rang: 'hoch' })],
    vorschlag: { gebe: eineSeite(1), bekomme: eineSeite(1) },
    hatVorschlag: true,
    befunde: [],
    fehler: null,
    ...anders,
  }
}

// bestaetigt ist ein Kontakt, an dem der Kasten erscheint.
export function bestaetigt() {
  return einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })
}

// zeigeHandel stellt den Bereich *Handeln* auf und schlägt das Gegenüber auf.
//
// **Über die Seite und nicht über den Kasten allein**: der Weg, den ein
// Mensch geht, führt über die Partnerwahl — und der `key` daran ist die
// Zusage, dass beim Wechsel nichts stehen bleibt (§7, Regel 1). Ein Lauf, der
// den Kasten unmittelbar aufstellte, prüfte sie nie.
export async function zeigeHandel(): Promise<
  ReturnType<typeof userEvent.setup>
> {
  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )

  const hand = userEvent.setup()

  await waehleGegenueber(hand, 'Fremder Rechner')

  return hand
}

// waehleGegenueber öffnet „Neuer Tausch …" und schlägt das Gegenüber auf —
// **seit UXR-15 steht die Wahl in der Schublade** (#1923) und nicht mehr
// offen in der Spalte links.
export async function waehleGegenueber(
  hand: ReturnType<typeof userEvent.setup>,
  name: string,
): Promise<void> {
  await hand.click(
    await screen.findByRole('button', { name: eigene.neuerTausch }),
  )

  await hand.click(
    await within(
      await screen.findByRole('dialog', { name: eigene.neuerTauschTitel }),
    ).findByRole('button', {
      name: einsetzen(eigene.partnerWaehlenName, { name }),
    }),
  )

  await waitFor(() => {
    expect(screen.queryByRole('dialog')).toBeNull()
  })
}

// setzeEin füllt das Feld und drückt den Knopf.
export async function setzeEin(
  hand: ReturnType<typeof userEvent.setup>,
  text: string,
): Promise<void> {
  await oeffneAbgleich(hand)

  const feld = await screen.findByLabelText(eigene.tauschStand)

  await hand.clear(feld)

  // **Eingesetzt und nicht getippt.** `type` liest `{` als Tastenbefehl, und
  // ein fremder Stand ist JSON — wer ihn Zeichen für Zeichen schickte, misst
  // die Tastatur statt des Abgleichs.
  if (text !== '') {
    await hand.click(feld)
    await hand.paste(text)
  }

  await hand.click(
    screen.getByRole('button', { name: eigene.tauschAbgleichen }),
  )
}

// oeffneHandel wählt einen Handel über seine Kennung — den Weg für
// Fachleute (UX-08.AK-04,
// #1680).
//
// **Für die Läufe der Kästen darunter**: sie messen, was ein Kasten mit einem
// gewählten Handel tut, und nicht die Wahl selbst — die steht in
// `handelswahl.test.tsx`. **Der Kasten wird nur geöffnet, wenn er zu ist**:
// ein zweiter Klick klappte ihn wieder zu.
export async function oeffneHandel(
  hand: ReturnType<typeof userEvent.setup>,
  vorgang: string,
): Promise<void> {
  // **Die Wahl steht in „Mit wem"** (TRADE-02, #1716).
  await zuPhase(hand, 'gegenueber')

  const schalter = await screen.findByRole('button', {
    name: eigene.wahlFachleute,
  })

  if (schalter.getAttribute('aria-expanded') !== 'true') {
    await hand.click(schalter)
  }

  const feld = screen.getByLabelText(eigene.wahlKennung)

  await hand.clear(feld)
  await hand.click(feld)
  await hand.paste(vorgang)
  await hand.click(screen.getByRole('button', { name: eigene.wahlOeffnen }))
}

// zuPhase schlägt eine Phase des Assistenten auf (TRADE-02, #1716).
//
// **Für die Läufe der Tafeln**: jede liegt hinter ihrem Schritt, und wer
// eine prüft, geht erst dorthin — so, wie ein Mensch es täte.
export async function zuPhase(
  hand: ReturnType<typeof userEvent.setup>,
  phase: keyof typeof eigene.assistent.phasen,
): Promise<void> {
  // **In der Schrittleiste und nicht irgendwo** — „Verbuchen" heißt auch
  // der Knopf der Buchung.
  const leiste = await screen.findByRole('navigation', {
    name: eigene.assistent.schrittleiste,
  })

  await hand.click(
    within(leiste).getByRole('button', {
      name: new RegExp(eigene.assistent.phasen[phase]),
    }),
  )
}

// oeffneUnterwegs schlägt auf, was zu diesem Gegenüber unterwegs ist — in
// „Verbuchen", im Kasten (TRADE-02, #1716).
export async function oeffneUnterwegs(
  hand: ReturnType<typeof userEvent.setup>,
): Promise<void> {
  await zuPhase(hand, 'verbuchen')
  await hand.click(
    await screen.findByRole('button', { name: eigene.assistent.unterwegs }),
  )
}

// oeffneAbgleich schlägt den Abgleich auf: „Was tauscht ihr", der Kasten
// darin — **geöffnet wird er nur, wenn er zu ist** (TRADE-02, #1716).
export async function oeffneAbgleich(
  hand: ReturnType<typeof userEvent.setup>,
): Promise<void> {
  await zuPhase(hand, 'angebot')

  const kasten = await screen.findByRole('button', {
    name: eigene.assistent.abgleich,
  })

  if (kasten.getAttribute('aria-expanded') !== 'true') {
    await hand.click(kasten)
  }
}

// keinAnhalt belegt den Preisanhalt ohne einen Preis (TRADE-02b, #1951) —
// für Läufe, die Zeilen anlegen und nicht nach Preisen fragen.
export const keinAnhalt = {
  HandelPreisanhalt: () =>
    Promise.resolve({
      anhalt: { zeilen: [], seiten: [], differenzen: [], hauptwaehrung: '' },
      fehler: null,
    }),
}

// oeffneZeile öffnet im Vorgang das Formular einer Spalte (TRADE-02b,
// #1951) — **erst, wenn ihr Knopf geht**: die eigene Spalte wartet auf den
// Ausweis, sonst ginge eine Zeile mit leerer Seite hinaus.
export async function oeffneZeile(
  hand: ReturnType<typeof userEvent.setup>,
  seite: 'ich' | 'gegenueber' = 'ich',
): Promise<void> {
  const raum = await screen.findByRole('region', { name: eigene.vorgangTitel })
  const knopf = await within(raum).findByRole('button', {
    name: einsetzen(eigene.assistent.zeileIn, {
      seite:
        seite === 'ich' ? eigene.positionVonIch : eigene.positionVonGegenueber,
    }),
  })

  await waitFor(() => {
    expect(knopf).toBeEnabled()
  })
  await hand.click(knopf)
}

// kennungsfeld schlägt im offenen Formular den Kasten „Kennung von Hand" auf
// und gibt sein Feld — P10 (#2378): seitdem steht die Kennung nicht mehr im
// Hauptweg, und Läufe, die einen Druck von Hand nennen, gehen durch den
// Kasten wie ein Mensch.
export async function kennungsfeld(
  hand: { click: (element: Element) => Promise<void> },
  raum: HTMLElement = document.body,
): Promise<HTMLElement> {
  await hand.click(
    within(raum).getByRole('button', { name: eigene.kennungVonHand }),
  )

  return within(raum).getByLabelText(eigene.positionPrinting)
}
