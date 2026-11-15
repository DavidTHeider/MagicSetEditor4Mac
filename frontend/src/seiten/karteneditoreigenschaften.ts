// **Was der Karteneditor von außen bekommt** — aus `Karteneditor.tsx`
// herausgelöst, damit die Komponente zeigt, was sie tut, und nicht zuerst,
// was sie annimmt (§2, die Zeilengrenze).
import type { Bildstand } from '../haken/useKartenvorschau'
import type { FeldbeschreibungDTO, KarteDTO } from '../api'
import type { Ansichtswahl } from './vorschauansicht'

// Feldwahl ist das gewählte Feld samt seinem Setzer — genau die Gestalt, die
// `useState<string>()` gibt, damit beide Wege dieselbe sind.
export type Feldwahl = [string | undefined, (feld: string | undefined) => void]

export interface KarteneditorEigenschaften {
  // setID ist das Set, gegen das der Dienst jede Kartenänderung hält
  // (#808).
  setID: string

  karte: KarteDTO

  // formen sagen, welches Steuerelement ein Feld bekommt — Art,
  // Mehrzeiligkeit, Wahlen, Beschreibung
  // (#560).
  //
  // **Von außen geladen**: sie gehören zum **Set**. Der Editor wird bei jedem
  // Kartenwechsel neu gebaut; ein eigener Lauf holte sie bei jedem Klick neu,
  // und bis dahin stünde jedes Feld einzeilig da und spränge dann um.
  formen: readonly FeldbeschreibungDTO[]

  // geaendert meldet, dass ein Feldwert **außerhalb der Eingabe** entstand —
  // beim Einsetzen eines Bildes vergibt die Ablage den Dateinamen (#246).
  // **Nicht beim gewöhnlichen Schreiben**: dort steht der Wert schon im Feld.
  geaendert: () => void

  // symbolZeichnen öffnet den Symboleditor für ein Feld der Art `symbol`.
  //
  // **Wahlfrei**, und ohne ihn steht kein Knopf da — der Grund steht bei
  // [Feldzeile]. Der Aufrufer bekommt den **Feldnamen**, denn nur er weiß,
  // wohin der Editor führt
  // (#1339).
  symbolZeichnen?: ((feld: string) => void) | undefined

  // feldwahl und ansicht halten **eine Ebene höher**, was den absichtlichen
  // Neuaufbau dieses Editors übersteht: das Feld, in dem jemand steht
  // (#1075),
  // und Maßstab und Überlagerungen der Vorschau (UXR-06, `vorschauansicht.ts`).
  //
  // **Wahlfrei**: wer den Editor allein fährt, hält beides selbst.
  feldwahl?: Feldwahl
  ansicht?: Ansichtswahl

  // formularZiel ist der Knoten des Inspectors im Arbeitsplatz — Werkstück
  // U5 von Phase 5.1 (#1416). Das Formular wird dorthin **portiert**, damit
  // der Editor der eine Ort seines Zustands bleibt und die Vorschau im
  // Zentrum steht. `undefined` heißt: kein Arbeitsplatz, beides nebeneinander;
  // `null` heißt: der Inspector ist eingeklappt, und das Formular bleibt weg.
  formularZiel?: HTMLElement | null | undefined

  // stand ist der Verlaufsstand: er liest nach, **ohne** neu zu bauen (U01).
  stand?: number

  // vorigesBild ist das zuletzt gezeigte Bild, über dem Schlüssel dieses
  // Editors gehalten, und bildGezeigt meldet jedes neue dorthin — damit die
  // Vorschau beim Kartenwechsel nicht leer blitzt (AP-4, #2361).
  vorigesBild?: Bildstand | undefined
  bildGezeigt?: ((stand: Bildstand) => void) | undefined
}
