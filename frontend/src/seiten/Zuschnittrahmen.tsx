// **Der Rahmen über dem Ursprungsbild** —
// #576.
//
// MSEs `ImageSliceSelector` (`src/gui/image_slice_window.cpp`, Zeilen 442–690,
// Fassung f9d9356 vom 2021-07-06): das Bild im Maßstab des Fensters, darüber
// der gewählte Bereich, an seinen Ecken sechs mal sechs Punkte große Griffe.
//
// **Was hier gerechnet wird, ist der Maßstab und sonst nichts.** Wohin ein Zug
// führt, entscheidet Go (`Zuschnittziehen`) — die Regel ist portiert, und eine
// zweite Fassung hier gäbe eine zweite Antwort (§15).
//
// **Eigene Datei**, weil `Zuschnittfenster.tsx` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, was das Fenster zeigt, hier,
// wie der Rahmen auf Zug reagiert.
import { useRef } from 'react'

import type { AuswahlDTO, ZuschnittAntwort } from '../api'
import { useTexte } from '../sprache/kontext'

interface ZuschnittrahmenEigenschaften {
  zuschnitt: ZuschnittAntwort
  auswahl: AuswahlDTO
  frei: boolean
  ziehen: (x: number, y: number, kanteX: number, kanteY: number) => void
}

// ecken sind die vier Griffe. **Nur Ecken**, solange das Seitenverhältnis
// steht — *„only drag corners if aspect fixed"* (Zeile 596).
const ecken: { x: number; y: number; wo: string }[] = [
  { x: -1, y: -1, wo: 'left-0 top-0 -translate-x-1/2 -translate-y-1/2' },
  { x: 1, y: -1, wo: 'right-0 top-0 translate-x-1/2 -translate-y-1/2' },
  { x: -1, y: 1, wo: 'left-0 bottom-0 -translate-x-1/2 translate-y-1/2' },
  { x: 1, y: 1, wo: 'right-0 bottom-0 translate-x-1/2 translate-y-1/2' },
]

// **Die Farben sind gemessen** (`docs/KONTRASTE.md`): der Rahmen steht auf dem
// Bild, und ein Bild kann jede Farbe haben. Deshalb zwei Linien — eine dunkle
// innen, eine helle außen —, wie MSE es tut:
//
//	dc.SetPen(*wxBLACK_PEN);
//	dc.DrawRectangle(left - 1, top - 1, width + 2, height + 2);
//	dc.SetPen(Color(64,64,64));
//	dc.DrawRectangle(left - 2, top - 2, width + 4, height + 4);
//
// `src/gui/image_slice_window.cpp`, Zeilen 504–507, Fassung f9d9356 vom
// 2021-07-06.
//
// **`outline-offset-2` ist der Punkt an der Sache**, und er ist von der
// Sichtprüfung erzwungen: MSE zeichnet die beiden Linien **konzentrisch** —
// eine bei einem Punkt Abstand, die andere bei zweien. Ohne den Versatz liegt
// die weiße Linie über der dunklen und deckt sie vollständig; auf einem hellen
// Bild war der Rahmen dann unsichtbar. Nachgemessen im laufenden Fenster:
// `outline-offset: 0px`, `box-shadow: … 0 0 0 2px`.
const rahmenlinien =
  // **Hier steht absichtlich keine Palettenfarbe** — Werkstück J von Phase
  // 4.10. Der Rahmen liegt über einem **Bild**, dessen Farben niemand kennt;
  // Weiß und Schwarz übereinander sind das einzige Paar, das auf jedem
  // Untergrund sichtbar bleibt. Eine Marke aus der Palette verschwände auf
  // einer Karte derselben Farbe.
  'outline outline-2 outline-bildhell outline-offset-2 ring-2 ring-bilddunkel'
const griff = 'absolute size-3 rounded-sm border border-bildhell bg-bilddunkel'

// **Zusammengesetzt oberhalb des JSX** und nicht darin: `keintextimjsx` weist
// eine Zeichenkette mit Platzhalter im Ausdruck ab — sie ist von einem
// Oberflächentext nicht zu unterscheiden.
const rahmenkasten = `absolute cursor-move ${rahmenlinien}`
const griffRechts = `${griff} right-0 top-1/2 translate-x-1/2 -translate-y-1/2`
const griffUnten = `${griff} left-1/2 bottom-0 -translate-x-1/2 translate-y-1/2`

export function Zuschnittrahmen({
  zuschnitt,
  auswahl,
  frei,
  ziehen,
}: ZuschnittrahmenEigenschaften) {
  const texte = useTexte()
  const flaeche = useRef<HTMLDivElement>(null)

  const zug =
    (kanteX: number, kanteY: number) => (ereignis: React.PointerEvent) => {
      ereignis.preventDefault()
      ereignis.stopPropagation()

      const ziel = ereignis.currentTarget
      const vonX = ereignis.clientX
      const vonY = ereignis.clientY

      // massstab sagt, wie viele Quellpunkte ein Bildschirmpunkt ist.
      //
      // **Aus dem gezeichneten Element und nicht aus einer Zahl im Code**: das
      // Bild ist über `max-w-full` schmaler als seine Punktzahl, sobald das
      // Fenster es ist — ein fester Maßstab machte den Rahmen bei jeder
      // anderen Fensterbreite falsch.
      //
      // **Hier gelesen und nicht in einer Hilfsfunktion**: einen Verweis an
      // eine Funktion zu reichen, die beim Zeichnen entsteht, hält React für
      // einen Zugriff **während** des Zeichnens — und es ist keiner, denn
      // dieser Rumpf läuft erst beim Zeigerdruck.
      //
      // **`?? 0` ist die Typprüfung, kein Fall**: der Verweis steht, sobald
      // gezeichnet ist. **Null heißt „noch nicht gezeichnet"** — dann ist eins
      // so gut wie jede andere Zahl, und die nächste Bewegung rechnet richtig.
      /* v8 ignore next */
      const breite = flaeche.current?.getBoundingClientRect().width ?? 0
      const mass = breite > 0 ? zuschnitt.quelleBreite / breite : 1

      ziel.setPointerCapture(ereignis.pointerId)

      const bewegt = (weiter: Event) => {
        const zeiger = weiter as PointerEvent

        ziehen(
          (zeiger.clientX - vonX) * mass,
          (zeiger.clientY - vonY) * mass,
          kanteX,
          kanteY,
        )
      }

      const beendet = () => {
        ziel.removeEventListener('pointermove', bewegt)
        ziel.removeEventListener('pointerup', beendet)
      }

      ziel.addEventListener('pointermove', bewegt)
      ziel.addEventListener('pointerup', beendet)
    }

  const anteil = (wert: number, ganz: number) =>
    `${String((wert / ganz) * 100)}%`

  return (
    <div ref={flaeche} className="relative max-w-md">
      <img
        src={zuschnitt.quelle}
        alt={texte.seiten.karten.zuschnitt.quelle}
        className="block w-full select-none"
        draggable={false}
      />

      {/* **Zeigerhilfen und keine Knöpfe** — G17 aus dem Zweitblick über
          Phase 6.3: die Griffe trugen `role="button"` und taten auf ein
          Aktivieren nichts; ein Vorleser bot sie als Knöpfe an. Wer ohne
          Zeiger zuschneidet, nimmt die vier Zahlenfelder daneben (links,
          oben, Breite, Höhe) — sie tun dasselbe. Deshalb verborgen. */}
      <div
        aria-hidden="true"
        data-griff="rahmen"
        className={rahmenkasten}
        style={{
          left: anteil(auswahl.x, zuschnitt.quelleBreite),
          top: anteil(auswahl.y, zuschnitt.quelleHoehe),
          width: anteil(auswahl.breite, zuschnitt.quelleBreite),
          height: anteil(auswahl.hoehe, zuschnitt.quelleHoehe),
        }}
        onPointerDown={zug(0, 0)}
      >
        {ecken.map((ecke) => (
          <span
            key={ecke.wo}
            data-griff="ecke"
            className={`${griff} ${ecke.wo}`}
            onPointerDown={zug(ecke.x, ecke.y)}
          />
        ))}

        {/* **Ohne festes Seitenverhältnis ziehen auch die Kantenmitten** —
            MSE zeichnet dann vier weitere Griffe (Zeilen 515–520). */}
        {frei && (
          <>
            <span
              data-griff="ecke"
              className={griffRechts}
              onPointerDown={zug(1, 0)}
            />
            <span
              data-griff="ecke"
              className={griffUnten}
              onPointerDown={zug(0, 1)}
            />
          </>
        )}
      </div>
    </div>
  )
}
