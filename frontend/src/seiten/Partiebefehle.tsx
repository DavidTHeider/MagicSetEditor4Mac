// **Die Handlungen am Tisch** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum die Handlungen nicht an der Karte kleben
//
// **Weil eine davon keine Karte meint.** Durchsehen und Mischen gelten einer
// **Zone**; ein Menü an jedem Kartenknopf hätte sie entweder doppelt oder gar
// nicht. Der Schnitt folgt der Frage (§2), und er ist derselbe wie im Kern:
// `sandkasten.go` kennt Objektbefehle, `sandkastenzonen.go` Zonenbefehle.
//
// # Was hier nicht mehr steht
//
// **Anhalten, Beenden, der Verlauf und die Freigabe** — sie stehen seit dem
// 2026-09-24 unter dem Tisch, in `Partiefuehrung`. Was hier bleibt, braucht
// man im selben Zug: die gewählte Karte, was mit ihr geht, und was der letzte
// Befehl ergeben hat.
import { handlungZu } from '../spiel/kartengriff'
import { useState } from 'react'

import { Ereignisliste } from './Ereignisliste'
import { Handlungsleiste } from '../bausteine/Handlungsleiste'
import { Kartenschau } from './Kartenschau'
import { useTexte } from '../sprache/kontext'

import type { Partie, Spielobjekt } from '../apipartie'
import type { Texturlauf } from '../haken/useSpieltexturen'
import { partievorgang } from '../haken/partiewege'

import type { Partielauf } from '../haken/usePartie'
import { Abschnittstitel } from '../bausteine/Abschnittstitel'

// wechseltDieZone sagt, ob ein Befehl das Objekt woandershin legt.
//
// **Eine Liste und keine Regel**: welcher Befehl eine Zone wechselt, weiß
// der Kern; hier steht nur, wann die **Wahl** im Fenster aufgehoben wird
// (§7, Regel 1). Eine zu kurze Liste lässt einen Knopf stehen, der ins Leere
// zeigt — der Kern antwortet dann „liegt nicht dort", und das ist ein
// Befund und kein Schaden.
function wechseltDieZone(art: string): boolean {
  return [
    'objekt.schieben',
    'objekt.zerstoeren',
    'objekt.opfern',
    'objekt.exilieren',
    'objekt.neutralisieren',
    'objekt.wirken',
    'objekt.land',
    'zug.abwerfen',
    'spieler.legende',
  ].includes(art)
}

interface Befehlseigenschaften {
  lauf: Partielauf
  partie: Partie
  objekt: Spielobjekt | undefined
  entwaehle: () => void

  texturen: Texturlauf
}

export function Partiebefehle({
  lauf,
  partie,
  objekt,
  entwaehle,
  texturen,
}: Befehlseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const laeuft = lauf.laeuft(partievorgang(partie.id))

  const ruht = partie.stand !== 'laeuft'

  // **Groß oder klein gilt über die Wahl hinaus**: wer Karten liest, liest
  // mehr als eine. Hier und nicht in `Kartenschau`, denn die wird bei jedem
  // Wechsel neu aufgebaut.
  const [gross, setzeGross] = useState(false)

  return (
    <div className="flex flex-col gap-4">
      <Abschnittstitel className="text-base font-semibold">
        {eigene.befehle}
      </Abschnittstitel>

      {objekt !== undefined && (
        /* **Ein `key` und kein Effekt** (§7, Regel 1): die betrachtete
           Seite gehört zu **diesem** Objekt, und wer ein anderes wählt,
           fängt bei der Vorderseite an. Der erste Anlauf setzte sie in einem
           `useEffect` zurück — `react-hooks/set-state-in-effect` hat es
           angestrichen, und zu Recht: ein Zustand, der nach dem Zeichnen
           zurückgesetzt wird, zeichnet einmal falsch. */
        <Kartenschau
          key={objekt.kennung}
          texturen={texturen}
          objekt={objekt}
          spiel={partie.spiel}
          gross={gross}
          umschalten={() => {
            setzeGross((war) => !war)
          }}
        />
      )}

      {/* **Die Handlungsleiste**
          (#1623)
          — **und es gibt sie nur einmal**: handgeschriebene Bedienungen
          daneben wären für dasselbe zwei Wege, die auseinanderlaufen können
          (§2). */}
      <Handlungsleiste
        partie={partie}
        erlaubt={lauf.erlaubt}
        objekt={objekt}
        amObjekt={
          objekt === undefined
            ? undefined
            : (handlungZu(lauf.handlungen, objekt.kennung)?.handlungen ?? [])
        }
        angehalten={laeuft || ruht}
        befehlen={(art, nutzlast) => {
          lauf.befehlen(art, nutzlast)

          // **Die Wahl wird aufgehoben, wo das Objekt weggeht** (§7,
          // Regel 1): nach einem Schieben, Zerstören oder Exilieren liegt
          // es woanders, und ein Knopf, der noch darauf zeigte, wäre ein
          // offener Kontext von vorhin.
          if (wechseltDieZone(art)) {
            entwaehle()
          }
        }}
      />

      <Ereignisliste lauf={lauf} />
    </div>
  )
}
