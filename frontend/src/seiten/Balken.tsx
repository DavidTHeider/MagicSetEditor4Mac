// Ein Balken einer Auszählung — eine Zahl im Verhältnis zum Ganzen.
//
// **Eine eigene Datei**, seit die erklärten Auszählungsdimensionen dazukamen
// (#1338):
// `Statistik.tsx` und `Auszaehlung.tsx` zeigen beide Balken, und zwei
// Fassungen desselben Bauteils laufen auseinander.
//
// **Und er ist ohne Farbe lesbar.** Jeder Balken trägt seine Zahl und seinen
// Anteil als Text; die Fläche ist die Zugabe, nicht die Auskunft. Ein
// Diagramm, dessen Aussage in der Farbe steckt, ist für jeden zwölften Mann
// keine.
import type { ZahlDTO } from '../api'
import { alsProzent } from '../lokal'
import { useDarstellung } from '../sprache/kontext'
import { ohneAuszeichnung } from './ohneauszeichnung'

// Balken zeigt eine Zahl im Verhältnis zum Ganzen — mit Text davor.
//
// **Zum Ganzen und nicht zur größten Zahl.** Der erste Anlauf maß am größten
// Balken: in der Verteilung stand dann „common 121 (100 %)", obwohl 121 von 312
// Karten gemeint waren. Eine Zahl, die 39 % ist und 100 % daneben stehen hat,
// ist schlimmer als gar keine.
// knopfklassen sind die Klassen eines wählbaren Balkens. **Außerhalb des JSX
// zusammengesetzt** — `keintextimjsx` weist jede Zeichenkette mit Platzhalter
// darin ab, und das zu Recht: so sieht ein übersehener Oberflächentext aus.
// Hier sind es Stilklassen. Gewählt trägt er den Rand der Markenfarbe, sonst
// nur beim Darüberfahren eine Fläche.
const balkenklassen =
  'block w-full rounded-sm border px-2 py-1 text-left ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:cursor-not-allowed disabled:opacity-60'

function knopfklassen(gewaehlt: boolean): string {
  return `${balkenklassen} ${
    gewaehlt ? 'border-marke bg-betont' : 'border-transparent hover:bg-betont'
  }`
}

// Waehlbarkeit macht aus einem Balken einen Schalter — UXR-F6
// (#2013).
//
// **Ein Balken ist ein Knopf, kein Bereich mit Klickhorcher**: er nimmt den
// Fokus, antwortet auf Enter und Leertaste und sagt einem Vorleseprogramm über
// `aria-pressed`, ob er gewählt ist. MSE2 wählt nur mit der Maus und den
// Pfeilen (`GraphControl::onMouseDown`, `onChar`,
// `src/gui/control/graph.cpp` Zeilen 1116–1161, Fassung f9d9356 vom
// 2021-07-06); die Tastatur ist hier die Gestalt, in der macOS es erwartet.
export interface Waehlbarkeit {
  gewaehlt: boolean
  waehlen: () => void
  // leerTitel sagt, warum ein Balken ohne Karten nichts wählt.
  leerTitel: string
}

export function Balken({
  name,
  karten,
  ganzes,
  anteil,
  feldwert = false,
  waehlbar,
}: ZahlDTO & {
  ganzes: number
  anteil: string
  feldwert?: boolean
  waehlbar?: Waehlbarkeit | undefined
}) {
  // **Prozent auf ganze Zahlen.** „37,4 %" behauptet eine Genauigkeit, die bei
  // zwölf Karten niemand hat.
  //
  // **Und das Ganze darf null sein**, seit eine Dimension ihre Gruppen auch
  // dann nennt, wenn keine Karte sie trägt: die erklärte Reihenfolge steht da,
  // mit Nullen. `0/0` wäre `NaN` und stünde als „NaN %" auf der Seite.
  const darstellung = useDarstellung()
  const verhaeltnis = ganzes === 0 ? 0 : karten / ganzes

  // **Die Breite ist keine Sprache** — sie ist CSS und bleibt eine Zahl;
  // der Text daneben folgt der Sprache des Fensters (LOC-03.AK-01).
  const prozent = Math.round(verhaeltnis * 100)

  const inhalt = (
    <>
      <div className="flex items-baseline justify-between gap-3">
        {/* **Ein Feldwert steht ohne seine Auszeichnung**
            (#952).
            Die Belegung zeigt **Feldnamen** — die kommen aus dem Spielpaket
            und tragen keine Marken; die Verteilung zeigt **Werte**, und die
            können welche tragen. Zwei Bedeutungen desselben Feldes, und
            deshalb sagt der Aufrufer, welche. */}
        <span>{feldwert ? ohneAuszeichnung(name) : name}</span>
        <span className="tabular-nums text-leise">
          {karten} ({alsProzent(verhaeltnis, darstellung)})
        </span>
      </div>

      {/* Die Fläche ist Zugabe: `aria-hidden`, weil die Zahl daneben dasselbe
          sagt — und zweimal dasselbe vorgelesen zu bekommen, hilft niemandem. */}
      <div
        aria-hidden="true"
        title={anteil}
        className="mt-1 h-2 rounded-sm bg-betont"
      >
        <div
          className="h-2 rounded-sm bg-marke"
          style={{ width: `${String(prozent)}%` }}
        />
      </div>
    </>
  )

  if (waehlbar === undefined) {
    return <li className="py-1">{inhalt}</li>
  }

  // **Ein Balken ohne Karten wählt nichts.** MSE2 wählt auch ihn und zeigt
  // dann eine leere Liste; hier ist er angehalten, und der Titel sagt warum
  // — eine leere Liste nach einem Klick läse sich wie ein Fehler.
  const leer = karten === 0

  return (
    <li className="py-0.5">
      <button
        type="button"
        aria-pressed={waehlbar.gewaehlt}
        disabled={leer}
        title={leer ? waehlbar.leerTitel : undefined}
        onClick={waehlbar.waehlen}
        className={knopfklassen(waehlbar.gewaehlt)}
      >
        {inhalt}
      </button>
    </li>
  )
}
