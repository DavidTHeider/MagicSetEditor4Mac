// **Die Werkbank für die Formatzeile** — sie stellt die Deckseite mit **einem**
// Format hin und wartet, bis es dasteht.
//
// **Eine eigene Datei**, weil zwei Testdateien sie brauchen und beide unter
// der Zeilengrenze aus §2 bleiben sollen (Werkstück U3 von Phase 5.1,
// #1414):
// `formatzeile` fragt, was die Zeile über **zwei Handgriffe** tut,
// `formatfelder`, was sie über eine **Eingabe** urteilt. Dieselbe Aufteilung
// wie bei `turnierwerkbank.tsx`.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import type { FormatAntwort, FormatDTO } from '../apidecks'

const texte = kataloge.de.seiten.decks

// baueFenster stellt die Deckseite mit **einem** Format hin — und gibt den
// Spion auf `FormatSetzen` zurück.
//
// **Die Liste kommt aus dem Stand, den das Fenster zuletzt geschrieben hat**,
// nicht aus einer festen Antwort. Eine Vorrichtung, die immer dasselbe Format
// zurückgibt, wäre großzügiger als der Betrieb (§3): sie könnte gar nicht
// zeigen, dass eine Fassung stehen bleibt.
export async function baueFenster(anfang: FormatDTO) {
  let stand = anfang

  const FormatSetzen = vi.fn((dto: FormatDTO): Promise<FormatAntwort> => {
    if (dto.fassung !== stand.fassung) {
      return Promise.resolve({
        format: null,
        fehler: {
          code: 'konflikt',
          schluessel: 'fehler.konflikt',
          meldung: 'Die Fassung stimmt nicht.',
        },
      })
    }

    stand = { ...dto, fassung: String(Number(dto.fassung) + 1) }

    return Promise.resolve({ format: stand, fehler: null })
  })

  belegeApi({
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [stand], fehler: null }),
    Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
    FormatSetzen,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )

  // **Die Formate gehen in einer Schublade auf** — seit AP-6 (#2363); bis
  // dahin lagen sie hinter einer Lasche (Werkstück I von Phase 4.10, #1325).
  await userEvent.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.decks.formateKnopf,
    }),
  )

  return FormatSetzen
}

// warteAufDieFormatzeile wartet, bis das Format **da** ist — nicht bis der
// Bereich da ist.
//
// # Warum das ein Unterschied ist
//
// `findByRole('region', { name: texte.formate })` trifft auch im **leeren**
// Zustand zu: der Bereich trägt seinen Namen dort ebenso, und darunter steht
// „Noch kein Format." Die Zusicherung galt damit, bevor die Antwort da war
// (§3), und der Griff danach ging in einen Baum, den es noch nicht gab.
//
// **Gemessen am 2026-09-02**: acht volle Frontendläufe, einer davon rot, und
// zwar mit
//
//	Unable to find a label with the text of: Sideboard ohne Grenze
//	<section aria-label="Formate">…Noch kein Format.…
//
// Nachgestellt hat es sich nicht durch Warten — `waitFor` wiederholt die
// Zusicherung und nicht den Griff —, sondern erst dadurch, dass die Läufe
// unter Last gezählt wurden.
//
// **Der Speicherknopf trägt den Formatnamen** und steht deshalb erst da, wenn
// die Antwort verarbeitet ist. Das ist das Abschlusssignal, das §3 verlangt.
// bearbeite klappt das Formular eines Formats auf.
//
// **Seit UXR-11** (#1919)
// steht ein Format als Zeile da und sein Formular erst auf Wunsch — wer es
// prüft, klappt es auf, wie ein Mensch es jetzt tut.
export async function bearbeite(name = 'Konstruiert'): Promise<void> {
  const bearbeiten = kataloge.de.bausteine.aktionskatalog.aktionen.bearbeiten

  await userEvent.click(
    await screen.findByRole('button', { name: `${bearbeiten.kurz}: ${name}` }),
  )
}

export async function warteAufDieFormatzeile(): Promise<HTMLElement> {
  await bearbeite()
  await screen.findByRole('button', {
    name: `${texte.formatSichern}: Konstruiert`,
  })

  // **Und dann noch die übrigen Antworten.** Die Seite fragt drei Bindungen;
  // die letzte, die eintrifft, zeichnet die Zeile neu. Ein Knoten, den man
  // vorher greift, ist danach abgehängt — `userEvent.clear` sagt dazu „The
  // element to be cleared could not be focused."
  await antwortAbwarten()

  return screen.getByRole('region', { name: texte.formate })
}
