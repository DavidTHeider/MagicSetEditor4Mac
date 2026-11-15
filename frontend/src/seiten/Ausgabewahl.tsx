// **Der erste Schritt des Ausgabeassistenten: wohin?** — UXR-09
// (#1917),
// nach dem Artboard *Ausgeben*: links die Fläche „Wohin", je Ziel ein Name
// und darunter, wofür es ist.
//
// # Optionsfelder und keine Auswahl
//
// **Weil jedes Ziel seinen Satz trägt.** R7 aus `docs/UIREVIEW.md` stellt
// über vier Werte eine Auswahl hin — die zeigte aber nur Namen, und
// „Druckbogen" oder „MTG JSON" sagt niemandem, der es zum ersten Mal sieht,
// wofür es ist. Das Artboard zeichnet die fünf als Zeilen mit Satz, und so
// stehen sie hier: eine Gruppe von Optionsfeldern, die Pfeiltasten wandern,
// und der Satz hängt über `aria-describedby` am Feld, statt Teil seines
// Namens zu werden.
import { useId } from 'react'

import { Satz } from '../bausteine/Satz'
import { useTexte } from '../sprache/kontext'
import { ausgabeziele, type Ausgabeziel } from '../ausgabe'
import { zielname } from './ausgabetexte'

interface AusgabewahlEigenschaften {
  gewaehlt: Ausgabeziel | undefined
  waehlen: (ziel: Ausgabeziel) => void
}

// Die Klassen einer Zeile — **fertige Zeichenketten**: was die gewählte
// bekommt, steht so auf einen Blick.
const zeile = {
  an: 'flex items-start gap-2 rounded-md border border-kante bg-betont p-2',
  aus: 'flex items-start gap-2 rounded-md border border-transparent p-2',
}

const wahlstil =
  'mt-1 accent-marke focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Ausgabewahl({ gewaehlt, waehlen }: AusgabewahlEigenschaften) {
  const texte = useTexte()
  const gruppe = useId()

  return (
    <fieldset className="flex flex-col gap-1 rounded-lg border border-kante bg-flaeche p-3">
      <legend className="px-1 font-medium">
        {texte.seiten.ausgeben.wohin}
      </legend>

      {ausgabeziele.map((ziel) => {
        const kennung = `${gruppe}-${ziel}`
        const satz = `${kennung}-satz`

        return (
          <div key={ziel} className={ziel === gewaehlt ? zeile.an : zeile.aus}>
            <input
              id={kennung}
              type="radio"
              name={gruppe}
              value={ziel}
              checked={ziel === gewaehlt}
              aria-describedby={satz}
              onChange={() => {
                waehlen(ziel)
              }}
              className={wahlstil}
            />
            <div className="min-w-0">
              <label htmlFor={kennung} className="font-medium">
                {zielname(texte, ziel)}
              </label>
              <p id={satz} className="text-sm text-leise">
                <Satz text={texte.seiten.ausgeben.ziele[ziel].kurz} />
              </p>
            </div>
          </div>
        )
      })}
    </fieldset>
  )
}
