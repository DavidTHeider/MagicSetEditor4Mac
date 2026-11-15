// **Der Arbeitskontext** — Werkstück O1 von Phase 5.2
// (#1425).
//
// **Eine eigene Datei**, weil `apinutzer.ts` mit ihm über 400 Zeilen ginge
// (§2) — derselbe Schnitt wie bei `apitoken.ts` und `apifaktor.ts`.
import type { Umschlag } from './umschlag'

// ArbeitskontextAntwort sagt, wer gerade in welchem Profil als was arbeitet —
// Werkstück O1 von Phase 5.2
// (#1425).
//
// **Ein Zug, ein Stand**: wer `WerArbeitet`, `GezeigtesProfil` und die Rolle
// nacheinander holte, konnte dazwischen einen Wechsel erwischen. Die
// `revision` steigt mit jedem Wechsel und jeder Freigabe — zwei Antworten mit
// derselben Zahl beschreiben denselben Stand.
export interface ArbeitskontextAntwort extends Umschlag {
  nutzerID: string
  profilID: string

  // eignerID ist der Mensch, dem das Profil gehört — derselbe wie nutzerID,
  // solange niemand in ein fremdes gewechselt ist.
  eignerID: string

  // rolle ist die wirksame Rolle in diesem Profil: `verwalter` für den
  // Eigner, sonst die Freigabe.
  rolle: string

  // nurLesend heißt: hier geht nichts, was schreibt — weil die Rolle `leser`
  // ist oder ein Token nur Lesen erlaubt.
  nurLesend: boolean

  // modus ist die Fläche, über die dieser Ruf kam — am Fenster `fenster`.
  modus: string
  revision: number
}
