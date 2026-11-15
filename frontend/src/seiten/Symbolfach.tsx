// **Das Fach *Symbol*** — der Editor und darunter die Bibliothek (DOM-02,
// #1697).
//
// **Die Bibliothek nur beim Setsymbol.** Mit Ziel bearbeitet der Editor das
// Symbolfeld einer Karte (#1339); einsetzen lässt die Bibliothek aber nur
// ins Set.
//
// **Ein eingesetztes Symbol baut den Editor neu.** Er steht auf geladenem
// Zustand und zeigte sonst weiter die Zeichnung von vorhin — derselbe Grund
// wie beim Schlüssel in `Lascheninhalt.tsx` (§7, Regel 1).
//
// **Und der Editor bearbeitet auch ein Symbol der Bibliothek** (AK-02): dann
// steht die Leiste darüber, und der Schlüssel trägt das Symbol.
//
// **Die Bibliothek steht eingeklappt** — UXR-07
// (#1915).
// Die Lasche maß bei 1.280 × 800 eine Rolltiefe von 4,1 (R4 aus
// `docs/UIREVIEW.md` erlaubt 2,5), und ihre führende Handlung ist das
// Zeichnen. Aufnehmen, Einsetzen und Bearbeiten bleiben einen Klick
// entfernt; der Kasten fragt den Bestand erst beim Öffnen und behält danach,
// was darin steht — auch während der Editor ein Symbol der Bibliothek
// bearbeitet.
import { useState } from 'react'

import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { useSymbolbearbeitung } from '../haken/useSymbolbearbeitung'
import { useTexte } from '../sprache/kontext'
import { Symbol } from './Symbol'
import { Symbolbearbeitung } from './Symbolbearbeitung'
import { Symbolbibliothek } from './Symbolbibliothek'

interface SymbolfachEigenschaften {
  setID: string
  ziel: { karteID: string; feld: string } | undefined
}

export function Symbolfach({ setID, ziel }: SymbolfachEigenschaften) {
  const titel = useTexte().seiten.karten.symbolbibliothek.ueberschrift
  const [stand, setzeStand] = useState(0)
  const [runde, setzeRunde] = useState(0)

  const neuZeichnen = () => {
    setzeStand((vorher) => vorher + 1)
  }
  const nachlesen = () => {
    setzeRunde((vorher) => vorher + 1)
  }

  const bearbeitung = useSymbolbearbeitung(nachlesen)
  const { bearbeitet } = bearbeitung

  return (
    <div className="flex flex-col gap-6">
      {bearbeitet === undefined ? (
        <Symbol key={stand} setID={setID} ziel={ziel} />
      ) : (
        <>
          <Symbolbearbeitung stand={bearbeitung} bearbeitet={bearbeitet} />
          <Symbol
            key={`${bearbeitet.id}/${String(stand)}`}
            setID={setID}
            bibliothek={{
              ...bearbeitet,
              gesichert: (veraltet) => {
                bearbeitung.gesichert(bearbeitet.id, veraltet)
              },
            }}
          />
        </>
      )}
      {ziel === undefined && (
        <Aufklappkasten titel={titel}>
          <Symbolbibliothek
            setID={setID}
            runde={runde}
            eingesetzt={neuZeichnen}
            bearbeiten={bearbeitung.oeffnen}
            zurueckgeholt={(id) => {
              if (bearbeitet?.id === id) {
                neuZeichnen()
              }
            }}
          />
        </Aufklappkasten>
      )}
    </div>
  )
}
