// **Eine Paarung, an eine Partie gebunden** — Werkstück G7 von Phase 5.5
// (#1464).
//
// # Warum das neben dem Ergebnisfeld steht und es nicht ersetzt
//
// **Weil beide Wege bleiben.** Wer am Küchentisch ein Turnier führt, tippt ein
// Ergebnis — daran ändert dieses Werkstück nichts. Wer **hier** spielt, bindet
// die Begegnung an die Partie, und dann melden beide Seiten; eingetragen wird,
// wenn sie dasselbe sagen.
//
// # Und warum die Grenze dabeisteht
//
// **Weil eine Unterschrift nach mehr aussieht, als sie ist.** Das Issue grenzt
// es selbst ab: „signierte private Runde ist keine Behauptung offizieller
// Turnierzertifizierung oder vollständiger Anti-Cheat-Sicherheit." Der
// Nachweis sagt, dass beide denselben Verlauf gesehen haben — nicht, dass
// richtig gespielt wurde.
import { useId, useState } from 'react'

import { Beschriftetesfeld } from './Beschriftetesfeld'
import { Ganzzahlfeld, ganzzahlImFeld } from './Ganzzahlfeld'
import { Knopf } from './Knopf'
import { useAnschluss } from '../anschlussraum'
import { useTurniermatch } from '../haken/useTurniermatch'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { MatchAntwort, Matchmeldung } from '../apiturniere'

interface Matcheigenschaften {
  turnierID: string
  paarungID: string

  // teilnehmer sind die beiden Kennungen der Paarung.
  teilnehmer: [string, string]
}

export function Turniermatch({
  turnierID,
  paarungID,
  teilnehmer,
}: Matcheigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  const lauf = useTurniermatch(turnierID, paarungID)
  const ueberschrift = useId()

  const [partieID, setzePartieID] = useState('')
  const [sitzA, setzeSitzA] = useState('')
  const [sitzB, setzeSitzB] = useState('')

  return (
    <section className="flex flex-col gap-2" aria-labelledby={ueberschrift}>
      <div className="flex items-center gap-3">
        <h4 id={ueberschrift} className="text-sm font-semibold">
          {eigene.match}
        </h4>

        <Knopf gewicht="leise" angehalten={lauf.laeuft} aufKlick={lauf.ansehen}>
          {eigene.matchAnsehen}
        </Knopf>
      </div>

      <Erklaerungsabsatz
        text={eigene.matchErklaerung}
        klasse="max-w-prose text-xs text-leise"
      />

      {lauf.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}

      {lauf.standfehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.standfehler}
        </p>
      )}

      {lauf.stand?.gebunden === false && (
        <div className="flex flex-col gap-2">
          <p className="text-sm text-leise">{eigene.matchNichtGebunden}</p>

          <Beschriftetesfeld
            bezeichnung={eigene.matchPartie}
            wert={partieID}
            setzeWert={setzePartieID}
          />

          <Beschriftetesfeld
            bezeichnung={eigene.matchSitzA}
            wert={sitzA}
            setzeWert={setzeSitzA}
          />

          <Beschriftetesfeld
            bezeichnung={eigene.matchSitzB}
            wert={sitzB}
            setzeWert={setzeSitzB}
          />

          <div>
            <Knopf
              gewicht="leise"
              angehalten={
                lauf.laeuft ||
                partieID.trim() === '' ||
                sitzA.trim() === '' ||
                sitzB.trim() === ''
              }
              aufKlick={() => {
                lauf.binden(partieID.trim(), [
                  {
                    teilnehmerID: teilnehmer[0],
                    sitz: sitzA.trim(),
                    peer: '',
                    deckabdruck: '',
                  },
                  {
                    teilnehmerID: teilnehmer[1],
                    sitz: sitzB.trim(),
                    peer: '',
                    deckabdruck: '',
                  },
                ])
              }}
            >
              {eigene.matchBinden}
            </Knopf>
          </div>
        </div>
      )}

      {lauf.stand?.gebunden === true && (
        <Gebundenes stand={lauf.stand} lauf={lauf} />
      )}

      <p className="max-w-prose text-xs text-leise">{eigene.matchGrenze}</p>
    </section>
  )
}

// Gebundenes zeigt, was an dieser Paarung hängt — und was gemeldet ist.
function Gebundenes({
  stand,
  lauf,
}: {
  stand: MatchAntwort
  lauf: ReturnType<typeof useTurniermatch>
}) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere
  const anschluss = useAnschluss()
  const [erster] = stand.sitze

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-center gap-3">
        <p className="text-sm">
          {eigene.matchGebunden}{' '}
          <span className="font-mono text-xs">{stand.partieID}</span>
        </p>

        {/* **Die Partie geht am Tisch auf** — DOM-04.AK-02 (#1699), aus der
            Sicht des ersten Sitzes: das Artboard *Turniere* zeichnet
            *Partie öffnen* an der Paarung. */}
        {erster !== undefined && (
          <Knopf
            gewicht="leise"
            aufKlick={() => {
              anschluss.uebergib({
                art: 'partieansicht',
                name: eigene.match,
                partieID: stand.partieID,
                sitz: erster.sitz,
              })
            }}
          >
            {eigene.partieOeffnen}
          </Knopf>
        )}
      </div>

      <ul className="flex flex-col gap-1 text-sm">
        {stand.sitze.map((sitz) => (
          <li key={sitz.sitz} className="flex flex-wrap gap-3">
            <span>
              {eigene.matchSitz} {sitz.sitz}
            </span>

            <span className="font-mono text-xs text-leise">
              {eigene.matchDeckabdruck}: {sitz.deckabdruck}
            </span>
          </li>
        ))}
      </ul>

      <Meldungsliste meldungen={stand.meldungen} />

      <Meldefeld lauf={lauf} sitze={stand.sitze.map((sitz) => sitz.sitz)} />

      <div>
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft}
          aufKlick={lauf.uebernehmen}
        >
          {eigene.matchUebernehmen}
        </Knopf>
      </div>

      {stand.uebernommen !== null && (
        <p role="status" className="text-sm">
          {eigene.matchUebernommen}
        </p>
      )}
    </div>
  )
}

// Meldungsliste zeigt, was die Seiten gemeldet haben.
function Meldungsliste({ meldungen }: { meldungen: Matchmeldung[] }) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  if (meldungen.length === 0) {
    return <p className="text-sm text-leise">{eigene.matchKeineMeldung}</p>
  }

  return (
    <ul className="flex flex-col gap-1 text-sm">
      {meldungen.map((meldung) => (
        <li key={meldung.von}>
          {meldung.ergebnis.spieleA}:{meldung.ergebnis.spieleB}:
          {meldung.ergebnis.unentschieden} {eigene.matchMeldungVon}{' '}
          {meldung.von}
        </li>
      ))}
    </ul>
  )
}

// Meldefeld nimmt die Meldung **einer** Seite entgegen.
function Meldefeld({
  lauf,
  sitze,
}: {
  lauf: ReturnType<typeof useTurniermatch>
  sitze: string[]
}) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  const [von, setzeVon] = useState(sitze[0] ?? '')
  const [kette, setzeKette] = useState('')
  const [spieleA, setzeSpieleA] = useState('')
  const [spieleB, setzeSpieleB] = useState('')

  // pruefen steht, sobald jemand melden wollte (F03, #2245): bis dahin wurde
  // „viele" zu null und ging als Ergebnis hinaus, das niemand getippt hatte.
  const [pruefen, setzePruefen] = useState(false)

  return (
    <div className="flex flex-col gap-2">
      <Beschriftetesfeld
        bezeichnung={eigene.matchSitz}
        wert={von}
        setzeWert={setzeVon}
      />

      <Beschriftetesfeld
        bezeichnung={eigene.matchNachweis}
        wert={kette}
        setzeWert={setzeKette}
      />

      <div className="flex gap-2">
        <Ganzzahlfeld
          beschriftung={eigene.spieleA}
          wert={spieleA}
          setzeWert={setzeSpieleA}
          pruefen={pruefen}
          leerIstNull
        />

        <Ganzzahlfeld
          beschriftung={eigene.spieleB}
          wert={spieleB}
          setzeWert={setzeSpieleB}
          pruefen={pruefen}
          leerIstNull
        />
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft || von.trim() === '' || kette.trim() === ''}
          aufKlick={() => {
            setzePruefen(true)

            const a = ganzzahlImFeld(spieleA, true)
            const b = ganzzahlImFeld(spieleB, true)

            if (a !== undefined && b !== undefined) {
              lauf.melden(von.trim(), kette.trim(), {
                spieleA: a,
                spieleB: b,
                unentschieden: 0,
              })
            }
          }}
        >
          {eigene.matchMelden}
        </Knopf>

        {/* **Der Weg ohne Tippen** — M17 (#1618): Spielstand und Nachweis
            kommen aus der beendeten Partie, und der Sitz sagt nur, wer
            meldet. Er steht **neben** dem Feld und nicht an seiner Stelle:
            wer am Küchentisch führt, trägt weiter von Hand ein. */}
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft || von.trim() === ''}
          aufKlick={() => {
            lauf.ausDerPartie(von.trim())
          }}
        >
          {eigene.matchAusDerPartie}
        </Knopf>
      </div>

      <p className="text-xs text-leise">{eigene.matchAusDerPartieHilfe}</p>
    </div>
  )
}
