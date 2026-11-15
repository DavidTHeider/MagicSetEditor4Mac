// **Suchen und Ersetzen über ein Set** — Werkstück 2 von Phase 3.7
// (#1168).
//
// # Warum eine eigene Lasche und kein Abschnitt der Suche
//
// MSE2 hat **einen** Dialog für beides (`wxFindReplaceDialog`,
// `src/gui/set/window.cpp`, Zeilen 770–795, Fassung f9d9356 vom 2021-07-06) —
// und das wäre hier die falsche Übernahme: die Volltextsuche geht über **alle**
// Sets und hat dafür eine Setwahl, das Ersetzen über **eines**. Beides
// nebeneinander lüde dazu ein, im falschen Set zu ersetzen; §6 verbietet
// verschachtelte Dialoge, und ein Dialog ist es hier ohnehin nicht.
//
// # Die Vorschau ist ein eigener Schritt
//
// Erst zeigen, dann tun. Eine Ersetzung über dreihundert Karten ist nicht zu
// überblicken, und ein vertipptes Wort schriebe sie alle. **Zurückgenommen
// wird sie als ein Zug** — das sagt die Zeile unter dem Knopf, damit niemand
// es raten muss.
import { useId, useState } from 'react'

import { api, type ErsetzungsAntwort, type ErsetzungsstelleDTO } from '../api'
import { Knopf } from '../bausteine/Knopf'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface ErsetzenEigenschaften {
  setID: string

  // geaendert meldet, dass der Bestand sich geändert hat.
  geaendert: () => void
}

// Stand ist, was zuletzt zurückkam — und ob es getan wurde.
interface Stand {
  antwort: ErsetzungsAntwort
  getan: boolean

  // fuer ist die Frage, zu der dieser Stand gehört.
  //
  // **Ohne sie war der Bestätigungsschritt keiner** (Befund 5 des Zweitblicks
  // über Phase 3.7): der Knopf „Jetzt ersetzen" stand weiter da, wenn jemand
  // nach der Vorschau den Suchtext änderte — und führte dann die **neue**,
  // nie gezeigte Frage aus, unter der Trefferzahl der alten.
  fuer: string
}

// alsSchluessel macht aus einer Frage eine vergleichbare Zeichenkette.
//
// **Alle vier Angaben, nicht nur der Suchtext.** Groß/klein und „ganzes Wort"
// ändern, was getroffen wird; ein Stand, der nur am Text hinge, gälte nach
// einem Klick auf das Kästchen weiter.
function alsSchluessel(frage: {
  text: string
  ersatz: string
  grossKleinBeachten: boolean
  ganzesWort: boolean
}): string {
  return JSON.stringify([
    frage.text,
    frage.ersatz,
    frage.grossKleinBeachten,
    frage.ganzesWort,
  ])
}

export function Ersetzen({ setID, geaendert }: ErsetzenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.ersetzen

  const [text, setzeText] = useState('')
  const [ersatz, setzeErsatz] = useState('')
  const [grossKleinBeachten, setzeGrossKlein] = useState(false)
  const [ganzesWort, setzeGanzesWort] = useState(false)

  const [stand, setzeStand] = useState<Stand>()
  const [fehler, setzeFehler] = useState<string>()

  const vorgang = useVorgang()

  const textID = useId()
  const ersatzID = useId()

  // **Ein Vorgang für beide Wege** (§7, Regel 4). Der Datensatz ist die
  // Ersetzung, und die Vorschau ist ihr erster Schritt: wer beides zugleich
  // anstieße, führte gegen einen Stand aus, den die Vorschau nicht kennt.
  const laeuft = vorgang.laeuft('ersetzen')

  // **`stand` ist beim Fragen leer** — er entsteht erst mit der Vorschau; beim
  // Bestätigen wird er unten eingesetzt (#1241).
  const frage = {
    text,
    ersatz,
    feld: '',
    grossKleinBeachten,
    ganzesWort,
    stand: '',
  }

  // **Der Stand gilt nur für die Frage, aus der er stammt.** Er wird nicht
  // geräumt, sondern **verglichen**: eine Vorschau, die noch dasteht, ist eine
  // Auskunft über die Frage von vorhin, und die bleibt wahr.
  const schluessel = alsSchluessel(frage)
  const giltNoch = stand?.fuer === schluessel

  // **Der Stand, den eine Bestätigung mitgibt** (#1241) — leer, solange keine
  // gültige Vorschau dasteht.
  //
  // **Hier und nicht im Rückruf**: dort liefe nur der eine Zweig, denn der
  // Knopf *Jetzt ersetzen* steht nur da, wo `giltNoch` gilt. Ein Zweig, den
  // kein Aufruf betritt, ist eine Behauptung darüber, was im Ernstfall
  // geschieht (§3).
  const bestaetigterStand = giltNoch ? stand.antwort.stand : ''

  const starte = (getan: boolean) =>
    vorgang.starte('ersetzen', async () => {
      try {
        // **Beim Bestätigen geht der Stand der Vorschau mit** (#1241).
        //
        // Zwischen Vorschau und Bestätigung kann eine Karte durch den Editor,
        // eine Rücknahme, den Dienst oder den unabhängigen MCP-Prozess
        // geändert worden sein — dann bestätigte der Knopf **andere Treffer
        // als die angezeigten**. Der Dienst weist das ab; hier steht, dass er
        // die Angabe bekommt.
        const antwort = getan
          ? await api().ErsetzungAusfuehren(setID, {
              ...frage,
              stand: bestaetigterStand,
            })
          : await api().ErsetzungVorschau(setID, frage)

        auspacken(antwort, () => antwort)
        setzeFehler(undefined)
        setzeStand({ antwort, getan, fuer: alsSchluessel(frage) })

        // **Regel 3**: geschrieben ist geschrieben. Das Melden ist
        // Bestenfalls-Arbeit — es nimmt die Änderung nicht zurück.
        if (getan) {
          geaendert()
        }
      } catch (ursache: unknown) {
        // **Nach einem gescheiterten Schreiben bleibt der alte Stand stehen**
        // (§7, Regel 3): eine geräumte Vorschau nähme die Angabe weg, an der
        // sich ablesen lässt, was schiefging.
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <div className="flex flex-col gap-4">
      <Erklaerungsabsatz text={eigene.erklaerung} klasse="text-sm text-leise" />

      <div className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1 text-sm" htmlFor={textID}>
          {eigene.suchtext}
          <input
            id={textID}
            type="search"
            value={text}
            onChange={(ereignis) => {
              setzeText(ereignis.target.value)
            }}
            className="w-64 rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
        </label>

        <label className="flex flex-col gap-1 text-sm" htmlFor={ersatzID}>
          {eigene.ersatz}
          <input
            id={ersatzID}
            type="text"
            value={ersatz}
            onChange={(ereignis) => {
              setzeErsatz(ereignis.target.value)
            }}
            className="w-64 rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
        </label>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={grossKleinBeachten}
            onChange={(ereignis) => {
              setzeGrossKlein(ereignis.target.checked)
            }}
            className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
          {eigene.grossKlein}
        </label>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={ganzesWort}
            onChange={(ereignis) => {
              setzeGanzesWort(ereignis.target.checked)
            }}
            className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
          {eigene.ganzesWort}
        </label>

        <Knopf
          gewicht="haupt"
          angehalten={laeuft || text.trim() === ''}
          aufKlick={() => {
            void starte(false)
          }}
        >
          {eigene.vorschauen}
        </Knopf>

        {/* **Ersetzen erscheint erst, wenn eine Vorschau dasteht.** Ein Knopf,
            der ohne Vorschau schreibt, macht den Schritt davor überflüssig —
            und genau dieser Schritt ist die Zusage des Werkstücks. */}
        {giltNoch && !stand.getan && stand.antwort.treffer > 0 && (
          <Knopf
            angehalten={laeuft}
            aufKlick={() => {
              void starte(true)
            }}
          >
            {einsetzen(eigene.ersetzen, {
              treffer: String(stand.antwort.treffer),
            })}
          </Knopf>
        )}
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {stand !== undefined && <Ergebnis stand={stand} />}
    </div>
  )
}

// Ergebnis zeigt, was sich ändert — oder was sich geändert hat.
function Ergebnis({ stand }: { stand: Stand }) {
  const texte = useTexte()
  const eigene = texte.seiten.ersetzen
  const { antwort, getan } = stand

  if (antwort.treffer === 0) {
    return <p className="text-sm text-leise">{eigene.nichtsGefunden}</p>
  }

  return (
    <section aria-label={eigene.stellen} className="flex flex-col gap-2">
      <h3 className="text-sm font-medium">
        {einsetzen(getan ? eigene.getan : eigene.zusammenfassung, {
          treffer: String(antwort.treffer),
          karten: String(antwort.karten),
        })}
      </h3>

      {/* **Die Zusage der Rücknahme steht dabei** — sie ist der Grund, warum
          man das hier wagen kann. */}
      {getan && <p className="text-sm text-leise">{eigene.ruecknahme}</p>}

      <ul className="flex flex-col gap-2">
        {antwort.stellen.map((stelle) => (
          <Stelle key={`${stelle.karteID}/${stelle.feld}`} stelle={stelle} />
        ))}
      </ul>

      {antwort.weitere && (
        <p className="text-sm text-leise">{eigene.weitere}</p>
      )}
    </section>
  )
}

// Stelle zeigt einen Feldwert vorher und nachher.
function Stelle({ stelle }: { stelle: ErsetzungsstelleDTO }) {
  const texte = useTexte()

  return (
    <li className="rounded-md border border-kante p-2 text-sm">
      <p className="text-leise">
        {einsetzen(texte.seiten.ersetzen.kartenfeld, {
          nummer: String(stelle.nummer),
          feld: stelle.feld,
        })}
      </p>

      {/* **Vorher und nachher untereinander**, nicht als Wortdiff: was sich
          ändert, ist oft ein Wort in einem Satz, und zwei Zeilen sind zu
          lesen. Ein Zeichenvergleich wäre die zweite Antwort auf eine Frage,
          die der Suchtext schon beantwortet. */}
      <p className="text-leise line-through">{stelle.vorher}</p>
      <p>{stelle.nachher}</p>
    </li>
  )
}
