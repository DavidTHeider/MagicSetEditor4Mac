// Die **Schlüsselwörter des Spiels** — eingeklappt, weil es 334 sind.
//
// **Aus `Schluesselwoerter.tsx` herausgelöst** (UXR-F3,
// #2009):
// die Seite wählt, legt an, schreibt und löscht ihre eigenen — hier steht,
// was sie von den festen zeigt. Der Schnitt folgt der Frage (§2).
//
// **Sie stehen weder in der Liste noch im Editor.** MSE mischt sie unter die
// eigenen und zeigt sie grau (`KeywordList::getItems`,
// `src/gui/control/keyword_list.cpp`, Zeilen 153–166, Fassung f9d9356 vom
// 2021-07-06); hier sind es 334 gegen eine Handvoll, und die Liste der
// eigenen soll die Liste der eigenen bleiben.
import type { SchluesselwortDTO } from '../api'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Schluesselwortauskunft } from './Schluesselwortauskunft'

interface FesteEigenschaften {
  woerter: SchluesselwortDTO[]
}

// Schluesselwortfeste zeigt die Schlüsselwörter des Spiels — eingeklappt, weil es 334 sind.
//
// **Der Knopf heißt, wie er aussieht**, und sagt, ob er offen ist (UXR-07,
// #1915): vorher trug er „Die des Spiels zeigen“ als Namen über der
// Aufschrift „334 aus dem Spiel“ und kein `aria-expanded` — gehört und
// gesehen waren zwei verschiedene Knöpfe (WCAG 2.5.3).
export function Schluesselwortfeste({ woerter }: FesteEigenschaften) {
  const texte = useTexte().seiten.karten.schluesselwoerter

  if (woerter.length === 0) {
    return null
  }

  return (
    <Aufklappkasten
      titel={einsetzen(texte.ausDemSpielZahl, {
        zahl: String(woerter.length),
      })}
    >
      <p className="mb-2 text-sm text-leise">{texte.ausDemSpiel}</p>

      <ul aria-label={texte.ausDemSpiel} className="flex flex-col gap-2">
        {woerter.map((wort, stelle) => (
          <Schluesselwortauskunft
            key={`${wort.wort}-${String(stelle)}`}
            wort={wort}
          />
        ))}
      </ul>
    </Aufklappkasten>
  )
}
