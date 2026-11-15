// **Die Typen der Leute-Seite, an einer Stelle** — Nutzer, Profil, Konto,
// Zugriffstoken, zweiter Faktor, Vorgaben, Freigaben und der Arbeitskontext.
//
// # Warum eine eigene Datei
//
// `api.ts` ist die Fassade und führt jeden Typ, den eine Seite aus einer
// Antwort braucht, **einmal** — Seiten importieren aus `./api` und nicht aus
// dem Modul, das den Typ zufällig hält. Mit Werkstück O2 von Phase 5.2
// (#1426)
// wuchs sie über die Zeilengrenze aus §2, und die Leute-Seite ist der Teil,
// der seit Phase 4.12 am schnellsten wächst. Er steht deshalb hier, und
// `api.ts` reicht ihn als Ganzes weiter.
//
// **Wer hier einen Typ hinzufügt, fügt ihn nur hier hinzu** — `export type *`
// nimmt ihn mit.

export type { ArbeitskontextAntwort } from './apikontext'
export type {
  AnbieterAntwort,
  AnbieterkontoDTO,
  EinNutzerAntwort,
  EinProfilAntwort,
  WerArbeitetAntwort,
  KontenAntwort,
  NutzerAntwort,
  NutzerDTO,
  NutzerangabenDTO,
  Nutzerbildantwort,
  NutzereinstellungDTO,
  NutzereinstellungenAntwort,
  ProfilangabenDTO,
  ProfilDTO,
} from './apinutzer'
export type {
  AblaufwahlDTO,
  AblaufwahlenAntwort,
  EinTokenAntwort,
  TokenAntwort,
  TokenangabenDTO,
  TokenDTO,
  TokenwiderrufAntwort,
} from './apitoken'
export type {
  Einrichtungsantwort,
  FaktorDTO,
  FaktorenAntwort,
} from './apifaktor'
export type {
  Vorgabeangebot,
  VorgabeangabenDTO,
  VorgabenAntwort,
  VorgabenstandDTO,
} from './apivorgaben'
export type {
  FreigabeDTO,
  FreigabeangabenDTO,
  FreigabenAntwort,
} from './apifreigaben'
