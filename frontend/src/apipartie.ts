// **Die Partie an der Naht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Was hier steht, ist die Sicht eines Sitzes und nicht der Zustand.** Es
// gibt keinen Weg von der Wahrheit zur Oberfläche, der an der Projektion
// vorbeiführt — `Vollzustand` hat mit Absicht kein Gegenstück hier.
import type { Sitz } from './apipartiesitz'
import type { Ausgang, Zug } from './apizug'
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// Regelwerk sagt, welche Regeln in welcher Strenge gelten.
export interface Regelwerk {
  kennung: string
  fassung: number

  // strenge ist `sandkasten`, `begleitet` oder `streng`.
  strenge: string

  // abgewaehlt sind die Bereiche, die die Runde **selbst** übernimmt —
  // PLAY-02.AK-02 und AK-03
  // (#1710).
  //
  // **Leer heißt: das Programm trägt, was es tragen kann.** Sie werden mit
  // dem Regelwerk **eingefroren** und ändern sich danach nicht mehr.
  abgewaehlt: string[]
}

// Sitz steht in `apipartiesitz.ts` — hier weitergereicht, damit jede
// Einfuhr bleibt, wie sie war.
export type { Sitz }

// Zuschauer sieht zu und hat keinen Sitz.
export interface Zuschauer {
  kennung: string
  name: string
}

// Spielobjekt ist ein Ding auf dem Tisch, so weit es gezeigt werden darf.
import type { Karteneigenschaften } from './apikarte'

export interface Spielobjekt {
  kennung: string
  art: string

  beherrscher: string
  eigner: string
  zone: string

  // karte und name sind leer, wo nicht hineingesehen werden darf.
  karte: string
  name: string

  // karteID ist die Zeile im Bestand — der Weg zum **Bild** (G2, #1459).
  //
  // **Leer, wo nicht hineingesehen werden darf**, und leer bei einem echten
  // Druck: dort gibt es keine eigene Karte, die dieses Programm zeichnet.
  karteID: string

  // bild ist die undurchsichtige Kennung des Kartenbilds — P9 (#2377),
  // Vermerk 0198. Leer, wo nicht hineingesehen werden darf; anders als
  // `karteID` zeigt sie in keinen Bestand. **`karteID` gilt nur für den
  // eigenen Bestand** — fehlt sie bei gesetztem `bild`, zeichnet das Bild der,
  // dem das Deck gehört.
  bild: string

  // ordnung ist **nur in offenen Zonen** gefüllt — die Reihenfolge einer
  // Bibliothek ist ihr eigentliches Geheimnis.
  ordnung: number

  getappt: boolean

  // verdeckt heißt: es **liegt** verdeckt — eine Tatsache für alle.
  verdeckt: boolean

  // geheim heißt: **dieser** Empfänger darf nicht hineinsehen.
  geheim: boolean

  haengtAn: string
  haufen: string
  zaehler: Record<string, number>

  // eigenschaften sind, was diese Karte **ist** — `null`, wo nicht
  // hineingesehen werden darf (M2, #1604).
  eigenschaften: Karteneigenschaften | null
}

export type { Karteneigenschaften, Kartenwert } from './apikarte'

// Mischpflicht sagt, dass eine verborgene Zone angesehen und seitdem nicht
// gemischt wurde.
//
// **Für jeden dieselbe** — sie ist die öffentliche Hälfte eines Blicks.
export interface Mischpflicht {
  sitz: string
  zone: string
}

// Legendenwahl ist eine offene Frage aus CR 704.5j — wer wählen muss und
// über welchen Namen.
export interface Legendenwahl {
  sitz: string
  name: string
}

// Ausloeser ist eine Fähigkeit, die ausgelöst hat und auf den Stapel wartet —
// CR 603.2 (M9, #1611).
//
// **Die Kennung ist der Griff**: der Befehl `spieler.reihung` nennt sie, und
// ohne sie ließen sich zwei Fähigkeiten derselben Art nicht unterscheiden.
export interface Ausloeser {
  kennung: string
  art: string
  beherrscher: string

  // wen ist, auf wen sie wirkt — leer, wo sie den Beherrscher meint.
  wen: string
}

// Partie ist eine Partie, wie einer sie sehen darf.
export interface Partie {
  id: string
  fassung: string

  spiel: string
  stand: string

  // fuer ist die Kennung, für die gerechnet wurde — damit ein Empfänger
  // merkt, wenn er die Sicht eines anderen bekommen hat.
  fuer: string

  // wirt steht offen da: die Vertrauensgrenze wird genannt, nicht versteckt.
  wirt: string

  regelwerk: Regelwerk

  sitze: Sitz[]
  zuschauer: Zuschauer[]
  objekte: Spielobjekt[]

  mischpflichten: Mischpflicht[]

  // legendenwahlen sind die offenen Fragen aus CR 704.5j (M12, #1614) —
  // **für jeden dieselben**: bis jemand geantwortet hat, steht das Spiel.
  legendenwahlen: Legendenwahl[]

  // ausgeloest sind die Fähigkeiten, die ausgelöst haben und auf den Stapel
  // warten — CR 603.2 (M9, #1611). **Für jeden dieselben.**
  ausgeloest: Ausloeser[]

  // ausloeserreihung ist der Sitz, der gerade ordnen soll — CR 603.3b.
  // **Leer heißt: niemand wird gefragt.**
  ausloeserreihung: string

  // vorrat sind die Kartenkennungen des **eigenen** Decks — damit die Bilder
  // da sind, bevor die Karte kommt. Leer für jeden, der nicht an diesem Sitz
  // sitzt.
  vorrat: string[]

  // zug ist, wo die Partie im Zug steht (G9, #1592).
  zug: Zug

  // ausgang ist, wie die Partie ausgegangen ist (M17, #1618).
  ausgang: Ausgang
}

// Spielereignis ist, was wirklich geschehen ist.
export interface Spielereignis {
  folge: string
  art: string

  oeffentlich: Record<string, unknown>

  // privat ist die Ergänzung für **diesen** Empfänger — leer, wo keine für
  // ihn dabei war.
  privat: Record<string, unknown>

  kette: string
}

// Objekthandlung sagt, was der Sitz an einem Objekt gerade tun kann — P3
// (#2371), aus dem Kern: `primaer` löst ein Klick aus, `zweit` ein
// Doppelklick (P4); leer heißt, dass es keine gibt.
export interface Objekthandlung {
  objekt: string
  handlungen: string[]
  primaer: string
  zweit: string
}

// PartieAntwort trägt eine Sicht und die erlaubten Befehle.
export interface PartieAntwort extends Umschlag {
  partie: Partie
  erlaubt: string[]

  // handlungen sind je Objekt, was dieser Sitz daran tun kann (P3, #2371).
  handlungen: Objekthandlung[]

  // befunde sind Auffälligkeiten, die die Partie **nicht** unbenutzbar machen
  // (G8, #1465) — ein Sitz etwa, dessen Profil es nicht mehr gibt.
  //
  // **Ein Befund ist kein Fehler**: gespielt wird weiter, und die Bindung
  // wird benannt statt verschwiegen.
  befunde: string[]

  // nachladefehler heißt: angelegt oder gewechselt **ist**, nur die Sicht
  // danach war nicht zu lesen (§7, Regel 3; #1970). `partie` trägt dann nur
  // Kennung und Sicht — genug, um selbst nachzulesen.
  nachladefehler?: FehlerDTO

  // absage sagt am fremden Tisch, dass der Wirt den letzten Wunsch abwies
  // (#2478) — mit Grund; sonst fehlt sie.
  absage?: Wunschabsage
}

// Wunschabsage ist ein Wunsch, den der Kern des Wirts abwies.
export interface Wunschabsage {
  befehl: string
  art: string
  grund: string
}

// PartieverlaufAntwort trägt den Verlauf einer Partie für **einen** Sitz.
//
// **Nicht `VerlaufAntwort`** — die gibt es schon für den Verlauf einer Karte
// (`apiverlauf.ts`), und zwei gleichnamige Antworten wären zwei verschiedene
// Dinge unter einem Namen.
export interface PartieverlaufAntwort extends Umschlag {
  // verlauf ist der Strom, **schon gekürzt**: was ein anderer Sitz privat
  // gesehen hat, steht nicht darin — auch nicht nachträglich.
  verlauf: Spielereignis[]
}

// BefehlAntwort trägt die Fassung danach und die Ereignisse.
export interface BefehlAntwort extends Umschlag {
  fassung: string
  ereignisse: Spielereignis[]
}

// Sitzvergabe ist ein Sitz und der Fingerabdruck seines Gegenübers.
//
// **Der Fingerabdruck und nicht die Kennung** — er ist die Gestalt, die ein
// Mensch vergleichen kann; die Kennung trägt den Schlüssel.
export interface Sitzvergabe {
  sitz: string
  fingerabdruck: string
}

// PartiefreigabeAntwort sagt, welche Partie für Gäste offensteht.
export interface PartiefreigabeAntwort extends Umschlag {
  laeuft: boolean
  partie: string
  wirt: string

  // vergeben ist, wer worauf sitzt — leer, solange niemand da ist.
  vergeben: Sitzvergabe[]
}

// PartiebeitrittAntwort sagt, an welchem fremden Tisch dieses Programm sitzt.
export interface PartiebeitrittAntwort extends Umschlag {
  laeuft: boolean
  partie: string

  // gegenueber ist der Fingerabdruck des Wirts — die Gestalt, die ein Mensch
  // vergleichen kann.
  gegenueber: string

  // pausiert heißt: die Leitung riss, ohne dass jemand aufgestanden ist.
  //
  // **Und nicht: die Partie ist vorbei.** Wem der Wirt wegbricht, dem gehört
  // sein Platz weiter; fortsetzen kann nur derselbe Wirt.
  pausiert: boolean

  // absage ist der Satz, mit dem der Wirt das Hinsetzen ablehnte (P9c,
  // #2377) — leer, wenn er es nicht tat. **Sein Satz, nicht unserer**: er
  // kommt vom anderen Rechner und steht unter einem eigenen.
  absage: string
}

// Sitzwunsch ist ein Platz am Tisch samt Deck.
export interface Sitzwunsch {
  kennung: string
  name: string
  deckID: string

  // leiheID ist eine **geliehene** Leihe statt eines eigenen Decks (G6,
  // #1463) — leer heißt: das eigene.
  leiheID: string

  // profilID bindet diesen Sitz an ein Profil **desselben Nutzers** (G8,
  // #1465) — leer heißt: keines.
  profilID: string

  // eingeladen ist die Peerkennung eines Gastes, der sein Deck mitbringt
  // (P9b, #2377) — dann bleiben Deck und Leihe leer.
  eingeladen?: string
}

// Partieentwurf ist der Wunsch nach einer Partie.
export interface Partieentwurf {
  spiel: string
  regelwerk: Regelwerk
  wirt: string
  sitze: Sitzwunsch[]
}

// Partiezeile ist eine Partie, wie die Liste sie zeigt —
// #1600.
//
// **Ohne Objekte und ohne Sitzkennungen**: eine Liste ist keine Sicht. Was in
// einer Partie liegt, sagt `Partie` — und die rechnet je Sitz.
export interface Partiezeile {
  id: string
  fassung: string
  spiel: string
  stand: string

  // sitze sind die Plätze am Tisch — Kennung und Name.
  sitze: { kennung: string; name: string }[]

  // objekte ist, wie viele Dinge in dieser Partie liegen.
  objekte: number

  // zug ist der wievielte Zug — null heißt: es geht nicht zugweise zu.
  zug: number

  // eigenerSitz ist der Sitz, der an das gezeigte Profil gebunden ist —
  // leer heißt: keiner, das Profil sieht hier zu (UXR-F16,
  // #2031).
  eigenerSitz: string

  // dran sagt, ob die Partie gerade auf den eigenen Sitz wartet. **Gerechnet
  // hat `service`** (Priorität, Abwurf, Legendenwahl, Reihung, Beginn); das
  // Fenster zeigt es nur — eine zweite Rechnung hier liefe von der ersten
  // auseinander (§2).
  dran: boolean
}

export interface PartienAntwort extends Umschlag {
  partien: Partiezeile[]
}

export type { Ausgang, Ausgeschieden, Zug } from './apizug'

export type { Partieapi } from './apipartieweg'
