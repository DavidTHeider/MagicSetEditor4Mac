// **Die Antworten der Parameterhilfen** —
// #1342.
//
// **Eine eigene Datei**, wie `antwortenauszaehlung.ts` und aus demselben Grund
// (§2, Zeilengrenze).
import type { Umschlag } from './api'

// ParameterartDTO ist eine Parameterart, wie ein Menü sie anbietet.
export interface ParameterartDTO {
  name: string
  beschreibung: string

  // einfuegen ist der Text, der ins **Muster** kommt.
  einfuegen: string
}

// ParameterverweisDTO ist eine Form, in der ein Erinnerungstext auf einen
// Parameter zeigt.
export interface ParameterverweisDTO {
  // nummer ist die Stelle des Parameters im Muster, ab 1.
  nummer: number

  // art ist der Name der Parameterart.
  art: string

  // name ist der Name des Verweisskripts — leer bei der schlichten Form.
  name: string
  beschreibung: string

  // einfuegen ist der Text, der in den **Erinnerungstext** kommt.
  einfuegen: string
}

// Parameterhilfenantwort trägt beide Menüs.
export interface Parameterhilfenantwort extends Umschlag {
  arten: ParameterartDTO[]
  verweise: ParameterverweisDTO[]
  befunde: string[]
}
