// Erscheinungsbild und Schriftgröße am Wurzelelement —
// #245.
//
// # Warum am Wurzelelement und nicht in React
//
// Beides gilt für **alles**: für den Rahmen, für jede Seite, für den
// Rollbalken, den die Anwendung gar nicht zeichnet. Ein React-Zustand, den
// jede Komponente lesen müsste, wäre dieselbe Angabe an hundert Stellen.
//
// # Warum die Größe in Prozent und nicht in Pixeln
//
// `rem` bezieht sich auf die Wurzelschrift. Wer sie in Prozent setzt, skaliert
// **jede** Angabe der Oberfläche mit — Abstände, Ränder und Fokusringe
// eingeschlossen, denn Tailwind gibt sie in `rem` aus. Eine feste Pixelzahl
// nähme der Systemeinstellung ihre Wirkung: wer im Betriebssystem größere
// Schrift eingestellt hat, bekäme sie hier wieder weggerechnet.

// groessen ist der Prozentsatz je Stufe.
//
// **Vier Stufen und keine freie Zahl.** #245 verlangt „ohne Umbruchschäden";
// das lässt sich für vier Stufen ansehen und für jede Zahl zwischen 50 und 300
// nicht. Die Schritte sind 12,5 % — groß genug, dass man sie sieht, klein
// genug, dass keine Zeile springt.
// **Ein Verbund und keine `Record<string, …>`.** Ein Index mit `string`
// zwänge unter `noUncheckedIndexedAccess` bei jedem Zugriff einen Rückfall —
// und ein `groessen['normal'] ?? '100%'` wäre ein Zweig, den kein Aufruf
// erreicht (§3).
export const groessen = {
  klein: '87.5%',
  normal: '100%',
  gross: '112.5%',
  sehrgross: '125%',
} as const

// Stufe ist eine der vier Größen.
export type Stufe = keyof typeof groessen

// istStufe prüft, was aus dem Bestand kommt.
//
// **Geprüft und nicht gecastet.** `domain.PruefeEinstellung` lässt nur die
// vier durch — aber im Bestand steht, was jemand hineingeschrieben hat, und
// eine fünfte Stufe wäre hier ein `undefined` als Schriftgröße.
export function istStufe(wert: string): wert is Stufe {
  return wert in groessen
}

// setzeErscheinungsbild schreibt die Wahl ans Wurzelelement.
//
// **`system` entfernt das Attribut**, statt es auf „system" zu setzen: der
// Selektor in `stil.css` fragt `html:not([data-erscheinungsbild])`, und ein
// Attribut mit irgendeinem Wert schlösse die Systemeinstellung aus.
export function setzeErscheinungsbild(wurzel: HTMLElement, wahl: string): void {
  if (wahl === 'hell' || wahl === 'dunkel') {
    wurzel.setAttribute('data-erscheinungsbild', wahl)

    return
  }

  wurzel.removeAttribute('data-erscheinungsbild')
}

// setzeSchriftgroesse schreibt die Stufe als CSS-Variable ans Wurzelelement.
//
// **Eine unbekannte Stufe fällt auf `normal` zurück.** Sie kann nur aus einem
// Bestand kommen, den jemand von Hand geändert hat — `domain.PruefeEinstellung`
// lässt sie nicht durch. Eine Oberfläche, die daraufhin gar keine Schriftgröße
// setzt, erbt die des Browsers und sieht zufällig richtig aus.
// schriftgroessenereignis meldet am Wurzelelement, dass die Schriftgröße neu
// gesetzt ist — Nachlauf zum Zweitblick über Phase 6 von 0.9.0, zweiter
// Teil, R5. **Wer in der Wurzelschrift rechnet, hört darauf** statt sie nur
// beim Zeichnen zu lesen: gesetzt wird sie in einem Effekt, also nach dem
// Zeichnen, das sie schon brauchte.
export const schriftgroessenereignis = 'mse-schriftgroesse'

export function setzeSchriftgroesse(wurzel: HTMLElement, stufe: string): void {
  wurzel.style.setProperty(
    '--schriftgroesse',
    istStufe(stufe) ? groessen[stufe] : groessen.normal,
  )
  wurzel.dispatchEvent(new Event(schriftgroessenereignis))
}

// anzeigeschriften ist die Schriftkette je Wahl — Werkstück J zu
// #1325.
//
// **Sieben Schriften, und jede liegt auf dem Zielrechner.** Nachgesehen am
// 2026-09-03 auf macOS 26: `Hoefler Text`, `Luminari`, `Herculanum` und
// `Trattatello` gehören zum Systemumfang; seit UX-09 dazu die drei
// Systemschriften `New York` (`ui-serif`), `SF Pro Rounded` (`ui-rounded`)
// und `SF Pro`. Das ist die Bedingung — die Anwendung arbeitet ohne Netz
// (§16, Regel 5), und eine geladene Schrift wäre entweder ein Binärartefakt
// im Baum oder eine Abhängigkeit vom Netz. **Und keine wird verteilt**
// (UX-09.AK-04): es gibt kein `@font-face`, nur Namen.
//
// **Der Fließtext bekommt nur eine Textschrift.** Eine Kalligraphie über
// einem Absatz ist keine Gestaltung, sondern eine Zumutung
// ([0065](../../docs/entscheidungen/0065-wie-diese-anwendung-aussieht.md),
// Festlegung 6); wer seit UX-09 die ganze Oberfläche umstellt, bekommt eine
// der vier aus [textschriften] und sonst die Systemschrift.
//
// **Jede Kette endet bei einer allgemeinen Familie.** macOS ist das Ziel, aber die Pipeline
// übersetzt auch für Windows und Linux (§12) — dort entscheidet der Rückfall,
// ob eine Überschrift überhaupt eine Serifenschrift ist.
//
// **Und jede trägt japanische und chinesische Rückfälle** — UX-09.AK-02
// (#1681).
// Ein Kartenname aus einem japanischen Set hat Glyphen, die keine dieser
// Schriften kennt; ohne Angabe wählte das System, und für eine Serifenschrift
// womöglich eine serifenlose. **Japanisch vor Chinesisch**, weil Kana nur
// japanische Schriften haben und ein japanischer Name mit einer chinesischen
// zuerst chinesische Zeichenformen bekäme. Alle liegen auf jedem Mac.
//
// **Die Karte bleibt davon unberührt**: sie rastert Go mit den Schriften
// ihrer Vorlage (`internal/render`), nicht das Fenster mit diesen.
//
// **Und in einer chinesischen Umgebung Chinesisch vorn** — G18 aus dem
// Zweitblick über Phase 6.3 (#1664, Phase 6.10). Steht vorn eine Schrift
// ohne Han-Zeichen, nimmt WebKit für ein Han-Zeichen die erste Schrift der
// Kette, die es hat — Hiragino —, und ein chinesischer Name bekäme
// japanische Zeichenformen. Deshalb gibt es je Kette drei Reihen
// ([schriftketten]), und `schriftreihen.css` wählt sie mit `:lang()`.
const cjkSerif = "'Hiragino Mincho ProN', 'Songti SC', 'Songti TC'"
const cjkSans = "'Hiragino Sans', 'PingFang SC', 'PingFang TC'"

export const anzeigeschriften = {
  pergament: `'Hoefler Text', 'Iowan Old Style', Charter, Georgia, ${cjkSerif}, serif`,
  buch: `ui-serif, 'New York', 'Iowan Old Style', Georgia, ${cjkSerif}, serif`,
  rund: `ui-rounded, 'SF Pro Rounded', -apple-system, system-ui, ${cjkSans}, sans-serif`,
  serifenlos: `-apple-system, BlinkMacSystemFont, system-ui, 'Helvetica Neue', ${cjkSans}, sans-serif`,
  fantasy: `Luminari, Papyrus, 'Copperplate', Georgia, ${cjkSerif}, serif`,
  antike: `Herculanum, 'Trajan Pro', 'Optima', Georgia, ${cjkSerif}, serif`,
  kalligrafie: `Trattatello, 'Zapfino', 'Snell Roundhand', Georgia, ${cjkSerif}, serif`,
} as const

// textschriften tragen einen Absatz — die drei übrigen sind Schmuck und
// bleiben an den Überschriften (UX-09.AK-03, Lesbarkeit).
export const textschriften: ReadonlySet<string> = new Set([
  'pergament',
  'buch',
  'rund',
  'serifenlos',
])

// schriftbereichGilt sagt ohne Setzen, was `setzeSchriftbereich` zurückgibt —
// **der Gestaltungsraum braucht es beim Zeichnen**, und ein Zustand, der erst
// im Effekt entsteht, zeichnete zweimal.
export function schriftbereichGilt(bereich: string, wahl: string): boolean {
  return (
    bereich !== 'ganz' || (textschriften.has(wahl) && istAnzeigeschrift(wahl))
  )
}

// setzeSchriftbereich legt die Anzeigeschrift über die ganze Oberfläche —
// oder nimmt sie dort wieder weg. **`false` heißt: eine Schmuckschrift wurde
// nicht über den Text gelegt**, und die Einstellungsseite sagt es.
export function setzeSchriftbereich(
  wurzel: HTMLElement,
  bereich: string,
  wahl: string,
): boolean {
  if (
    bereich === 'ganz' &&
    textschriften.has(wahl) &&
    istAnzeigeschrift(wahl)
  ) {
    setzeKetten(wurzel, '--font-text', schriftketten(wahl))

    return true
  }

  for (const name of reihennamen('--font-text')) {
    wurzel.style.removeProperty(name)
  }

  return bereich !== 'ganz'
}

// Anzeigeschrift ist eine der vier Wahlen.
export type Anzeigeschrift = keyof typeof anzeigeschriften

// istAnzeigeschrift prüft, was aus dem Bestand kommt.
export function istAnzeigeschrift(wert: string): wert is Anzeigeschrift {
  return wert in anzeigeschriften
}

// setzeAnzeigeschrift schreibt die Kette als CSS-Variable ans Wurzelelement.
//
// **Eine unbekannte Wahl fällt auf Pergament zurück** — aus demselben Grund
// wie bei der Schriftgröße: sie kann nur aus einem von Hand geänderten Bestand
// kommen, und gar keine Schrift zu setzen sähe zufällig richtig aus, bis
// jemand die Vorgabe seines Browsers ändert.
export function setzeAnzeigeschrift(wurzel: HTMLElement, wahl: string): void {
  setzeKetten(
    wurzel,
    '--font-anzeige',
    schriftketten(istAnzeigeschrift(wahl) ? wahl : 'pergament'),
  )
}

// CJKReihe sagt, welche Han-Schrift vorn steht: japanisch, vereinfacht oder
// traditionell chinesisch.
export type CJKReihe = 'ja' | 'hans' | 'hant'

// cjkReihe ordnet einer Sprache nach BCP 47 ihre Reihe zu. **Traditionell**
// sind die Schrift `Hant` und die Regionen Taiwan, Hongkong und Macau —
// dieselben, die `stil.css` mit `:lang()` nennt.
export function cjkReihe(sprache: string): CJKReihe {
  const klein = sprache.toLowerCase()

  if (klein !== 'zh' && !klein.startsWith('zh-')) {
    return 'ja'
  }

  return /-(hant|tw|hk|mo)(-|$)/.test(klein) ? 'hant' : 'hans'
}

// cjkBlock trifft die Han-Schriften einer Kette, die hintereinander stehen.
const cjkName = "'(?:Hiragino [^']+|Songti [ST]C|PingFang [ST]C)'"
const cjkBlock = new RegExp(`${cjkName}(?:, ${cjkName})*`)

// rang ist die Stelle einer Han-Schrift in einer Reihe.
const rang: Readonly<Record<CJKReihe, readonly [number, number, number]>> = {
  // japanisch, vereinfacht, traditionell
  ja: [0, 1, 2],
  hans: [2, 0, 1],
  hant: [2, 1, 0],
}

// cjkUmordnen stellt die Han-Schriften einer Kette in die Reihenfolge der
// Reihe; alles andere bleibt, wo es steht.
export function cjkUmordnen(kette: string, reihe: CJKReihe): string {
  return kette.replace(cjkBlock, (block) => {
    const stelle = (name: string) =>
      rang[reihe][name.includes('Hiragino') ? 0 : name.endsWith("SC'") ? 1 : 2]

    return block
      .split(', ')
      .sort((a, b) => stelle(a) - stelle(b))
      .join(', ')
  })
}

// schriftketten gibt eine Wahl in allen drei Reihen.
export function schriftketten(wahl: Anzeigeschrift): Record<CJKReihe, string> {
  const kette = anzeigeschriften[wahl]

  return {
    ja: kette,
    hans: cjkUmordnen(kette, 'hans'),
    hant: cjkUmordnen(kette, 'hant'),
  }
}

// reihennamen sind die drei Variablen einer Kette — die japanische trägt den
// Namen ohne Endung, weil sie gilt, wo keine Sprache etwas anderes sagt.
function reihennamen(name: string): readonly [string, string, string] {
  return [name, `${name}-hans`, `${name}-hant`]
}

// setzeKetten schreibt die drei Reihen ans Wurzelelement.
function setzeKetten(
  wurzel: HTMLElement,
  name: string,
  ketten: Record<CJKReihe, string>,
): void {
  const [ja, hans, hant] = reihennamen(name)

  wurzel.style.setProperty(ja, ketten.ja)
  wurzel.style.setProperty(hans, ketten.hans)
  wurzel.style.setProperty(hant, ketten.hant)
}

// setzeBewegung übersteuert `prefers-reduced-motion` am Wurzelelement —
// Werkstück 8 von Phase 4.12
// (#1351),
// Punkt 11.
//
// # Warum die Übersteuerung nur strenger machen kann
//
// **Weil das System recht hat, wenn es weniger verlangt.** Wer im
// Betriebssystem „weniger Bewegung" eingestellt hat, hat einen Grund; eine
// Anwendung, die das mit einer eigenen Einstellung **aufhebt**, arbeitet
// gegen ihn. Deshalb gibt es nur `reduziert` und `system` — kein „mehr
// Bewegung, egal was das System sagt".
//
// **`system` entfernt das Attribut**, statt es zu setzen: dieselbe Form wie
// bei [setzeErscheinungsbild], und aus demselben Grund — der Selektor in
// `barrierefreiheit.css` fragt auf `[data-bewegung='reduziert']`, und ein Attribut mit
// irgendeinem anderen Wert wäre eine Zusage, die niemand liest.
export function setzeBewegung(wurzel: HTMLElement, wahl: string): void {
  if (wahl === 'reduziert') {
    wurzel.setAttribute('data-bewegung', wahl)

    return
  }

  wurzel.removeAttribute('data-bewegung')
}

// setzeKontrast übersteuert `prefers-contrast` am Wurzelelement.
//
// **Dieselbe Richtung wie [setzeBewegung]**: sie schaltet die Faserstruktur
// zu, nie ab.
export function setzeKontrast(wurzel: HTMLElement, wahl: string): void {
  if (wahl === 'hoch') {
    wurzel.setAttribute('data-kontrast', wahl)

    return
  }

  wurzel.removeAttribute('data-kontrast')
}
