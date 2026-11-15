// **Die achtundzwanzig Wege der Partie** — die Naht, nicht ihre Gestalten.
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, *was* eine Partie ist** — Sitz, Objekt, Zug, Ausgang und
// die Antworten, in denen sie hinausgehen. Hier steht, **wie man sie
// erreicht**: je Fähigkeit eine Methode der Wails-Fassade.
//
// **`apipartie.ts` stand mit beidem bei 410 Zeilen** (§2) — die wartenden
// Fähigkeiten aus `M9`
// (#1611)
// waren die Zeilen, die darüber hinausgingen.
import type {
  BefehlAntwort,
  PartieAntwort,
  PartiebeitrittAntwort,
  Partieentwurf,
  PartiefreigabeAntwort,
  PartienAntwort,
  PartieverlaufAntwort,
} from './apipartie'
import type {
  NebenbeiAntwort,
  PartiediagnoseAntwort,
  PartieoffenlegungAntwort,
} from './apinebenbei'
import type { DeckleiheAntwort, Leihwunsch, Leihziel } from './apileihe'

// Partiemitbring ist, was ein eingeladener Gast an den Tisch mitbringt (P9b,
// #2377): ein eigenes Deck **oder** eine Leihe des Wirts — beides leer, wenn
// sein Sitz schon ein Deck hat.
export interface Partiemitbring {
  deckID: string
  leiheID: string
}

// Partieapi sind die achtundzwanzig Wege.
export interface Partieapi {
  PartieAnlegen: (entwurf: Partieentwurf) => Promise<PartieAntwort>
  Partie: (kennung: string, fuer: string) => Promise<PartieAntwort>

  // Partien nennt die eigenen Partien — die Frage, die **vor** dem Ansehen
  // steht (#1600).
  Partien: () => Promise<PartienAntwort>

  PartieBefehl: (
    partieID: string,
    befehlID: string,
    absender: string,
    art: string,
    nutzlast: Record<string, unknown>,
    erwartet: string,
  ) => Promise<BefehlAntwort>

  // PartieVerlauf gibt, was in dieser Partie geschehen ist — aus der Sicht
  // **eines** Sitzes (G5,
  // #1462).
  PartieVerlauf: (
    kennung: string,
    fuer: string,
  ) => Promise<PartieverlaufAntwort>

  // **Die Freigabe für Gäste** (G5, #1462) — nur am Fenster, und der Grund
  // steht in `docs/FAEHIGKEITEN.md`.
  PartieFreigeben: (
    kennung: string,
    fuer: string,
    gastsitze: string[],
  ) => Promise<PartiefreigabeAntwort>
  PartieFreigabeBeenden: () => Promise<PartiefreigabeAntwort>
  PartieFreigabeStand: () => Promise<PartiefreigabeAntwort>

  // **Der Beitritt an einem fremden Tisch** (G5, #1462).
  PartieBeitreten: (
    gegenueber: string,
    partie: string,
    mit: Partiemitbring,
  ) => Promise<PartiebeitrittAntwort>
  PartieBeitrittBeenden: () => Promise<PartiebeitrittAntwort>
  PartieBeitrittStand: () => Promise<PartiebeitrittAntwort>

  // PartieWuenschen gibt **kein** Ergebnis: was aus einem Wunsch wird,
  // entscheidet der Wirt und meldet es über die Wails-Ereignisse.
  PartieWuenschen: (
    befehlID: string,
    art: string,
    nutzlast: Record<string, unknown>,
  ) => Promise<PartiebeitrittAntwort>

  // **Am Tisch reden und zeigen** (G5, #1462) — beide für **beide** Seiten:
  // wer freigegeben hat, ruft rund; wer beigetreten ist, schickt an seinen
  // Wirt. Welche der beiden gilt, entscheidet die Go-Seite aus ihrem Stand.
  PartieSagen: (text: string) => Promise<NebenbeiAntwort>
  PartieZeigen: (
    art: string,
    objekt: string,
    x: number,
    y: number,
  ) => Promise<NebenbeiAntwort>

  // **Das Protokoll danach** (G5, #1462): wer es lesen darf, und wann es
  // weg ist.
  PartieOffenlegungen: (kennung: string) => Promise<PartieoffenlegungAntwort>
  PartieOffenlegen: (
    kennung: string,
    von: string,
    an: string[],
  ) => Promise<PartieoffenlegungAntwort>
  PartieOffenlegungZuruecknehmen: (
    kennung: string,
    von: string,
  ) => Promise<PartieoffenlegungAntwort>
  PartieVergessen: (kennung: string) => Promise<PartieoffenlegungAntwort>

  // **Und wie es um die Leitung steht** (G5, #1462) — ohne Kartengeheimnisse.
  Partiediagnose: () => Promise<PartiediagnoseAntwort>

  // **Die digitale Deckleihe** (G6, #1463) — nur am Fenster, und der Grund
  // steht in `docs/FAEHIGKEITEN.md`. Ihre Gestalten stehen in `apileihe.ts`.
  Deckleihen: () => Promise<DeckleiheAntwort>
  DeckleiheAnbieten: (
    wunsch: Leihwunsch,
    ziel: string,
  ) => Promise<DeckleiheAntwort>
  DeckleiheAnnehmen: (pfad: string) => Promise<DeckleiheAntwort>

  // **Den Ort wählt ein Dialog** (UXR-F15, #2030): der Sichern-Dialog für
  // das Ziel; für den Umschlag, der angenommen wird, gilt `UmschlagWaehlen`.
  DeckleiheZielWaehlen: () => Promise<Leihziel>
  DeckleiheWiderrufen: (kennung: string) => Promise<DeckleiheAntwort>
  DeckleiheZurueckgeben: (kennung: string) => Promise<DeckleiheAntwort>

  PartiePausieren: (kennung: string, fuer: string) => Promise<PartieAntwort>
  PartieFortsetzen: (kennung: string, fuer: string) => Promise<PartieAntwort>
  PartieBeenden: (kennung: string, fuer: string) => Promise<PartieAntwort>
}
