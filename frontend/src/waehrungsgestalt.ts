// **Die Gestalt eines Währungscodes** — UX-06.AK-02
// (#1678).
//
// # Dieselbe Regel wie in Go
//
// `istWaehrungscode` in `internal/domain/geld.go`: drei lateinische
// Buchstaben, **gezählt als Zeichen und nicht als Bytes** — `€` hat in UTF-8
// drei davon und ist keine Währung. Das Fenster prüft vor, damit der Satz am
// Feld steht, solange der Rohtext noch dasteht (AK-03); **entschieden wird in
// Go**, an jeder Fläche.
//
// # Plausibel heißt nicht bekannt
//
// **Drei Buchstaben sind die Gestalt eines Codes, nicht seine Existenz**
// (AK-04). `XAU` hat die Gestalt und ist eine; `ABC` hat sie und ist keine —
// und keine Stelle dieser Anwendung behauptet, das zu wissen. Die Vorschläge
// am Feld nennen, was das System kennt; sie schließen nichts aus.

// Waehrungsbefund ist, was an einem Code nicht stimmt.
export type Waehrungsbefund =
  { art: 'gestalt' } | { art: 'gross'; gross: string }

// waehrungsbefund urteilt über einen getippten Code.
//
// **Leer ist hier kein Befund**: ob eine Währung fehlen darf, weiß der
// Aufrufer — beim Exemplar nur ohne Preis, bei einer Grenze nie.
//
// **`kleinErlaubt`** folgt der Stelle, an die der Wert geht: das Exemplar hebt
// einen klein geschriebenen Code an (Vermerk 0057), die Grenze eines
// Kostenkontos weist ihn ab und sagt, wie er richtig heißt.
export function waehrungsbefund(
  text: string,
  kleinErlaubt: boolean,
): Waehrungsbefund | undefined {
  const getrimmt = text.trim()
  if (getrimmt === '') {
    return undefined
  }

  const gross = getrimmt.toUpperCase()
  if (!/^[A-Z]{3}$/.test(gross)) {
    return { art: 'gestalt' }
  }

  if (!kleinErlaubt && getrimmt !== gross) {
    return { art: 'gross', gross }
  }

  return undefined
}

// Waehrungsvorschlag ist ein Code mit seinem Namen in der Sprache des Fensters.
export interface Waehrungsvorschlag {
  wert: string
  bezeichnung: string
}

// waehrungsvorschlaege nennt die Währungen, die das System kennt.
//
// **Aus `Intl` und nicht aus einer eigenen Liste**: welche Währungen es gibt,
// ändert sich, und eine Liste hier wäre an dem Tag falsch, an dem ISO 4217
// eine neue aufnimmt. `Intl` kennt, was die Plattform kennt — und wo es
// fehlt, bleibt das Feld ein Freitext mit Prüfung, nur ohne Vorschläge.
export function waehrungsvorschlaege(sprache: string): Waehrungsvorschlag[] {
  let codes: string[]
  let namen: Intl.DisplayNames

  try {
    codes = Intl.supportedValuesOf('currency')
    // **`fallback: 'none'`**: ein Code, den die Namensliste nicht kennt,
    // bekommt keinen erfundenen Namen, sondern steht für sich.
    namen = new Intl.DisplayNames([sprache], {
      type: 'currency',
      fallback: 'none',
    })
  } catch {
    return []
  }

  return codes.map((wert) => ({ wert, bezeichnung: namen.of(wert) ?? wert }))
}
