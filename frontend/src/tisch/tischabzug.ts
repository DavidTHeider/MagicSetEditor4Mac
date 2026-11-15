// **Der Weg vom Ausschnitt zum Bild** — PLAY-05.AK-03 und AK-04
// (#1713).
//
// # Warum das hier steht und nicht im Bauteil
//
// **Weil es drei Fehlerwege hat und keinen Baum braucht** (§3): kein Maler,
// kein Ausschnitt, keine gültigen Maße. Ein Bauteil, das sie in einem
// Rückruf zusammenrechnete, wäre nur über ein gezeichnetes Fenster zu
// prüfen — und ein gezeichnetes Fenster gibt es unter jsdom nicht.
//
// # Und warum die Maße **vor** der Grafikkarte geprüft werden
//
// **Weil ein zu großer Abzug den Kontext kostet.** AK-04 nennt es: „große
// Flächen und DPI-Grenzen sind begrenzt und fehlerrobust". Ein Ruf, der erst
// hinausgeht und dann scheitert, nimmt den ganzen Tisch mit
// (#1460
// hat das einmal gekostet).
import { ausschnittFuer, bildmasse } from './tischausschnitt'

import type { Ausschnittwahl, Massfehler } from './tischausschnitt'
import type { Kamera } from './kamera'
import type { Tischmalerei } from './maler'
import type { Tischplan } from './tischplan'

// Zoll ist, wie viele Millimeter ein Zoll hat.
const Zoll = 25.4

// Abzugsfehler sagt, warum es kein Bild gibt.
//
// **`keinMaler` und `keinAusschnitt` neben den Maßfehlern**: drei Gründe,
// drei Sätze — einer für alle nähme dem Lesenden die Auskunft, ob ein
// zweiter Versuch lohnt (dieselbe Überlegung wie bei [Malhindernis]).
//
// **`zieht` kommt von der Fläche und nicht von hier**: solange eine Karte
// gefasst ist, zeigt der Szenenbaum den gezogenen Plan, und ein Abzug davon
// hielte einen Stand fest, den niemand bestätigt hat (Befund 3 des
// Zweitblicks über Phase 6.9, PLAY-05.AK-04).
//
// **`aufbau` heißt: der Maler kommt noch** — der Tisch wird bei jedem
// Sichtwechsel neu aufgebaut, und bis dahin gibt es keinen (Nachlauf über
// Phase 6.9).
export type Abzugsfehler =
  Massfehler | 'keinMaler' | 'keinAusschnitt' | 'zieht' | 'aufbau'

// Abzug ist ein fertiges Bild samt seinen Maßen.
export interface Abzug {
  // bild ist ein PNG als Data-URL.
  bild: string

  breite: number
  hoehe: number
}

// abzugMachen gibt den Abzug — oder den Grund, warum es keinen gibt.
export async function abzugMachen(
  malerei: Tischmalerei,
  plan: Tischplan,
  kamera: Kamera,
  flaeche: { breite: number; hoehe: number },
  wahl: Ausschnittwahl,
  dpi: number,
): Promise<Abzug | { fehler: Abzugsfehler }> {
  if (malerei.art === 'ohne') {
    return { fehler: 'keinMaler' }
  }

  const ausschnitt = ausschnittFuer(plan, kamera, flaeche, wahl)
  if (ausschnitt === undefined) {
    return { fehler: 'keinAusschnitt' }
  }

  const masse = bildmasse(ausschnitt, dpi)
  if ('fehler' in masse) {
    return { fehler: masse.fehler }
  }

  const bild = await malerei.maler.abzug(ausschnitt, dpi / Zoll)

  return { bild, breite: masse.breite, hoehe: masse.hoehe }
}
