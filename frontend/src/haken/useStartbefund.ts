// **Was beim Start geschah** —
// #962.
//
// # Warum das Fenster danach fragt
//
// Bis zum 2026-08-25 beendete sich die Anwendung, wenn der Bestand nicht
// aufging: kein Fenster, keine Bindung, kein Knopf. Der Rückweg über eine
// Sicherung war damit **im gescheiterten Fall** nicht erreichbar — und das ist
// genau der Fall, für den es ihn gibt.
//
// Jetzt öffnet das Fenster trotzdem und fragt hier nach, was es zeigen soll.
//
// # Warum nicht `useLaden`
//
// Der bräuchte einen `Sprachraum` — er übersetzt Fehlermeldungen. Diese Frage
// steht **vor** allem anderen und hat keine; ihr einziger Fehlerfall ist eine
// kaputte Naht, und dann gibt es nichts zu übersetzen. Regel 2 aus §7 hält
// [useLauf], wie überall sonst.
import { useEffect, useState } from 'react'

import { api } from '../fassade'
import type { Rueckweg } from '../antwortenbestand'
import { useLauf } from './useLauf'

export interface Startstand {
  // geladen sagt, ob **je** eine Antwort da war.
  //
  // **Vorher wird nichts gezeichnet.** Sonst stünde einen Wimpernschlag die
  // gewöhnliche Oberfläche da, und dann spränge sie auf die Wiederherstellung.
  //
  // **Und danach nie wieder nichts.** Der erste Anlauf setzte es beim
  // Nachfragen zurück — nach einem geglückten Zurückspielen wird neu gefragt,
  // und die ganze Anwendung verschwand für einen Augenblick samt der Meldung,
  // wohin der bisherige Bestand kam. Gefunden hat es der Lauf über die
  // Bestandspflege, nicht dieser hier (#962).
  geladen: boolean

  offen: boolean
  grund: string

  // schluessel ordnet den Grund einer Klasse zu — leer, wenn keine passt.
  //
  // **Er trägt den Satz, den jemand liest** (Werkstück 3 von Phase 4.13,
  // #1352);
  // `grund` bleibt die Einzelheit darunter.
  schluessel: string

  // fassung ist die laufende Fassung dieser Anwendung (#1168) — leer, solange
  // keine Antwort da war.
  fassung: string

  // rueckspielrest sagt, dass ein abgebrochenes Zurückspielen beim Start nicht
  // abgeschlossen werden konnte — sonst leer.
  //
  // **Im Klartext aus Go**, wie `grund` und aus demselben Grund: der Satz nennt
  // einen Pfad und einen Systemfehler, und beides steht in keinem Katalog (#1170).
  rueckspielrest: string

  // rueckweg nennt die Sicherung, die diese Fassung öffnen kann — sonst
  // `null` (Werkstück 5 von Phase 4.13, #1352).
  rueckweg: Rueckweg | null
}

export function useStartbefund(stand: number): Startstand {
  const lauf = useLauf()

  // **Der Stand steht im Zustand**, statt ihn beim Wechsel zurückzusetzen.
  //
  // Ein `setzeBefund` zu Beginn des Effekts wäre ein Zustandswechsel im Effekt
  // selbst — er löste eine zweite Zeichnung aus, bevor irgendetwas geschehen
  // ist. Wer den Stand mitführt, beantwortet „ist die Antwort von heute?"
  // ohne ihn.
  const [befund, setzeBefund] = useState<
    Omit<Startstand, 'geladen'> & { fuer: number }
  >({
    fuer: -1,
    offen: false,
    grund: '',
    schluessel: '',
    fassung: '',
    rueckspielrest: '',
    rueckweg: null,
  })

  useEffect(() => {
    const meiner = lauf.naechster()

    void (async () => {
      try {
        const antwort = await api().Startbefund()

        if (lauf.istAktuell(meiner)) {
          setzeBefund({ fuer: stand, ...antwort })
        }
      } catch {
        // **Scheitert die Frage selbst, gilt der Bestand als offen.**
        //
        // Sie geht über dieselbe Naht wie alles andere; kommt sie nicht durch,
        // ist das Fenster ohnehin nicht benutzbar, und eine
        // Wiederherstellungsseite, die auch nicht arbeiten kann, verdeckte nur
        // die eigentliche Meldung.
        if (lauf.istAktuell(meiner)) {
          setzeBefund({
            fuer: stand,
            offen: true,
            grund: '',
            schluessel: '',
            fassung: '',
            rueckspielrest: '',
            rueckweg: null,
          })
        }
      }
    })()
  }, [lauf, stand])

  // **Der alte Befund bleibt stehen, während der neue lädt.** Er ist einen
  // Augenblick alt und im Zweifel richtig; ein leeres Fenster ist es nie.
  return {
    geladen: befund.fuer >= 0,
    offen: befund.offen,
    grund: befund.grund,
    schluessel: befund.schluessel,
    fassung: befund.fassung,
    rueckspielrest: befund.rueckspielrest,
    // **`?? null` und nicht bloß durchgereicht**: eine Antwort ohne das Feld
    // — eine ältere Naht, eine belegte Fassade — gäbe sonst `undefined`, und
    // die Seite unterschiede „liegt keine da" nicht von „nicht gefragt".
    rueckweg: befund.rueckweg ?? null,
  }
}
