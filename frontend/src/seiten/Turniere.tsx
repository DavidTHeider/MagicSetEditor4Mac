// **Die Turniere** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// # Warum eine eigene Seite und keine Lasche der Decks
//
// Weil ein Turnier etwas anderes ist als ein Deck: ein **Verlauf über die
// Zeit** statt eines Datenmodells mit einer Prüfung. Ein Deck steht still,
// ein Turnier läuft — und beide auf eine Seite zu legen hieße, zwei
// Fehlerklassen zu mischen.
//
// # Was hier steht und was nicht
//
// *Wie es aussieht* steht hier; *was die Seite tut*, in `useTurniere` (§2).
import { useCallback, useRef, useState } from 'react'

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Ladefehler } from '../bausteine/Ladehinweise'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { Turnierassistent } from '../bausteine/Turnierassistent'
import { Einlesebericht, Turniereinlesen } from '../bausteine/Turnierdatei'
import { Turnierliste } from '../bausteine/Turnierliste'
import { Turnierwerkbank } from '../bausteine/Turnierwerkbank'
import { useTurniere } from '../haken/useTurniere'
import { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { Verdeckung, useVerdeckt } from '../verdeckung'

// bestandsstand zählt fremde Schreibvorgänge — die Liste lädt dann nach
// (Befund 3 des Gesamtreviews über 0.6.0).
export function Turniere({ bestandsstand = 0 }: { bestandsstand?: number }) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  const stand = useTurniere(bestandsstand)

  // **Welches Turnier der Assistent gerade führt** — DOM-04.AK-01 (#1699).
  // Für dieses zeigt die Seite keine Werkbank, bis er fertig ist.
  const [imAssistenten, setzeImAssistenten] = useState<string | null>(null)

  // **Der Assistent steht in einer Schublade** — UXR-12 (#1920). Der Fokus
  // geht vorher auf den Auslöser, damit er nach dem Schließen dorthin
  // zurückkehrt — in WebKit gibt ein Klick keinen Fokus (wie bei den
  // Paketen).
  const [anlage, setzeAnlage] = useState(false)
  const ausloeser = useRef<HTMLButtonElement>(null)

  // **Was ein Einlesen fand**, unter dem Kopf — der Knopf steht darin.
  const [befunde, setzeBefunde] = useState<string[] | null>(null)

  // **Welches Turnier eben aus der Liste geöffnet wurde** — dorthin geht der
  // Fokus, sobald es dasteht (UXR-12, #1920).
  const [ansprung, setzeAnsprung] = useState<string | null>(null)
  const angesprungen = useCallback(() => {
    setzeAnsprung(null)
  }, [])

  // **Eine Seite unter einer Schublade ist verdeckt** (`verdeckung.ts`).
  const verdeckt = useVerdeckt() || anlage

  const offen =
    stand.offen !== null && stand.offen.id !== imAssistenten
      ? stand.offen
      : null

  const liste = (
    <Turnierliste
      stand={stand}
      oeffnen={(kennung) => {
        setzeAnsprung(kennung)
        void stand.waehle(kennung)
      }}
    />
  )

  const meldung = stand.fehler !== undefined && (
    <p role="alert" className="text-fehler">
      {stand.fehler}
    </p>
  )

  return (
    <>
      {/* **Gesperrt, solange die Schublade steht**: `inert` nimmt Fokus,
          Klick und Vorleseprogramm, `Verdeckung` die Horcher am Fenster. */}
      <Verdeckung value={verdeckt}>
        <div className="flex flex-col gap-6" inert={anlage}>
          {/* **Eine führende Handlung** (R1): *Turnier anlegen …*, solange
              kein Turnier offen ist. Steht eines da, führt dessen Runde, und
              das Anlegen tritt zurück. **Einlesen daneben** — DOM-04.AK-03
              (#1699): beides bringt ein neues Turnier, und auch in einer
              leeren Liste. */}
          <Seitenkopf titel={eigene.titel} erklaerung={eigene.erklaerung}>
            <Turniereinlesen stand={stand} eingelesen={setzeBefunde} />

            <Aktionsknopf
              aktion="anlegen"
              gewicht={offen === null ? 'haupt' : 'neben'}
              schreibt
              griff={ausloeser}
              aufKlick={() => {
                ausloeser.current?.focus()
                setzeAnlage(true)
              }}
            >
              {eigene.neuOeffnen}
            </Aktionsknopf>
          </Seitenkopf>

          {stand.ladefehler !== undefined && (
            <Ladefehler
              fehler={stand.ladefehler}
              wiederholen={stand.neuLaden}
            />
          )}

          {!anlage && meldung}

          <Einlesebericht befunde={befunde} />

          {/* **Das offene Turnier vor der Liste** — UXR-12 (#1920). Die
              Liste wächst mit jedem Turnierabend um eine Zeile; unter ihr
              begann die Werkbank bei acht Turnieren erst unterhalb der Falz.
              **Der `key` trägt die Fassung mit** — dieselbe Lehre wie bei der
              Formatzeile (Befund 3 des Nachlaufs über Phase 4.6): hinge er
              allein an der Kennung, bliebe jedes halb ausgefüllte
              Ergebnisfeld über ein Nachladen hinweg stehen. */}
          {offen !== null && (
            <Turnierwerkbank
              key={`${offen.id}:${String(offen.runden.length)}`}
              stand={stand}
              turnier={offen}
              ansprung={ansprung === offen.id}
              angesprungen={angesprungen}
            />
          )}

          {/* **Ist ein Turnier offen, faltet die Liste** — gemessen am
              Fenster (`uxr12`): mit acht Turnieren und einem Abend zu
              sechzehn rollte die Seite 2,7 bis 2,8 Höhen tief, über R4.
              Wer an einem Abend arbeitet, braucht die anderen nicht. */}
          {offen === null ? (
            liste
          ) : (
            <Aufklappkasten
              titel={einsetzen(eigene.alleTurniere, {
                anzahl: String(stand.turniere.length),
              })}
            >
              {liste}
            </Aufklappkasten>
          )}
        </div>
      </Verdeckung>

      <Turnierassistent
        stand={stand}
        aufLauf={setzeImAssistenten}
        offen={anlage}
        schliessen={() => {
          setzeAnlage(false)
        }}
        meldung={anlage && meldung}
      />
    </>
  )
}
