// **Suchen, Ordnen und Filtern über der Partienliste** — PLAY-03.AK-05
// (#1711).
//
// # Warum ein eigener Baustein
//
// **Weil es zwei Fragen sind** (§2): *was steht da* beantwortet
// `Partienliste`, *wonach ausgewählt wird* diese Leiste. Die Rechnung dahinter
// steht noch einmal daneben, in `partienauswahl.ts` — sie ist ohne einen Baum
// zu prüfen.
//
// # Und warum sie ab zwei Partien erscheint
//
// **Weil sie über einer einzigen Zeile nur etwas wegnehmen kann.** Ein
// Suchfeld, das eine Liste von eins auf null bringt, ist kein Werkzeug,
// sondern eine Falle — und drei Bedienelemente über einer Zeile sind
// derselbe aufgeblähte Bildschirm, gegen den
// #1771
// steht. Die Entscheidung trifft der Aufrufer; hier steht nur, wie sie
// aussieht.
import { useId } from 'react'

import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { standwort } from '../seiten/partiewoerter'
import { useTexte } from '../sprache/kontext'

import type { Auswahl, Ordnung } from './partienauswahl'

interface Leisteneigenschaften {
  auswahl: Auswahl
  setzeAuswahl: (auswahl: Auswahl) => void

  // staende sind die Stände, die in dieser Liste vorkommen — und nur die.
  staende: readonly string[]

  gezeigt: number
  alle: number
}

export function Partienauswahlleiste({
  auswahl,
  setzeAuswahl,
  staende,
  gezeigt,
  alle,
}: Leisteneigenschaften) {
  const eigene = useTexte().seiten.spielen

  const suchID = useId()
  const standID = useId()
  const ordnungID = useId()

  // **Zurückgenommen wird nur, was gewählt ist** — ein Knopf ohne Wirkung
  // ist eine Zusage, die niemand einlöst (§10).
  const gewaehlt =
    auswahl.frage !== '' ||
    auswahl.stand !== '' ||
    auswahl.ordnung !== 'zuletzt'

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-end gap-3">
        <label htmlFor={suchID} className="flex min-w-48 flex-1 flex-col gap-1">
          <span className="text-sm font-medium">{eigene.partienSuche}</span>
          <input
            id={suchID}
            type="search"
            value={auswahl.frage}
            placeholder={eigene.partienSuchePlatzhalter}
            className={feld}
            onChange={(ereignis) => {
              setzeAuswahl({ ...auswahl, frage: ereignis.target.value })
            }}
          />
        </label>

        <label htmlFor={standID} className="flex flex-col gap-1">
          <span className="text-sm font-medium">{eigene.partienStand}</span>
          <select
            id={standID}
            value={auswahl.stand}
            className={feld}
            onChange={(ereignis) => {
              setzeAuswahl({ ...auswahl, stand: ereignis.target.value })
            }}
          >
            <option value="">{eigene.partienStandAlle}</option>
            {staende.map((stand) => (
              <option key={stand} value={stand}>
                {standwort(eigene, stand)}
              </option>
            ))}
          </select>
        </label>

        <label htmlFor={ordnungID} className="flex flex-col gap-1">
          <span className="text-sm font-medium">{eigene.partienOrdnung}</span>
          <select
            id={ordnungID}
            value={auswahl.ordnung}
            className={feld}
            onChange={(ereignis) => {
              setzeAuswahl({
                ...auswahl,
                ordnung: ereignis.target.value as Ordnung,
              })
            }}
          >
            <option value="zuletzt">{eigene.partienOrdnungZuletzt}</option>
            <option value="name">{eigene.partienOrdnungName}</option>
            <option value="zug">{eigene.partienOrdnungZug}</option>
          </select>
        </label>

        {gewaehlt && (
          <Knopf
            gewicht="leise"
            aufKlick={() => {
              setzeAuswahl(keineAuswahl)
            }}
          >
            {eigene.partienZuruecksetzen}
          </Knopf>
        )}
      </div>

      {/*
        **Die Zahl nennt das Ganze, sobald etwas fehlt.** Eine gefilterte
        Liste sieht sonst aus wie eine kurze, und niemand sähe, dass zwei
        Partien hinter der Auswahl liegen.
      */}
      <p className="text-xs text-leise">
        {gezeigt === alle
          ? einsetzen(eigene.partienZahl, { zahl: String(alle) })
          : einsetzen(eigene.partienZahlGefiltert, {
              gezeigt: String(gezeigt),
              alle: String(alle),
            })}
      </p>
    </div>
  )
}

// keineAuswahl ist der Anfangszustand — und der, auf den zurückgenommen wird.
export const keineAuswahl: Auswahl = {
  frage: '',
  stand: '',
  ordnung: 'zuletzt',
}

const feld =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'
