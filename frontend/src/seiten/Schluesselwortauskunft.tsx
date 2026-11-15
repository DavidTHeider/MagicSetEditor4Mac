// **Ein Schlüsselwort des Spiels, als Text** — UXR-07
// (#1915).
//
// # Warum kein abgeschaltetes Feld
//
// **Weil sich nichts daran ändern lässt.** Ein Schlüsselwort des Spiels
// gehört dem Spielpaket; MSE zeigt es grau und lässt es nicht bearbeiten
// (`KeywordList::OnGetItemAttr`, `src/gui/control/keyword_list.cpp`, Zeilen
// 225–235, Fassung f9d9356 vom 2021-07-06). Hier stand es bis dahin als
// ganze Bearbeitungszeile mit fünf abgeschalteten Feldern — bei
// `magic.mse-game` 368 Zeilen, 1.840 Felder (R6 und R7 aus
// `docs/UIREVIEW.md`). Text sagt dasselbe, lässt sich markieren und
// kopieren und kostet keinen Knoten, den der Webprozess halten muss.
//
// # Warum eine Beschreibungsliste
//
// **Damit ein Vorleseprogramm sagt, was was ist** — „Muster: Fliegend“ statt
// zweimal „Fliegend“. Leere Angaben fallen weg; ein Name ohne Wert wäre eine
// Zeile, die nichts sagt.
import type { SchluesselwortDTO } from '../api'
import { useTexte } from '../sprache/kontext'

interface AuskunftEigenschaften {
  wort: SchluesselwortDTO
}

export function Schluesselwortauskunft({ wort }: AuskunftEigenschaften) {
  const texte = useTexte().seiten.karten.schluesselwoerter

  const alle: [string, string][] = [
    [texte.modus, wort.modus],
    [texte.muster, wort.muster],
    [texte.erinnerung, wort.erinnerung],
    [texte.regeln, wort.regeln],
  ]
  const angaben = alle.filter(([, wert]) => wert !== '')

  return (
    <li className="flex flex-col gap-1 rounded-sm border border-kante px-3 py-2 text-sm">
      <span className="font-medium">{wort.wort}</span>
      <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5">
        {angaben.map(([name, wert]) => (
          <div key={name} className="contents">
            <dt className="text-leise">{name}</dt>
            <dd className="break-words">{wert}</dd>
          </div>
        ))}
      </dl>
    </li>
  )
}
