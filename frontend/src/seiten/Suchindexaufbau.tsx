// **Der Aufbau des Suchindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum er neben der Suche steht und nicht darin
//
// **Weil er etwas anderes tut.** Suchen liest; Aufbauen schreibt, läuft
// Minuten und tauscht am Ende die aktive Generation. Beides in einer
// Komponente hieße, dass `Suchindex.tsx` über die Zeilengrenze ginge — und
// der Schnitt folgt der Frage, nicht der Zahl (§2).
//
// # Der Pfad steht als Text da
//
// **Und nicht hinter einem Dateidialog.** Eine Sammeldatei von Scryfall wird
// geladen und liegt danach an einem festen Ort; wer den Index baut, kennt
// ihn. Ein Dialog wäre die bequemere Gestalt — er gehört zu der Produktfläche,
// die A16 entwirft, und nicht zu einem Verwaltungsaufruf.
import { useId, useState } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { api } from '../fassade'
import type { IndexaufbauAntwort } from '../apiindex'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

// feldklassen sind die Klassen eines freien Eingabefelds.
//
// **Mit sichtbarem Fokusring** (§6): eine Fläche, die nur mit der Maus
// bedienbar aussieht, ist für jemanden, der mit der Tastatur arbeitet, nicht
// bedienbar — und ein Wächter zählt es nach.
const feldklassen =
  'rounded-sm border border-kante bg-grund px-2 py-1 ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Suchindexaufbau({ danach }: { danach: () => void }) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const vorgang = useVorgang()
  const ueberschrift = useId()

  const [modell, setzeModell] = useState('')
  const [quelle, setzeQuelle] = useState('')
  const [revision, setzeRevision] = useState('')
  const [fortschreiben, setzeFortschreiben] = useState(false)
  const [bericht, setzeBericht] = useState<IndexaufbauAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  const aufbauen = () =>
    vorgang.starte('indexaufbau', async () => {
      try {
        const antwort = await api().IndexAufbauen(
          modell,
          quelle,
          revision,
          '',
          fortschreiben,
        )

        setzeBericht(auspacken(antwort, () => antwort))
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      } finally {
        // **Der Stand wird danach neu geholt — auch nach einem Fehlschlag**
        // (§7 Regel 3): ein gescheiterter Aufbau hinterlässt einen
        // Fehlerzustand, und der ist die Auskunft, die sagt, was war.
        danach()
      }
    })

  return (
    <div
      className="flex flex-col gap-2 border-t border-kante pt-3"
      aria-labelledby={ueberschrift}
      role="group"
    >
      <h4 id={ueberschrift} className="text-sm font-semibold">
        {eigene.suchindexAufbau}
      </h4>

      <Erklaerungsabsatz
        text={eigene.suchindexAufbauErklaerung}
        klasse="text-sm text-leise"
      />

      <div className="flex flex-wrap items-end gap-2">
        <Aufbaufeld
          aufschrift={eigene.suchindexModellfeld}
          wert={modell}
          setze={setzeModell}
        />

        <Aufbaufeld
          aufschrift={eigene.suchindexQuellfeld}
          wert={quelle}
          setze={setzeQuelle}
        />

        <Aufbaufeld
          aufschrift={eigene.suchindexRevisionfeld}
          wert={revision}
          setze={setzeRevision}
        />

        {/* **Der Schalter ist kein Feld** — er sagt, *wie* gebaut wird,
            und nicht *woraus*. */}
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={fortschreiben}
            onChange={(ereignis) => {
              setzeFortschreiben(ereignis.target.checked)
            }}
          />

          {eigene.suchindexDelta}
        </label>

        <Knopf
          gewicht="neben"
          aufKlick={() => void aufbauen()}
          angehalten={vorgang.laeuft('indexaufbau')}
        >
          {eigene.suchindexAufbauen}
        </Knopf>
      </div>

      {bericht && (
        <p className="text-sm">
          {eigene.suchindexFertig} {bericht.generation}: {bericht.geschrieben}{' '}
          {eigene.suchindexGeschrieben}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}

// Aufbaufeld ist eines der drei Eingabefelder.
function Aufbaufeld({
  aufschrift,
  wert,
  setze,
}: {
  aufschrift: string
  wert: string
  setze: (frisch: string) => void
}) {
  return (
    <label className="flex flex-col gap-1 text-sm">
      {aufschrift}

      <input
        className={feldklassen}
        value={wert}
        onChange={(ereignis) => {
          setze(ereignis.target.value)
        }}
      />
    </label>
  )
}
