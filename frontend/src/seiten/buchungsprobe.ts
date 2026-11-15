// **Die Vorrichtung der Buchungsläufe** — Werkstück T6 von Phase 5.6
// (#1471),
// zusammengezogen mit UXR-15
// (#1923).
//
// **Eine eigene Datei**, weil drei Läufe sie brauchen und `handelsbuchung`
// sonst über die Zeilengrenze aus §2 ginge — und weil die Buchung seit
// UXR-15 am **Stand des Handels** hängt: welche Knöpfe stehen, sagt der
// Vorgang, den der Raum kennt (`HandelVorgaenge`), und nicht ein Feld.
import { screen, waitFor, within } from '@testing-library/react'
import { expect } from 'vitest'

import { einKontakt } from './peerkontaktprobe'
import { einVorgang } from './handelslisteprobe'
import { oeffneHandel, zuPhase, type zeigeHandel } from './tauschprobe'
import { kataloge } from '../sprache/kataloge'

import type { Handelsvorgang } from '../apihandelsleitung'
import type { belegeApi } from '../test/api'
import { mitStandnachlauf } from '../test/standnachlauf'
import type {
  Journalantwort,
  Verbuchungsantwort,
} from '../apihandelsverbuchung'

const eigene = kataloge.de.seiten.handeln

// bestaetigt ist ein Kontakt, an dem der Kasten erscheint.
export function bestaetigt() {
  return einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })
}

// einStand baut eine Antwort der Verbuchung.
export function einStand(
  anders: Partial<Verbuchungsantwort> = {},
): Verbuchungsantwort {
  return {
    vorgang: 'v-1',
    zustand: 'lokalVerbucht',
    beleg: {
      vorgang: 'v-1',
      abdruck: 'abdruck-1',
      unterschrift: 'deadbeef',
      beidseitig: false,
      gegenzeichner: '',
      gegengezeichnet: '',
      gebucht: '2026-09-18T12:00:00Z',
      befund: null,
    },
    hatBeleg: true,
    fehler: null,
    ...anders,
  }
}

// einJournal baut eine Antwort des Journals.
export function einJournal(
  zeilen: Journalantwort['zeilen'] = [],
): Journalantwort {
  return { vorgang: 'v-1', zeilen, fehler: null }
}

// eineBewegung baut eine Zeile des Journals.
export function eineBewegung(
  anders: Partial<Journalantwort['zeilen'][number]> = {},
): Journalantwort['zeilen'][number] {
  return {
    id: 1,
    art: 'abgang',
    anlass: 'tausch',
    vorgang: 'v-1',
    bezug: 0,
    grund: '',
    printing: 'bolt-m10',
    bezeichnung: 'Lightning Bolt · M10 #146',
    menge: 2,
    zustand: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    exemplarId: 7,
    lagerort: 'Kiste 4',
    notiz: '',
    gebucht: '2026-09-18T12:00:00Z',
    ...anders,
  }
}

// imAbschluss ist der Handel `v-1` in einem Zustand der Phase „Verbuchen" —
// **vorgabemäßig nach der Übergabe**, wo als Nächstes verbucht wird.
export function imAbschluss(anders: Partial<Handelsvorgang> = {}) {
  return einVorgang({
    vorgang: 'v-1',
    zustand: 'uebergabeOffen',
    phase: 'verbuchen',
    lage: 'dran',
    naechster: 'verbuchen',
    moegliche: ['verbuchen', 'klaeren'],
    ...anders,
  })
}

// dieVorgaenge gibt dem Raum die Handel mit dem Gegenüber — jeder Ruf den
// nächsten Stand, der letzte bleibt.
//
// **Mindestens einer** — ohne Stand gäbe es keinen Raum, und eine Verzweigung
// für den leeren Fall wäre eine, die kein Lauf betritt (§3).
export function dieVorgaenge(
  erster: Handelsvorgang,
  ...weitere: Handelsvorgang[]
) {
  const folge = [...weitere]
  let stand = erster

  return () => {
    const antwort = Promise.resolve({ vorgaenge: [stand], fehler: null })
    stand = folge.shift() ?? stand

    return antwort
  }
}

// eineFreigabe gibt die Zustimmung, die ein Mensch am Fenster erteilt (T10,
// #1475).
//
// **Sie steht in jeder Belegung, die verbucht oder bestätigt**: ohne sie
// scheitert die Fassade mit dem Namen der Methode (§3), und dieser Lauf misst
// die Verbuchung und nicht den Riegel davor.
export const eineFreigabe = () =>
  Promise.resolve({ kennung: 'f-1', laeuftab: '', fehler: null })

// buchung schlägt den Handel `v-1` auf, geht nach „Verbuchen" und gibt den
// Kasten der Buchung.
export async function buchung(
  hand: Awaited<ReturnType<typeof zeigeHandel>>,
): Promise<HTMLElement> {
  await oeffneHandel(hand, 'v-1')

  // **Die Buchung steht in „Verbuchen"** (TRADE-02, #1716).
  await zuPhase(hand, 'verbuchen')

  return screen.findByRole('region', { name: eigene.buchungTitel })
}

// trageEin drückt den genannten Knopf im Kasten der Buchung — **erst, wenn
// er dasteht**: welche stehen, sagt der Stand, den der Raum geladen hat.
export async function trageEin(
  hand: Awaited<ReturnType<typeof zeigeHandel>>,
  knopf: string,
) {
  await hand.click(
    await within(await buchung(hand)).findByRole('button', { name: knopf }),
  )
}

// buchungsstand ist die Zeile, die den Stand der Buchung ansagt — **im
// Kasten** und nicht irgendwo: auch die Lage über dem Assistenten sagt an.
export async function buchungsstand(): Promise<HTMLElement> {
  return within(
    await screen.findByRole('region', { name: eigene.buchungTitel }),
  ).findByRole('status')
}

// standSagt wartet, bis die Zeile des Kastens den Text ansagt — **auf die
// Wirkung** (§3): seit der Hüllenprobe `nat04handel` (#1672) liest der
// Kasten den Beleg nach jedem Schritt neu, und die Zeile von vorher steht
// bis dahin da.
export async function standSagt(text: string): Promise<void> {
  await waitFor(async () => {
    expect(await buchungsstand()).toHaveTextContent(text)
  })
}

type Belegung = Parameters<typeof belegeApi>[0]

// mitBelegnachlauf gibt einer Belegung den lesenden Weg dazu:
// `Handelsbeleg` sagt, was der letzte **geglückte** Schritt hinterließ — wie
// der Bestand. Vor dem ersten steht kein Beleg da.
export function mitBelegnachlauf(
  belegung: Belegung,
  anfang: Verbuchungsantwort = einStand({
    zustand: 'uebergabeOffen',
    hatBeleg: false,
  }),
): Belegung {
  let zuletzt = anfang

  // **Nur, was belegt ist, wird umhüllt** — ein unbelegter Schritt scheitert
  // weiter mit dem Namen seiner Methode (§3).
  const merke =
    <A extends unknown[]>(
      ruf: (...argumente: A) => Promise<Verbuchungsantwort>,
    ) =>
    async (...argumente: A) => {
      const antwort = await ruf(...argumente)

      if (antwort.fehler === null) {
        zuletzt = antwort
      }

      return antwort
    }

  const { Handelsverbuchung, Handelsabgleich, Handelsbestaetigung } = belegung

  // **Ein Zug legt in den Ausgang und überträgt** (P13, Vermerk 0200) —
  // wer bucht, legt die Übergabe vorher hinein. Wo ein Lauf den Ausgang
  // selbst belegt, gilt seine Belegung.
  return mitStandnachlauf({
    HandelSchicken: () => Promise.resolve({ nachrichtId: 'n-1', fehler: null }),
    HandelZustellen: (kennung: string) =>
      Promise.resolve({ gegenueber: kennung, getragen: 1, fehler: null }),
    ...belegung,
    ...(Handelsverbuchung && { Handelsverbuchung: merke(Handelsverbuchung) }),
    ...(Handelsabgleich && { Handelsabgleich: merke(Handelsabgleich) }),
    ...(Handelsbestaetigung && {
      Handelsbestaetigung: merke(Handelsbestaetigung),
    }),
    Handelsbeleg: () => Promise.resolve(zuletzt),
  })
}
