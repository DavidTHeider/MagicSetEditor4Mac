// Bemerken, dass ein **anderer** Prozess geschrieben hat (#156).
//
// # Warum das Fenster fragt
//
// Der MCP-Server ist ein eigener Prozess auf derselben Datei
// ([0010](../../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)); was
// ein Modell dort ändert, sah das Fenster bisher nicht. `PRAGMA data_version`
// steigt, sobald eine andere Verbindung bestätigt hat — und seit #1485 hat
// jeder Prozess genau eine, sodass „andere Verbindung" wirklich „anderer
// Prozess" heißt. Bis dahin zählte das Fenster sein eigenes Schreiben mit und
// lud nach jedem Feld die Seite neu.
//
// **Gefragt wird von hier und nicht von einem Faden in Go**: der führe auch
// dann, wenn niemand hinsieht. Der Grund steht ausführlich in
// [0017](../../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md).
import { useEffect, useRef, useState } from 'react'

import { api } from '../api'

// abstand ist der Abstand zweier Abfragen.
//
// **Zwei Sekunden**: kurz genug, dass niemand es merkt, lang genug, dass es
// nichts kostet. Die Zahl steht hier und nicht in Go, weil sie zur Anzeige
// gehört und nicht zur Datenbank.
export const abstand = 2000

// useBestandsstand gibt einen Zähler, der nach jeder fremden Änderung steigt.
//
// **Als Zahl und nicht als Rückruf**, aus demselben Grund wie bei `useVerlauf`:
// die Seiten legen ihn in ihre Ladeabhängigkeiten. Was sich geändert hat, steht
// im Bestand, und gelesen wird es von dort.
export function useBestandsstand(): number {
  const [stand, setzeStand] = useState(0)
  const gesehen = useRef<number | undefined>(undefined)

  useEffect(() => {
    let laeuft = true

    // **Eine Abfrage zur Zeit.** Dauert eine länger als der Abstand, stieße das
    // Intervall eine zweite an — und eine spät eintreffende Antwort setzte
    // `gesehen` auf einen älteren Stand zurück und löste eine Aktualisierung
    // aus, die es nicht gab. Gefunden vom Zweitblick durch Codex über Phase 6.
    let unterwegs = false

    const fragen = async () => {
      // **Nicht fragen, wenn niemand hinsieht.** Ein Fenster im Hintergrund
      // lädt nichts neu, und die Abfrage wäre Arbeit ohne Betrachter.
      if (unterwegs || document.visibilityState === 'hidden') {
        return
      }

      unterwegs = true

      try {
        const antwort = await api().Datenstand()

        // **Ein Fehler bleibt still.** Er beträfe eine Abfrage, die niemand
        // angefordert hat; sie zwei Sekunden später zu wiederholen ist die
        // bessere Antwort als eine Meldung, die im Weg steht.
        if (!laeuft || antwort.fehler) {
          return
        }

        if (gesehen.current === undefined) {
          gesehen.current = antwort.stand

          return
        }

        if (gesehen.current !== antwort.stand) {
          gesehen.current = antwort.stand
          setzeStand((bisher) => bisher + 1)
        }
      } catch {
        // Dito: eine Fassade ohne diese Methode ist ein Test, kein Betrieb.
      } finally {
        unterwegs = false
      }
    }

    void fragen()

    const uhr = setInterval(() => {
      void fragen()
    }, abstand)

    return () => {
      laeuft = false
      clearInterval(uhr)
    }
  }, [])

  return stand
}
