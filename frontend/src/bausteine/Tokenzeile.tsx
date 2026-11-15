// **Eine Zeile der Tokenliste** — herausgelöst aus `Zugriffstoken.tsx` in
// Werkstück O2 von Phase 5.2
// (#1426).
//
// # Was sie sagt
//
// Bezeichnung, Principal, Erlaubnis, Ablauf — und seit O2 die **letzte
// Nutzung** (ein Zugang, den niemand mehr braucht, wird sichtbar) und die
// **Frist des alten Geheimnisses** nach einer Rotation.
//
// # Die Rotation fragt nach
//
// **An Ort und Stelle** (`Entscheidung`), nicht in einem Dialog: sie erzeugt
// ein Geheimnis, das nur einmal erscheint, und das soll niemand aus Versehen
// auslösen. Der neue Klartext erscheint oben bei der Liste, wo auch der eines
// neuen Tokens steht — ein Ort für dieselbe Sache.
import { useState } from 'react'

import { Entscheidung } from './Entscheidung'
import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { TokenDTO } from '../api'

interface Eigenschaften {
  eines: TokenDTO
  laeuft: boolean
  aufErlaubnis: (scope: string) => void
  aufRotieren: () => void
  aufZurueckziehen: () => void
}

export function Tokenzeile({
  eines,
  laeuft,
  aufErlaubnis,
  aufRotieren,
  aufZurueckziehen,
}: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer
  // **Eine Frage je Zeile, und sie gilt einer Handlung.** Erneuern und
  // Zurückziehen fragen beide — das Zurückziehen erst seit dem 2026-09-25
  // (UX-10.AK-04, #1682): wer den Token benutzt, kommt danach nicht mehr
  // herein, und ein Klick genügte dafür.
  const [gefragt, setzeGefragt] = useState<'rotieren' | 'zurueckziehen'>()
  const rotationGefragt = gefragt === 'rotieren'

  const gueltigkeit = eines.abgelaufen
    ? eigene.tokenAbgelaufen
    : eines.ablauf === ''
      ? eigene.tokenUnbefristet
      : eines.ablauf

  const nutzung =
    eines.letzteNutzung === ''
      ? eigene.tokenNieBenutzt
      : einsetzen(eigene.tokenLetzteNutzung, { zeit: eines.letzteNutzung })

  return (
    <li className="flex flex-col gap-2">
      <div className="flex flex-wrap items-center gap-3">
        <span className="font-medium">{eines.bezeichnung}</span>
        <span className="text-leise">{eines.principal}</span>

        <select
          aria-label={`${eigene.tokenScope} — ${eines.bezeichnung}`}
          value={eines.scope}
          onChange={(ereignis) => {
            aufErlaubnis(ereignis.target.value)
          }}
          disabled={laeuft}
          className="rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          <option value="lesen">{eigene.tokenLesen}</option>
          <option value="schreiben">{eigene.tokenSchreiben}</option>
        </select>

        <span className="text-leise">{gueltigkeit}</span>
        <span className="text-leise">{nutzung}</span>

        {eines.vorgaengerBis !== '' && (
          <span className="text-leise">
            {einsetzen(eigene.tokenVorgaengerBis, {
              zeit: eines.vorgaengerBis,
            })}
          </span>
        )}

        <Knopf
          gewicht="leise"
          angehalten={laeuft || rotationGefragt}
          aufKlick={() => {
            setzeGefragt('rotieren')
          }}
        >
          {eigene.tokenRotieren}
        </Knopf>

        <Aktionsknopf
          aktion="loeschen"
          gewicht="leise"
          angehalten={laeuft || gefragt === 'zurueckziehen'}
          aufKlick={() => {
            setzeGefragt('zurueckziehen')
          }}
        >
          {eigene.tokenLoeschen}
        </Aktionsknopf>
      </div>

      {rotationGefragt && (
        <Entscheidung
          frage={eigene.tokenRotierenFrage}
          erklaerung={eigene.tokenRotierenErklaerung}
          bestaetigen={eigene.tokenRotieren}
          abbrechen={texte.bausteine.aktionen.abbrechen}
          laeuft={laeuft}
          aufBestaetigen={() => {
            setzeGefragt(undefined)
            aufRotieren()
          }}
          aufAbbrechen={() => {
            setzeGefragt(undefined)
          }}
        />
      )}

      {gefragt === 'zurueckziehen' && (
        <Entscheidung
          frage={einsetzen(eigene.tokenLoeschenFrage, {
            name: eines.bezeichnung,
          })}
          erklaerung={eigene.tokenLoeschenErklaerung}
          bestaetigen={eigene.tokenLoeschenJa}
          aktion="loeschen"
          abbrechen={texte.bausteine.aktionen.abbrechen}
          laeuft={laeuft}
          aufBestaetigen={() => {
            setzeGefragt(undefined)
            aufZurueckziehen()
          }}
          aufAbbrechen={() => {
            setzeGefragt(undefined)
          }}
        />
      )}
    </li>
  )
}
