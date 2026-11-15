// **Die gewählte Karte** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Klein, bis jemand sie groß sehen will
//
// **Wie im Artboard *Tisch***: unter „Mit dieser Karte" steht die Karte
// klein, daneben ihr Name und ihr Zustand, darunter die Knöpfe. In
// Betrachtungsgröße schöbe sie die Handlungen unter die Falz (UX-08,
// #1680).
//
// **Groß gibt es auf Zuruf**, und erst dann wird die Betrachtungsgüte
// geholt — „Hohe Auflösung bei Inspektion/Zoom nachladen" (#1459). Ob groß
// oder klein, hält der Aufrufer: es gilt für die nächste Karte mit, und
// dieses Bauteil wird bei jedem Wechsel neu aufgebaut.
//
// # Warum ein eigenes Bauteil
//
// **Wegen des `key`.** Welche Seite betrachtet wird, gehört zu **diesem**
// Objekt; wer ein anderes wählt, fängt bei der Vorderseite an. Das ist §7,
// Regel 1 — und ihre sauberste Gestalt ist ein `key`, der das Bauteil beim
// Wechsel abräumt.
//
// **Der erste Anlauf setzte den Zustand in einem `useEffect` zurück.**
// `react-hooks/set-state-in-effect` hat es angestrichen, und zu Recht: ein
// Zustand, der erst **nach** dem Zeichnen zurückgesetzt wird, zeichnet einmal
// falsch — die zweite Seite der vorigen Karte, für einen Bildlauf lang.
//
// # Und warum die Auskunft von der Textur kommt
//
// **Weil sie eine über die Karte ist.** Wer ein Bild bekommt, darf die Karte
// sehen — und dann darf er wissen, dass sie sich umdrehen lässt. Die
// Projektion müsste dafür in die Karte sehen; sie kennt nur den
// eingefrorenen Abzug.
import { useState } from 'react'

import { Kartenbild } from '../bausteine/Kartenbild'
import { Knopf } from '../bausteine/Knopf'
import { bildwunsch } from './tischbild'
import { useTexte } from '../sprache/kontext'
import { kartenwort } from '../spiel/feldwerte'
import { zonenwort } from './partiewoerter'

import type { Spielobjekt } from '../apipartie'
import type { Texturlauf } from '../haken/useSpieltexturen'

export function Kartenschau({
  texturen,
  objekt,
  spiel,
  gross,
  umschalten,
}: {
  texturen: Texturlauf
  objekt: Spielobjekt
  spiel: string
  gross: boolean
  umschalten: () => void
}) {
  const eigene = useTexte().seiten.spielen

  const [seite, setzeSeite] = useState('vorne')

  const wunsch = bildwunsch(
    objekt,
    spiel,
    gross ? 'betrachten' : 'tisch',
    seite,
  )

  // **Der Zustand in Worten**, wie das Artboard ihn unter den Namen setzt —
  // wo sie liegt, und ob sie getappt ist.
  const zustand = [
    zonenwort(eigene, objekt.zone),
    ...(objekt.getappt ? [eigene.getappt] : []),
  ].join(' · ')

  // **Groß steht die Karte über ihrem Namen**: neben 224 Punkten Bild blieben
  // in der Spalte keine hundert für den Text.
  return (
    <div
      className={
        gross ? 'flex flex-col items-start gap-2' : 'flex items-start gap-3'
      }
    >
      {/* **Die Tischgüte steht daneben als Ausweich**: sie liegt schon da,
          also ist sofort etwas zu sehen, während die große entsteht. */}
      <Kartenbild texturen={texturen} wunsch={wunsch} ausweichguete="tisch" />

      <div className="flex min-w-0 flex-col items-start gap-1 text-sm">
        <span className="font-semibold">
          {objekt.geheim ? eigene.geheim : kartenwort(objekt)}
        </span>

        <span className="text-leise">{zustand}</span>

        <div className="flex flex-wrap gap-1">
          <Knopf gewicht="leise" gedrueckt={gross} aufKlick={umschalten}>
            {eigene.grossAnsehen}
          </Knopf>

          {/* **Die andere Seite gibt es nur, wo es eine gibt.** Gefragt wird
              die Seite, die gerade dasteht, und nicht immer die Vorderseite:
              `Kartenbild` meldet sie beim Wechsel ab, und dann wäre die
              Auskunft weg, kaum dass jemand umgedreht hat. Beide Seiten
              tragen dieselbe Antwort, denn sie ist eine über die **Karte**. */}
          {texturen.zweiseitig(wunsch) && (
            <Knopf
              gewicht="leise"
              aufKlick={() => {
                setzeSeite(seite === 'vorne' ? 'zweite' : 'vorne')
              }}
            >
              {seite === 'vorne' ? eigene.andereSeite : eigene.ersteSeite}
            </Knopf>
          )}
        </div>
      </div>
    </div>
  )
}
