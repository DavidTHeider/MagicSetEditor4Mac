// **Die Parameterhilfen der Schlüsselwörter** —
// #1342.
//
// # Was MSE2 hier tut
//
// Zwei Knöpfe unter zwei Feldern: „insert parameter" klappt die
// Parameterarten des Spiels auf und schreibt `<atom-param>art</atom-param>`
// ins **Muster**; „refer parameter" klappt die Parameter **dieses**
// Schlüsselworts auf und schreibt in den **Erinnerungstext** — `{param1}`,
// oder das Ergebnis eines `refer script:`
// (`src/gui/set/keywords_panel.cpp`, Zeilen 190–232, Fassung f9d9356 vom
// 2021-07-06).
//
// **Hier ist es eine Auswahlliste statt eines Aufklappmenüs an einem Knopf.**
// Ein Menü über einem Knopf ist ein Menü in einem Menü, und §6 verbietet
// verschachtelte Dialoge; eine Auswahlliste ist dieselbe Wahl mit fertiger
// Tastaturbedienung.
import { useTexte } from '../sprache/kontext'
import { eingabe } from './Schluesselwortfeld'

// Wahleintrag ist ein Eintrag beider Menüs.
export interface Wahleintrag {
  // schluessel ist der Wert des `<option>` — er muss eindeutig sein.
  schluessel: string

  // aufschrift ist, was dasteht.
  aufschrift: string

  // einfuegen ist der Text, der ins Feld kommt.
  einfuegen: string
}

interface ParameterwahlEigenschaften {
  aufschrift: string
  eintraege: readonly Wahleintrag[]
  angehalten: boolean
  aufWahl: (text: string) => void
}

export function Parameterwahl({
  aufschrift,
  eintraege,
  angehalten,
  aufWahl,
}: ParameterwahlEigenschaften) {
  const texte = useTexte().seiten.karten.schluesselwoerter

  // **Ohne Einträge steht sie nicht da** (§6). Ein Spiel ohne Parameterarten
  // hat keine, und eine leere Liste erklärte keinen nächsten Schritt.
  if (eintraege.length === 0) {
    return null
  }

  return (
    <label className="flex items-center gap-2 text-xs">
      {aufschrift}
      <select
        className={eingabe}
        // **Sie zeigt immer den leeren Eintrag.** Sie ist ein Befehl und kein
        // Zustand: was gewählt wurde, steht danach im Feld darüber, nicht hier.
        value=""
        disabled={angehalten}
        onChange={(ereignis) => {
          const gewaehlt = eintraege.find(
            (eintrag) => eintrag.schluessel === ereignis.target.value,
          )

          /* v8 ignore next 3 -- der leere Eintrag findet nichts, und einen
             anderen Wert kann die Liste nicht liefern. */
          if (gewaehlt === undefined) {
            return
          }

          aufWahl(gewaehlt.einfuegen)
        }}
      >
        <option value="">{texte.parameterWaehlen}</option>
        {eintraege.map((eintrag) => (
          <option key={eintrag.schluessel} value={eintrag.schluessel}>
            {eintrag.aufschrift}
          </option>
        ))}
      </select>
    </label>
  )
}
