// **Was eine Karte ist** — Werkstück M2 der Regelmaschine
// (#1604).
//
// **Eine eigene Datei, und der Schnitt folgt der Frage** (§2): `apipartie.ts`
// beantwortet, **was auf dem Tisch liegt** — hier steht, **was eine Karte
// ist**. Der Anlass war die Zeilengrenze: mit diesen beiden Typen ging die
// eine Datei auf 401 Zeilen.

// Karteneigenschaften sind Farbe, Typ und Werte einer Karte — CR 109.3.
//
// **Die Werte tragen ihre Marken schon** (CR 122.1a): eine 2/3 mit zwei
// +1/+1-Marken kommt hier als 4/5 an. Gerechnet wird in der Domäne, damit es
// **eine** Rechnung gibt statt einer je Fläche.
export interface Karteneigenschaften {
  // name ist der Name der Karte — CR 201.1.
  //
  // **Nicht `Spielobjekt.name`**: der trägt den Namen eines Spielsteins und
  // ist bei einer Karte leer.
  name: string

  // farben sind die Buchstaben der Manasymbole — `W`, `U`, `B`, `R`, `G`.
  //
  // **Eine leere Liste heißt farblos** (CR 105.2c).
  farben: string[]

  manawert: number

  supertypen: string[]
  typen: string[]
  untertypen: string[]

  staerke: Kartenwert
  widerstandskraft: Kartenwert
  loyalitaet: Kartenwert
  verteidigung: Kartenwert

  // unbekannteTypen sind die Wörter der Typzeile, die das Regelwerk nicht
  // kennt — ein **Befund** und kein Fehler (§8). Die Karte steht trotzdem da.
  unbekannteTypen: string[]
}

// Kartenwert ist eine Zahl auf einer Karte, die auch fehlen oder ein Stern
// sein darf — CR 208.2, CR 208.3.
export interface Kartenwert {
  // steht unterscheidet „keine" von „null": eine Nichtkreatur hat **keine**
  // Stärke.
  steht: boolean

  // stern heißt: eine Fähigkeit bestimmt die Zahl (CR 604.3).
  stern: boolean

  zahl: number
}
