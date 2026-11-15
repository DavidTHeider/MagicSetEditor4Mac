// Die gemeinsame Hülle jeder Antwort aus `internal/app` — und wie man sie
// öffnet.
//
// **Eine eigene Datei**, weil `api.ts` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht die Fassade, hier die Form, in der jede
// ihrer Antworten ankommt.
import { Bindungsfehler } from './bindungsfehler'

import type { FehlerDTO } from './formen'

// Umschlag ist die gemeinsame Hülle jeder Antwort.
//
// **Exportiert, seit die Typen des Paketkatalogs in einer eigenen Datei
// stehen** (§2, Zeilengrenze): sie erweitern ihn, und ein nicht exportierter
// Typ ließe sich nur hier erweitern — also wieder in dieser Datei.
export interface Umschlag {
  fehler?: FehlerDTO | null
}

// auspacken gibt den Inhalt eines Umschlags — oder lehnt ab.
//
// Der Inhalt kommt als Rückruf ohne Argument, damit der Aufrufer ihn aus dem
// Umschlag greift, den er ohnehin in der Hand hat. Ein Rückruf **mit** dem
// Umschlag zwänge TypeScript, seinen Typ aus dem engen `Umschlag` zu schließen,
// und jedes Feld daneben wäre ein Fehler.
export function auspacken<T>(umschlag: Umschlag, inhalt: () => T): T {
  if (umschlag.fehler) {
    throw new Bindungsfehler(umschlag.fehler, befundeAus(umschlag))
  }

  return inhalt()
}

// befundeAus holt die Begleitbefunde aus einem Umschlag — wenn er welche hat.
//
// # Warum das hier steht und nicht im Typ
//
// **Weil nicht jede Antwort welche trägt.** Ein `SchreibAntwort` hat keine,
// eine `Preisantwort` schon; `befunde` in den `Umschlag` zu heben zwänge jede
// Bindung, ein leeres Feld mitzuschicken — und die Go-Seite müsste es füllen,
// damit es nicht als `null` ankommt (§3).
//
// **Gelesen wird deshalb, was dasteht.** Ein Umschlag ohne `befunde` gibt eine
// leere Liste, und der Aufrufer prüft auf eines statt auf zwei
// (#1202).
function befundeAus(umschlag: Umschlag): readonly string[] {
  const roh: unknown = (umschlag as { befunde?: unknown }).befunde

  if (!Array.isArray(roh)) {
    return []
  }

  return roh.filter((eintrag): eintrag is string => typeof eintrag === 'string')
}
