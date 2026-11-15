// **Der eingereichte Deckabzug eines Teilnehmers** — Befund 2 des Nachlaufs
// über Phase 4.7.
//
// Er wurde beim Anmelden gebaut, im Bestand abgelegt und im DTO ausgeliefert
// — und das Fenster zeigte davon eine **nackte Zahl** neben dem Namen. Wozu
// es ihn gibt (Vermerk 0062, Entscheidung 3: Text und keine Verweise, damit
// er jedes Löschen überlebt), war nirgends zu sehen.
//
// **Aufklappbar und nicht immer offen.** Acht Teilnehmer mit je sechzig
// Zeilen wären eine Wand; wer eine Liste prüfen will, prüft eine.
import { useState } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import type { AbzugsplatzDTO } from '../apiturniere'

export function Turnierabzug({
  abzug,
  wer,
}: {
  abzug: AbzugsplatzDTO[]

  // wer benennt den Teilnehmer — die Beschriftung muss ihn nennen, sonst
  // heißen bei acht Teilnehmern acht Knöpfe gleich.
  wer: string
}) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  const [offen, setzeOffen] = useState(false)
  const listenkennung = `abzug-${wer}`

  if (abzug.length === 0) {
    return null
  }

  const karten = kartenzahl(abzug)

  return (
    <>
      <Knopf
        bezeichnung={`${eigene.deckabzug}: ${wer}`}
        ausgeklappt={offen}
        steuert={listenkennung}
        aufKlick={() => {
          setzeOffen(!offen)
        }}
      >
        {karten}
      </Knopf>

      {offen && (
        <ul
          id={listenkennung}
          aria-label={`${eigene.deckabzug}: ${wer}`}
          className="basis-full pl-6 text-sm text-leise"
        >
          {abzug.map((platz) => (
            <li key={platz.id}>
              {platz.menge}× {platz.bezeichnung || eigene.deckabzugUnbenannt}
              {platz.printing !== '' && ` (${platz.printing})`}
            </li>
          ))}
        </ul>
      )}
    </>
  )
}

// kartenzahl zählt die Karten eines Abzugs — **Karten, nicht Zeilen.** Vier
// mal dasselbe ist eine Zeile und vier Karten; die Zahl daneben meinte
// bisher die Zeilen, und niemand konnte es sehen, weil nichts danebenstand.
// Seit UXR-12 (#1920) steht sie auch als Text an einer ruhenden Zeile.
export function kartenzahl(abzug: AbzugsplatzDTO[]): number {
  return abzug.reduce((summe, platz) => summe + platz.menge, 0)
}
