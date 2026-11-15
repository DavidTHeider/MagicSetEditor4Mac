// **Die Zeichenfläche des Symboleditors** —
// #890.
//
// # Warum SVG und nicht ein Canvas
//
// Weil **jeder Punkt ein Element sein muss**. §6 verlangt durchgängige
// Tastaturbedienbarkeit, und das Issue nennt sie bei einem Zeichenwerkzeug
// „die schwerste Zusage dieser Phase".
//
// Auf einem Canvas gibt es nichts, worauf ein Tabulator springen könnte — man
// müsste die Auswahl daneben führen und jede Taste selbst deuten. In einem SVG
// ist ein Punkt ein `<circle>` mit `tabIndex`, und der Browser macht den Rest:
// Fokusreihenfolge, Fokusring, Bildschirmleser.
//
// **Der Preis steht dabei**: ein SVG mit tausend Punkten wäre langsamer als
// ein Canvas. Ein Setsymbol hat Dutzende.

import {
  useCallback,
  useRef,
  type KeyboardEvent,
  type PointerEvent,
} from 'react'

import type { SymbolteilDTO, SymbolzugDTO } from '../antworten'
import { Symbolanfasser } from './Symbolanfasser'
import { Teilumriss } from './Teilumriss'
import { useTexte } from '../sprache/kontext'

// groesse ist die Kantenlänge der Fläche in Bildpunkten.
//
// **Eine `.mse-symbol` rechnet in `[0…1]`** — die Zahl hier ist reine
// Darstellung und steht deshalb an einer Stelle.
const groesse = 320

// Die Klassen der Fläche, klein und im größeren Zustand der Schublade.
const flaechenrahmen =
  'shrink-0 rounded-sm border border-kante bg-flaeche dark:border-kante'
const flaecheKlein = 'h-80 w-80 ' + flaechenrahmen
const flaecheGross = 'h-[32rem] w-[32rem] ' + flaechenrahmen

// schritt ist, wie weit eine Pfeiltaste einen Punkt bewegt.
//
// **Ein Hundertstel**, und mit Umschalt ein Zehntel. Bei 320 Bildpunkten sind
// das gut drei Punkte — fein genug, um etwas auszurichten, grob genug, um in
// wenigen Anschlägen über die Fläche zu kommen.
const schritt = 0.01
const grosserSchritt = 0.1

interface SymbolflaecheEigenschaften {
  // gross zeichnet die Fläche im größeren Zustand der Schublade (AP-12,
  // #2369). Die Punkte rechnen über die Maße des Elements, nicht über
  // `groesse` — die Fläche wächst also nur im Bild.
  gross?: boolean
  teile: SymbolteilDTO[]
  gewaehlt: number

  // mitgewaehlt sind **alle** gewählten Stellen — die Fläche hebt sie hervor.
  //
  // **Zwei Angaben und keine Verdopplung**: `gewaehlt` ist die eine Stelle,
  // an der gearbeitet wird (Griffe, Symmetrieanfasser), `mitgewaehlt` ist die
  // Wahl. So hält es MSE auch: `getAShape()` greift eine aus der Menge
  // (`src/gui/symbol/selection.cpp`, Zeilen 66–71, Fassung `f9d9356` vom
  // 2021-07-06), gezeichnet wird die ganze.
  mitgewaehlt: readonly number[]
  verschieben: (teil: number, punkt: number, dx: number, dy: number) => void

  // setzenAuf stellt einen Punkt **auf** eine Stelle im Einheitsquadrat.
  //
  // **Beides nebeneinander, und beides mit Grund.** Die Tastatur meint einen
  // Schritt; der Zeiger meint eine Stelle. Wer den Zeiger als Folge von
  // Schritten liest, braucht einen Ausgangswert — und der ist ab der zweiten
  // Bewegung alt (#890).
  setzenAuf: (teil: number, punkt: number, x: number, y: number) => void

  // zugBeginnt sagt an, dass ein neuer Zug anfängt — MSEs `resetActions()`
  // (`src/gui/symbol/point_editor.cpp` Zeile 358, Fassung `f9d9356` vom
  // 2021-07-06).
  //
  // **Ohne diesen Ruf wären zwei Züge an demselben Punkt ein Schritt der
  // Rücknahme.** Die vielen Bewegungen **eines** Zuges sollen zusammenfallen;
  // was zwei Züge trennt, ist das Aufsetzen.
  zugBeginnt: () => void

  waehlen: (teil: number) => void

  // einfuegen setzt einen Punkt hinter den genannten, loeschen entfernt ihn.
  //
  // **Sie stehen hier und nicht nur in der Hilfe.** Die Tastaturhilfe nannte
  // Eingabetaste und Rückschritt, bevor die Seite sie konnte — eine Zusage in
  // einem Oberflächentext ist eine Zusage (§10).
  einfuegen: (teil: number, punkt: number) => void
  loeschen: (teil: number, punkt: number) => void

  // griffAuf zieht einen der beiden Bézierhandgriffe eines Punktes auf eine
  // Stelle.
  //
  // **Sie sind sichtbar und bedienbar, oder es gibt keine Kurven.** Ein
  // Kreispunkt hat zwei Griffe; ohne sie ließe sich eine Rundung nur über die
  // Grundform herstellen und danach nie wieder ändern.
  griffAuf: (
    teil: number,
    punkt: number,
    davor: boolean,
    x: number,
    y: number,
  ) => void

  // **Die Kopien der gewählten Symmetrie** —
  // #1092.
  //
  // Sie kommen fertig gerechnet von der Go-Seite; hier werden sie nur gemalt.
  kopien: SymbolzugDTO[][]

  mitteAuf: (x: number, y: number) => void
  symmetriegriffAuf: (x: number, y: number) => void
}

export function Symbolflaeche({
  gross = false,
  teile,
  gewaehlt,
  mitgewaehlt,
  verschieben,
  setzenAuf,
  zugBeginnt,
  waehlen,
  einfuegen,
  loeschen,
  griffAuf,
  kopien,
  mitteAuf,
  symmetriegriffAuf,
}: SymbolflaecheEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbol

  // ziehen merkt sich, welcher Punkt gerade am Zeiger hängt.
  //
  // **Ein `ref` und kein Zustand**: er ändert sich bei jeder Mausbewegung, und
  // ein Neuzeichnen je Pixel wäre der Preis für eine Auskunft, die niemand
  // sieht.
  const amZeiger = useRef<{
    teil: number
    punkt: number
    griff: 'davor' | 'danach' | null
  } | null>(null)

  // ausSicht rechnet Zeigerkoordinaten in das Einheitsquadrat um.
  //
  // **Über die Maße des Elements und nicht über `groesse`.** Das SVG skaliert
  // mit seinem Kasten; wer die Zahl aus dem `viewBox` nähme, verschöbe den
  // Punkt um den Skalierungsfaktor.
  const ausSicht = useCallback(
    (ereignis: PointerEvent<SVGSVGElement | SVGCircleElement>) => {
      const kasten =
        ereignis.currentTarget.ownerSVGElement?.getBoundingClientRect()
      const eigener = ereignis.currentTarget.getBoundingClientRect()
      const flaeche = kasten ?? eigener

      return {
        x: (ereignis.clientX - flaeche.left) / flaeche.width,
        y: (ereignis.clientY - flaeche.top) / flaeche.height,
      }
    },
    [],
  )

  const zeigerRunter = (
    ereignis: PointerEvent<SVGCircleElement>,
    teilstelle: number,
    punktstelle: number,
    griff: 'davor' | 'danach' | null = null,
  ) => {
    amZeiger.current = { teil: teilstelle, punkt: punktstelle, griff }
    zugBeginnt()
    waehlen(teilstelle)

    // **Der Zeiger wird eingefangen.** Ohne das verliert das Element ihn,
    // sobald er es verlässt — und der Punkt bliebe am Rand des Kreises kleben.
    ereignis.currentTarget.setPointerCapture(ereignis.pointerId)
  }

  // zeigerBewegt reicht die **Stelle** weiter, an der der Zeiger steht.
  //
  // **Und nicht die Strecke, die er zurückgelegt hat.** Eine Differenz gegen
  // den Punkt, wie er beim Aufsetzen dastand, veraltete mit der ersten
  // Bewegung: die zweite rechnete wieder gegen den Anfang und addierte das
  // Ergebnis auf den schon gezogenen Punkt. Von 0,5 über 0,6 nach 0,7 käme
  // 0,8 heraus, und mit jeder weiteren Bewegung liefe der Punkt dem Zeiger
  // weiter davon.
  //
  // **Wer die Stelle umrechnet, hat den heutigen Wert in der Hand**: für den
  // Punkt die Seite, die den Bauplan hält, für den Griff die Go-Seite, wo
  // ohnehin die Sperre sitzt.
  const zeigerBewegt = (ereignis: PointerEvent<SVGSVGElement>) => {
    const gezogen = amZeiger.current
    if (gezogen === null) {
      return
    }

    const wohin = ausSicht(ereignis)

    if (gezogen.griff === null) {
      setzenAuf(gezogen.teil, gezogen.punkt, wohin.x, wohin.y)

      return
    }

    griffAuf(
      gezogen.teil,
      gezogen.punkt,
      gezogen.griff === 'davor',
      wohin.x,
      wohin.y,
    )
  }

  const zeigerHoch = () => {
    amZeiger.current = null
  }

  const taste = (
    ereignis: KeyboardEvent<SVGCircleElement>,
    teilstelle: number,
    punktstelle: number,
  ) => {
    const weite = ereignis.shiftKey ? grosserSchritt : schritt

    const richtungen: Record<string, [number, number]> = {
      ArrowLeft: [-weite, 0],
      ArrowRight: [weite, 0],
      ArrowUp: [0, -weite],
      ArrowDown: [0, weite],
    }

    // **Einfügen und Löschen kommen zuerst**, weil sie keine Richtung haben.
    if (ereignis.key === 'Enter' || ereignis.key === 'Backspace') {
      ereignis.preventDefault()
      waehlen(teilstelle)

      if (ereignis.key === 'Enter') {
        einfuegen(teilstelle, punktstelle)
      } else {
        loeschen(teilstelle, punktstelle)
      }

      return
    }

    const richtung = richtungen[ereignis.key]
    if (richtung === undefined) {
      return
    }

    // **Der Browser scrollt sonst mit.** Eine Pfeiltaste auf einem
    // fokussierten Element bewegt bei ihm die Seite; hier bewegt sie den
    // Punkt, und beides zugleich wäre eine Fläche, die unter dem Zeiger
    // wegläuft.
    ereignis.preventDefault()

    waehlen(teilstelle)
    verschieben(teilstelle, punktstelle, richtung[0], richtung[1])
  }

  return (
    // **Eine benannte Gruppe und kein Bild** — G17 aus dem Zweitblick über
    // Phase 6.3: die Kinder eines `role="img"` sind für Hilfsmittel nur
    // Darstellung, und die Punkte, die mit den Pfeiltasten gehen,
    // verschwanden für VoiceOver darin.
    <svg
      role="group"
      aria-label={eigene.zeichnung}
      viewBox={`0 0 ${String(groesse)} ${String(groesse)}`}
      className={gross ? flaecheGross : flaecheKlein}
      onPointerMove={zeigerBewegt}
      onPointerUp={zeigerHoch}
      onPointerCancel={zeigerHoch}
    >
      {rueckwaerts(teile).map(({ teil, teilstelle }) => (
        <Teilumriss
          key={teil.name + String(teilstelle)}
          teil={teil}
          groesse={groesse}
          klassen={
            mitgewaehlt.includes(teilstelle)
              ? 'fill-auswahl/60 stroke-auswahlkante'
              : 'fill-betont/60 stroke-leise'
          }
        />
      ))}

      {/* **Die Anfasser kommen zuletzt** — MSEs `drawEditingHints` nach allen
          Teilen (`src/render/symbol/viewer.cpp`, Zeilen 109–116, Fassung
          `f9d9356` vom 2021-07-06). In Listenreihenfolge, damit die des
          gewählten Teils nicht unter der Füllung eines anderen liegen. */}
      {teile.map((teil, teilstelle) => (
        <Symbolanfasser
          key={teil.name + String(teilstelle)}
          teil={teil}
          teilstelle={teilstelle}
          gewaehlt={teilstelle === gewaehlt}
          groesse={groesse}
          schritt={schritt}
          grosserSchritt={grosserSchritt}
          kopien={kopien}
          mitteAuf={mitteAuf}
          symmetriegriffAuf={symmetriegriffAuf}
          griffAuf={griffAuf}
          zeigerRunter={zeigerRunter}
          taste={taste}
          waehlen={waehlen}
        />
      ))}
    </svg>
  )
}

// rueckwaerts gibt die Teile in **Zeichenreihenfolge** — das letzte zuerst.
//
// **Das erste Teil der Liste liegt im Bild oben**, und SVG malt das Spätere
// über das Frühere; der Grund und die Fundstelle stehen in `Teilumriss.tsx`.
// Die Stelle wird mitgeführt, weil Griffe und Wahl an ihr hängen und nicht an
// der Malreihenfolge.
function rueckwaerts(
  teile: SymbolteilDTO[],
): { teil: SymbolteilDTO; teilstelle: number }[] {
  return teile.map((teil, teilstelle) => ({ teil, teilstelle })).reverse()
}
