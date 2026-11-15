// **Was am Tisch gerade geschehen ist** — PLAY-04.AK-02
// (#1712).
//
// # Warum aus dem Stand und nicht aus dem Ereignisstrom
//
// **Weil der Strom nicht sagt, woher etwas kam.** Gemessen am 2026-09-22 an
// `internal/service/sandkastenschieben_test.go`: ein `objekt.schieben` trägt
// `objekt`, `zone` und manchmal `stelle` — die **Ausgangszone steht nicht
// darin**. „Eintritt" ließe sich daraus lesen, „Austritt" nicht.
//
// **Und weil ein Gast gar keinen Strom hat.** Er bekommt nach jedem Zug das
// Tischbild, das der Wirt für ihn gerechnet hat (G5,
// #1462) —
// eine Wirkung, die am Ereignis hinge, gäbe es an seinem Tisch nicht. Der
// Vergleich zweier Stände gilt an beiden.
//
// # Was hier nicht steht
//
// **Wie eine Wirkung aussieht.** `wirkungsbild` nennt Farbe, Zeichen und
// Dauer; **gezeichnet** wird sie in `pixinaht.ts` und in der Zonenliste. Die
// Trennung ist dieselbe wie zwischen `tischplan` und `pixinaht`: hier die
// Entscheidung, dort der Ruf an die Grafikkarte.
import type { Partie } from '../apipartie'

// Wirkungsart sind die acht Klassen, die PLAY-04.AK-02 aufzählt:
// „Kartenbewegung, Eintritt/Austritt, Tappen, Leben, Gift, Zugwechsel und
// Spielende".
export type Wirkungsart =
  | 'bewegung'
  | 'eintritt'
  | 'austritt'
  | 'drehung'
  | 'leben'
  | 'gift'
  | 'zugwechsel'
  | 'spielende'

// Wirkung ist eine Klasse und das, woran sie hängt.
export interface Wirkung {
  art: Wirkungsart

  // woran ist eine Objekt- oder Sitzkennung — **leer bei Zugwechsel und
  // Spielende**: sie hängen am Tisch und nicht an einem Ding darauf.
  woran: string
}

// Tischstand ist die kleine Projektion, auf der verglichen wird.
//
// **Klein mit Absicht**: was hier nicht drinsteht, kann keine Wirkung
// auslösen. Ein Vergleich über die ganze Partie meldete jedes neu gerechnete
// Feld als Änderung.
export interface Tischstand {
  objekte: Record<string, { zone: string; getappt: boolean }>
  sitze: Record<string, { leben: number; vergiftung: number }>
  zug: number
  vorbei: boolean
}

// Schlachtfeld ist die Zone, in der Bleibendes liegt — CR 403.1.
const Schlachtfeld = 'tisch'

// standVon zieht aus einer Partie, worauf es ankommt.
export function standVon(partie: Partie): Tischstand {
  const objekte: Tischstand['objekte'] = {}
  for (const objekt of partie.objekte) {
    objekte[objekt.kennung] = {
      zone: objekt.zone,
      getappt: objekt.getappt,
    }
  }

  const sitze: Tischstand['sitze'] = {}
  for (const sitz of partie.sitze) {
    sitze[sitz.kennung] = {
      leben: sitz.leben,
      vergiftung: sitz.vergiftung,
    }
  }

  return {
    objekte,
    sitze,
    zug: partie.zug.nummer,

    // **Vorbei heißt: ein Sieger oder ein Unentschieden** — CR 104.2 und
    // CR 104.4. Dieselbe Frage stellt `Partieausgang` für seine Zeile.
    vorbei: partie.ausgang.sieger !== '' || partie.ausgang.unentschieden,
  }
}

// wirkungen sagt, was sich seit dem vorigen Stand getan hat.
//
// **Ohne vorigen Stand gibt es keine** — wer einen Tisch aufschlägt, hat
// nichts verpasst, und ein Brett, das beim Öffnen einmal ganz aufblitzt, wäre
// kein Hinweis, sondern ein Blendwerk.
export function wirkungen(
  vorher: Tischstand | undefined,
  jetzt: Tischstand,
): Wirkung[] {
  if (vorher === undefined) {
    return []
  }

  return [
    ...objektwirkungen(vorher, jetzt),
    ...sitzwirkungen(vorher, jetzt),
    ...tischwirkungen(vorher, jetzt),
  ]
}

// objektwirkungen vergleicht, was auf dem Tisch liegt.
function objektwirkungen(vorher: Tischstand, jetzt: Tischstand): Wirkung[] {
  const gefunden: Wirkung[] = []

  for (const [kennung, stand] of Object.entries(jetzt.objekte)) {
    const davor = vorher.objekte[kennung]

    // **Aus dem Nichts auf den Tisch ist ein Eintritt** — ein Spielstein
    // entsteht dort (CR 111.1). Anderswo ist es kein Ereignis: eine Karte,
    // die dieser Sitz zum ersten Mal sieht, hat nichts getan.
    if (davor === undefined) {
      if (stand.zone === Schlachtfeld) {
        gefunden.push({ art: 'eintritt', woran: kennung })
      }

      continue
    }

    if (davor.zone !== stand.zone) {
      gefunden.push({
        art: zonenwirkung(davor.zone, stand.zone),
        woran: kennung,
      })
    }

    // **Und die Drehung daneben, nicht danach.** Eine Karte, die getappt
    // aufs Feld kommt, tut zweierlei; wer nur eines zeigte, ließe das andere
    // verschwinden.
    if (davor.getappt !== stand.getappt) {
      gefunden.push({ art: 'drehung', woran: kennung })
    }
  }

  // **Was es nicht mehr gibt, ist ausgetreten** — es kann in eine Zone
  // gegangen sein, die dieser Sitz nicht sieht, und dann ist sein
  // Verschwinden alles, was man davon erfährt.
  for (const [kennung, stand] of Object.entries(vorher.objekte)) {
    if (jetzt.objekte[kennung] === undefined && stand.zone === Schlachtfeld) {
      gefunden.push({ art: 'austritt', woran: kennung })
    }
  }

  return gefunden
}

// zonenwirkung sagt, welche Klasse ein Zonenwechsel hat.
function zonenwirkung(von: string, nach: string): Wirkungsart {
  if (nach === Schlachtfeld) return 'eintritt'
  if (von === Schlachtfeld) return 'austritt'

  return 'bewegung'
}

// sitzwirkungen vergleicht, was an den Spielern steht.
function sitzwirkungen(vorher: Tischstand, jetzt: Tischstand): Wirkung[] {
  const gefunden: Wirkung[] = []

  for (const [kennung, stand] of Object.entries(jetzt.sitze)) {
    const davor = vorher.sitze[kennung]

    // **Ein neuer Sitz bringt keine Wirkung mit**: sein Lebensstand ist sein
    // Anfangsstand und keine Änderung.
    if (davor === undefined) continue

    if (davor.leben !== stand.leben) {
      gefunden.push({ art: 'leben', woran: kennung })
    }

    if (davor.vergiftung !== stand.vergiftung) {
      gefunden.push({ art: 'gift', woran: kennung })
    }
  }

  return gefunden
}

// tischwirkungen sind die zwei, die an keinem Ding hängen.
function tischwirkungen(vorher: Tischstand, jetzt: Tischstand): Wirkung[] {
  const gefunden: Wirkung[] = []

  if (vorher.zug !== jetzt.zug) {
    gefunden.push({ art: 'zugwechsel', woran: '' })
  }

  // **Genau einmal**: ein Ende, das bei jedem Bild aufblitzte, wäre nach dem
  // dritten kein Hinweis mehr.
  if (!vorher.vorbei && jetzt.vorbei) {
    gefunden.push({ art: 'spielende', woran: '' })
  }

  return gefunden
}

// Bild ist, wie eine Wirkung aussieht.
export interface Bild {
  // farbe ist der Ton, in dem gezeichnet wird — als Zahl, denn die Leinwand
  // rechnet damit.
  farbe: number

  // zeichen ist, was **ohne** Farbe unterscheidet — WCAG 1.4.1 verlangt, dass
  // sie nicht das einzige Mittel ist. Es steht in der Zonenliste am Eintrag
  // und auf der Leinwand als Marke.
  zeichen: string

  // dauer ist, wie lange die Wirkung zu sehen ist — in Millisekunden.
  dauer: number
}

// bilder ordnet jeder Wirkung ihr Aussehen zu.
//
// **Acht verschiedene Farben und acht verschiedene Zeichen**, und ein Lauf
// daneben hält beides auseinander: zwei Wirkungen, die gleich aussehen, sind
// eine — und das Kriterium verlangt wörtlich „unterscheidbare Effekte".
//
// **Die Zahlen sind gewählt und nicht gemessen.** Wer sie ändert, ändert eine
// Gestaltung und keine Zusage — dieselbe Ansage wie bei `Bewegungsdauer`.
const bilder: Record<Wirkungsart, Bild> = {
  // **Grün kommt, rot geht** — die Richtung, die ein Brett ohnehin hat.
  eintritt: { farbe: 0x35a26a, zeichen: '▲', dauer: 420 },
  austritt: { farbe: 0xc2453a, zeichen: '▼', dauer: 420 },

  // **Blau bewegt sich** — ein Wechsel ohne Brett, etwa Hand auf Friedhof.
  bewegung: { farbe: 0x3b7dd8, zeichen: '→', dauer: 320 },

  // **Die Drehung ist die kürzeste**: sie hat ihre eigene Bewegung schon,
  // und der Ring bestätigt sie nur.
  drehung: { farbe: 0xb8862b, zeichen: '↻', dauer: 260 },

  // **Leben und Gift sind zwei Zahlen an einem Sitz** (CR 119.1, CR 122.1f)
  // und tragen deshalb zwei deutlich verschiedene Töne — wer sie verwechselt,
  // verwechselt zwei Verlustbedingungen.
  leben: { farbe: 0xd4426f, zeichen: '♥', dauer: 520 },
  gift: { farbe: 0x6b8f24, zeichen: '☣', dauer: 520 },

  // **Der Zugwechsel hängt am Tisch und dauert länger** — er betrifft alles,
  // was daraufliegt.
  zugwechsel: { farbe: 0x8a5cd6, zeichen: '⟳', dauer: 680 },

  // **Und das Spielende am längsten**: es ist die Antwort auf die einzige
  // Frage, die dann noch zählt.
  spielende: { farbe: 0x2f3542, zeichen: '■', dauer: 900 },
}

// wirkungsbild gibt das Aussehen einer Wirkung.
export function wirkungsbild(art: Wirkungsart): Bild {
  return bilder[art]
}

// Wirkungsarten sind alle acht, in der Reihenfolge der Tabelle darüber.
//
// **Aus der Tabelle und nicht als zweite Aufzählung daneben** — die liefe bei
// der neunten Wirkung auseinander, und der Lauf über die Unterscheidbarkeit
// prüfte dann sieben von acht (§10). Dass die Tabelle **vollständig** ist,
// hält `Record<Wirkungsart, Bild>` schon beim Übersetzen.
export const Wirkungsarten = Object.keys(bilder) as Wirkungsart[]
