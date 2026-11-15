// **Die `auto replace:`-Regeln des Spiels, an der Schreibmarke** —
// #897.
//
// # Warum ein Haken und kein Aufruf in der Seite
//
// Die Seite kennt das Set, das Feld kennt die Schreibmarke — und dazwischen
// liegt ein Aufruf, den zwei Bauteile brauchen (die Karten- und die
// Setangabenseite). Was ein Bauteil **tut**, gehört in einen Haken; wie es
// aussieht, in die Komponente (§2).
//
// # Was er nicht tut
//
// **Er hält keinen Zustand.** Eine Ersetzung ist eine Frage und eine Antwort;
// ein Zustand dazwischen wäre einer, der nach einem Setwechsel noch dastünde
// (§7 Regel 1).
//
// **Und er zeigt einen Fehler nicht an — mit Absicht, und die Begründung hat
// zwei Anläufe gebraucht.**
//
// Bis zum 2026-08-23 behauptete dieser Kopf, er verschlucke keinen Fehler; er
// liest `antwort.fehler` nicht einmal
// (#918).
// Der zweite Anlauf begründete das Schweigen damit, der Grund erreiche die
// Nutzenden über die **Vorschau** — und der Nachlauf des Zweitblicks hat es
// widerlegt: `Setangaben` benutzt denselben Haken und hat keine Vorschau
// (#932).
//
// # Warum das Schweigen trotzdem bleibt
//
// **Die Ersetzung läuft bei jedem Tastendruck.** Ein Meldeweg an dieser Stelle
// stapelte dieselbe Meldung je Zeichen — und verdeckte damit gerade die, um
// derentwillen es ihn gäbe.
//
// **Der Fehler hat trotzdem ein Ziel.** `app.Uebersetze` schreibt jeden
// **unbekannten** Fehler ins Protokoll (`internal/app/fehler.go`); erwartete
// Ausgänge — Validierung, fehlender Datensatz, Konflikt, Abbruch — schreibt es
// bewusst nicht, weil sie den einen Eintrag verdeckten, um dessentwillen es da
// ist.
//
// # Was das kostet, und diesmal ohne Ausflucht
//
// **Ein vorübergehender Bestandsfehler während des Tippens ist auf dem
// Bildschirm unsichtbar** — er steht im Protokoll und sonst nirgends. Das gilt
// für beide Seiten, die diesen Haken benutzen. Wer es ändern will, braucht
// einen Meldeweg, der **entprellt** — und der ist hier nicht gebaut.
//
// **Text und Marke bleiben in jedem Fall stehen.** Die Go-Seite gibt beides
// auch im Fehlerfall unverändert zurück; wer hier den Text verlöre, machte aus
// einer fehlenden Vorlage einen Datenverlust (§7 Regel 3).
import { useCallback } from 'react'

import { api } from '../api'

// **Das Set ist Pflicht, nicht wahlfrei.** Beide Seiten, die Textfelder
// zeigen, haben eines; ein Zweig für „kein Set" wäre einer, den kein Aufruf
// betreten kann (§3).
export function useAutoersetzung(
  setID: string,
): (text: string, stelle: number) => Promise<{ text: string; stelle: number }> {
  return useCallback(
    async (text: string, stelle: number) => {
      const antwort = await api().AutoErsetzen(setID, text, stelle)

      return { text: antwort.text, stelle: antwort.stelle }
    },
    [setID],
  )
}
