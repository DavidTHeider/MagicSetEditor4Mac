// **Das Formular für ein neues Zugriffstoken** — herausgelöst aus
// `Zugriffstoken.tsx` in Werkstück O2 von Phase 5.2
// (#1426),
// als der Principal und die Ablaufwahlen dazukamen.
//
// # Der Principal ist Pflicht
//
// **Ein Token nennt, für wen es arbeitet** — ein Mensch oder eine Maschine.
// `domain.Zugriffstoken.Validate` weist ein leeres ab; das Formular sagt es
// vorher, indem der Knopf ruht, solange eines der beiden Pflichtfelder leer
// ist (§6).
//
// # Die Ablaufwahlen kommen aus Go
//
// **Vier fertige Zeitpunkte und ein eigener.** Gerechnet wird in `domain`,
// damit alle vier Flächen dieselben Zeitpunkte nennen; das Formular wählt
// nur aus. »Eigener Zeitpunkt« gibt das freie Feld frei.
import { useState } from 'react'

import { Beschriftetesfeld } from './Beschriftetesfeld'
import { Aktionsknopf } from './Aktionsknopf'
import { useTexte } from '../sprache/kontext'

import type { AblaufwahlDTO, TokenangabenDTO } from '../api'

interface Eigenschaften {
  profilID: string
  wahlen: readonly AblaufwahlDTO[]
  laeuft: boolean
  aufAnlegen: (angaben: TokenangabenDTO) => void
}

const leererEntwurf = {
  bezeichnung: '',
  principal: '',
  scope: 'lesen',
  ablauf: '',
  wahl: 'nie',
}

const auswahl =
  'rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Tokenformular({
  profilID,
  wahlen,
  laeuft,
  aufAnlegen,
}: Eigenschaften) {
  const eigene = useTexte().seiten.nutzer
  const [entwurf, setzeEntwurf] = useState(leererEntwurf)

  // wahlname ist der Text zu einer Wahl — eine unbekannte bleibt beim Namen.
  const wahlname = (name: string): string => {
    const namen: Record<string, string | undefined> = eigene.tokenAblaufwahlen

    return namen[name] ?? name
  }

  const unvollstaendig =
    entwurf.bezeichnung.trim() === '' || entwurf.principal.trim() === ''

  return (
    <div className="flex flex-wrap items-end gap-3">
      <Beschriftetesfeld
        bezeichnung={eigene.tokenBezeichnung}
        wert={entwurf.bezeichnung}
        breite={20}
        setzeWert={(wert) => {
          setzeEntwurf((bisher) => ({ ...bisher, bezeichnung: wert }))
        }}
      />

      <Beschriftetesfeld
        bezeichnung={eigene.tokenPrincipal}
        wert={entwurf.principal}
        breite={16}
        setzeWert={(wert) => {
          setzeEntwurf((bisher) => ({ ...bisher, principal: wert }))
        }}
      />

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">{eigene.tokenScope}</span>
        <select
          value={entwurf.scope}
          onChange={(ereignis) => {
            setzeEntwurf((bisher) => ({
              ...bisher,
              scope: ereignis.target.value,
            }))
          }}
          className={auswahl}
        >
          <option value="lesen">{eigene.tokenLesen}</option>
          <option value="schreiben">{eigene.tokenSchreiben}</option>
        </select>
      </label>

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">{eigene.tokenAblaufwahl}</span>
        <select
          value={entwurf.wahl}
          onChange={(ereignis) => {
            const wahl = ereignis.target.value
            const gewaehlt = wahlen.find((eine) => eine.name === wahl)

            setzeEntwurf((bisher) => ({
              ...bisher,
              wahl,
              ablauf: gewaehlt?.ablauf ?? bisher.ablauf,
            }))
          }}
          className={auswahl}
        >
          {wahlen.map((wahl) => (
            <option key={wahl.name} value={wahl.name}>
              {wahlname(wahl.name)}
            </option>
          ))}
          <option value="eigen">{eigene.tokenAblaufwahlen.eigen}</option>
        </select>
      </label>

      {entwurf.wahl === 'eigen' && (
        <Beschriftetesfeld
          bezeichnung={eigene.tokenAblauf}
          wert={entwurf.ablauf}
          breite={22}
          setzeWert={(wert) => {
            setzeEntwurf((bisher) => ({ ...bisher, ablauf: wert }))
          }}
        />
      )}

      <Aktionsknopf
        aktion="anlegen"
        gewicht="haupt"
        angehalten={laeuft || unvollstaendig}
        aufKlick={() => {
          aufAnlegen({
            id: '',
            profilID,
            bezeichnung: entwurf.bezeichnung,
            principal: entwurf.principal,
            scope: entwurf.scope,
            ablauf: entwurf.ablauf,
            partie: '',
            sitz: '',
            fassung: '',
          })
        }}
      >
        {eigene.tokenAnlegen}
      </Aktionsknopf>
    </div>
  )
}
