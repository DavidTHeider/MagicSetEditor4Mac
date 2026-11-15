// **Ein Satz im Editor einer eigenen Übersetzung** — LOC-02
// (#1688).
//
// **Ein eigener Stand je Zeile**: was getippt ist, gehört dem Feld, bis es
// übernommen wird. Die Vorschau zeigt den Satz, wie das Fenster ihn setzt —
// **fett** und `Code` inklusive.
import { useId, useState } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Aktionsknopf } from './Aktionsknopf'
import { Satz } from './Satz'

export function Uebersetzungszeile({
  pfad,
  deutsch,
  basistext,
  text,
  veraltet,
  befund,
  angehalten,
  uebernehmen,
}: {
  pfad: string
  deutsch: string
  basistext: string
  text: string | undefined
  veraltet: boolean
  befund: string | undefined
  angehalten: boolean
  uebernehmen: (text: string) => void
}) {
  const eigene = useTexte().seiten.einstellungen.eigenuebersetzung
  const [entwurf, setzeEntwurf] = useState(text ?? '')
  const befundID = useId()

  return (
    <li className="flex flex-col gap-2 py-3">
      <code className="text-xs text-leise">{pfad}</code>
      <div className="grid gap-2 text-sm md:grid-cols-2">
        <div>
          <span className="text-xs text-leise">{eigene.quelle}</span>
          <p>
            <Satz text={deutsch} />
          </p>
        </div>
        <div>
          <span className="text-xs text-leise">{eigene.basistext}</span>
          <p>
            <Satz text={basistext} />
          </p>
        </div>
      </div>
      {veraltet && <p className="text-sm text-leise">{eigene.veraltetMarke}</p>}
      <textarea
        value={entwurf}
        rows={Math.min(6, Math.max(1, Math.ceil(deutsch.length / 80)))}
        aria-label={einsetzen(eigene.uebersetzung, { pfad })}
        aria-invalid={befund !== undefined}
        aria-describedby={befund === undefined ? undefined : befundID}
        onChange={(ereignis) => {
          setzeEntwurf(ereignis.target.value)
        }}
        className="rounded-md border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      />
      {befund !== undefined && (
        <div role="alert">
          <p id={befundID} className="text-sm text-fehler">
            {befund}
          </p>
        </div>
      )}
      <div className="flex gap-2">
        <Aktionsknopf
          aktion="sichern"
          gewicht="neben"
          angehalten={angehalten || entwurf.trim() === '' || entwurf === text}
          aufKlick={() => {
            uebernehmen(entwurf)
          }}
        >
          {eigene.uebernehmen}
        </Aktionsknopf>
        {text !== undefined && (
          <Aktionsknopf
            aktion="zuruecknehmen"
            gewicht="neben"
            angehalten={angehalten}
            aufKlick={() => {
              setzeEntwurf('')
              uebernehmen('')
            }}
          >
            {eigene.zuruecknehmen}
          </Aktionsknopf>
        )}
      </div>
    </li>
  )
}
