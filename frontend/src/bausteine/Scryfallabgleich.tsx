// **Kartendaten von Scryfall übernehmen**
// (#1013).
//
// # Warum es diesen Kasten gibt
//
// `ScryfallVorschlag` und `ScryfallUebernehmen` waren angemeldet, standen im
// Fähigkeitskatalog und waren über MCP, CLI und REST erreichbar — **kein
// Bauteil rief sie**. Dieselbe Gestalt wie
// #852: ein
// Weg mit voller Abdeckung, den niemand erreichen kann.
//
// # Warum zwei Schritte
//
// §16, Regel 6: „Nichts wird still übernommen. Was von außen in ein Kartenfeld
// will, **zeigt vorher, was sich ändert**." Der Kasten zeigt je Feld **alt und
// neu**; ein Vorschlag, der nur den neuen Wert zeigte, ließe niemanden
// entscheiden.
//
// **Und nichts ist vorbelegt.** Angehakt wird von Hand — sonst wäre die
// Rückfrage eine Höflichkeit statt einer Grenze.
import { useState } from 'react'

import { Knopf } from './Knopf'
import type { Feldvorschlag, Scryfallvorschlag } from '../antworten'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

interface ScryfallabgleichEigenschaften {
  // kartenname ist der Name, mit dem gesucht wird.
  kartenname: string

  laeuft: boolean

  // vorschlag ist der geholte Stand — `undefined`, solange keiner da ist.
  vorschlag: Scryfallvorschlag | undefined

  // meldung steht **in diesem Kasten** (§6): der Karteneditor zeigt eine
  // Meldung nur am Feld, das sie verursacht hat, und ein Abgleich verursacht
  // keines. Sie fiele dort still aus.
  meldung: string | undefined

  holen: () => void
  uebernehmen: (felder: string[]) => void
}

export function Scryfallabgleich({
  kartenname,
  laeuft,
  vorschlag,
  meldung,
  holen,
  uebernehmen,
}: ScryfallabgleichEigenschaften): React.JSX.Element {
  const texte = useTexte().seiten.karten.abgleich

  const [angehakt, setzeAngehakt] = useState<string[]>([])

  return (
    <section className="mt-4 border-t border-kante pt-4">
      <h3 className="text-sm font-semibold">{texte.ueberschrift}</h3>
      <Erklaerungsabsatz
        text={texte.erklaerung}
        klasse="mt-1 text-xs text-leise"
      />

      <div className="mt-2 flex flex-wrap gap-2">
        <Knopf
          angehalten={laeuft || kartenname.trim() === ''}
          aufKlick={() => {
            setzeAngehakt([])
            holen()
          }}
        >
          {texte.holen}
        </Knopf>
      </div>

      {meldung !== undefined && (
        <p role="alert" className="mt-2 text-sm text-fehler">
          {meldung}
        </p>
      )}

      {vorschlag !== undefined && (
        <Vorschlagsliste
          angehakt={angehakt}
          anhaken={(feld) => {
            setzeAngehakt((bisher) =>
              bisher.includes(feld)
                ? bisher.filter((eines) => eines !== feld)
                : [...bisher, feld],
            )
          }}
          laeuft={laeuft}
          uebernehmen={() => {
            uebernehmen(angehakt)
          }}
          vorschlag={vorschlag}
        />
      )}
    </section>
  )
}

interface VorschlagslisteEigenschaften {
  vorschlag: Scryfallvorschlag
  angehakt: string[]
  laeuft: boolean
  anhaken: (feld: string) => void
  uebernehmen: () => void
}

// Vorschlagsliste zeigt, was sich ändern würde.
//
// **Eine eigene Funktion**, weil der Kasten sonst über die Zeilengrenze ginge
// (§2) — und der Schnitt folgt einer Frage: oben steht, **wie** man holt, hier,
// **was** dabei herauskam.
function Vorschlagsliste({
  vorschlag,
  angehakt,
  laeuft,
  anhaken,
  uebernehmen,
}: VorschlagslisteEigenschaften): React.JSX.Element {
  const texte = useTexte().seiten.karten.abgleich

  if (vorschlag.aenderungen.length === 0) {
    return <p className="mt-2 text-sm">{texte.nichtsZuTun}</p>
  }

  return (
    <div className="mt-3">
      <p className="text-xs text-leise">
        {vorschlag.alterStunden === 0
          ? texte.frisch
          : einsetzen(texte.alter, {
              stunden: String(Math.round(vorschlag.alterStunden)),
            })}{' '}
        <a
          className="underline focus-visible:outline-2"
          href={vorschlag.herkunft}
          rel="noreferrer"
          target="_blank"
        >
          {texte.quelle}
        </a>
      </p>

      <ul className="mt-2 space-y-2">
        {vorschlag.aenderungen.map((aenderung) => (
          <Zeile
            aenderung={aenderung}
            angehakt={angehakt.includes(aenderung.feld)}
            anhaken={anhaken}
            key={aenderung.feld}
          />
        ))}
      </ul>

      {vorschlag.befunde.length > 0 && (
        <ul className="mt-2 space-y-1 text-xs text-leise">
          {vorschlag.befunde.map((befund) => (
            <li key={befund}>{befund}</li>
          ))}
        </ul>
      )}

      <div className="mt-3">
        <Knopf
          angehalten={laeuft || angehakt.length === 0}
          aufKlick={uebernehmen}
          gewicht="haupt"
        >
          {texte.uebernehmen}
        </Knopf>
      </div>
    </div>
  )
}

// Zeile ist ein Feld mit seinem alten und seinem neuen Wert.
function Zeile({
  aenderung,
  angehakt,
  anhaken,
}: {
  aenderung: Feldvorschlag
  angehakt: boolean
  anhaken: (feld: string) => void
}): React.JSX.Element {
  const texte = useTexte().seiten.karten.abgleich

  const kennung = `abgleich-${aenderung.feld}`

  return (
    <li className="flex items-start gap-2 text-sm">
      <input
        checked={angehakt}
        className="mt-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        id={kennung}
        onChange={() => {
          anhaken(aenderung.feld)
        }}
        type="checkbox"
      />
      <label htmlFor={kennung}>
        <span className="font-medium">{aenderung.feld}</span>
        <span className="block text-xs text-leise">
          {texte.alt}: {aenderung.alt === '' ? texte.leer : aenderung.alt}
        </span>
        <span className="block text-xs">
          {texte.neu}: {aenderung.neu === '' ? texte.leer : aenderung.neu}
        </span>
      </label>
    </li>
  )
}
