// **Was Magic Set Editor beim Speichern über ein Set notiert hat** —
// #1325.
//
// # Der Befund
//
// Diese fünf Angaben standen im Kasten „Nicht gedeutet" — in **jedem** der
// zehn kanonischen Sets dieselben fünf. Das ist keine unverstandene Zeile,
// sondern ein Feld, das MSE2 kennt; was jede bedeutet, steht in
// `internal/domain/setkopf.go`, nachgelesen im C++-Quelltext.
//
// # Warum hier und nicht auf der Kartenlasche
//
// Sie handeln vom **Set**. Der Kasten stand auf der Kartenseite, und wer
// Karten baut, sucht dort keine Formatfassung.
//
// # Was hier nicht steht
//
// **Keine Bearbeitung.** MSE2 schreibt `mse_version` bei jedem Speichern neu
// und die beiden Paketfassungen aus den installierten Paketen; diese Anwendung
// sichert über die mitgeführte Quelldatei
// ([0011](../../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md))
// und lässt sie stehen. Ein Feld, das man ändern kann, ohne dass es ankommt,
// wäre schlimmer als keines.
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { SetkopfDTO } from '../api'

interface Kopfeigenschaften {
  kopf: SetkopfDTO
}

export function Setkopfangaben({ kopf }: Kopfeigenschaften) {
  const texte = useTexte().seiten.setangaben

  // **Leere Angaben fallen weg.** Sie standen nicht in der Datei; eine Zeile
  // mit einem Gedankenstrich behauptete, es gäbe sie und sie sei leer.
  const zeilen = [
    [texte.mseFassung, kopf.mseFassung],
    [texte.spielfassung, kopf.spielfassung],
    [texte.stilfassung, kopf.stilfassung],
    [texte.versionskontrolle, kopf.versionskontrolle],
    [texte.apprenticecode, kopf.apprenticecode],
  ].filter(([, wert]) => wert !== '')

  // **Ohne eine einzige Angabe gibt es den Kasten nicht.** Ein Set, das in
  // dieser Anwendung entsteht, hat keine — und eine Überschrift ohne Inhalt
  // ist ein Versprechen, das nicht eingelöst wird.
  if (zeilen.length === 0) {
    return null
  }

  // **Ein Abschnitt und kein Kasten** — AP-12 (#2369), B-28: in der
  // Schublade *Eigenschaften* stand er als einziger gerahmt zwischen
  // ungerahmten Feldern. Die Überschrift ist eine h3, wie die Feldgruppen.
  return (
    <section aria-label={texte.kopf} className="flex flex-col gap-2">
      <h3 className="font-semibold">{texte.kopf}</h3>
      <Erklaerungsabsatz
        text={texte.kopfHinweis}
        klasse="max-w-prose text-sm text-leise"
      />

      <dl className="grid gap-x-6 gap-y-2 sm:grid-cols-[max-content_minmax(0,1fr)]">
        {zeilen.map(([aufschrift, wert]) => (
          <div key={aufschrift} className="contents">
            <dt className="text-leise">{aufschrift}</dt>
            <dd className="tabular-nums">{wert}</dd>
          </div>
        ))}
      </dl>
    </section>
  )
}
