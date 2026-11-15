// **Beiwerk in der Werkstatt** — FMT-02.AK-01
// (#1691).
//
// Ein Bild oder eine Schrift ist kein Text; statt des Editors steht hier, was
// es ist, und dieselbe Zeile Handgriffe wie über einer Vorlage.
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { useTexte } from '../sprache/kontext'
import { Dateiaktionen } from './Dateiaktionen'

import type { Dateiaktionen as Handgriffe } from '../haken/useDateiaktionen'

export function Beiwerktafel({
  datei,
  aktionen,
}: {
  datei: string
  aktionen: Handgriffe
}) {
  const texte = useTexte().seiten.vorlagen

  return (
    <section className="flex flex-col gap-2">
      <h2 className="font-mono text-sm">{datei}</h2>
      <Erklaerungsabsatz text={texte.beiwerkHinweis} />
      <Dateiaktionen
        aktionen={aktionen}
        datei={datei}
        hauptdatei={false}
        gesperrt={false}
      />
    </section>
  )
}
