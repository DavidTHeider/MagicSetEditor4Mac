// **Der lokale Suchindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum er hier steht und nicht auf einer eigenen Seite
//
// **Weil er heute ein Diagnose- und Verwaltungsaufruf ist.** Das Issue sagt
// es selbst: „Begrenzter tatsächlicher Diagnose-/Verwaltungsaufruf für
// Indexstatus, Aufbau/Abbruch und Suche vorhanden; **A16 erweitert ihn um
// Produktflächen**." Eine eigene Seite jetzt wäre die Produktfläche, die dort
// entworfen wird — und danach zweimal dieselbe Sache.
//
// # Der Modus steht über den Treffern
//
// **Weil man ihn an ihnen nicht sieht.** Eine lexikalische Antwort ist eine
// andere Auskunft als eine hybride; wer das verschwiege, verkaufte den
// Rückfall als die Sache selbst.
import { useId, useState } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { api } from '../fassade'
import type { IndexstandAntwort, IndexsucheAntwort } from '../apiindex'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { Suchindexaufbau } from './Suchindexaufbau'
import { auspacken } from '../umschlag'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

// derLeerfilter ist die Filterstruktur ohne jede Einschränkung.
//
// **Ausgeschrieben und nicht `{}`**: die Go-Seite erwartet jedes Feld, und
// ein fehlendes wäre dort ein Nullwert, den niemand geschrieben hat.
const derLeerfilter = {
  quellen: [],
  sprachen: [],
  sets: [],
  farben: [],
  farbidentitaet: [],
  typen: [],
  manawertVon: null,
  manawertBis: null,
  format: '',
  formatlage: '',
  formatstand: '',
}

// feldklassen sind die Klassen eines freien Eingabefelds.
//
// **Mit sichtbarem Fokusring** (§6): eine Fläche, die nur mit der Maus
// bedienbar aussieht, ist für jemanden, der mit der Tastatur arbeitet, nicht
// bedienbar — und ein Wächter zählt es nach.
const feldklassen =
  'rounded-sm border border-kante bg-grund px-2 py-1 ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Suchindex() {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const vorgang = useVorgang()
  const ueberschrift = useId()

  // **Der Stand kommt über den Ladehaken** (§7): er führt den Lauf mit und
  // trägt eine Antwort nur ein, wenn sie noch zum aktuellen Stand gehört —
  // von Hand geschriebene Wächter dafür sind die Stelle, an der es immer
  // wieder fehlt.
  const auskunft = useLaden<IndexstandAntwort>(() => api().Indexstand(), [])

  const [ergebnis, setzeErgebnis] = useState<IndexsucheAntwort>()
  const [frage, setzeFrage] = useState('')
  const [fehler, setzeFehler] = useState<string>()

  const stand = auskunft.daten

  const suchen = () =>
    vorgang.starte('indexsuche', async () => {
      try {
        const antwort = await api().Indexsuche(frage, 10, derLeerfilter)

        setzeErgebnis(auspacken(antwort, () => antwort))
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const abbrechen = () =>
    vorgang.starte('indexabbruch', async () => {
      await api().IndexAbbrechen()

      auskunft.nachladen()
    })

  return (
    <section
      className="flex flex-col gap-3 border-t border-kante pt-4"
      aria-labelledby={ueberschrift}
    >
      <h3 id={ueberschrift} className="text-lg font-semibold">
        {eigene.suchindex}
      </h3>

      <Erklaerungsabsatz
        text={eigene.suchindexErklaerung}
        klasse="text-sm text-leise"
      />

      {stand && !stand.aktiv && (
        <p className="text-sm text-leise">{eigene.suchindexKeiner}</p>
      )}

      {stand?.aktiv && <Indexauskunft stand={stand} />}

      {stand?.aufbau.laeuft && (
        <p className="flex items-center gap-2 text-sm">
          <span>
            {eigene.suchindexLaeuft}: {stand.aufbau.geschrieben}{' '}
            {eigene.suchindexGeschrieben}
          </span>

          <Knopf
            gewicht="leise"
            aufKlick={() => void abbrechen()}
            angehalten={vorgang.laeuft('indexabbruch')}
          >
            {eigene.suchindexAbbrechen}
          </Knopf>
        </p>
      )}

      {/* **Kein `<form>`, und das ist kein Versehen.** `Knopf` ist immer
          `type="button"`; ein Formular darum herum hätte keinen Absender.
          Die Tastatur erreicht die Suche über die Eingabetaste im Feld und
          über den Knopf selbst — beides ist ein Weg (§6). */}
      <div className="flex items-end gap-2">
        <label className="flex flex-1 flex-col gap-1 text-sm">
          {eigene.suchindexFrage}

          <input
            className={feldklassen}
            value={frage}
            onChange={(ereignis) => {
              setzeFrage(ereignis.target.value)
            }}
            onKeyDown={(ereignis) => {
              if (ereignis.key === 'Enter') {
                void suchen()
              }
            }}
          />
        </label>

        <Knopf
          gewicht="haupt"
          aufKlick={() => void suchen()}
          angehalten={vorgang.laeuft('indexsuche')}
        >
          {eigene.suchindexSuchen}
        </Knopf>
      </div>

      {ergebnis && <Trefferliste ergebnis={ergebnis} />}

      <Suchindexaufbau danach={auskunft.nachladen} />

      {(fehler ?? auskunft.fehler) !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler ?? auskunft.fehler}
        </p>
      )}
    </section>
  )
}

// Indexauskunft zeigt, wogegen gesucht wird.
function Indexauskunft({ stand }: { stand: IndexstandAntwort }) {
  const eigene = useTexte().seiten.technik

  return (
    <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-sm">
      <dt className="text-leise">{eigene.suchindexModell}</dt>
      <dd>{stand.manifest.modell}</dd>

      <dt className="text-leise">{eigene.suchindexLaufzeit}</dt>
      <dd>{stand.manifest.laufzeit}</dd>

      <dt className="text-leise">{eigene.suchindexVerfahren}</dt>
      <dd>
        {stand.manifest.suchkonfiguration} · {stand.manifest.abstandsmass} ·{' '}
        {stand.manifest.dimension}
      </dd>

      <dt className="text-leise">{eigene.suchindexStaende}</dt>
      <dd>
        <ul>
          {stand.manifest.korpusstaende.map((korpus) => (
            <li key={korpus.quelle}>
              {korpus.quelle} · {korpus.revision} {korpus.datum}
            </li>
          ))}
        </ul>
      </dd>

      <dt className="text-leise">{eigene.suchindexSemantisch}</dt>
      <dd>
        {stand.suchbar ? eigene.suchindexJa : eigene.suchindexNein}{' '}
        {stand.grund}
      </dd>
    </dl>
  )
}

// Trefferliste zeigt die Funde samt Modus.
function Trefferliste({ ergebnis }: { ergebnis: IndexsucheAntwort }) {
  const eigene = useTexte().seiten.technik

  const modus = {
    hybrid: eigene.suchindexModusHybrid,
    lexikalisch: eigene.suchindexModusLexikalisch,
    'ohne-index': eigene.suchindexModusOhneIndex,
  }[ergebnis.modus]

  return (
    <>
      <p className="text-sm text-leise">
        {eigene.suchindexModus}: {modus ?? ergebnis.modus} {ergebnis.grund}
      </p>

      {ergebnis.treffer.length === 0 ? (
        <p className="text-sm text-leise">{eigene.suchindexLeer}</p>
      ) : (
        <ol className="flex flex-col gap-2 text-sm">
          {ergebnis.treffer.map((treffer) => (
            <li key={treffer.kennung} className="flex flex-col">
              <span className="font-semibold">{treffer.titel}</span>

              <span className="text-leise">
                {treffer.ausschnitt || treffer.text}
              </span>

              {/* **Die Herkunft steht am Treffer und nicht in einer
                  Fußnote**: Quelle, Stand und Abrufzeit sind das, was ihn
                  belegbar macht. */}
              <span className="text-xs text-leise">
                {treffer.quelle} · {treffer.quellrevision} {treffer.quelldatum}{' '}
                · {treffer.abgerufen}
              </span>
            </li>
          ))}
        </ol>
      )}
    </>
  )
}
