// **Wer das Protokoll später liest** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum die Grenzen dabeistehen und nicht bloß gelten
//
// **Weil eine Freigabe nach mehr aussieht, als sie ist.** Drei Sätze stehen
// deshalb neben den Knöpfen und nicht in einer Hilfeseite:
//
//   - Sie wirkt **erst nach** der Partie. Am Tisch sieht weiter jeder nur,
//     was ihm gehört.
//   - Eine Rücknahme **holt nichts zurück**. Was jemand gelesen hat, hat er
//     gelesen.
//   - Das Löschen betrifft **diesen Rechner**. Über Kopien anderswo sagt es
//     nichts.
//
// Alle drei sind Zusagen, die dieses Programm hält — und alle drei wären
// missverständlich, wenn man sie verschwiege (§6, §10).
import { useId, useState } from 'react'

import { usePartieprotokoll } from '../haken/usePartieprotokoll'
import { useTexte } from '../sprache/kontext'
import { Beschriftetesfeld } from './Beschriftetesfeld'
import { Aktionsknopf } from './Aktionsknopf'
import { Entscheidung } from './Entscheidung'
import { Knopf } from './Knopf'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { Offenlegung } from '../apinebenbei'
import { Abschnittstitel } from './Abschnittstitel'

interface Protokolleigenschaften {
  kennung: string

  // eigenerSitz füllt das Feld vor — freigeben kann nur, wem etwas gehört.
  eigenerSitz: string

  // beendet sagt, ob das Protokoll gelöscht werden darf.
  beendet: boolean
}

export function Protokollfreigabe({
  kennung,
  eigenerSitz,
  beendet,
}: Protokolleigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const lauf = usePartieprotokoll(kennung)
  const ueberschrift = useId()

  const [an, setzeAn] = useState('')

  // **Gelöscht wird erst nach einer Frage** — UX-10.AK-04 (#1682): Zustand,
  // Verlauf und Freigaben sind danach auf diesem Rechner weg.
  const [fragt, setzeFragt] = useState(false)

  return (
    <section className="flex flex-col gap-2" aria-labelledby={ueberschrift}>
      <Abschnittstitel id={ueberschrift} className="text-base font-semibold">
        {eigene.protokoll}
      </Abschnittstitel>

      <Erklaerungsabsatz
        text={eigene.protokollErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <Erklaerungsabsatz
        text={eigene.protokollNurSpaeter}
        klasse="max-w-prose text-xs text-leise"
      />

      {lauf.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}

      <Beschriftetesfeld
        bezeichnung={eigene.protokollEmpfaenger}
        wert={an}
        setzeWert={setzeAn}
      />

      <div className="flex flex-wrap items-center gap-3">
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft || an.trim() === '' || eigenerSitz === ''}
          aufKlick={() => {
            lauf.freigeben(eigenerSitz, empfaengerAus(an))
          }}
        >
          {eigene.protokollFreigeben}
        </Knopf>

        <Aktionsknopf
          aktion="zuruecknehmen"
          gewicht="leise"
          angehalten={lauf.laeuft || eigenerSitz === ''}
          aufKlick={() => {
            lauf.zuruecknehmen(eigenerSitz)
          }}
        >
          {eigene.protokollZuruecknehmen}
        </Aktionsknopf>

        <Knopf gewicht="leise" angehalten={lauf.laeuft} aufKlick={lauf.ansehen}>
          {eigene.protokollAnsehen}
        </Knopf>
      </div>

      <Erklaerungsabsatz
        text={eigene.protokollKeineRueckholung}
        klasse="max-w-prose text-xs text-leise"
      />

      {/* **Der Grund, aus dem die Liste nach einem geglückten Schritt fehlt**
          (#1970) — an ihrer Stelle und nicht oben am Schritt. */}
      {lauf.listenfehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.listenfehler}
        </p>
      )}

      <Freigabenliste offenlegungen={lauf.offenlegungen} />

      <Erklaerungsabsatz
        text={eigene.protokollAufbewahrung}
        klasse="max-w-prose text-xs text-leise"
      />

      {beendet && (
        <div className="flex flex-col gap-2">
          <Aktionsknopf
            aktion="loeschen"
            gewicht="leise"
            angehalten={lauf.laeuft || fragt}
            aufKlick={() => {
              setzeFragt(true)
            }}
          >
            {eigene.protokollLoeschen}
          </Aktionsknopf>

          {/* **Die Erklärung steht in der Frage, sobald sie gestellt ist.** */}
          {fragt ? (
            <Entscheidung
              frage={eigene.protokollLoeschenFrage}
              erklaerung={eigene.protokollLoeschenErklaerung}
              bestaetigen={eigene.protokollLoeschenJa}
              aktion="loeschen"
              abbrechen={texte.bausteine.aktionen.abbrechen}
              laeuft={lauf.laeuft}
              aufBestaetigen={() => {
                setzeFragt(false)
                lauf.vergessen()
              }}
              aufAbbrechen={() => {
                setzeFragt(false)
              }}
            />
          ) : (
            <Erklaerungsabsatz
              text={eigene.protokollLoeschenErklaerung}
              klasse="max-w-prose text-xs text-leise"
            />
          )}
        </div>
      )}

      {lauf.istVergessen && (
        <p role="status" className="text-sm">
          {eigene.protokollGeloescht}
        </p>
      )}
    </section>
  )
}

// Freigabenliste zeigt, wer wem etwas freigegeben hat.
//
// **Drei Zustände und nicht zwei**: „noch nicht geholt" und „geholt und leer"
// sehen gleich aus und sind es nicht (§6) — beim einen muss jemand klicken,
// beim anderen gibt es nichts zu sehen.
function Freigabenliste({
  offenlegungen,
}: {
  offenlegungen: Offenlegung[] | undefined
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  if (offenlegungen === undefined) {
    return null
  }

  if (offenlegungen.length === 0) {
    return <p className="text-sm text-leise">{eigene.protokollNiemand}</p>
  }

  return (
    <ul className="flex flex-col gap-1 text-sm">
      {offenlegungen.map((gabe) => (
        <li key={gabe.von}>
          {eigene.protokollSitz} {gabe.von} — {eigene.protokollFuer}{' '}
          {gabe.an.map(lesbar(eigene.protokollAlle)).join(', ')}
        </li>
      ))}
    </ul>
  )
}

// lesbar ersetzt den Stern durch ein Wort.
//
// **Ein Wort und kein Zeichen**: `*` ist die Form auf der Leitung; wer sie im
// Fenster stehen ließe, verlangte von einem Menschen, sie zu kennen.
function lesbar(alle: string): (empfaenger: string) => string {
  return (empfaenger) => (empfaenger === '*' ? alle : empfaenger)
}

// empfaengerAus zerlegt das Feld.
//
// **Leere Stücke fallen weg**: „ben, " ist ein Empfänger und nicht zwei, und
// die Go-Seite wiese einen leeren ab — mit einer Meldung über etwas, das
// niemand eingetippt hat.
function empfaengerAus(eingabe: string): string[] {
  return eingabe
    .split(',')
    .map((stueck) => stueck.trim())
    .filter((stueck) => stueck !== '')
}
