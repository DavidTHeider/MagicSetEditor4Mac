// **Eine gemessene Liste unter jsdom** — Werkstück 2 von Phase 3.5
// (#1166).
//
// # Warum es das braucht
//
// jsdom rechnet kein Layout: `clientHeight` ist 0, und
// `getBoundingClientRect()` gibt lauter Nullen. `useSichtfenster` sieht damit
// **nie** ein Maß und zeichnet, was es ohne Maß zeichnet.
//
// Bis zum 2026-08-27 war das die **ganze** Liste — und deshalb sah jede
// Prüfung alle 3.788 Zeilen, während das Fenster sie ebenfalls alle zeichnete,
// nur um eine davon zu messen. Der Test war grün, weil die Vorrichtung
// großzügiger war als der Betrieb; §3 nennt genau das als die Klasse von
// Fehler, die man so nicht sieht.
//
// **Jetzt zeichnet der Haken ohne Maß eine Probezeile**, und wer eine Liste
// prüfen will, gibt ihr eines. Das ist die richtige Richtung: die Vorrichtung
// stellt her, was der Betrieb ohnehin hat — ein Fenster mit einer Höhe.

// zeilenhoehe und sichthoehe sind die Maße, die diese Vorrichtung vorgibt.
//
// **40 und 4.000**: eine Zeile, wie sie im Fenster gemessen wurde, und ein
// Fenster, das hundert davon fasst. Damit sieht eine Prüfung ihre ganze
// Beispielliste, ohne dass der Haken abgeschaltet wird — gemessen wird
// wirklich, nur eben an Zahlen, die diese Datei setzt.
const zeilenhoehe = 40
const sichthoehe = 4000

// **Was von diesen beiden Zahlen wirklich trägt** — nachgemessen am 2026-08-27
// über zwei Gegenproben (§10):
//
//   - `zeilenhoehe = 0` macht `karten.test.tsx` und `karteanlegen.test.tsx`
//     rot. Sie ist die Zahl, an der das Maß hängt.
//   - `sichthoehe = 200` ändert **nichts**. Das ist keine Nachlässigkeit,
//     sondern Rechnung: `rand = 8` je Seite heißt, dass immer mindestens
//     `1 + 16` Zeilen gezeichnet werden, und keine Beispielliste dieses
//     Projekts ist so lang.
//
// **Die zweite Gegenprobe ist damit keine** — sie belegt nichts über die
// Sichthöhe. Wer den Überhang ändert, muss sie neu führen.

// gibJsdomEinLayout lässt `ul` und `li` eine Höhe haben.
//
// **Auf den Prototypen und nicht auf einem Knoten**: die Zeilen entstehen und
// vergehen beim Rollen, und eine Vorrichtung, die einen einzelnen Knoten
// bestückt, verlöre sie beim ersten Neuzeichnen.
//
// Gibt die Funktion zurück, die es rückgängig macht — `einrichtung.ts` ruft
// sie nach jedem Lauf.
export function gibJsdomEinLayout(): () => void {
  // **Über die Beschreibung und nicht über den Zugriff.** Ein
  // `const x = Element.prototype.getBoundingClientRect` löst die Methode von
  // ihrem Objekt, und `@typescript-eslint/unbound-method` weist es zu Recht
  // ab: wer sie später ohne Empfänger ruft, bekommt ein `this`, das er nicht
  // meinte. Die Beschreibung trägt dieselbe Funktion und diesen Zweifel nicht.
  const vorherRect = Object.getOwnPropertyDescriptor(
    Element.prototype,
    'getBoundingClientRect',
  )
  const vorherHoehe = Object.getOwnPropertyDescriptor(
    Element.prototype,
    'clientHeight',
  )

  Element.prototype.getBoundingClientRect = function messen(
    this: Element,
  ): DOMRect {
    const hoehe = this.tagName === 'LI' ? zeilenhoehe : sichthoehe

    return {
      x: 0,
      y: 0,
      top: 0,
      left: 0,
      right: 0,
      bottom: hoehe,
      width: 0,
      height: hoehe,
      toJSON: () => ({}),
    }
  }

  Object.defineProperty(Element.prototype, 'clientHeight', {
    configurable: true,
    get(this: Element) {
      return this.tagName === 'LI' ? zeilenhoehe : sichthoehe
    },
  })

  return () => {
    /* v8 ignore next 3 -- jsdom bringt beide Eigenschaften mit; der Zweig
       steht für den Fall, dass eine Umgebung das einmal nicht tut */
    if (vorherRect !== undefined) {
      Object.defineProperty(
        Element.prototype,
        'getBoundingClientRect',
        vorherRect,
      )
    }

    if (vorherHoehe === undefined) {
      // **Ohne Beschreibung gab es die Eigenschaft nicht** — dann muss sie
      // wieder verschwinden, statt mit einem Ausweichwert dazubleiben.
      delete (Element.prototype as unknown as Record<string, unknown>)
        .clientHeight
    } else {
      Object.defineProperty(Element.prototype, 'clientHeight', vorherHoehe)
    }
  }
}
