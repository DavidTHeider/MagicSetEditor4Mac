// **Wie viele Nachkommastellen eine Währung hat** — dieselbe Antwort wie
// `domain.Waehrungsstellen` in `internal/domain/geld.go`, LOC-03.AK-02
// (#1689).
//
// # Warum eine zweite Liste und nicht `Intl`
//
// **Weil `Intl` bei siebzehn Währungen eine andere Antwort gibt als Go.** Aus
// `Intl` gelesen wären es zwei Quellen für dieselbe Zahl: Go rechnet nach
// ISO 4217, `Intl` nach CLDR. Gemessen am 2026-09-22 über alle 162 Codes, die
// `Intl` kennt: **17 Abweichungen**, darunter HUF (Go 2, `Intl` 0) und IQD
// (Go 3, `Intl` 0). Ein Betrag von 1.999 Fillér stünde damit als »1.999 Ft«
// im Fenster — hundertfach.
//
// **Die Skala gehört dem, der speichert.** Go legt fest, was eine Zahl im
// Bestand bedeutet; das Fenster teilt durch dieselbe Stellenzahl und lässt
// `Intl` nur noch **zeigen** — beim Forint ohne Nachkommastellen, wie es
// dort üblich ist.
//
// # Was die beiden Listen zusammenhält
//
// `internal/werkzeuge/waehrungsstellen_test.go` fragt Go für jeden der
// 17.576 möglichen Codes und vergleicht mit dieser Datei. **Was er nicht
// hält**: ob ISO 4217 selbst sich geändert hat — das sieht er in keiner der
// beiden Listen.

// nullstellige sind die Währungen ohne Nachkommastellen — ISO 4217.
const nullstellige = new Set([
  'BIF',
  'CLP',
  'DJF',
  'GNF',
  'ISK',
  'JPY',
  'KMF',
  'KRW',
  'PYG',
  'RWF',
  'UGX',
  'UYI',
  'VND',
  'VUV',
  'XAF',
  'XOF',
  'XPF',
])

// dreistellige sind die Währungen mit drei Nachkommastellen — ISO 4217,
// **ohne CLF und UYW**: dort vier, hier wie in Go zwei (Vermerk 0167).
const dreistellige = new Set(['BHD', 'IQD', 'JOD', 'KWD', 'LYD', 'OMR', 'TND'])

// waehrungsstellen gibt die Nachkommastellen einer Währung.
//
// **Zwei für alles andere** — auch für einen leeren oder unbekannten Code.
// Dieselbe Vorgabe wie in Go: dort sagt ein zweiter Wert, dass geraten
// wurde, und wer hier einen unbekannten Code hat, bekommt von Go ohnehin
// eine Meldung dazu.
export function waehrungsstellen(waehrung: string): number {
  const code = waehrung.trim().toUpperCase()

  if (nullstellige.has(code)) {
    return 0
  }

  return dreistellige.has(code) ? 3 : 2
}
