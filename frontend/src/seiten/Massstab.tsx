// **Der Maßstab der Vorschau** — UXR-06
// (#1914),
// vorher die Zoomleiste aus Werkstück U5 von Phase 5.1
// (#1416).
//
// # Eine Auswahl und keine Knopfleiste
//
// **Sechs Werte sind eine Auswahl** — R7 aus `docs/UIREVIEW.md`: Segmente
// bis vier, darüber eine Auswahl. Die Leiste davor stellte sechs Knöpfe vor
// die Karte, und die gedrückte Stufe trug das Hauptgewicht einer führenden
// Handlung; gemessen am nativen Fenster waren das sechs der 27
// Bedienelemente im Bild (R2 erlaubt zwanzig). Das Artboard *Karten* zeichnet
// eine Auswahl.
//
// **Was die Leiste zusagte, bleibt**: eine Tabstelle — das gibt die Auswahl
// von selbst, und die Pfeiltasten wählen in ihr —, die Prozentzahl nach der
// Sprache und die Druckgröße nur, wenn die Vorlage eine Auflösung nennt.
//
// # Der Grund steht an der Auswahl
//
// Eine angehaltene Option sagt nicht, **warum**. Der Satz dazu hängt über
// `aria-describedby` an der Auswahl und steht nur da, wenn es ihn braucht.
import { useId } from 'react'

import { alsProzent } from '../lokal'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { stufen, type Zoomwahl } from './zoom'

interface MassstabEigenschaften {
  zoom: Zoomwahl
  waehlen: (zoom: Zoomwahl) => void
  druckMoeglich: boolean
}

// wahlAus macht aus dem Wert einer Option die Wahl, die er meint. Die
// Prozentstufen stehen als Zahl im Wert, die übrigen als ihr Name.
function wahlAus(wert: string): Zoomwahl {
  if (wert === 'anpassen' || wert === 'hundert' || wert === 'druck') {
    return wert
  }

  return Number(wert) as 50 | 150 | 200
}

export function Massstab({
  zoom,
  waehlen,
  druckMoeglich,
}: MassstabEigenschaften) {
  const worte = useTexte().seiten.karten.zoom

  // **Die Stufen sind Zahlen mit Zeichen** — und das Zeichen folgt der
  // Sprache: »150 %« im Deutschen, »150%« im Englischen (LOC-03.AK-01,
  // #1689).
  const darstellung = useDarstellung()
  const grund = useId()

  return (
    <span className="inline-flex flex-wrap items-center gap-2">
      <select
        aria-label={worte.bezeichnung}
        aria-describedby={druckMoeglich ? undefined : grund}
        value={String(zoom)}
        onChange={(ereignis) => {
          waehlen(wahlAus(ereignis.target.value))
        }}
        className="rounded-sm border border-kante bg-transparent px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      >
        <option value="anpassen">{worte.anpassen}</option>
        {stufen.map((stufe) => (
          <option key={stufe} value={String(stufe)}>
            {alsProzent(stufe === 'hundert' ? 1 : stufe / 100, darstellung)}
          </option>
        ))}
        <option value="druck" disabled={!druckMoeglich}>
          {worte.druck}
        </option>
      </select>

      {!druckMoeglich && (
        <span id={grund} className="text-xs text-leise">
          {worte.ohneDruck}
        </span>
      )}
    </span>
  )
}
