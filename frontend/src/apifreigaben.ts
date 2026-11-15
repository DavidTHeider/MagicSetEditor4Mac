// **Rechte, Rollen und Sichtbarkeit über Nutzer hinweg** — Werkstück 11 von
// Phase 4.12
// (#1351).
//
// # Drei Rollen, und die erste heißt weniger, als das Wort verspricht
//
// **`leser` sieht, dass es das Profil gibt** — nicht, was darin steht. Der
// Grund steht in
// [0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md).
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// FreigabeDTO ist die Rolle eines Nutzers an einem Profil.
export interface FreigabeDTO {
  profilID: string
  nutzerID: string

  // name ist der Anzeigename — eine Kennung allein sagt nichts.
  name: string

  // rolle ist `leser`, `mitarbeiter` oder `verwalter`.
  rolle: string

  angelegt: string
}

// FreigabenAntwort trägt die Freigaben eines Profils samt der eigenen Rolle.
export interface FreigabenAntwort extends Umschlag {
  freigaben: FreigabeDTO[]

  // sichtbarkeit ist `privat` oder `bestand`.
  sichtbarkeit: string

  // eigeneRolle sagt, was der arbeitende Mensch darf — leer heißt: nichts.
  //
  // **Sie kommt in derselben Antwort**, weil die Seite daran entscheidet, was
  // sie überhaupt anbietet (§7, Regel 2).
  eigeneRolle: string

  // nachladefehler heißt: **die Änderung gilt**, nur die Liste danach war
  // nicht zu lesen — etwa nach dem eigenen Herabstufen (§7, Regel 3; #1970).
  // Er steht neben `fehler`; die Seite liest danach selbst neu.
  nachladefehler?: FehlerDTO
}

// FreigabeangabenDTO ist, was das Fenster an einer Freigabe setzen darf.
export interface FreigabeangabenDTO {
  profilID: string
  nutzerID: string
  rolle: string
}

// Freigabenapi sind die drei Bindungen dazu.
export interface Freigabenapi {
  // Freigaben gibt sie samt der eigenen Rolle.
  Freigaben: (profilID: string) => Promise<FreigabenAntwort>

  // FreigabeSetzen vergibt eine Rolle — **nur ein Verwalter darf das**.
  FreigabeSetzen: (angaben: FreigabeangabenDTO) => Promise<FreigabenAntwort>

  // FreigabeLoeschen zieht eine Freigabe zurück.
  FreigabeLoeschen: (
    profilID: string,
    nutzerID: string,
  ) => Promise<FreigabenAntwort>
}
