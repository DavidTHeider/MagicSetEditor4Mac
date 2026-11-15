// **Marken vergeben und pflegen** — UX-12.AK-02
// (#1684).
//
// Zwei Bausteine für den Navigator: die **Vergabe** einer Marke an alle
// gewählten Objekte — ganz oder gar nicht — und die **Pflege** einer Marke
// über alle Arten: umbenennen, zusammenführen, überall entfernen. Die Regeln
// dafür stehen in Go (`service.Objektmarken`): auf einen vorhandenen Namen
// nur ausdrücklich, und eine Marke ist nie ein Recht.
import { useId, useState } from 'react'

import { api } from '../api'
import type { Markenart } from '../apiobjektmarken'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Aktionsknopf } from './Aktionsknopf'
import { Entscheidung } from './Entscheidung'

const feldklasse =
  'min-w-0 flex-1 rounded-sm border border-kante bg-transparent px-2 py-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'
const knopfklasse =
  'rounded-sm border border-kante px-2 py-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60'

interface MarkenvergabeEigenschaften {
  art: Markenart
  kennungen: readonly string[]
  geaendert: () => void
}

// Markenvergabe hängt eine Marke an alle gewählten Objekte.
export function Markenvergabe({
  art,
  kennungen,
  geaendert,
}: MarkenvergabeEigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.objektmarken
  const vorgang = useVorgang()
  const feld = useId()
  const [marke, setzeMarke] = useState('')
  const [fehler, setzeFehler] = useState<string>()

  return (
    <form
      aria-label={eigene.vergebenUeberschrift}
      className="flex flex-col gap-1 text-sm"
      onSubmit={(ereignis) => {
        ereignis.preventDefault()
        void vorgang.starte('vergabe', async () => {
          try {
            auspacken(
              await api().MarkeVergeben(art, [...kennungen], marke),
              () => undefined,
            )
            setzeMarke('')
            setzeFehler(undefined)
            geaendert()
          } catch (ursache: unknown) {
            setzeFehler(meldungAus(ursache, texte))
          }
        })
      }}
    >
      <label htmlFor={feld}>
        {einsetzen(eigene.vergebenName, { anzahl: String(kennungen.length) })}
      </label>
      <div className="flex gap-2">
        <input
          id={feld}
          value={marke}
          onChange={(ereignis) => {
            setzeMarke(ereignis.target.value)
          }}
          className={feldklasse}
        />
        <button
          type="submit"
          disabled={marke.trim() === '' || vorgang.laeuft('vergabe')}
          className={knopfklasse}
        >
          {eigene.vergeben}
        </button>
      </div>
      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </form>
  )
}

interface MarkenpflegeEigenschaften {
  // runde lädt den Katalog neu, geaendert sagt es dem Aufrufer.
  runde: number
  geaendert: () => void
}

// Markenpflege benennt um, führt zusammen und entfernt überall.
export function Markenpflege({ runde, geaendert }: MarkenpflegeEigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.objektmarken
  const vorgang = useVorgang()
  const wahlfeld = useId()
  const namensfeld = useId()

  const katalog = useLaden(async () => {
    const antwort = await api().Markenkatalog()

    return auspacken(antwort, () => antwort.marken)
  }, [runde])

  const [gewaehlt, setzeGewaehlt] = useState('')
  const [neu, setzeNeu] = useState('')
  const [zusammen, setzeZusammen] = useState(false)
  // fragt hält die Marke, **für die** gefragt wird — nicht die gerade
  // gewählte (§7, Regel 1).
  const [fragt, setzeFragt] = useState<{ marke: string; anzahl: number }>()
  const [meldung, setzeMeldung] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()

  const posten = katalog.daten ?? []
  const marke = posten.find((einer) => einer.marke === gewaehlt) ?? posten[0]

  if (marke === undefined) {
    return <p className="text-sm text-leise">{eigene.pflegeKeine}</p>
  }

  const ausfuehren = (lauf: () => Promise<{ anzahl: number }>) =>
    vorgang.starte('pflege', async () => {
      try {
        const antwort = await lauf()
        setzeMeldung(
          einsetzen(eigene.erledigt, { anzahl: String(antwort.anzahl) }),
        )
        setzeFehler(undefined)
        setzeNeu('')
        geaendert()
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const umbenennen = () => {
    if (neu.trim() === '') {
      return
    }

    void ausfuehren(async () => {
      const antwort = await api().MarkeUmbenennen(marke.marke, neu, zusammen)

      return auspacken(antwort, () => antwort)
    })
  }

  return (
    // **Die Überschrift trägt der Kasten darum** — die Pflege steht seit
    // UXR-10 eingeklappt im Navigator der Sammlung.
    <section
      aria-label={eigene.pflegeUeberschrift}
      className="flex flex-col gap-2 text-sm"
    >
      <label htmlFor={wahlfeld}>{eigene.pflegeWahl}</label>
      <select
        id={wahlfeld}
        value={marke.marke}
        onChange={(ereignis) => {
          setzeGewaehlt(ereignis.target.value)
          setzeFragt(undefined)
          setzeNeu('')
          setzeZusammen(false)
        }}
        className="rounded-sm border border-kante bg-transparent px-2 py-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      >
        {posten.map((einer) => (
          <option key={einer.marke} value={einer.marke}>
            {einsetzen(eigene.pflegePosten, {
              marke: einer.marke,
              anzahl: String(einer.anzahl),
            })}
          </option>
        ))}
      </select>

      <form
        className="flex flex-col gap-1"
        onSubmit={(ereignis) => {
          ereignis.preventDefault()
          umbenennen()
        }}
      >
        <label htmlFor={namensfeld}>{eigene.umbenennenName}</label>
        <div className="flex gap-2">
          <input
            id={namensfeld}
            value={neu}
            onChange={(ereignis) => {
              setzeNeu(ereignis.target.value)
            }}
            className={feldklasse}
          />
          <Aktionsknopf
            aktion="bearbeiten"
            klein
            angehalten={neu.trim() === '' || vorgang.laeuft('pflege')}
            aufKlick={umbenennen}
          >
            {eigene.umbenennen}
          </Aktionsknopf>
        </div>
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={zusammen}
            onChange={(ereignis) => {
              setzeZusammen(ereignis.target.checked)
            }}
          />
          {eigene.zusammenfuehren}
        </label>
      </form>

      <div>
        <Aktionsknopf
          aktion="entfernen"
          klein
          angehalten={vorgang.laeuft('pflege')}
          aufKlick={() => {
            setzeFragt(marke)
          }}
        >
          {eigene.ueberallEntfernen}
        </Aktionsknopf>
      </div>

      {fragt !== undefined && (
        <Entscheidung
          frage={einsetzen(eigene.ueberallFrage, {
            marke: fragt.marke,
            anzahl: String(fragt.anzahl),
          })}
          erklaerung={eigene.ueberallErklaerung}
          bestaetigen={eigene.ueberallLoeschenJa}
          aktion="loeschen"
          abbrechen={texte.bausteine.aktionen.abbrechen}
          laeuft={vorgang.laeuft('pflege')}
          aufBestaetigen={() => {
            setzeFragt(undefined)
            void ausfuehren(async () => {
              const antwort = await api().MarkeUeberallEntfernen(fragt.marke)

              return auspacken(antwort, () => antwort)
            })
          }}
          aufAbbrechen={() => {
            setzeFragt(undefined)
          }}
        />
      )}

      {meldung !== undefined && fehler === undefined && (
        <p role="status">{meldung}</p>
      )}
      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
