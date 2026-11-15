// **Was eine Partie als Ganzes angeht** — Anhalten, Fortsetzen, Beenden,
// ihr Verlauf und ihre Freigabe.
//
// # Warum das nicht neben dem Tisch steht
//
// **Weil man es nicht in jedem Zug braucht.** Das Artboard *Tisch* hat in
// seiner Spalte den Zug, die gewählte Karte und ihre Handlungen und sonst
// nichts; dort unter den Handlungen schöben diese vier die Ereignisse aus dem
// Blick (UX-08,
// #1680).
// Unter dem Tisch sind sie einen Bildlauf entfernt — so weit wie die
// Freigabe des Protokolls.
//
// # Und warum das Beenden eine Rückfrage hat
//
// **Weil es endgültig ist.** Eine beendete Partie lässt sich nicht
// fortsetzen — „ein Abbruch, den man zurücknehmen kann, ist keiner". §6
// verlangt die Frage dort, wo geklickt wird.
import { useState } from 'react'

import { Entscheidung } from '../bausteine/Entscheidung'
import { Knopf } from '../bausteine/Knopf'
import { Partiefreigabe } from './Partiefreigabe'
import { Partieverlauf } from './Partieverlauf'
import { Protokollfreigabe } from '../bausteine/Protokollfreigabe'
import { partievorgang } from '../haken/partiewege'
import { useTexte } from '../sprache/kontext'

import type { Partie } from '../apipartie'
import type { Partielauf } from '../haken/usePartie'

export function Partiefuehrung({
  lauf,
  partie,
  meldeGastsitze,
}: {
  lauf: Partielauf
  partie: Partie
  meldeGastsitze?: (gastsitze: readonly string[] | undefined) => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const [fragt, setzeFragt] = useState(false)

  const laeuft = lauf.laeuft(partievorgang(partie.id))

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="leise"
          angehalten={laeuft || partie.stand !== 'laeuft'}
          aufKlick={() => {
            lauf.wechseln('pausieren')
          }}
        >
          {eigene.pausieren}
        </Knopf>

        <Knopf
          gewicht="leise"
          angehalten={laeuft || partie.stand !== 'pausiert'}
          aufKlick={() => {
            lauf.wechseln('fortsetzen')
          }}
        >
          {eigene.fortsetzen}
        </Knopf>

        <Knopf
          gewicht="leise"
          angehalten={laeuft || partie.stand === 'beendet'}
          aufKlick={() => {
            setzeFragt(true)
          }}
        >
          {eigene.beenden}
        </Knopf>
      </div>

      {fragt && (
        <Entscheidung
          frage={eigene.beendenFrage}
          bestaetigen={eigene.beendenJa}
          abbrechen={eigene.beendenNein}
          laeuft={laeuft}
          aufBestaetigen={() => {
            setzeFragt(false)
            lauf.wechseln('beenden')
          }}
          aufAbbrechen={() => {
            setzeFragt(false)
          }}
        />
      )}

      {/*
        **Der `key` löst Regel 1 und 2 aus §7 zugleich.** Ein Sitzwechsel
        wechselt das Bezugsobjekt: die Rückschau gehört einem Empfänger, und
        was für Anna geholt wurde, darf für Bert nicht stehen bleiben. Mit
        einem neuen `key` baut React das Bauteil neu auf, und damit ist auch
        jede noch laufende Antwort ins Leere gelaufen statt in den neuen
        Stand. **Wer ihn entfernt, nimmt beide Regeln weg.**
      */}
      <Partieverlauf
        key={`${partie.id}:${partie.fuer}`}
        kennung={partie.id}
        fuer={partie.fuer}
      />

      {/*
        **Derselbe `key` und derselbe Grund** wie eine Zeile darüber: die
        Freigabe gehört zu **dieser** Partie und diesem Wirt, und ein
        Sitzwechsel macht aus dem Wirt einen anderen.
      */}
      <Partiefreigabe
        key={`freigabe:${partie.id}:${partie.fuer}`}
        partie={partie}
        meldeGastsitze={meldeGastsitze}
      />

      {/* **Was von dieser Partie bleibt** (G5, #1462) — die Einwilligung
          gehört an den Tisch und nicht in eine Einstellungsseite: wer sie
          gibt, gibt sie über **diese** Partie und über seinen eigenen Sitz.
          Der `key` räumt sie beim Wechsel der Partie mit ab (§7, Regel 1). */}
      <Protokollfreigabe
        key={partie.id}
        kennung={partie.id}
        eigenerSitz={partie.fuer}
        beendet={partie.stand === 'beendet'}
      />
    </div>
  )
}
