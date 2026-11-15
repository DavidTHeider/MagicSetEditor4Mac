// Die Tastenkürzel des Rahmens.
//
// **macOS-Eigenheiten werden bedient, nicht nachgeahmt** (AGENTS.md §6): ⌘K für
// die Palette, ⌘F für die Suche, ⌘, für die Einstellungen, ⌘Z und ⇧⌘Z für
// Rückgängig, ⌘S und ⇧⌘S für Sichern. Alle sind gewohnte Bedeutungen; eine andere Belegung wäre eine
// Anwendung, die sich falsch anfühlt, ohne dass jemand sagen könnte, warum.
import { seiten, type Seitenschluessel } from '../navigation'
import {
  vorgabebelegung,
  ziffern,
  type Ziffernbelegung,
} from './ziffernbelegung'

// **Ein unterschiedener Verbund**, kein Verbund mit optionalem Feld.
//
// Der erste Anlauf hatte `seite?: Seitenschluessel`, und der Aufrufer musste
// danach noch einmal prüfen, ob es da ist — ein Zweig, den kein Test betreten
// kann, weil `art: 'seite'` ohne Seite gar nicht entsteht. So sagt der Typ, was
// gilt, und der Zweig fällt weg.
export type Kuerzeltreffer =
  | { art: 'palette' }
  | { art: 'suche' }
  | { art: 'zurueck' }
  | { art: 'vor' }
  | { art: 'sichern' }
  | { art: 'sichernUnter' }
  | { art: 'oeffnen' }
  | { art: 'neu' }
  | { art: 'schliessen' }
  | { art: 'drucken' }
  // **Die Regionen eines Arbeitsplatzes** — Werkstück U2 von Phase 5.1
  // (#1413):
  // ⌥⌘[ klappt den Navigator, ⌥⌘] den Inspector. Der Rahmen tut damit
  // nichts — der Arbeitsplatz horcht selbst, denn nur er weiß, ob er eine
  // solche Region hat.
  | { art: 'navigator' }
  | { art: 'inspector' }
  | { art: 'seitenleiste' }
  // **Der Seitengang** — UX-03.AK-02 (#1675): ⌘[ und ⌘].
  | { art: 'seiteZurueck' }
  | { art: 'seiteVor' }
  | { art: 'seite'; seite: Seitenschluessel }

// Gewohnt sind die Kürzel, die eine Handlung von außen rufen — ⌘O, ⌘N, ⌘W,
// ⌘P (#891) und der Seitengang (UX-03.AK-02, #1675). Der Rahmen fängt sie ab,
// auch wenn niemand etwas anzubieten hat.
export type Gewohnt = Extract<
  Kuerzeltreffer,
  {
    art:
      'oeffnen' | 'neu' | 'schliessen' | 'drucken' | 'seiteZurueck' | 'seiteVor'
  }
>

const gewohnteArten: ReadonlySet<string> = new Set<Gewohnt['art']>([
  'oeffnen',
  'neu',
  'schliessen',
  'drucken',
  'seiteZurueck',
  'seiteVor',
])

export function istGewohnt(treffer: Kuerzeltreffer): treffer is Gewohnt {
  return gewohnteArten.has(treffer.art)
}

// Tastenlage ist alles, was eine Deutung braucht.
//
// **Als Verbund und nicht als vier Stellen**, weil vier Wahrheitswerte
// hintereinander am Aufrufort nichts mehr sagen: `deuten(taste, true, false,
// true)` ist nicht zu lesen.
export interface Tastenlage {
  taste: string

  // code ist die **physische** Taste — `"Digit1"`, `"KeyS"`.
  //
  // **Sie wird für die zweite Reihe gebraucht** (#1291): mit ⌥ schreibt macOS
  // das Zeichen um. ⌥1 gibt nicht `"1"`, sondern `"¡"`; wer auf `taste`
  // prüfte, fände die Ziffer nie. `code` bleibt, was es ist.
  code: string

  meta: boolean

  // ctrl sagt, ob ⌃ gedrückt war. **Keines unserer Kürzel trägt es** — eine
  // Kombination mit ⌃ gehört macOS (#1994). Wahlweise, weil sie fast nie
  // gedrückt ist.
  ctrl?: boolean

  // alt sagt, ob ⌥ gedrückt war — die zweite Kürzelreihe.
  alt: boolean

  umschalt: boolean

  // imEingabefeld sagt, ob der Fokus in einem Textfeld steht.
  imEingabefeld: boolean
}

// deuten sagt, was eine Tastenkombination bedeutet — oder nichts.
//
// **Getrennt vom Ereignis**, damit es ohne DOM prüfbar ist: eine Tabelle aus
// Taste und Bedeutung ist zu lesen, ein Test über `fireEvent.keyDown` nicht.
// **Die Ziffern kommen aus der Belegung** (UX-02.AK-03, #1674) — ohne eine
// gilt die Vorgabe aus `navigation.ts`. ⌘, ⌘' und ⌘? sind keine Ziffern und
// bleiben, wo sie sind.
export function deuten(
  lage: Tastenlage,
  belegung: Ziffernbelegung = vorgabebelegung,
): Kuerzeltreffer | null {
  // **⌃⌘ gehört dem System** (#1994): ⌃⌘F ist das Vollbild, ⌃⌘Leertaste
  // die Emoji. Gemessen am 2026-10-01 in der Abnahme-VM: ⌃⌘F schlug die Suche
  // auf, weil hier nur ⌘ und die Taste gelesen wurden.
  if (!lage.meta || lage.ctrl === true) {
    return null
  }

  // **Die zweite Reihe zuerst und für sich.** Mit ⌥ trägt `taste` ein anderes
  // Zeichen; jede Prüfung darunter liefe ins Leere, und ein ⌥⌘S würde
  // stillschweigend nichts tun statt zu sichern.
  if (lage.alt) {
    return zweiteReihe(lage.code)
  }

  // **⌘[ und ⌘] gehen den Seitengang** (UX-03.AK-02, #1675) — über die Taste
  // wie ⌥⌘[: auf einer deutschen Tastatur liegt „[" unter ⌥5, und ⌥⌘5 ist
  // eine Seite. **Auch im Eingabefeld**, wie ⌘S: dort bedeutet die Taste
  // nichts, und das offene Feld verlässt der Gang selbst.
  const schritt = seitenschritte[lage.code]
  if (schritt !== undefined) {
    return schritt
  }

  const klein = lage.taste.toLowerCase()

  if (klein === 'k') {
    return { art: 'palette' }
  }

  if (klein === 'f') {
    return { art: 'suche' }
  }

  if (klein === 'z') {
    // **Im Textfeld gehört ⌘Z dem Textfeld.**
    //
    // Geschrieben wird beim Verlassen des Feldes, nicht bei jedem Tastendruck:
    // was jemand gerade tippt, steht noch in keinem Verlauf. Wer ⌘Z hier
    // abfinge, nähme eine **frühere** Änderung zurück, während der Cursor in
    // einer ungesicherten steht — die Rücknahme träfe etwas anderes als das,
    // was die tippende Person vor Augen hat.
    if (lage.imEingabefeld) {
      return null
    }

    return lage.umschalt ? { art: 'vor' } : { art: 'zurueck' }
  }

  if (klein === 's') {
    // **Auch im Eingabefeld**, anders als ⌘Z.
    //
    // Der Grund für die Ausnahme dort ist, dass das Textfeld selbst eine
    // Rücknahme kennt und die gemeinte näher liegt. ⌘S kennt es nicht: in einem
    // Textfeld sichert niemand etwas, und wer im Kartentext steht und ⌘S
    // drückt, meint das Set (#162).
    return lage.umschalt ? { art: 'sichernUnter' } : { art: 'sichern' }
  }

  // **⌘O, ⌘N, ⌘W und ⌘P** (#891) — gewohnte Bedeutungen unter macOS, und bis zum
  // 2026-08-25 unbelegt.
  //
  // **Auch im Eingabefeld**, aus demselben Grund wie ⌘S: keines der drei hat
  // in einem Textfeld eine eigene Bedeutung, und wer im Kartentext steht und
  // ⌘P drückt, meint den Druckbogen.
  const gewohnt: Record<string, Kuerzeltreffer> = {
    o: { art: 'oeffnen' },
    n: { art: 'neu' },
    w: { art: 'schliessen' },
    p: { art: 'drucken' },
  }

  const gewohnter = gewohnt[klein]
  if (gewohnter !== undefined) {
    return gewohnter
  }

  if (ziffern.includes(klein)) {
    const seite = belegung[klein]

    return seite === undefined ? null : { art: 'seite', seite }
  }

  const treffer = seiten.find((seite) => seite.kuerzel === klein)
  if (treffer) {
    return { art: 'seite', seite: treffer.schluessel }
  }

  return null
}

// Kuerzelangabe ist ein Kürzel, wie eine Übersicht es zeigt.
//
// **Für die Tafel in den Einstellungen**
// (#245):
// „⌘1…⌘8, ⌘K, ⌘S, ⌘Z, ⌘, stehen — **sichtbar** sind sie nirgends."
export interface Kuerzelangabe {
  taste: string
  umschalt: boolean
  art: Kuerzeltreffer['art']

  // alt sagt, dass das Kürzel in der zweiten Reihe liegt — ⌥⌘ statt ⌘ —,
  // und code nennt dann die Taste, über die es gedeutet wird (#1413).
  alt?: boolean
  code?: string
}

// zweiteReihe deutet ⌥⌘ **über den Code**, weil ⌥ das Zeichen umschreibt.
//
// **Sie führt nur auf Seiten.** Ein ⌥⌘ auf etwas anderes gäbe eine zweite
// Bedeutung für dieselbe Handlung, und die erste hat schon eine.
function zweiteReihe(code: string): Kuerzeltreffer | null {
  // **Die zwei Klammern gehören den Regionen** (#1413) — über den Code,
  // weil ⌥ auch sie umschreibt: auf einer deutschen Tastatur gibt ⌥⌘[ ein
  // anderes Zeichen als auf einer amerikanischen, die Taste ist dieselbe.
  const region = regionen[code]
  if (region !== undefined) {
    return region
  }

  return zweiteReiheAus(code, seiten)
}

// regionen ordnet die Klammertasten den Regionen zu.
const regionen: Record<string, Kuerzeltreffer> = {
  BracketLeft: { art: 'navigator' },
  BracketRight: { art: 'inspector' },

  // **⌥⌘S klappt die Seitenleiste** — Werkstück U14 von Phase 5.1 (#1490),
  // das Kürzel des Finders und von Mail. ⌘S und ⇧⌘S bleiben das Sichern.
  KeyS: { art: 'seitenleiste' },
}

// seitenschritte ordnet die Klammertasten ohne ⌥ dem Seitengang zu.
const seitenschritte: Record<string, Kuerzeltreffer> = {
  BracketLeft: { art: 'seiteZurueck' },
  BracketRight: { art: 'seiteVor' },
}

// zweiteReiheAus ist derselbe Weg über eine gegebene Seitenliste.
//
// **Die Liste ist ein Parameter, weil der Zweig sonst unerreichbar wäre.**
// Seit die Turnierseite ⌥⌘0 trägt (#1294), ist **jede** Ziffer vergeben —
// über eine echte Taste lässt sich „Ziffer ohne Seite" nicht mehr herstellen.
//
// **Der Zweig bleibt trotzdem nötig**: er schützt nicht gegen eine Eingabe,
// sondern gegen eine **Konfiguration**. Genau so war `navigation.ts` bis zum
// 2026-09-01, und genau so wäre sie wieder, wenn jemand eine Seite entfernt.
// Ein Zweig, den kein Test betreten kann, ist nach §3 keiner, den es geben
// darf — also wird er dort betreten, wo er wirkt.
export function zweiteReiheAus(
  code: string,
  liste: readonly { zweitkuerzel: string; schluessel: Seitenschluessel }[],
): Kuerzeltreffer | null {
  const ziffer = code.startsWith('Digit') ? code.slice('Digit'.length) : ''
  if (ziffer === '') {
    return null
  }

  const treffer = liste.find((seite) => seite.zweitkuerzel === ziffer)
  if (!treffer) {
    return null
  }

  return { art: 'seite', seite: treffer.schluessel }
}

// globaleKuerzel sind die Kürzel, die nicht auf eine Seite führen.
//
// **Und `kuerzel.test.ts` hält sie gegen [deuten].** Eine Liste neben einer
// Funktion ist eine Kopie, und Kopien laufen auseinander; der Test schickt
// jede Zeile durch die Deutung und prüft, dass dabei die genannte Art
// herauskommt. Was `deuten` kann und hier fehlt, fällt damit nicht auf — was
// hier steht und `deuten` nicht kann, sofort.
export const globaleKuerzel: readonly Kuerzelangabe[] = [
  { taste: 'k', umschalt: false, art: 'palette' },
  { taste: 'f', umschalt: false, art: 'suche' },
  { taste: 'z', umschalt: false, art: 'zurueck' },
  { taste: 'z', umschalt: true, art: 'vor' },
  { taste: 's', umschalt: false, art: 'sichern' },
  { taste: 's', umschalt: true, art: 'sichernUnter' },
  { taste: 'o', umschalt: false, art: 'oeffnen' },
  { taste: 'n', umschalt: false, art: 'neu' },
  { taste: 'w', umschalt: false, art: 'schliessen' },
  { taste: 'p', umschalt: false, art: 'drucken' },
  {
    taste: '[',
    umschalt: false,
    alt: true,
    code: 'BracketLeft',
    art: 'navigator',
  },
  {
    taste: ']',
    umschalt: false,
    alt: true,
    code: 'BracketRight',
    art: 'inspector',
  },
  { taste: '[', umschalt: false, code: 'BracketLeft', art: 'seiteZurueck' },
  { taste: ']', umschalt: false, code: 'BracketRight', art: 'seiteVor' },
]

// kuerzelschluessel unterscheidet zwei Zeilen der Tafel — **samt ⌥**, denn
// ⌘[ und ⌥⌘[ sind dieselbe Taste und zwei Kürzel.
export function kuerzelschluessel(angabe: Kuerzelangabe): string {
  return `${angabe.taste}/${String(angabe.umschalt)}/${String(angabe.alt ?? false)}`
}
