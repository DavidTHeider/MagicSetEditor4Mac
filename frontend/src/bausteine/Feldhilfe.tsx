// **Das Hilfesymbol am Feld** — UX-15.AK-02
// (#1771).
//
// # Wofür es das gibt
//
// **Weil unter jedem Feld zwei bis fünf Zeilen standen** und die Seite
// dadurch lang wurde. Was beim Entscheiden hilft, bleibt sichtbar; die
// Begründung steht hier — erreichbar über Zeiger **und** Tastatur, und für
// eine Vorlesehilfe als Beschreibung des Knopfes.
//
// # Was es nicht ist
//
// **Kein Ort für etwas, das man wissen muss.** Das Issue sagt es: der Text
// hinter dem Symbol „ist nicht die einzige Quelle einer Information, die zum
// Entscheiden nötig ist". Wer eine Grenze, eine Einheit oder eine Bedingung
// dorthin verschiebt, hat sie versteckt und nicht aufgeräumt — das ist eine
// Frage ans Review und keine an einen Wächter.
//
// # Warum ein Knopf und kein Zeichen mit `title`
//
// **Weil ein `title` unter dem Zeiger erscheint und nirgends sonst.** Der
// Grund steht ausführlich im Kopf von [Tooltip]; hier steht nur, dass ein
// Fragezeichen ohne Bedienelement keinen Fokus bekommt.
import { Tooltip } from './Tooltip'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

// Hilfeeigenschaften nennt **entweder** das Feld **oder** den fertigen Namen.
//
// **Als Vereinigung und nicht als zwei Felder mit Vorbehalt**: ein `zu ?? ''`
// daneben wäre ein Zweig, den kein Aufruf erreicht — die Abdeckung hat ihn
// angestrichen (§3).
//
// **Warum es beides gibt**: am Feld heißt der Knopf nach dem Feld, am
// Erläuterungsabsatz nach dem Satz, den er fortsetzt — dreißig Abschnitte
// heißen nicht überall gleich.
type Hilfeeigenschaften = { text: string } & (
  | { zu: string; bezeichnung?: undefined }
  | { bezeichnung: string; zu?: undefined }
)

export function Feldhilfe(eigenschaften: Hilfeeigenschaften) {
  const eigene = useTexte().bausteine

  const { text } = eigenschaften
  const name = hilfename(eigene, eigenschaften)

  return (
    // **Der Druck will die Erklärung sehen** (#2068) — anders als an einem
    // Knopf, der etwas tut, räumt er sie hier nicht.
    <Tooltip text={text} breite="breit" druck="bleibt">
      {(griffe) => (
        <button
          type="button"
          aria-label={name}
          className={knopfklassen}
          {...griffe}
        >
          {/* **Ein Fragezeichen und kein Bild** — es trägt keine Bedeutung
              für sich; die steht im Namen des Knopfes, und den liest eine
              Vorlesehilfe. Deshalb `aria-hidden`. */}
          <span aria-hidden>{Fragezeichen}</span>
        </button>
      )}
    </Tooltip>
  )
}

// hilfename gibt den Namen des Knopfes.
//
// **Als `if` und nicht als `? :`**: die Vereinigung wird am Vorhandensein
// von `bezeichnung` unterschieden, und nur eine Verzweigung verengt `zu` auf
// `string`. Ein `??` bräuchte daneben ein `zu ?? ''` — und das wäre der
// Zweig, den kein Aufruf erreicht (§3).
function hilfename(
  eigene: { feldhilfe: string },
  eigenschaften: Hilfeeigenschaften,
): string {
  if (eigenschaften.bezeichnung !== undefined) {
    return eigenschaften.bezeichnung
  }

  return einsetzen(eigene.feldhilfe, { feld: eigenschaften.zu })
}

// Fragezeichen ist das Zeichen auf dem Knopf.
//
// **Als Konstante und nicht im JSX**: `keintextimjsx` weist jeden Textknoten
// ab, und das zu Recht — dies ist die benannte Ausnahme, weil ein
// Fragezeichen in jeder Sprache eines ist (§9, wie der Strich am
// Mengenknopf).
const Fragezeichen = '?'

const knopfklassen =
  'inline-flex h-5 w-5 items-center justify-center rounded-full border ' +
  'border-kante text-xs leading-none text-leise ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'
