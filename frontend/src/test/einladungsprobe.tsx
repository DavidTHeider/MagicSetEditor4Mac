// **Die Vorrichtung der Einladungsläufe** — NET-01
// (#1705).
//
// **Eine eigene Datei**, weil drei Läufe sie brauchen und jeder sonst über
// die Zeilengrenze aus §2 ginge.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

import { Einladungstausch } from '../bausteine/Einladungstausch'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type {
  Einladungsvorschau,
  Einladungsvorschauantwort,
} from '../apikontakteinladung'
import type { Peerkontakteantwort } from '../apipeerkontakt'

export const technik = kataloge.de.seiten.technik
export const eigene = technik.einladung

export const derCode = 'mse4mac-kontakt:1:eyJrZW5udW5nIjoiYSJ9.c2ln'

// eineVorschau baut, was eine Prüfung sagt.
export function eineVorschau(
  anders: Partial<Einladungsvorschau> = {},
): Einladungsvorschau {
  return {
    kennung: '12D3KooWAnna',
    name: 'Anna',
    zweck: 'handel',
    adresse: '192.168.1.42:41234',
    ablauf: '2026-10-01T12:00:00Z',
    fingerabdruck: '9f4b-8c8e-9025',
    vergleichscode: '49311 67386 01529 95503 63193 56021',
    lage: 'unbekannt',
    vorgaenger: '',
    angenommen: false,
    einloesbar: false,
    ...anders,
  }
}

// alsAntwort packt eine Vorschau in ihren Umschlag.
export function alsAntwort(
  vorschau: Einladungsvorschau,
): Promise<Einladungsvorschauantwort> {
  return Promise.resolve({ vorschau, fehler: null })
}

// eineLeereListe ist die Antwort einer geglückten Annahme.
export function eineLeereListe(): Promise<Peerkontakteantwort> {
  return Promise.resolve({ kontakte: [], zwecke: [], fehler: null })
}

// einFehler ist eine Antwort, die scheitert.
export function einFehler(meldung: string) {
  return {
    fehler: { code: 'validierung' as const, schluessel: '', meldung },
  }
}

// stelleAuf zeigt den Tausch und gibt die Hand und das Nachladen.
export function stelleAuf() {
  const nachAufnahme = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Einladungstausch
        zwecke={['paarung', 'teilen', 'handel', 'partie']}
        nachAufnahme={nachAufnahme}
      />
    </Sprachraum>,
  )

  return { hand: userEvent.setup(), nachAufnahme }
}

// pruefe öffnet „Code eingeben", fügt den Code ein und prüft.
export async function pruefe(
  hand: ReturnType<typeof userEvent.setup>,
  code = derCode,
) {
  await hand.click(screen.getByRole('button', { name: eigene.eingeben }))
  await hand.click(screen.getByLabelText(eigene.eingabe))
  await hand.paste(code)
  await hand.click(screen.getByRole('button', { name: eigene.pruefen }))
}
