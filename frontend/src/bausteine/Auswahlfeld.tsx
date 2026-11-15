// **Ein Auswahlfeld: nativ bis vierzehn Werte, mit Suche ab fünfzehn** —
// UXR-F1 (#2007),
// R7 aus `docs/UIREVIEW.md`: *„Auswahl, mit Suche ab 15 Werten"*.
//
// # Wofür es das gibt
//
// **Weil die Stellen es nicht selbst entscheiden sollen.** Das Deck am Sitz,
// das Set in der Suche, die Vorlage eines Turniers, das Feld einer Karte: ihre
// Listen wachsen mit dem Bestand, und niemand weiß beim Bauen, ob es drei
// Werte werden oder dreihundert. Eine Stelle, die die Schwelle selbst prüfte,
// trüge den Zweig — und seine Kehrseite in ihren Tests — zehnmal.
//
// - **Unter `suchschwelle`** (15) bleibt es das **native Feld**: das Menü von
//   macOS, die Tastatur und die Vorlesung kommen mit, und jeder bestehende
//   Weg dahin — ein Test, ein Drehbuch — bleibt, was er war.
// - **Ab `suchschwelle`** ist es die [`Suchauswahl`](./Suchauswahl.tsx).
//
// **Die Beschriftung stellt der Aufrufer**, wie bei der `Suchauswahl` —
// `<label htmlFor={id}>` oder die eines `Feld` —, oder er bittet mit
// `benennen` um ein `aria-label`.
//
// **Was es nicht kann**: Gruppen (`optgroup`) und deaktivierte Zeilen. Wer sie
// braucht (die `Statistikwahl`), hält seinen nativen Zweig selbst und nimmt
// ab der Schwelle die `Suchauswahl`.
import { Suchauswahl } from './Suchauswahl'
import { suchschwelle, type Suchwahl } from './suchwahl'

const nativ =
  'rounded-sm border border-kante bg-transparent px-2 py-1 ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring ' +
  'disabled:opacity-60'

interface AuswahlfeldEigenschaften {
  id: string
  bezeichnung: string
  eintraege: readonly Suchwahl[]
  wert: string
  aendern: (wert: string) => void
  angehalten?: boolean
  beschriebenVon?: string | undefined
  ungueltig?: boolean
  benennen?: boolean

  // klasse ersetzt die Klassen des **nativen** Feldes. Die Suchauswahl hat
  // ihre eigenen; sie sehen ähnlich aus, aber nicht gleich — ein Rand in
  // derselben Farbe, eine Fläche, die nicht durchscheint.
  klasse?: string
}

export function Auswahlfeld({
  id,
  bezeichnung,
  eintraege,
  wert,
  aendern,
  angehalten = false,
  beschriebenVon,
  ungueltig = false,
  benennen = false,
  klasse = nativ,
}: AuswahlfeldEigenschaften) {
  if (eintraege.length >= suchschwelle) {
    return (
      <Suchauswahl
        id={id}
        bezeichnung={bezeichnung}
        eintraege={eintraege}
        wert={wert}
        aendern={aendern}
        angehalten={angehalten}
        beschriebenVon={beschriebenVon}
        ungueltig={ungueltig}
        benennen={benennen}
      />
    )
  }

  return (
    <select
      id={id}
      value={wert}
      disabled={angehalten}
      aria-label={benennen ? bezeichnung : undefined}
      aria-describedby={beschriebenVon}
      aria-invalid={ungueltig ? true : undefined}
      onChange={(ereignis) => {
        aendern(ereignis.target.value)
      }}
      className={klasse}
    >
      {eintraege.map((eintrag) => (
        <option key={eintrag.wert} value={eintrag.wert}>
          {eintrag.name}
        </option>
      ))}
    </select>
  )
}
