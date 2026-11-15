// **Der Inhalt einer Seite** — herausgelöst aus `Rahmen`, als er mit seinem
// Namen über die Zeilengrenze aus §2 wuchs.
import type { ReactNode } from 'react'

import {
  beschriftung,
  inhaltsflaeche,
  type Seitenschluessel,
} from '../navigation'
import { useTexte } from '../sprache/kontext'
import { Bereichsfehler } from './Bereichsfehler'
import { Fehlergrenze } from './Fehlergrenze'

interface InhaltEigenschaften {
  aktuell: Seitenschluessel
  children: ReactNode
}

export function Inhalt({ aktuell, children }: InhaltEigenschaften) {
  const texte = useTexte()

  return (
    // **Ein Fokusziel und kein Halt** (UX-10.AK-02, #1682): der Sprunglink und
    // ein Seitenwechsel setzen ihn hierher.
    //
    // **Mit dem Namen der Seite** (H01, #2255): ein Vorleser sagt dann »Sets,
    // Hauptbereich« statt nur »Hauptbereich« — gefunden am nativen Fenster,
    // nach ⌘2 bei jeder Rolle.
    <main
      id="inhalt"
      data-seite={aktuell}
      aria-label={beschriftung(texte, aktuell)}
      tabIndex={-1}
      className="flex-1 px-8 pb-8 pt-4 focus:outline-none"
    >
      {/* **Der Inhalt hat eine Breite.** Ohne sie liefen Fließtext und
          Tabellen über einen 34-Zoll-Bildschirm, und niemand fände die
          nächste Zeile wieder.

          **Und sie hängt an der Seite** ([inhaltsbreite]): eine Lasche des
          Sets ist eine Arbeitsfläche, kein Fließtext
          (#697). */}
      <div className={inhaltsflaeche(aktuell)}>
        {/* **Ein Wurf hier nimmt nur diesen Bereich mit** (F01, #2340) —
            Seitenleiste und Kopf bleiben. **Ohne Schlüssel**: ein `key` je
            Seite zeichnete den ganzen Inhalt bei jedem Wechsel neu, und der
            Seitengang verlöre, was im Formular stand; die Grenze räumt ihren
            Fehler beim Wechsel selbst. */}
        <Fehlergrenze
          ebene="bereich"
          bereich={aktuell}
          zeige={(stand) => <Bereichsfehler {...stand} />}
        >
          {children}
        </Fehlergrenze>
      </div>
    </main>
  )
}
