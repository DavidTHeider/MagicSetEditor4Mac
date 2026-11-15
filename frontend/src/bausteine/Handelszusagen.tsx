// **Welche Besitzzeilen weggehen** —
// #1643.
//
// # Warum ein Mensch das benennt
//
// **Weil ein Angebot einen Druck nennt und keine Zeile des eigenen
// Bestands.** Wer drei `NM`-Exemplare desselben Drucks in drei Lagerorten
// hat, muss sagen, **welches** weggeht; das rät kein Weg, und einer, der es
// täte, legte die falsche Karte auf den Tisch.
//
// # Und warum die Fassung dabeisteht
//
// **Weil sie die Bedingung ist und nicht die Buchführung** (§7, Regel 2):
// hat jemand die Zeile geändert, während der Handel lief, rechnet die Zusage
// gegen einen Stand, den es nicht mehr gibt — und bekommt einen Konflikt
// statt eines stillen Erfolgs.
import { useId, useState } from 'react'

import { Aufklappkasten } from './Aufklappkasten'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Ganzzahlfeld } from './Ganzzahlfeld'
import { Zusagenwahl } from './Zusagenwahl'

import { Knopf } from './Knopf'
import { Satz } from './Satz'
import { api } from '../fassade'
import { Bindungsfehler } from '../bindungsfehler'
import { ganzzahlAus } from '../ganzzahl'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import type { Vorgang } from '../haken/useVorgang'

import type {
  Handelsvorgang as Vorgangsstand,
  Handelszusage,
} from '../apihandelsleitung'

interface Handelszusageneigenschaften {
  kennung: string

  // lauf ist der Riegel des Raums — **derselbe wie für Schritte, Vorschläge
  // und Aktionen** (§7, Regel 4; Befund N3 des Nachlaufs über Phase 6.8).
  // Mit einem eigenen konnten ein Schritt und dieser Kasten zugleich
  // schreiben.
  lauf: Vorgang

  // vorgang ist der gewählte Handel — leer heißt: noch keiner (UX-08.AK-04).
  vorgang: string

  // stand nennt die Positionen; die eigenen sind die, die nicht vom
  // Gegenüber kommen — ein Tausch hat zwei Seiten.
  stand: Vorgangsstand | undefined
}

export function Handelszusagen({
  kennung,
  vorgang,
  lauf,
  stand,
}: Handelszusageneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln

  const ueberschrift = useId()

  const [exemplar, setzeExemplar] = useState('')
  const [menge, setzeMenge] = useState('1')
  const [fassung, setzeFassung] = useState('1')

  // pruefen steht, sobald jemand festhalten wollte — **und erst dann** zeigt
  // ein Feld, dass sein Text keine ganze Zahl ist (F03, #2245). Bis dahin las
  // `parseInt` „12a" als 12 und hielt ein anderes Exemplar fest.
  const [pruefen, setzePruefen] = useState(false)
  const [fehler, setzeFehler] = useState<string>()
  const [zusagen, setzeZusagen] = useState<Handelszusage[]>()

  // listenfehler ist der Grund, aus dem die Liste nach einem **geglückten**
  // Festhalten fehlt (§7, Regel 3; #1970) — er steht an ihrer Stelle.
  const [listenfehler, setzeListenfehler] = useState<string>()

  const kennzeichen = `handelsvorgang/${kennung}`

  // fahre führt einen Ruf und räumt vorher, was er beantwortet.
  const fahre = (ruf: () => Promise<void>): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)
      setzeListenfehler(undefined)

      try {
        await ruf()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
        setzeZusagen(undefined)
      }
    })

  const festhalten = (
    exemplarID: number,
    stueck: number,
    gesehen: number,
  ): Promise<void> =>
    fahre(async () => {
      const antwort = await api().HandelFesthalten(
        vorgang,
        exemplarID,
        stueck,
        gesehen,
      )

      const gelesen = auspacken(antwort, () => antwort.zusagen)

      // **Festgehalten ist festgehalten** (§7, Regel 3; #1970): die leere
      // Liste daneben sagt nichts — sie behauptete „nichts festgehalten".
      if (antwort.nachladefehler) {
        setzeZusagen(undefined)
        setzeListenfehler(
          meldungAus(new Bindungsfehler(antwort.nachladefehler), texte),
        )

        return
      }

      setzeZusagen(gelesen)
    })

  // **Die eigenen Positionen** — was nicht vom Gegenüber kommt.
  const eigenePositionen = (stand?.positionen ?? []).filter(
    (position) => position.von !== kennung,
  )

  const nachsehen = (): Promise<void> =>
    fahre(async () => {
      const antwort = await api().Handelszusagen(vorgang)

      setzeZusagen(auspacken(antwort, () => antwort.zusagen))
    })

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-2">
      <h3 id={ueberschrift} className="text-sm font-medium">
        {eigene.zusagenTitel}
      </h3>

      <p className="max-w-prose text-sm text-leise">
        <Satz text={eigene.zusagenErklaerung} />
      </p>

      {vorgang === '' ? (
        <p className="text-sm text-leise">{eigene.zusagenOhneVorgang}</p>
      ) : (
        <>
          {eigenePositionen.map((position) => (
            <Zusagenwahl
              key={`${position.printing}/${String(position.menge)}`}
              position={position}
              angehalten={lauf.laeuft(kennzeichen)}
              festhalten={(exemplarID, stueck, gesehen) => {
                void festhalten(exemplarID, stueck, gesehen)
              }}
            />
          ))}

          <div className="flex flex-wrap gap-2">
            <Knopf
              gewicht="leise"
              angehalten={lauf.laeuft(kennzeichen)}
              aufKlick={() => {
                void nachsehen()
              }}
            >
              {eigene.zusagenNachsehen}
            </Knopf>
          </div>

          <Aufklappkasten titel={eigene.zusagenFachleute}>
            <div className="flex flex-col gap-2">
              <Erklaerungsabsatz
                text={eigene.zusagenFachleuteErklaerung}
                klasse="max-w-prose text-sm text-leise"
              />

              <div className="flex flex-wrap items-end gap-2">
                <Ganzzahlfeld
                  beschriftung={eigene.zusagenExemplar}
                  wert={exemplar}
                  setzeWert={setzeExemplar}
                  pruefen={pruefen}
                />

                <Ganzzahlfeld
                  beschriftung={eigene.zusagenMenge}
                  wert={menge}
                  setzeWert={setzeMenge}
                  pruefen={pruefen}
                />

                <Ganzzahlfeld
                  beschriftung={eigene.zusagenFassung}
                  wert={fassung}
                  setzeWert={setzeFassung}
                  pruefen={pruefen}
                />

                <Knopf
                  gewicht="leise"
                  angehalten={lauf.laeuft(kennzeichen)}
                  aufKlick={() => {
                    setzePruefen(true)

                    const exemplarID = ganzzahlAus(exemplar)
                    const stueck = ganzzahlAus(menge)
                    const gesehen = ganzzahlAus(fassung)

                    if (
                      exemplarID !== undefined &&
                      stueck !== undefined &&
                      gesehen !== undefined
                    ) {
                      void festhalten(exemplarID, stueck, gesehen)
                    }
                  }}
                >
                  {eigene.zusagenFesthalten}
                </Knopf>
              </div>
            </div>
          </Aufklappkasten>
        </>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {listenfehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {listenfehler}
        </p>
      )}

      {zusagen !== undefined && (
        <div role="status" className="flex flex-col gap-1 text-sm text-leise">
          {zusagen.length === 0 ? (
            <p>{eigene.zusagenLeer}</p>
          ) : (
            <ul className="flex flex-col gap-1">
              {zusagen.map((zusage) => (
                <li
                  key={`${String(zusage.exemplarId)}/${String(zusage.revision)}`}
                  className="font-mono text-xs"
                >
                  #{zusage.exemplarId} · {zusage.menge} ·{' '}
                  {zusage.uebergeben
                    ? eigene.zusagenUebergeben
                    : eigene.zusagenOffen}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </section>
  )
}
