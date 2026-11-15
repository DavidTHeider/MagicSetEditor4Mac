// **Die Vorrichtung des KI-Vorschlags** — herausgezogen für
// #1539.
//
// # Warum eine eigene Datei
//
// **Weil `kivorschlag.test.tsx` mit der Freigabe über die Zeilengrenze aus §2
// ging.** Dieselbe Bewegung wie bei `test/kibild.tsx`: die Vorrichtung steht
// an einer Stelle, die Fragen in mehreren.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { KIVorschlag } from '../bausteine/KIVorschlag'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { KIVorschlagAntwort } from '../antworten'

export const vorschlagstexte = kataloge.de.seiten.einstellungen

// einVorschlag baut eine Antwort.
export function einVorschlag(teil = {}): KIVorschlagAntwort {
  return {
    kennung: 'ki-4-1',
    anbieter: 'ollama',
    modell: 'gemma4:8b',
    felder: [{ feld: 'name', werte: ['Blitzschlag'], bisher: 'Flammenball' }],
    befunde: [],
    ablauf: '2026-09-12T12:00:00Z',
    uebernommen: 0,
    belege: [],
    korpusstaende: [],
    einschraenkungen: [],
    weg: 'gewaehlt',
    ausgang: '',
    rueckfaelle: null,
    fehler: null,
    ...teil,
  }
}

// einKatalog baut die Modellauskunft, aus der die Wahl entsteht.
//
// **Seit A8** (#1447)
// liest dieses Bauteil den Katalog: Anbieter und Modell werden **gewählt**,
// nicht getippt. Ohne Beleg stünde hier eine leere Wahl.
export function einKatalog(
  ...modelle: { anbieter: string; alias: string; ort?: string }[]
) {
  return {
    nurLokal: false,
    vorrang: [],
    vorrangbefunde: [],
    einrichtung: 'bereit',
    anbieter: [],
    cloud: { stand: 'ungewiss', beleg: 'keine Serverdatei' },
    befunde: [],
    modelle: modelle.map((modell) => ({
      ...modell,
      laufzeit: 'ollama 0.12.3',
      ort: modell.ort ?? 'lokal',
      faehigkeiten: [],
      herkunft: [],
      bindung: `${modell.anbieter}/${modell.alias}`,
      kontextfenster: 0,
      kontextbeleg: '',
    })),
    fehler: null,
  }
}

// aufklappen schlägt den Kasten auf — erst dann wird der Katalog gelesen.
//
// **Zugeklappt geht nichts hinaus** (A8, #1447): `KIModelle` fragt die
// eingeschalteten Anbieter, und bei Anthropic heißt das einen Ruf nach
// draußen. Das Aufschlagen einer Karte darf ihn nicht auslösen.
export async function aufklappen(hand: ReturnType<typeof userEvent.setup>) {
  await hand.click(screen.getByText(vorschlagstexte.kiVorschlagTitel))
}

// zeigeVorschlagskasten rendert das Bauteil und gibt den Spion auf »geaendert«.
export function zeigeVorschlagskasten() {
  const geaendert = vi.fn()

  render(
    <Sprachraum sprache="de">
      <KIVorschlag karteID="4" geaendert={geaendert} />
    </Sprachraum>,
  )

  return geaendert
}

// ausfuellenUndVorschlagen wählt das Modell, tippt den Auftrag und drückt.
//
// **Gewartet wird auf die Möglichkeit, nicht auf die Wahl** (§3). Das
// `select` steht sofort da — mit »Bitte wählen« und sonst nichts —, und ein
// `findByLabelText` darauf ist eine Zusicherung, die schon vor der Antwort
// zutrifft. Gewartet wird deshalb auf den **Eintrag**, den der Katalog
// bringt; erst dann gibt es etwas zu wählen.
export async function ausfuellenUndVorschlagen(wahl = 'ollama/gemma4:8b') {
  const hand = userEvent.setup()

  await aufklappen(hand)

  await screen.findByRole('option', { name: wahl })

  await hand.selectOptions(
    screen.getByLabelText(vorschlagstexte.kiVorschlagAnbieter),
    wahl,
  )
  await hand.type(
    screen.getByLabelText(vorschlagstexte.kiVorschlagAuftrag),
    'Schreibe.',
  )
  await hand.click(
    screen.getByRole('button', { name: vorschlagstexte.kiVorschlagen }),
  )

  return hand
}
