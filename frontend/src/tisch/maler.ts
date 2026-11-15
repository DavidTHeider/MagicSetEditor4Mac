// **Die Naht zwischen Plan und Grafikkarte** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Warum hier kein `pixi.js` steht
//
// **Weil dann jeder Test es lüde.** Diese Datei entscheidet, *ob* gezeichnet
// werden kann, und reicht das Zeichnen an eine Naht weiter; die Naht selbst
// holt PixiJS **dynamisch** und ist die einzige Stelle im Fenster, die den
// Namen kennt. Dieselbe Gestalt wie `internal/app/fensternaht.go`: die
// Entscheidung wird gemessen, der Ruf nach draußen ist ein Feld.
//
// # Und warum es überhaupt ein „ob" gibt
//
// **Weil WebGL fehlen kann.** #1460 verlangt es wörtlich: „UI bleibt bei
// fehlendem Renderer mit verständlichem Fehler bedienbar". Ein Tisch, der
// ohne Grafikkarte einfach leer bliebe, wäre die halbe Zusage, und die ist
// schlimmer als gar keine (§10).
import type { Kamera } from './kamera'
import type { Tischfarben } from './tischfarben'
import type { Tischlage, Tischplan } from './tischplan'
import type { Rechteck } from './tischausschnitt'
import type { Wirkungsart } from './tischwirkungen'

// Bildquelle sagt, welches Bild zu einer Lage gehört — oder keines.
//
// **Sie wird gefragt und nicht mitgeliefert**, weil Texturen später kommen
// als der Plan: G2 holt sie über die Naht, und bis dahin zeichnet der Maler
// einen Platzhalter.
// Bildquelle gibt die Adresse des Bildes zu einer Lage.
//
// **Der zweite Parameter sagt, was gemeint ist** — Befund (a) des Entwurfs
// der Zieloberfläche, PLAY-03.AK-07
// (#1711).
// Ist er wahr, liegt die Karte verdeckt oder darf diese Sicht nicht
// hineinsehen; gefragt wird dann nach dem **Rücken des Spiels**, der keiner
// Karte gehört. So bleibt die Geheimhaltung an einer Stelle, die man prüfen
// kann, statt daran zu hängen, dass niemand fragt.
export type Bildquelle = (
  lage: Tischlage,
  verdeckt: boolean,
) => string | undefined

// Zeichenwunsch ist alles, was ein Bild braucht.
//
// **Als Verbund und nicht als Reihe von Parametern.** Der erste Anlauf reichte
// drei Werte durch; mit der Auswahl wären es vier, mit dem Tischbild fünf, und
// jeder neue änderte jede Naht und jedes Doppel. Ein Feld dazu ändert nichts.
export interface Zeichenwunsch {
  plan: Tischplan
  kamera: Kamera
  bild: Bildquelle

  // gewaehlt ist die Karte, die hervorgehoben wird — oder keine.
  //
  // **Sie steht hier und nicht im Plan**, weil sie niemandem außer diesem
  // Fenster gehört: wer eine Karte ansieht, ändert am Spiel nichts.
  gewaehlt?: string | undefined

  // hintergrund ist die Adresse eines eigenen Tischbilds — leer heißt: der
  // gewählte Farbsatz.
  //
  // **Eine Adresse und keine Bytes**, aus demselben Grund wie bei den
  // Kartenbildern: ein Hintergrund von 1920 Punkten käme als Base64 bei jedem
  // Öffnen des Tisches über die Brücke.
  hintergrund?: string | undefined

  // bewegt sagt, ob Karten an ihren neuen Platz **fliegen** sollen.
  //
  // **Ein Bild ohne Bewegung zeigt denselben Stand** — #1460: „Animation
  // aus/ein erzeugt denselben fachlichen Eventstrom." Hier ist das keine
  // Zusage, sondern eine Folge: eine Bewegung ist Darstellung, und die
  // schickt nichts an den Wirt.
  bewegt?: boolean | undefined

  // wirkungen sagt je Kennung, was gerade geschehen ist — PLAY-04.AK-02
  // (#1712).
  //
  // **Hier und nicht im Plan**, aus demselben Grund wie `gewaehlt` darüber:
  // eine Wirkung gehört dem Auge und nicht der Partie. Zwei Fenster am
  // selben Tisch sehen verschiedene — wer gerade erst aufgeschlagen hat,
  // sieht keine.
  wirkungen?: Map<string, Wirkungsart> | undefined

  // beschriftung gibt den Namen einer Zone — PLAY-03.AK-06.
  //
  // **Von außen, weil eine Zone in zwei Sprachen zwei Namen hat.** Die
  // Kataloge liegen in `sprache/`; ein Zeichner, der sie läse, hätte die
  // Oberflächensprache im Renderer. Fehlt sie, bleibt das Feld unbeschriftet
  // — der Tisch ist dann so, wie er vor PLAY-03 war, und nicht kaputt.
  beschriftung?: ((zone: string, kurz?: boolean) => string) | undefined

  // sitzkopf gibt, was am Sitz steht — PLAY-03.
  //
  // **Von außen, aus demselben Grund wie `beschriftung`**: „Leben" heißt in
  // zwei Sprachen zweierlei. Fehlt es, trägt der Tisch keine Sitzköpfe — er
  // ist dann so, wie er vorher war, und nicht kaputt.
  sitzkopf?: ((sitz: string) => string) | undefined

  // beschriftungsstufe ist der Wert von `tisch.beschriftung` — aus, klein,
  // normal, groß (`schildmass.ts`). Fehlt sie, steht die Schrift mittel.
  beschriftungsstufe?: string | undefined
}

// Tischmaler bringt einen Plan auf eine Fläche.
export interface Tischmaler {
  zeichne(wunsch: Zeichenwunsch): void

  // masse sind die Maße der Zeichenfläche in Punkten.
  masse(): { breite: number; hoehe: number }

  // passeAn bringt die Zeichenfläche **jetzt** auf die Größe ihres Behälters
  // und gibt die neuen Maße.
  //
  // **Jetzt und nicht im nächsten Bild**: PixiJS folgt dem Behälter von
  // selbst erst dort (`queueResize`, pixi.js 8.20.1), und wer die Maße im
  // selben Augenblick las, bekam die alten — der Tisch stand nach einer
  // Größenänderung abgeschnitten da, gesehen am 2026-09-24 in der
  // WebKit-Probe (#1828).
  passeAn(): { breite: number; hoehe: number }

  // faerbe tauscht die Farbtafel — für einen Wechsel des Tischbilds.
  //
  // **Und wirft dafür keinen Renderer weg**: das erste Bild eines neuen
  // Kontexts kostet das vierundvierzigfache eines gewöhnlichen
  // (`docs/TISCHLEISTUNG.md`).
  faerbe(farben: Tischfarben): void

  // abzug gibt den Ausschnitt als PNG-Data-URL — PLAY-05.AK-03 und AK-04
  // (#1713).
  //
  // **Aus dem stehenden Bild und nicht aus einer laufenden Bewegung** — AK-04
  // wörtlich: „aus einem konsistenten Snapshot, nicht aus halb laufender
  // Animation". Der Zeichner hält dafür an: er rendert den Ausschnitt in eine
  // eigene Textur, und die entsteht aus dem Szenenbaum, wie er **jetzt**
  // steht.
  //
  // **Der Ausschnitt ist in Millimetern** — derselben Welt, in der der Plan
  // rechnet —, und die Auflösung sagt, wie viele Bildpunkte ein Millimeter
  // wird.
  abzug(ausschnitt: Rechteck, punkteJeMillimeter: number): Promise<string>

  // beiVerlust meldet, dass der Grafikkontext weggefallen ist.
  beiVerlust(melde: () => void): void

  raeume(): void
}

// Malhindernis sagt, **warum** nicht gezeichnet wird.
//
// Drei Gründe, und sie sind verschieden: der Rechner kann kein WebGL, der
// Renderer kam nicht hoch, oder er ist unterwegs weggefallen. Ein einziger
// Satz für alle drei nähme dem Lesenden die Auskunft, ob ein neuer Versuch
// lohnt.
export type Malhindernis = 'kein-webgl' | 'kein-renderer' | 'verloren'

export type Tischmalerei =
  { art: 'da'; maler: Tischmaler } | { art: 'ohne'; grund: Malhindernis }

// Malnaht ist der Weg zu einem echten Maler. **Sie bekommt eine eigene
// Leinwand** — siehe `tischmalerei`.
export type Malnaht = (leinwand: HTMLCanvasElement) => Promise<Tischmaler>

// Probeshader ist der Fragmentshader, bei dem `checkMaxIfStatementsInShader`
// aus PixiJS 8.20.1 spätestens anhält — der mit einem einzigen Zweig, also
// ohne ein `if`. Übersetzt er nicht, hält PixiJS nie an.
const Probeshader =
  'precision mediump float;\nvoid main(void){\nfloat test = 0.1;\n\ngl_FragColor = vec4(0.0);\n}'

// grafikprobe sagt, ob diese Umgebung einen Tisch zeichnen kann — und wenn
// nicht, woran es liegt.
//
// **Zwei Fassungen werden gefragt**, weil `webgl2` auf älteren Rechnern fehlt
// und PixiJS dann auf `webgl` zurückfällt. Und alles steht in einem Riegel:
// manche Browser werfen beim Anlegen des Kontexts, statt `null` zu geben.
//
// **Und der Kontext muss übersetzen** — UI-N01 (#2089). PixiJS 8.20.1
// halbiert bei jedem gescheiterten Shader bis null und versucht es dann ohne
// Ende weiter; ein Kontext, der nichts übersetzt, hält die Webansicht an.
// Ein Versuch hier ist begrenzte Arbeit, und danach kommt PixiJS gar nicht
// erst dran. **Was das nicht hält**: einen Kontext, der erst während des
// Aufbaus verloren geht — dann hilft nur die eigene Leinwand je Anlauf.
//
// **Der Probekontext wird freigegeben.** WebKit hält nur eine begrenzte Zahl
// offen und nimmt dann dem ältesten seinen — und das wäre der des Tischs.
export function grafikprobe(dokument: Document): Malhindernis | undefined {
  let gl: WebGLRenderingContext | WebGL2RenderingContext | null
  try {
    const probe = dokument.createElement('canvas')
    gl = probe.getContext('webgl2') ?? probe.getContext('webgl')
  } catch {
    return 'kein-webgl'
  }
  if (!gl) return 'kein-webgl'

  try {
    if (gl.isContextLost()) return 'kein-renderer'

    const shader = gl.createShader(gl.FRAGMENT_SHADER)
    if (!shader) return 'kein-renderer'
    gl.shaderSource(shader, Probeshader)
    gl.compileShader(shader)
    const uebersetzt = gl.getShaderParameter(shader, gl.COMPILE_STATUS) === true
    gl.deleteShader(shader)

    return uebersetzt ? undefined : 'kein-renderer'
  } finally {
    gl.getExtension('WEBGL_lose_context')?.loseContext()
  }
}

// tischmalerei versucht einen Maler und sagt sonst, woran es lag.
//
// **Jeder Anlauf bekommt eine eigene Leinwand im Ort**, und sein Abräumen
// nimmt sie wieder heraus — UI-N01 (#2089). Bis zum 2026-10-06 bekam
// PixiJS die Leinwand, die React hält. Sein `destroy` nahm sie aus dem
// Dokument und verlor ihren Kontext (`ViewSystem`, `GlContextSystem`,
// 8.20.1); `Activity` behielt den Baum, und der nächste Anlauf bekam
// dieselbe Leinwand mit dem verlorenen Kontext. Gemessen am nativen
// Fenster: nach dem Verlassen nicht mehr im Dokument, `isContextLost()`
// wahr, `COMPILE_STATUS` `null` — und der zweite Einstieg stand.
export async function tischmalerei(
  ort: HTMLElement,
  naht: Malnaht,
): Promise<Tischmalerei> {
  const hindernis = grafikprobe(ort.ownerDocument)
  if (hindernis) return { art: 'ohne', grund: hindernis }

  const leinwand = ort.ownerDocument.createElement('canvas')
  leinwand.className = 'block h-full w-full'
  ort.append(leinwand)

  try {
    const maler = await naht(leinwand)

    return {
      art: 'da',
      maler: {
        ...maler,
        raeume() {
          maler.raeume()
          leinwand.remove()
        },
      },
    }
  } catch {
    // **Der Grund wird nicht durchgereicht.** Was PixiJS wirft, ist für
    // Menschen unlesbar; was hier zählt, ist die Auskunft „kein Renderer",
    // und die trägt die Oberfläche in ihren eigenen Worten vor.
    leinwand.remove()

    return { art: 'ohne', grund: 'kein-renderer' }
  }
}
