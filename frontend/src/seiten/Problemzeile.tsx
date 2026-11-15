// **Die Zeile unter den Kacheln: wie alt die Prüfung ist, oder der Anstoß**
// — AP-8, Teil 2
// (#2365),
// Entwurf `Start09`: „geprüft vor 4 min · 3.788 von 3.788“.
//
// # Auf Anstoß, nie von selbst
//
// **Die Prüfung läuft nie je Aufruf der Startseite** (PERF02,
// `LEISTUNGSSCHNITTE.md` §11): bei 3.788 Karten dauert sie eine knappe
// Minute und nimmt die Hälfte der Kerne. Ist nicht jedes Set aktuell
// geprüft, steht hier, wie viele fehlen, und ein Knopf stößt sie an. Sie
// läuft als Hintergrundaufgabe — mit Fortschritt und Abbruch darunter.
//
// **Geschrieben ist geschrieben** (§7, Regel 3): nach der Prüfung werden die
// Zahlen nachgeladen; scheitert sie, steht die Meldung da, und es wird nicht
// nachgeladen.
import { useState } from 'react'

import { api } from '../api'
import { Knopf } from '../bausteine/Knopf'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import type { Problemlage } from './problemlage'
import { alterSatz } from './altersatz'

// zeile ist der Rahmen — **immer da und höher als ein Knopf** (mit 36 Punkten
// sprang die Seite am Fenster um zwei, als der Knopf kam), auch beim
// Laden und ohne Sets: sonst rückte, was darunter steht, beim Eintreffen der
// Zahlen nach unten (AP-4, Laden in Endhöhe).
const zeile =
  '-mt-3 mb-6 flex min-h-10 flex-wrap items-center gap-3 text-sm text-leise'

export function Problemzeile({
  lage,
  geladen,
  nachladen,
}: {
  // lage fehlt, solange die Zahlen ausstehen oder die Stände nicht zu lesen
  // waren.
  lage: Problemlage | undefined
  geladen: boolean
  nachladen: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.start.ueberblick
  const [laeuft, setzeLaeuft] = useState(false)
  const [fehler, setzeFehler] = useState<string>()

  if (lage === undefined || lage.gesamt === 0) {
    return (
      <div className={zeile} aria-hidden={geladen ? undefined : true}>
        {!geladen && <span className="invisible">{eigene.pruefen}</span>}
      </div>
    )
  }

  const pruefen = async () => {
    setzeLaeuft(true)
    setzeFehler(undefined)

    try {
      const antwort = await api().SetproblemePruefen()
      auspacken(antwort, () => undefined)
      nachladen()
    } catch (ursache: unknown) {
      setzeFehler(meldungAus(ursache, texte))
    } finally {
      setzeLaeuft(false)
    }
  }

  return (
    <div className={zeile}>
      <p>
        {lage.vollstaendig
          ? einsetzen(eigene.problemeGeprueft, {
              wann: alterSatz(lage.alter, eigene.alter),
              geprueft: String(lage.kartenGeprueft),
              gesamt: String(lage.kartenGesamt),
            })
          : einsetzen(eigene.problemeOffen, {
              offen: String(lage.gesamt - lage.geprueft),
              gesamt: String(lage.gesamt),
            })}
      </p>

      {!lage.vollstaendig && (
        <Knopf
          angehalten={laeuft}
          aufKlick={() => {
            void pruefen()
          }}
        >
          {eigene.pruefen}
        </Knopf>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}
