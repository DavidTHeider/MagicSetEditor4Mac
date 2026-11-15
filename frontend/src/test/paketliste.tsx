// **Die Vorrichtungen der Paketliste** — geteilt von `paketliste.test.tsx`
// und `paketlistenplan.test.tsx`, seit die Liste mit dem Plan (DOM-01,
// #1696) über die Zeilengrenze aus §2 ging.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { PaketlisteMitBefunden } from './abhaengigkeiten'

// **Die Anlage steht auf derselben Seite** (#1350) und liest beim Aufbau,
// welche Arten sich anlegen lassen. Ohne diese Belegung scheiterte die
// Vorrichtung mit „Die Methode PaketArten ist nicht belegt" — und die
// Ablehnung stünde als unbehandelt im Lauf.
export const anlegbar = () =>
  Promise.resolve([
    { art: 'mse-game', brauchtSpiel: false },
    { art: 'mse-style', brauchtSpiel: true },
  ])

export const texte = kataloge.de

// katalog gibt eine Antwort mit einem Paket.
export function katalog() {
  return {
    pakete: [
      {
        art: 'mse-style',
        name: 'probe-schlicht',
        anzeigename: 'probe-schlicht',
        version: '2026-08-15',
        pfad: '/pakete/probe-schlicht.mse-style',
        herkunft: 'aus der Saat',
        pruefsumme: 'sha256:abc',
        braucht: ['probe.mse-game 2026-08-15'],
        gruppe: '',
        rang: 1000000,
        artrang: 30,
        herkunftsart: 'mitgeliefert' as const,
        hauptdatei: 'style',
        geaendert: false,
        gebrauchtVon: [],
        benutzt: 0,
      },
    ],
    befunde: [],
  }
}

export function zeige() {
  render(
    <Sprachraum sprache="de">
      <PaketlisteMitBefunden />
    </Sprachraum>,
  )
}

// plan gibt die Antwort auf einen Plan.
export function plan(vorgang: 'entfernen' | 'einbinden', quelle = '') {
  return {
    plan: {
      vorgang,
      paket: 'probe-schlicht.mse-style',
      quelle,
      von: '2026-08-15',
      zu: '',
      geaendert: false,
      moeglich: true,
      sets: ['Erstes'],
      konflikte: [],
    },
    abgebrochen: false,
  }
}

// entferne geht den Weg, den ein Mensch geht: das Menü der Zeile, der
// Eintrag, der Plan im Inspector, das Ja (#1413, DOM-01).
//
// **Erst die Zeile wählen**: seit UXR-17 (#1925) steht das Menü nur an der
// aktiven Zeile, und die Maus macht eine Zeile über die Wahl aktiv.
export async function entferne(name: string): Promise<void> {
  await userEvent.click(screen.getByRole('row', { name }))
  await userEvent.click(
    screen.getByRole('button', { name: `Handlungen — ${name}` }),
  )
  await userEvent.click(
    screen.getByRole('menuitem', { name: texte.seiten.pakete.plan.loeschen }),
  )
  await userEvent.click(await screen.findByRole('button', { name: 'Löschen' }))
}

export const fehler = (meldung: string) => ({
  code: 'unbekannt' as const,
  schluessel: 'fehler.unbekannt',
  meldung,
})
