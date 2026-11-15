// **Die Proben der Deckseite** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `decks.test.tsx` unter der Zeilengrenze aus §2
// bleibt — dieselbe Aufteilung wie bei `kartenproben.ts`.
import { keineHerkunft } from '../apiherkunft'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { deckmenue } from '../test/deckhandlung'
import { kataloge } from '../sprache/kataloge'

import type {
  BauartDTO,
  DeckDTO,
  DeckplatzDTO,
  DeckurteilDTO,
  FormatDTO,
} from '../apidecks'

export function deckprobe(teil: Partial<DeckDTO> = {}): DeckDTO {
  return {
    id: '1',
    fassung: '1',
    name: 'Aggro',
    spiel: 'magic',
    notiz: '',
    formatID: '',
    plaetze: [],
    hauptdeck: 0,
    sideboard: 0,
    herkunft: keineHerkunft,
    ...teil,
  }
}

export function platzprobe(teil: Partial<DeckplatzDTO> = {}): DeckplatzDTO {
  return {
    id: '3',
    karteID: '',
    printing: 'a1b2c3d4-0000-4000-8000-00000000000a',
    druckID: '',
    vorlaeufig: false,
    typzeile: '',
    manawert: 0,
    menge: 4,
    bereich: 'hauptdeck',
    bezeichnung: '',
    manakosten: '',
    ...teil,
  }
}

export function formatprobe(teil: Partial<FormatDTO> = {}): FormatDTO {
  return {
    id: '5',
    fassung: '1',
    name: 'Konstruiert',
    spiel: 'magic',
    mindestgroesse: 60,
    hoechstgroesse: 0,
    sideboardgroesse: 15,
    sideboardOffen: false,
    hoechstzahlGleich: 4,
    scryfallname: 'standard',
    ...teil,
  }
}

export function bauartprobe(): BauartDTO {
  return { bauart: 'konstruiert', format: formatprobe({ id: '', fassung: '' }) }
}

export function urteilprobe(teil: Partial<DeckurteilDTO> = {}): DeckurteilDTO {
  return {
    formatname: 'Konstruiert',
    hauptdeck: 4,
    sideboard: 0,
    urteile: [],
    groessen: [],
    zulaessig: true,
    ungeprueft: 0,
    ...teil,
  }
}

// zumDeckgebiet wählt eine Lasche des offenen Decks.
//
// **Seit Werkstück I von Phase 4.10 hat die Werkbank drei Gebiete**
// (#1325):
// Karten, Auswertung, Prüfung.
export async function zumDeckgebiet(
  gebiet: 'karten' | 'auswertung' | 'pruefung',
): Promise<void> {
  const namen = kataloge.de.seiten.decks

  const name = {
    karten: namen.gebietKarten,
    auswertung: namen.gebietAuswertung,
    pruefung: namen.gebietPruefung,
  }[gebiet]

  await userEvent.click(await screen.findByRole('tab', { name }))
}

// aufschlagen öffnet ein Deck der Liste und wartet, bis seine Handlungen
// dastehen.
//
// **Seit UXR-11** (#1919)
// stehen Umbenennen, Duplizieren, Teilen und Löschen nur am **offenen** Deck
// — vorher an jeder Zeile. Wer eines davon prüft, schlägt das Deck zuerst
// auf, wie ein Mensch es jetzt tut. **Seit AP-6** (#2363) stehen sie hinter
// dem Menü „…“; auf das wird gewartet.
export async function aufschlagen(name = 'Aggro'): Promise<void> {
  const namen = kataloge.de.seiten.decks

  await userEvent.click(
    await screen.findByRole('button', { name: `${namen.waehlen}: ${name}` }),
  )
  await deckmenue(name)
}
