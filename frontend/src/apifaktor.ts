// **Der zweite Faktor** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, weil `apinutzer.ts` mit ihm über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, **wer** hier arbeitet, hier,
// **wie er sich ein zweites Mal ausweist**.
import type { Umschlag } from './umschlag'

// FaktorDTO ist ein zweiter Faktor.
export interface FaktorDTO {
  // art ist `totp` oder `biometrie`.
  art: string

  // bezeichnung sagt, welches Gerät gemeint ist — leer erlaubt.
  bezeichnung: string
}

// FaktorenAntwort trägt die Faktoren eines Nutzers **samt** der Auskunft, ob
// dieser Rechner Biometrie kann.
//
// **Beides in einer Antwort**, weil die Seite beides nebeneinander zeigt: was
// eingerichtet ist und was sich einrichten ließe (§7, Regel 2).
export interface FaktorenAntwort extends Umschlag {
  faktoren: readonly FaktorDTO[]
  biometrie: boolean
  biometrieart: string
  biometriegrund: string
}

// Einrichtungsantwort trägt, was **einmal** zu sehen ist.
//
// **Danach steht nichts davon mehr in einer Antwort dieser Anwendung.**
export interface Einrichtungsantwort extends Umschlag {
  // adresse ist die `otpauth://`-Adresse für den Scanner.
  adresse: string

  // geheimnis ist dieselbe Angabe zum Abtippen.
  geheimnis: string

  // codes sind die Wiederherstellungscodes im Klartext.
  codes: readonly string[]
}

// Faktorapi sind die fünf Bindungen dazu.
export interface Faktorapi {
  // Faktoren gibt die zweiten Faktoren eines Nutzers.
  Faktoren: (nutzerID: string) => Promise<FaktorenAntwort>

  // FaktorEinrichten zieht ein TOTP-Geheimnis.
  //
  // **Der Faktor gilt danach noch nicht** — erst `FaktorBestaetigen`. Wer ihn
  // vorher einträgt und den Scan vergisst, sperrt sich selbst aus.
  FaktorEinrichten: (nutzerID: string) => Promise<Einrichtungsantwort>

  // FaktorBestaetigen trägt den TOTP-Faktor ein, wenn ein Code aus ihm stimmt.
  FaktorBestaetigen: (nutzerID: string, code: string) => Promise<Umschlag>

  // BiometrieEinrichten trägt Touch ID oder Face ID ein.
  //
  // **Es verlangt einmal den Finger** — sonst trüge jemand einen Faktor ein,
  // den er nicht bedienen kann.
  BiometrieEinrichten: (nutzerID: string) => Promise<Umschlag>

  // FaktorEntfernen nimmt einen zweiten Faktor heraus.
  FaktorEntfernen: (nutzerID: string, art: string) => Promise<Umschlag>
}
