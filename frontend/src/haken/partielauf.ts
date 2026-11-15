// **Was `usePartie` dem Fenster gibt** — ausgelagert, damit der Haken unter
// der Zeilengrenze aus §2 bleibt (#2388); was er **tut**, steht dort.
import type {
  Objekthandlung,
  Partie,
  Partieentwurf,
  Spielereignis,
} from '../api'
import type { Standwechsel } from './partiewege'

export interface Partielauf {
  // anfangen baut eine Partie aus Sitzwünschen und zeigt die Sicht des Wirts.
  anfangen: (entwurf: Partieentwurf) => void

  // ansehen holt die Sicht **eines** Sitzes — auch für den Perspektivwechsel.
  ansehen: (kennung: string, fuer: string) => void

  // befehlen stellt einen Spielbefehl aus der Sicht, die gerade dasteht.
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void

  // wechseln hält an, nimmt wieder auf oder beendet.
  wechseln: (wohin: Standwechsel) => void

  raeumen: () => void

  laeuft: (was: string) => boolean

  partie: Partie | undefined

  // wechselt sagt, dass eine andere Sicht gewählt ist und noch nicht da.
  //
  // **Solange steht die alte nicht mehr da** — Befund 1 des Nachlaufs über
  // Phase 6.9: sie zeigte weiter die Hand des vorigen Sitzes und ließ an
  // ihr handeln. **Und nach einem Fehlschlag auch nicht**: die Hand von
  // vorhin kommt nicht zurück, wer weiter will, wählt die Sicht noch einmal.
  wechselt: boolean

  // erlaubt sind die Befehlsarten, die **dieser** Sitz gerade stellen darf.
  erlaubt: string[]

  // handlungen sagen je Objekt, was dieser Sitz daran tun kann, und welche
  // Handlung Klick und Doppelklick auslösen — aus dem Kern (P3, #2371;
  // bedient in P4, #2372).
  handlungen: Objekthandlung[]

  // wartend ist, wie viele Befehle noch nicht beantwortet sind — P5 (#2373):
  // ein Wunsch, der wartet, ist kein bestätigter Spielstand (0126), aber er
  // ist zu sehen.
  wartend: number

  // befunde sind Auffälligkeiten, die die Partie **nicht** unbenutzbar machen
  // (G8, #1465)
  // — ein Sitz etwa, dessen Profil es nicht mehr gibt.
  //
  // **Ein Befund ist kein Fehler**: `fehler` daneben hält an, dieser hier
  // nicht. Wer beides in ein Feld legte, machte aus einem Hinweis eine
  // Störung (§8).
  befunde: string[]

  // ereignisse ist, was der letzte Befehl bewirkt hat — aus der Sicht seines
  // Absenders und nicht aus der aller.
  ereignisse: Spielereignis[]

  // ereignisluecke: vorn fehlen Züge, die verborgen kamen (Z3, `nachholen.ts`).
  ereignisluecke: boolean

  fehler: string | undefined

  // fehlerVorgang sagt, **welche** Handlung die Meldung erzeugt hat.
  //
  // **Damit sie dort steht, wo sie entstand** — CLAUDE.md §6. Ohne ihn hing
  // sie am Seitenkopf, und der Knopf, der sie auslöste, stand in einem
  // 370 Pixel breiten Fenster fünf Bildschirmhöhen darunter (gemessen am
  // 2026-09-17 an der laufenden Anwendung).
  fehlerVorgang: string | undefined
}

// Regelstand ist, was der Kern zu einem Stand über das Erlaubte sagt —
// **in einem Zustand**, damit Erlaubnis und Auskunft nie aus zwei Antworten
// stammen.
export interface Regelstand {
  erlaubt: string[]
  handlungen: Objekthandlung[]
}

export const keineRegeln: Regelstand = { erlaubt: [], handlungen: [] }
