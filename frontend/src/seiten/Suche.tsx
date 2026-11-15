// Die Volltextsuche.
//
// **Gesucht wird in der Datenbank, nicht hier** — über FTS5, und das ist der
// Grund, warum SQLite das Arbeitsformat ist (`docs/entscheidungen/0002`). Ein
// Durchlauf im Frontend bräuchte alle Karten im Arbeitsspeicher, und genau das
// tut MSE2.
//
// **Die Filterwerte kommen aus dem Bestand**, nicht aus einer Liste hier: ein
// fremdes Spiel führt Felder, die wir nicht kennen, und eine selbst
// geschriebene Aufzählung ließe ausgerechnet die nicht filtern.
import { useCallback, useId, useState } from 'react'

import { api, auspacken, type SetDTO, type TrefferDTO } from '../api'
import { useLaden } from '../haken/useLaden'
import { Seitenladung } from '../bausteine/Seitenladung'
import { useTexte } from '../sprache/kontext'
import { Auswahlfeld } from '../bausteine/Auswahlfeld'
import { Knopf } from '../bausteine/Knopf'
import { seitengroesse } from '../bausteine/Listenansicht'
import { einsetzen } from '../sprache/einsetzen'
import { Ladefeld } from './Ladefeld'
import { ohneAuszeichnung } from './ohneauszeichnung'
import { Wissensrecherche } from './Wissensrecherche'

interface SucheEigenschaften {
  // oeffnen bringt einen Treffer in den Karteneditor.
  oeffnen: (setID: string, karteID: string) => void

  // offenesSet ist das Set, aus dem gesucht wird — der Umfang beginnt dort
  // (AP-11, #2368, B-17); „Alle Sets“ steht daneben.
  offenesSet?: string
}

interface Ergebnis {
  treffer: TrefferDTO[]
  gekuerzt: boolean
}

export function Suche({ oeffnen, offenesSet = '' }: SucheEigenschaften) {
  const texte = useTexte()
  const [text, setzeText] = useState('')
  const [feld, setzeFeld] = useState('')
  const [setID, setzeSetID] = useState(offenesSet)

  // **Die Recherche geht auf, statt dazustehen** (AP-11, B-17, `Ruhe09`):
  // sie ist eine eigene Aufgabe neben dem Feld. Einmal offen, bleibt sie
  // gezeichnet — eine halb getippte Frage geht beim Zuklappen nicht verloren.
  const [recherche, setzeRecherche] = useState(false)
  const [rechercheGezeichnet, setzeRechercheGezeichnet] = useState(false)
  if (recherche && !rechercheGezeichnet) {
    setzeRechercheGezeichnet(true)
  }
  const rechercheID = useId()

  const feldwahlID = useId()
  const setzwahlID = useId()

  const holenSets = useCallback(async () => {
    const antwort = await api().Sets()

    return auspacken(antwort, () => antwort.sets)
  }, [])

  // **Die Feldliste hängt am gewählten Set** (§7, Regel 1): wer das Set
  // wechselt, bekommt dessen Felder — und nicht die des vorigen.
  const holenFelder = useCallback(async () => {
    const antwort = await api().Suchfelder(setID)

    return auspacken(antwort, () => antwort.felder)
  }, [setID])

  const holenTreffer = useCallback(async (): Promise<Ergebnis> => {
    if (text.trim() === '') {
      return { treffer: [], gekuerzt: false }
    }

    const antwort = await api().Suchen(text, feld, setID)

    return auspacken(antwort, () => ({
      treffer: antwort.treffer,
      gekuerzt: antwort.gekuerzt,
    }))
  }, [text, feld, setID])

  const sets = useLaden<SetDTO[]>(holenSets, [])
  const felder = useLaden<string[]>(holenFelder, [setID])

  // **Alle drei Filter stehen in den Abhängigkeiten.** Wer einen ändert, sucht
  // neu — und die Antwort auf die vorige Frage trägt sich nicht ein (§7,
  // Regel 2).
  const ergebnis = useLaden<Ergebnis>(holenTreffer, [text, feld, setID])

  const gesucht = text.trim() !== ''

  return (
    <>
      {/* **Eine Lasche ist ein Abschnitt, keine zweite Seite**
          (#812).
          Der `Setarbeitsplatz` trägt die erste Ebene mit dem Namen des Sets. */}
      <h2 className="text-2xl font-semibold tracking-tight">
        {texte.seiten.suche.titel}
      </h2>

      {/* **Alle Filter sind per Tastatur erreichbar** und stehen in der
          Reihenfolge, in der man sie benutzt: erst wonach, dann wo. */}
      <div className="mt-4 flex flex-wrap items-end gap-4">
        {/* **Gedeckelt.** Ein Suchfeld, das die ganze Fensterbreite nimmt,
            schiebt die beiden Filter an den rechten Rand — und dann liegt
            zwischen zusammengehörenden Bedienelementen ein halber Bildschirm. */}
        <label className="flex max-w-xl flex-1 flex-col gap-1">
          <span className="text-sm font-medium">
            {texte.seiten.suche.eingabe}
          </span>
          <input
            type="search"
            value={text}
            onChange={(ereignis) => {
              setzeText(ereignis.target.value)
            }}
            className="min-w-48 rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
        </label>

        {/* **Feld und Set sind Auswahlfelder** — UXR-F1
            (#2007),
            R7: ab fünfzehn Werten mit Suche, darunter nativ. **Die
            Beschriftung steht neben dem Feld und nicht um es**, damit der
            Name des Feldes nur sie ist. */}
        <div className="flex flex-col gap-1">
          <label htmlFor={feldwahlID} className="text-sm font-medium">
            {texte.seiten.suche.feld}
          </label>
          <Auswahlfeld
            id={feldwahlID}
            bezeichnung={texte.seiten.suche.feld}
            eintraege={[
              { wert: '', name: texte.seiten.suche.alleFelder },
              ...(felder.daten ?? []).map((name) => ({ wert: name, name })),
            ]}
            wert={feld}
            aendern={setzeFeld}
          />
        </div>

        <div className="flex flex-col gap-1">
          <label htmlFor={setzwahlID} className="text-sm font-medium">
            {texte.seiten.suche.set}
          </label>
          <Auswahlfeld
            id={setzwahlID}
            bezeichnung={texte.seiten.suche.set}
            eintraege={[
              { wert: '', name: texte.seiten.suche.alleSets },
              ...(sets.daten ?? []).map((satz) => ({
                wert: satz.id,
                name: satz.name,
              })),
            ]}
            wert={setID}
            aendern={(neu) => {
              // **Das Feld wird mitgeräumt** (§7, Regel 1): ein Feld, das nur
              // das vorige Set führt, filterte im neuen alles weg — und die
              // leere Trefferliste sähe aus, als gäbe es dort nichts.
              setzeSetID(neu)
              setzeFeld('')
            }}
          />
        </div>

        {/* **„Recherche …“ nach den Filtern**, in der Tabfolge danach. */}
        <Knopf
          gewicht="leise"
          ausgeklappt={recherche}
          steuert={rechercheID}
          aufKlick={() => {
            setzeRecherche((war) => !war)
          }}
        >
          {texte.seiten.suche.recherche}
        </Knopf>
      </div>

      {/* **Die Filter laden mit der Seite** — G11 (#2277): ohne diesen
          Hinweis stand die Seite bei langsamen oder gescheiterten Feldern
          da wie immer, nur mit leerer Auswahl. */}
      <Seitenladung zustaende={[sets, felder]} />

      {gesucht ? (
        <>
          <Ladefeld
            laedt={ergebnis.laedt}
            fehler={ergebnis.fehler}
            leer={ergebnis.daten?.treffer.length === 0}
          >
            <p className="mt-4 max-w-prose text-leise">
              {texte.seiten.suche.nichts}
            </p>
          </Ladefeld>

          {ergebnis.daten && ergebnis.daten.treffer.length > 0 && (
            <>
              {/* **Die Kürzung ist zu sehen.** Eine stille sähe aus wie ein
                  vollständiges Ergebnis. */}
              <p role="status" className="mt-4 text-sm text-leise">
                {ergebnis.daten.gekuerzt
                  ? texte.seiten.suche.gekuerzt
                  : `${String(ergebnis.daten.treffer.length)} ${texte.seiten.suche.treffer}`}
              </p>

              {/* **Der Schlüssel ist die Frage** (§7, Regel 1): eine neue
                  beginnt wieder bei den ersten Treffern. */}
              <Trefferliste
                key={`${text}/${feld}/${setID}`}
                treffer={ergebnis.daten.treffer}
                oeffnen={oeffnen}
              />
            </>
          )}
        </>
      ) : (
        <p className="mt-4 max-w-prose text-leise">{texte.seiten.suche.leer}</p>
      )}

      {/* **Die Recherche steht im selben Arbeitsplatz** (A16, #1515) — unter
          der Volltextsuche, und seit AP-11 erst, wenn „Recherche …“ sie
          aufschlägt. */}
      {rechercheGezeichnet && (
        <div id={rechercheID} hidden={!recherche}>
          <Wissensrecherche oeffnen={oeffnen} />
        </div>
      )}
    </>
  )
}

// Trefferliste zeigt die Treffer — **gefaltet**, UXR-05
// (#1913).
//
// Der Dienst schickt bis zu 200 (`domain.SearchLimit`), und jeder ist ein
// Knopf. Mit den Bedienelementen der Seite standen damit über 200 da — R6
// aus `docs/UIREVIEW.md` —, und jeder war ein Tabstopp. Gefaltet wird wie in
// der `Listenansicht`, mit derselben Seitengröße und demselben Satz.
function Trefferliste({
  treffer,
  oeffnen,
}: {
  treffer: readonly TrefferDTO[]
  oeffnen: (setID: string, karteID: string) => void
}) {
  const texte = useTexte()
  const [seiten, setzeSeiten] = useState(1)
  const gezeigt = treffer.slice(0, seiten * seitengroesse)

  return (
    <>
      <ul className="mt-3 divide-y divide-kante">
        {gezeigt.map((einer) => (
          <li
            key={`${einer.karteID}/${einer.feld}`}
            className="flex flex-wrap items-baseline gap-3 py-2"
          >
            <button
              type="button"
              onClick={() => {
                oeffnen(einer.setID, einer.karteID)
              }}
              className="text-marke underline underline-offset-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              {einer.kartenname === ''
                ? texte.seiten.karten.ohneName
                : ohneAuszeichnung(einer.kartenname)}
            </button>
            <span className="text-sm text-leise">{einer.feld}</span>
            {/* **Ohne die Auszeichnung** (#952): ein Treffer im Aromatext
                zeigte `<i-flavor>…</i-flavor>` statt des Satzes. Hier steht
                der Wert als Text, nicht als Karte. */}
            <span className="basis-full text-leise">
              {ohneAuszeichnung(einer.wert)}
            </span>
          </li>
        ))}
      </ul>

      {treffer.length > gezeigt.length && (
        <div className="mt-3">
          <Knopf
            aufKlick={() => {
              setzeSeiten(seiten + 1)
            }}
          >
            {einsetzen(texte.bausteine.listenansicht.weitere, {
              anzahl: String(
                Math.min(seitengroesse, treffer.length - gezeigt.length),
              ),
            })}
          </Knopf>
        </div>
      )}
    </>
  )
}
