// **Die Bindungen für Nutzer und Profile** — Phase 4.12
// (#1351).
import type { ArbeitskontextAntwort } from './apikontext'
import type { Umschlag } from './umschlag'

// SchreibAntwort ist ein Umschlag ohne Nutzlast — `api.ts` führt sie.
type SchreibAntwort = Umschlag

// NutzerDTO ist ein Nutzer samt seinen Profilen.
export interface NutzerDTO {
  id: string
  name: string

  // bild ist der Dateiname im Bildverzeichnis — leer heißt: keines.
  bild: string

  // **Woran die Seite die Kontoart erkennt** — Werkstück 6 (#1351).
  //
  // **Das Kennwort steht hier nicht**, und das ist keine Auslassung: es geht
  // hinein und nie zurück. Wer es vergisst, setzt ein neues.
  email: string
  telefon: string
  anonym: boolean

  // **Die Angaben aus Werkstück 8** (#1351) — alle freiwillig.
  anrede: string
  namenszusatz: string
  vorname: string
  weiterenamen: string
  nachname: string
  slogan: string
  kurzebio: string
  langebio: string

  // fassung ist die zuletzt gelesene — sie geht beim Ändern zurück.
  fassung: string

  profile: ProfilDTO[]
}

// NutzerangabenDTO ist, was das Fenster an einem Nutzer ändern darf.
//
// **Ein Verbund und keine elf Argumente**: die Felder gehören zusammen, und
// wer sie einzeln reicht, vertauscht irgendwann zwei.
//
// **Was hier nicht steht, ändert man anderswo**: Adresse, Rufnummer und
// Kennwort gehören zur Anmeldung, die zweiten Faktoren zu ihrem eigenen Weg.
export interface NutzerangabenDTO {
  id: string
  name: string
  bild: string
  anrede: string
  namenszusatz: string
  vorname: string
  weiterenamen: string
  nachname: string
  slogan: string
  kurzebio: string
  langebio: string

  // fassung ist die, die der Rufer gelesen hat.
  //
  // **Ohne sie ginge eine Änderung verloren** (§7, Regel 2 über die Flächen
  // hinweg): zwei gleichzeitige Schreibvorgänge läsen dieselbe Fassung, und
  // die zweite überschriebe die erste.
  fassung: string
}

// Nutzerbildantwort trägt die Profilbilder als Daten-URI, je Nutzerkennung.
//
// **Eine Tabelle und keine Liste**: wer kein Bild hat, steht gar nicht darin
// — der Normalfall, und ein leerer Eintrag wäre die teurere Art, das zu sagen.
export interface Nutzerbildantwort extends Umschlag {
  bilder: Record<string, string>

  // badges sind die Abzeichen, je **Profil**kennung — Werkstück 9 (#1351).
  //
  // **In derselben Antwort** (§7, Regel 2): ein Abzeichen liegt **über** dem
  // Gesicht, und zwei Rufe wären zwei Stände.
  badges: Record<string, string>
}

// NutzereinstellungDTO ist eine Einstellung, die dem **Menschen** gehört.
//
// **Neben den Anwendungseinstellungen und nicht in ihnen**: jene gelten für
// das Profil, in dem gearbeitet wird, diese für den Menschen — er nimmt sie
// in jedes seiner Profile mit.
export interface NutzereinstellungDTO {
  schluessel: string
  wert: string

  // werte sind die erlaubten — ohne sie stünden sie hier noch einmal, und
  // die zweite Liste wäre die ungeprüfte.
  werte: string[]
}

// NutzereinstellungenAntwort trägt sie alle — auch das Ungesetzte.
export interface NutzereinstellungenAntwort extends Umschlag {
  einstellungen: NutzereinstellungDTO[]
}

// ProfilDTO ist eine Arbeitsumgebung eines Nutzers.
export interface ProfilDTO {
  id: string
  nutzerID: string
  name: string

  // **Die freiwilligen Angaben** — Werkstück 9 (#1351), Punkte 10 bis 13.
  //
  // **`subslogan` und nicht `slogan`**: er steht **unter** dem des Menschen
  // und ersetzt ihn nicht.
  subslogan: string
  kurzebeschreibung: string
  langebeschreibung: string

  // badge ist der Dateiname des Abzeichens — leer heißt: keines.
  badge: string

  // **Ein technisches Profil und sein Ablauf** — Werkstück 10 (#1351),
  // Punkt 12.
  //
  // **`technisch` heißt: dahinter steht kein Mensch, sondern ein Automat.**
  // Wer einen Dienst automatisch bedienen lässt, will nicht, dass er in
  // **seinem** Profil schreibt — was dort entsteht, entsteht nicht von seiner
  // Hand. Es gehört trotzdem einem Nutzer: irgendwer muss es abschalten
  // können.
  //
  // **`ablauf` ist ein Zeitpunkt und kein Löschauftrag** (RFC 3339, leer heißt
  // „läuft nicht ab"). Nichts verschwindet von selbst.
  technisch: boolean
  ablauf: string

  // fassung ist die zuletzt gelesene — sie geht beim Ändern zurück.
  fassung: string
}

// ProfilangabenDTO ist, was das Fenster an einem Profil ändern darf.
//
// **Ein Verbund und keine sechs Argumente** — dieselbe Bauform und derselbe
// Grund wie bei [NutzerangabenDTO].
//
// **Die Nutzerkennung steht nicht darin**: ein Profil wechselt seinen Nutzer
// nicht, und wer sie mitschickte, könnte es.
export interface ProfilangabenDTO {
  id: string
  name: string
  subslogan: string
  kurzebeschreibung: string
  langebeschreibung: string
  badge: string

  // **Technisch und Ablauf** — Werkstück 10 (#1351). Der Grund steht bei
  // [ProfilDTO] und wird hier nicht wiederholt.
  technisch: boolean
  ablauf: string

  // fassung ist die, die der Rufer gelesen hat.
  fassung: string
}

// NutzerAntwort trägt die Übersicht.
export interface NutzerAntwort extends Umschlag {
  nutzer: NutzerDTO[]
}

// EinNutzerAntwort trägt einen einzelnen Nutzer.
export interface EinNutzerAntwort extends Umschlag {
  nutzer: NutzerDTO | null
}

// AnbieterkontoDTO ist ein Anbieterkonto, wie das Fenster es sieht.
export interface AnbieterkontoDTO {
  anbieter: string

  // anzeigename ist, wie der Anbieter diesen Menschen nennt — leer erlaubt.
  anzeigename: string
}

// AnbieterAntwort trägt die Anbieter, über die man sich anmelden kann.
export interface AnbieterAntwort extends Umschlag {
  anbieter: readonly string[]

  // staende sagt je Anbieter, ob er eingerichtet ist und wogegen sein Weg
  // geprüft ist — ID-05.AK-01 (#1704). **Optional**, damit eine Belegung ohne
  // ihn im Test „nichts bekannt" heißt und nicht einen Absturz.
  staende?: readonly AnmeldeanbieterDTO[]
}

// AnmeldeanbieterDTO ist ein Anmeldeanbieter samt seinem Stand.
export interface AnmeldeanbieterDTO {
  anbieter: string
  eingerichtet: boolean
  geprueft: 'simuliert' | 'realGeprueft'
}

// KontenAntwort trägt die Konten eines Nutzers.
export interface KontenAntwort extends Umschlag {
  konten: readonly AnbieterkontoDTO[]
}

// EinProfilAntwort trägt ein einzelnes Profil.
export interface EinProfilAntwort extends Umschlag {
  profil: ProfilDTO | null
}

export interface Nutzerapi {
  // Nutzer gibt alle Nutzer samt ihren Profilen.
  Nutzer: () => Promise<NutzerAntwort>

  // NutzerAnlegen legt einen Nutzer **samt seinem ersten Profil** an — ein
  // Nutzer ohne Arbeitsumgebung könnte sich anmelden und nichts tun.
  NutzerAnlegen: (name: string, bild: string) => Promise<EinNutzerAntwort>

  // ProfilAnlegen legt ein weiteres Profil für einen Nutzer an.
  ProfilAnlegen: (nutzerID: string, name: string) => Promise<EinProfilAntwort>

  // NutzerAendern schreibt die Angaben eines Nutzers fort.
  //
  // **Mit der Fassung**, die der Rufer gelesen hat — ohne sie ginge eine
  // gleichzeitige Änderung an einer anderen Fläche verloren.
  NutzerAendern: (angaben: NutzerangabenDTO) => Promise<EinNutzerAntwort>

  // ProfilAendern schreibt die Angaben eines Profils fort.
  //
  // **Ein Verbund seit Werkstück 9** (#1351). Es **setzt** und verschmilzt
  // nicht: was fehlt, steht danach leer da.
  ProfilAendern: (angaben: ProfilangabenDTO) => Promise<EinProfilAntwort>

  // NutzerLoeschen nimmt einen Nutzer samt seinen Profilen heraus.
  //
  // **Der Name bestätigt** — dieselbe Form wie beim Löschen eines Sets: eine
  // Kennung allein ist eine Zahl, und wer sich vertippt, löscht den falschen
  // Menschen.
  NutzerLoeschen: (nutzerID: string, name: string) => Promise<SchreibAntwort>

  // ProfilLoeschen nimmt ein Profil heraus.
  ProfilLoeschen: (profilID: string, name: string) => Promise<SchreibAntwort>

  // GezeigtesProfil sagt, in welchem Profil dieses Fenster arbeitet.
  //
  // **Jede andere Antwort gilt für dieses Profil** — Sets, Decks, Sammlung,
  // Einstellungen.
  GezeigtesProfil: () => Promise<EinProfilAntwort>

  // Anbieter nennt die Anbieter, über die man sich anmelden kann.
  //
  // **Ohne Kontakt nach draußen**: die Liste steht im Programm, und wer sie
  // abfragt, hat noch nichts angestoßen.
  Anbieter: () => Promise<AnbieterAntwort>

  // Anmelden führt den Anmeldefluss — **der Browser geht auf**.
  //
  // Er geht auf den Rechner des Anbieters, kommt auf `127.0.0.1` zurück, und
  // erst danach gibt es einen Nutzer. Wer abbricht, bekommt einen Fehler und
  // keinen halben Zustand.
  Anmelden: (anbieter: string) => Promise<EinNutzerAntwort>

  // Anbieterkonten gibt die Konten eines Nutzers.
  Anbieterkonten: (nutzerID: string) => Promise<KontenAntwort>

  // KontoLoesen nimmt eine fremde Kennung von einem Nutzer.
  //
  // **Der Rückweg gehört dazu**: wer sein Konto beim Anbieter schließt, darf
  // nicht ausgesperrt sein.
  KontoLoesen: (nutzerID: string, anbieter: string) => Promise<SchreibAntwort>

  // KontoEintragen legt ein Konto unter einer Adresse oder einer Rufnummer an.
  //
  // **Eine Kennung und nicht zwei Felder**: ob da eine Adresse oder eine
  // Nummer steht, sieht man ihr an. Der Name ist freiwillig — leer heißt:
  // einen aus der Kennung.
  KontoEintragen: (
    kennung: string,
    kennwort: string,
    name: string,
  ) => Promise<EinNutzerAntwort>

  // KontoAnmelden lässt einen Nutzer unter seiner Kennung herein.
  //
  // **Die Meldung sagt nicht, welche Hälfte falsch war** — die Auskunft hülfe
  // niemandem außer dem, der rät.
  KontoAnmelden: (
    kennung: string,
    kennwort: string,
    nachweis: string,
  ) => Promise<EinNutzerAntwort>

  // KontoAnonym legt ein Konto ohne jede Angabe an.
  //
  // **Ohne Argument, weil es nichts anzugeben gibt.**
  KontoAnonym: () => Promise<EinNutzerAntwort>

  // Nutzerbilder gibt zu jedem Nutzer mit Bild einen Daten-URI.
  //
  // **Nur am Fenster.** Ein Daten-URI ist die Darstellung einer Liste und kein
  // Weg zu einer Auskunft.
  Nutzerbilder: () => Promise<Nutzerbildantwort>

  // NutzerbildWaehlen fragt nach einer Bilddatei und legt sie ab.
  //
  // **Der Dialog steht in Go** — ein Fenster bekäme einen Namen und keinen
  // Pfad. Ein Abbruch gibt eine leere Antwort und keinen Fehler.
  NutzerbildWaehlen: (
    nutzerID: string,
    fassung: string,
  ) => Promise<EinNutzerAntwort>

  // NutzerbildEntfernen nimmt das Bild eines Nutzers wieder weg.
  NutzerbildEntfernen: (
    nutzerID: string,
    fassung: string,
  ) => Promise<EinNutzerAntwort>

  // Nutzereinstellungen gibt die Einstellungen eines Menschen.
  Nutzereinstellungen: (nutzerID: string) => Promise<NutzereinstellungenAntwort>

  // NutzereinstellungSetzen schreibt eine fort und gibt **alle** zurück.
  //
  // **Alle, weil sonst zweimal zu fragen wäre** (§7, Regel 2).
  NutzereinstellungSetzen: (
    nutzerID: string,
    schluessel: string,
    wert: string,
  ) => Promise<NutzereinstellungenAntwort>

  // PersonendatenSichern fragt nach einem Ziel und schreibt die Auskunft.
  //
  // **Der Dialog steht in Go.** Ein Abbruch gibt einen leeren Umschlag und
  // keinen Fehler.
  PersonendatenSichern: (nutzerID: string) => Promise<SchreibAntwort>

  // ProfilbadgeWaehlen fragt nach einer Bilddatei und legt sie als Abzeichen
  // ab.
  //
  // **Das Abzeichen gehört dem Profil**: es liegt rechts unten **über** dem
  // Nutzerfoto und zeigt die Rolle, in der jemand gerade arbeitet.
  ProfilbadgeWaehlen: (
    profilID: string,
    fassung: string,
  ) => Promise<EinProfilAntwort>

  // ProfilbadgeEntfernen nimmt das Abzeichen eines Profils wieder weg.
  ProfilbadgeEntfernen: (
    profilID: string,
    fassung: string,
  ) => Promise<EinProfilAntwort>

  // ProfilWechseln stellt es um.
  //
  // **Danach baut sich die Anwendung neu auf** (§7, Regel 1) — siehe
  // `profil.tsx`. Es wird nichts gelöscht: was im alten Profil liegt, ist nach
  // dem Zurückwechseln wieder da.
  ProfilWechseln: (
    profilID: string,
    nachweis: string,
  ) => Promise<EinProfilAntwort>

  // WerArbeitet sagt, **wer** gerade arbeitet — Werkstück 17 von Phase 4.12
  // (#1400).
  //
  // **Nicht dasselbe wie der Eigner aus `GezeigtesProfil`.** Wer in ein
  // fremdes Profil gewechselt ist, bleibt, wer er ist — und seine eigenen
  // Freigaben entscheiden, wohin er weiter darf.
  WerArbeitet: () => Promise<WerArbeitetAntwort>

  // NutzerWechseln stellt um, wer arbeitet.
  //
  // **Der Nachweis ist der zweite Faktor DIESES Nutzers**, nicht der eigene.
  // Wer den eigenen prüfte, ließe jeden zu jedem werden, der seinen eigenen
  // Code kennt.
  //
  // **Der Profilzeiger zieht mit**: danach wird im ersten Profil des neuen
  // Menschen gearbeitet, und die Anwendung baut sich neu auf (§7, Regel 1).
  NutzerWechseln: (
    nutzerID: string,
    nachweis: string,
  ) => Promise<EinProfilAntwort>

  // Arbeitskontext sagt in **einem** Zug, wer gerade in welchem Profil als
  // was arbeitet — Werkstück O1 von Phase 5.2
  // (#1425).
  //
  // **An der Stelle von `Leseart`** (#1351, Werkstück 16): `nurLesend` steht
  // darin, und dazu, warum — die Rolle oder ein Token, das nur lesen darf.
  Arbeitskontext: () => Promise<ArbeitskontextAntwort>
}

// WerArbeitetAntwort nennt den Menschen, der gerade arbeitet.
//
// **Die Kennung und nicht der ganze Nutzer**: wer die Angaben braucht, hat sie
// aus der Übersicht — und eine zweite Stelle, die einen Nutzer zusammensetzt,
// liefe mit der ersten auseinander.
export interface WerArbeitetAntwort extends Umschlag {
  nutzerID: string
}
