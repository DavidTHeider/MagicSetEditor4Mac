// Die Stiloptionen eines Sets (#41).
//
// **Sie kommen aus dem Paket, nicht aus einer Liste bei uns.** Was hier steht,
// ist die Darstellung: eine Auswahl wird ein `select`, ein Schalter ein
// `checkbox`, alles andere ein Textfeld. Die Beschreibung stammt aus der
// Vorlage und bleibt deshalb unübersetzt (§9).
import { useCallback, useState } from 'react'

import { api, auspacken, type StiloptionDTO } from '../api'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { Stiloption } from './Stiloption'
import { falte } from './stiloptionsfalte'
import { useTexte } from '../sprache/kontext'

interface Eigenschaften {
  setID: string
  geaendert: () => void

  // getippt meldet den Stand eines Textfeldes, während jemand darin tippt —
  // `undefined`, sobald geschrieben ist
  // (#590).
  //
  // **Der Zustand liegt oben, nicht hier.** Die Vorschau steht neben dieser
  // Liste und nicht darin; ein zweiter Halter wäre ein zweiter Stand derselben
  // Sache (§7, Regel 2).
  getippt?: (probe: { name: string; wert: string } | undefined) => void
}

export function Stiloptionen({ setID, geaendert, getippt }: Eigenschaften) {
  const texte = useTexte()
  const vorgang = useVorgang()

  const holen = useCallback(async () => {
    const antwort = await api().Stiloptionen(setID)

    return auspacken(antwort, () => antwort.optionen)
  }, [setID])
  const { daten, fehler, nachladen } = useLaden<StiloptionDTO[]>(holen, [setID])

  const [schreibfehler, setzeSchreibfehler] = useState<string>()

  // **Ein Vorgang je Datensatz** (§7, Regel 4): der Datensatz ist die Option.
  // Zwei Läufe auf dieselbe gingen gegen einen Stand, den das Backend noch
  // nicht kennt.
  const setzen = async (option: StiloptionDTO, wert: string) => {
    await vorgang.starte(option.name, async () => {
      setzeSchreibfehler(undefined)

      try {
        const antwort = await api().StiloptionSetzen(setID, option.name, wert)

        auspacken(antwort, () => undefined)

        // **Erst räumen, dann nachladen.** Ab hier steht der Wert im Bestand,
        // und die Vorschau soll ihn von dort nehmen — dieselbe Reihenfolge wie
        // im Karteneditor beim Verlassen eines Feldes.
        getippt?.(undefined)

        nachladen()
        geaendert()
      } catch (ursache: unknown) {
        setzeSchreibfehler(meldungAus(ursache, texte))
      }
    })
  }

  // **Zurück auf die Vorgabe** — es entfernt den gespeicherten Wert
  // (#764).
  //
  // **Derselbe Vorgangsschlüssel wie [setzen]** (§7, Regel 4): der Datensatz
  // ist die Option, und ob sie gesetzt oder geräumt wird, ändert daran nichts.
  const zuruecksetzen = async (option: StiloptionDTO) => {
    await vorgang.starte(option.name, async () => {
      setzeSchreibfehler(undefined)

      try {
        const antwort = await api().StiloptionZuruecksetzen(setID, option.name)

        auspacken(antwort, () => undefined)

        getippt?.(undefined)

        nachladen()
        geaendert()
      } catch (ursache: unknown) {
        setzeSchreibfehler(meldungAus(ursache, texte))
      }
    })
  }

  if (daten === undefined || daten.length === 0) {
    return fehler === undefined ? null : (
      <p role="alert" className="mt-3 text-fehler">
        {texte.bausteine.zustaende.fehler} {fehler}
      </p>
    )
  }

  // zeile ist eine Option samt ihren Wegen — vorn wie hinter dem Falz
  // dieselbe.
  const zeile = (option: StiloptionDTO) => (
    <Stiloption
      key={option.name}
      setID={setID}
      option={option}
      laeuft={vorgang.laeuft(option.name)}
      setzen={(wert) => {
        void setzen(option, wert)
      }}
      tippt={(wert) => {
        getippt?.({ name: option.name, wert })
      }}
      zuruecksetzen={() => {
        void zuruecksetzen(option)
      }}
    />
  )

  // **Vier vorn, der Rest dahinter** — R3 aus `docs/UIREVIEW.md` und das
  // Artboard *Eigenschaften* (UXR-07, #1915); die Regel steht in
  // `stiloptionsfalte.ts`.
  const { vorn, weitere, gezeigt, gesamt } = falte(daten)

  return (
    <div className="mt-4">
      {/* **Die Zahl neben der Überschrift und nicht darin** — sonst hieße
          die Überschrift für ein Vorleseprogramm „Stiloptionen4 von 7“. */}
      <div className="flex items-baseline gap-2">
        <h3 className="font-semibold">{texte.seiten.karten.stiloptionen}</h3>
        {weitere.length > 0 && (
          <span className="text-sm text-leise">
            {einsetzen(texte.seiten.karten.stiloptionenSichtbar, {
              zahl: String(gezeigt),
              gesamt: String(gesamt),
            })}
          </span>
        )}
      </div>

      {schreibfehler !== undefined && (
        <p role="alert" className="mt-2 text-fehler">
          {texte.bausteine.zustaende.fehler} {schreibfehler}
        </p>
      )}

      <div className="mt-2 grid max-w-md gap-3">{vorn.map(zeile)}</div>

      {/* **Eingeklappt, nicht entfernt** — der Kasten zeichnet erst beim
          Öffnen und behält danach, was darin steht. */}
      {weitere.length > 0 && (
        <div className="mt-3">
          <Aufklappkasten
            titel={einsetzen(texte.seiten.karten.stiloptionenWeitere, {
              zahl: String(gesamt - gezeigt),
            })}
          >
            <div className="grid max-w-md gap-3">{weitere.map(zeile)}</div>
          </Aufklappkasten>
        </div>
      )}
    </div>
  )
}
