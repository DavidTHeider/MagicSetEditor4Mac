// **Die Hilfe führt zur Frage** — UXR-22
// (#1930).
//
// # Der Anlass
//
// Gemessen am nativen Fenster stand die Hilfe bei einer Rolltiefe von 4,1:
// das erste Gebiet zählte alle Seiten mit je einem Absatz auf, und wer eine
// Frage hatte, las von vorn. Das Artboard *Hilfe* stellt ein Suchfeld über
// die Begriffe und „Wo du gerade warst" daneben.
//
// # Was hier steht
//
// **Drei Bausteine und kein Zustand.** Das Feld, die Treffer und die
// Herkunft; welche Frage gestellt ist und welches Gebiet offen steht, hält
// die Seite.
import { useId } from 'react'

import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Flaeche } from '../bausteine/Flaeche'
import { Knopf } from '../bausteine/Knopf'
import { Ohnetreffer } from '../bausteine/Ohnetreffer'
import { Satz } from '../bausteine/Satz'
import type { Seitenschluessel } from '../navigation'
import { einsetzen } from '../sprache/einsetzen'
import { useSprache, useTexte } from '../sprache/kontext'
import { falte } from '../sprache/vergleich'

// Hilfepunkt ist ein Begriff mit seinem Satz.
export interface Hilfepunkt {
  name: string
  satz: string
}

// Hilfegebiet ist, was die Suche durchsucht: ein Gebiet mit seinen Punkten.
export interface Hilfegebiet {
  name: string
  punkte: Hilfepunkt[]
}

const feld =
  'w-full max-w-md rounded-md border border-kante bg-flaeche px-3 py-2 ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Hilfesuchfeld({
  suche,
  setzeSuche,
}: {
  suche: string
  setzeSuche: (suche: string) => void
}) {
  const { hilfe } = useTexte()
  const feldID = useId()

  return (
    <label htmlFor={feldID} className="mb-4 flex flex-col gap-1">
      <span className="text-sm font-medium">{hilfe.suche.bezeichnung}</span>
      <input
        id={feldID}
        type="search"
        value={suche}
        className={feld}
        onChange={(ereignis) => {
          setzeSuche(ereignis.target.value)
        }}
      />
    </label>
  )
}

// passt sagt, ob ein Punkt zur Frage gehört — im Namen oder im Satz, ohne
// Rücksicht auf Groß und Klein.
function passt(punkt: Hilfepunkt, frage: string, sprache: string): boolean {
  return (
    falte(punkt.name, sprache).includes(frage) ||
    falte(punkt.satz, sprache).includes(frage)
  )
}

export function Hilfetreffer({
  suche,
  gebiete,
  leeren,
}: {
  suche: string
  gebiete: Hilfegebiet[]

  // leeren nimmt die Suche weg — der Weg aus „nichts passt" (G02, #2249).
  leeren: () => void
}) {
  const { hilfe } = useTexte()
  const sprache = useSprache()
  const frage = falte(suche.trim(), sprache)
  const getroffen = gebiete
    .map((gebiet) => ({
      name: gebiet.name,
      punkte: gebiet.punkte.filter((punkt) => passt(punkt, frage, sprache)),
    }))
    .filter((gebiet) => gebiet.punkte.length > 0)

  return (
    <Flaeche ueberschrift={hilfe.suche.gefunden}>
      {getroffen.length === 0 ? (
        <Ohnetreffer
          satz={einsetzen(hilfe.suche.nichts, { suche: suche.trim() })}
          leeren={leeren}
        />
      ) : (
        <div className="flex flex-col gap-4">
          {getroffen.map((gebiet) => (
            <div key={gebiet.name}>
              <h3 className="mb-2 font-semibold">{gebiet.name}</h3>
              <dl className="grid gap-3 sm:grid-cols-[10rem_minmax(0,1fr)]">
                {gebiet.punkte.map((punkt) => (
                  <div key={punkt.name} className="contents">
                    <dt className="font-semibold">{punkt.name}</dt>
                    <dd className="max-w-prose text-leise">
                      <Satz text={punkt.satz} />
                    </dd>
                  </div>
                ))}
              </dl>
            </div>
          ))}
        </div>
      )}
    </Flaeche>
  )
}

// Hilfeherkunft zeigt die Hilfe der Seite, von der man kam — und den Weg zu
// allen.
export function Hilfeherkunft({
  herkunft,
  zurUebersicht,
}: {
  herkunft: Seitenschluessel
  zurUebersicht: () => void
}) {
  const texte = useTexte()
  const seite = texte.hilfe.seiten[herkunft]

  return (
    <Flaeche ueberschrift={texte.hilfe.suche.herkunftUeberschrift}>
      <dl className="mb-3 grid gap-3 sm:grid-cols-[10rem_minmax(0,1fr)]">
        <div className="contents">
          <dt className="font-semibold">{seite.ueberschrift}</dt>
          {/* **Der erste Satz, der Rest hinter dem Fragezeichen** (UX-15):
              der ganze Absatz hob die Rolltiefe im engen Fenster auf 2,8 —
              gemessen am 2026-10-02, R4 erlaubt 2,5. */}
          <dd>
            <Erklaerungsabsatz
              text={seite.text}
              klasse="max-w-prose text-leise"
            />
          </dd>
        </div>
      </dl>
      <Knopf gewicht="neben" aufKlick={zurUebersicht}>
        {texte.seiten.hilfe.seitenuebersicht}
      </Knopf>
    </Flaeche>
  )
}
