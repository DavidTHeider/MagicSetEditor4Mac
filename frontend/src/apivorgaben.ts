// **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// **Nicht dasselbe wie eine Einstellung.** Jene hält einen **aufzählbaren**
// Wert, dessen Liste in `internal/domain` steht; diese fünf zeigen auf Zeilen
// im Bestand — ein Format, ein Spielpaket, ein Stilpaket, Präferenzdecks,
// zugeordnete Lagerorte.
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// Vorgabeangebot ist ein Wert, den ein Profil wählen kann.
//
// **Mit Beschriftung**, weil eine Kennung allein nichts sagt: `7` ist kein
// Format, und daraus ließe sich keine Auswahl bauen, die jemand bedienen kann.
export interface Vorgabeangebot {
  wert: string
  beschriftung: string
}

// VorgabenstandDTO ist eine Art samt ihrem Stand.
export interface VorgabenstandDTO {
  // art ist `format`, `spiel`, `vorlage`, `deck` oder `lagerort`.
  art: string

  // mehrwertig sagt, ob mehrere Werte gelten dürfen.
  mehrwertig: boolean

  // werte sind die gewählten, **in ihrer Reihenfolge** — bei den
  // Präferenzdecks ist sie die Aussage.
  werte: string[]

  // moeglich ist, was dieses Profil wählen kann.
  //
  // **Dieselbe Liste prüft danach**: eine zweite hier ließe sich mit der ersten
  // nicht abgleichen.
  moeglich: Vorgabeangebot[]

  // verwaist sind gewählte Werte, die es nicht mehr gibt.
  //
  // **Sie stehen weiter in `werte`.** Diese Liste ist ein Befund und kein
  // zweiter Stand: wegzuräumen nähme jemandem seine Wahl, während er nicht
  // hinsieht.
  verwaist: string[]
}

// VorgabenAntwort trägt alle fünf Arten.
export interface VorgabenAntwort extends Umschlag {
  vorgaben: VorgabenstandDTO[]

  // nachladefehler heißt: **gesetzt ist**, nur der Stand danach war nicht zu
  // lesen (§7, Regel 3; #1970). Die Seite liest ohnehin selbst nach.
  nachladefehler?: FehlerDTO
}

// VorgabeangabenDTO ist, was das Fenster an einer Art setzen darf.
export interface VorgabeangabenDTO {
  profilID: string
  art: string

  // werte sind die gewählten — **eine leere Liste räumt ab**.
  werte: string[]
}

// Vorgabenapi sind die beiden Bindungen dazu.
export interface Vorgabenapi {
  // Profilvorgaben gibt alle fünf Arten samt ihrem Angebot.
  Profilvorgaben: (profilID: string) => Promise<VorgabenAntwort>

  // ProfilvorgabeSetzen ersetzt die Werte **einer** Art und gibt den ganzen
  // Stand zurück (§7, Regel 2).
  ProfilvorgabeSetzen: (angaben: VorgabeangabenDTO) => Promise<VorgabenAntwort>
}
