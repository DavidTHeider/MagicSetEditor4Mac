// **Die Zugangsdaten fremder Dienste** — Werkstück 8 von Phase 4.8
// (#1321),
// entschieden in [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).
//
// **Eine eigene Datei**, damit `api.ts` unter der Zeilengrenze aus §2 bleibt.
import type { FehlerDTO } from './formen'

// ZugangsstandDTO sagt, welche Angaben ein Dienst braucht und welche liegen.
//
// **Namen und keine Werte.** Der Schlüssel selbst verlässt den Schlüsselbund
// nie in Richtung Fenster.
export interface ZugangsstandDTO {
  dienst: string

  // angaben sind die Namen der nötigen Angaben — leer, wenn der Dienst keine
  // braucht.
  angaben: string[]

  // hinterlegt sind die, die im Schlüsselbund liegen.
  hinterlegt: string[]

  // vollstaendig sagt, ob alle nötigen Angaben da sind.
  //
  // **Gerechnet wird es in Go**, damit die Rechnung nicht an zwei Stellen
  // steht.
  vollstaendig: boolean
}

export interface ZugaengeAntwort {
  zugaenge: ZugangsstandDTO[]
  fehler: FehlerDTO | null
}

// Zugangsapi sind die drei Bindungen — Werkstück 8 von Phase 4.8.
//
// **`Zugaenge` gibt Namen und keine Werte.** Ein Weg, der den Schlüssel ins
// Fenster zurückgäbe, machte aus dem Schlüsselbund eine Durchreiche: der Wert
// stünde im JSON der Bindung, im Speicher des Browsers und in jedem
// Mitschnitt.
export interface Zugangsapi {
  Zugaenge: () => Promise<ZugaengeAntwort>

  ZugangSetzen: (
    dienst: string,
    angabe: string,
    geheimnis: string,
  ) => Promise<SchreibAntwortZugang>

  ZugangLoeschen: (
    dienst: string,
    angabe: string,
  ) => Promise<SchreibAntwortZugang>
}

// SchreibAntwortZugang ist eine Antwort ohne Nutzlast.
//
// **Eine eigene Angabe und kein Import aus `api.ts`**: die Kante zeigte sonst
// zurück, und zwei Dateien, die einander kennen, sind eine.
export interface SchreibAntwortZugang {
  fehler: FehlerDTO | null
}

// **Das Geheimnis eines Anmeldeanbieters** —
// #1401,
// Befund 3 des Zweitblicks über Phase 4.12.
//
// **Hier und nicht in einer eigenen Datei**, weil es dieselbe Frage ist wie
// oben — was liegt im Schlüsselbund, und wie kommt es hinein — nur für die
// vier Anmeldeanbieter statt für die Kartendienste.

// AnbietergeheimnisDTO sagt, was ein Anbieter braucht und was vorliegt.
//
// **Nie den Wert**, aus demselben Grund wie bei [ZugangsstandDTO].
export interface AnbietergeheimnisDTO {
  anbieter: string

  // pflicht und verboten kommen aus Go und sind dort je Anbieter
  // nachgelesen: GitHub verlangt eines, Microsoft verbietet es.
  pflicht: boolean
  verboten: boolean

  hinterlegt: boolean
}

export interface AnbietergeheimnisseAntwort {
  anbieter: AnbietergeheimnisDTO[]
  fehler: FehlerDTO | null
}

// Anbietergeheimnisapi sind die drei Bindungen.
export interface Anbietergeheimnisapi {
  Anbietergeheimnisse: () => Promise<AnbietergeheimnisseAntwort>

  AnbietergeheimnisSetzen: (
    anbieter: string,
    geheimnis: string,
  ) => Promise<SchreibAntwortZugang>

  AnbietergeheimnisLoeschen: (anbieter: string) => Promise<SchreibAntwortZugang>
}
