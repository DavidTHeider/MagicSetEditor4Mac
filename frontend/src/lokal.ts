// **Zahlen in der Sprache des Fensters — und zurück** — LOC-03.AK-01 und
// AK-03 (#1689):
// „lokalisierte Eingaben müssen verlustfrei in kanonische interne Werte
// übergehen", und „mehrdeutige Eingaben werden erklärt und nicht still
// falsch interpretiert."
//
// # Die Grenze in einem Satz
//
// **Was in dieser Sprache geschrieben ist, wird übersetzt; was zwei
// Lesarten hat, wird erklärt; alles andere geht unverändert weiter — und Go
// erklärt es.** Über die Naht geht nur die kanonische Form
// ([`docs/LOKALISIERUNG.md`](../../docs/LOKALISIERUNG.md)).
//
// # Warum eine Lesart nur gilt, wenn die Sprache sie selbst so schriebe
//
// **Weil eine Regel über Trennzeichen an der nächsten Sprache zerbricht.**
// Gemessen am 2026-09-22: Hindi gruppiert zu zweien (»12,34,567.89«),
// Schweizerdeutsch trennt mit einem Apostroph, Französisch mit einem
// schmalen geschützten Leerzeichen. Wer stattdessen fragt, ob `Intl` die
// Zahl **genau so** schriebe, hat alle diese Fälle auf einmal — und liest
// »19,99« im Englischen **nicht** als 1999: dort schriebe niemand so.
//
// **Was dabei nicht gelesen wird**, und das ist Absicht: Ziffern einer
// anderen Schrift (»١٬٢٣٤« im Arabischen). Sie gehen unverändert weiter,
// und Go nennt, was dasteht.

// Lesart ist, was aus einer Eingabe wird.
export type Lesart =
  // **Eine Deutung** — `kanonisch` geht so über die Naht.
  | { art: 'eindeutig'; kanonisch: string }

  // **Zwei Deutungen mit verschiedenem Wert** — »1.234« ist im Deutschen
  // tausendzweihundertvierunddreißig und in kanonischer Form eins Komma
  // zwei drei vier. Wer hier eine wählte, riete (AK-03).
  | { art: 'mehrdeutig'; lesarten: readonly [string, string] }

  // **Keine Deutung** — der Text geht unverändert weiter.
  | { art: 'fremd' }

// kanonisch ist die Form über die Naht: optionales Minus, Ziffern, optional
// ein Punkt mit Ziffern dahinter — dieselbe wie in `internal/domain/geld.go`.
const kanonisch = /^-?\d+(?:\.\d+)?$/u

// lesarten sagt, was eine Eingabe bedeutet.
//
// **`stellen` ist die Zahl der Nachkommastellen, die gelten dürfen** — bei
// einem Betrag die seiner Währung. Sie entscheidet, ob die kanonische Form
// überhaupt eine Lesart ist: »1.234« ist beim Euro keine, beim Dinar schon.
// Ohne Angabe gilt jede.
export function lesarten(
  text: string,
  sprache: string,
  stellen = Number.POSITIVE_INFINITY,
): Lesart {
  const roh = raumgleich(text.trim())

  // **Kein Text ist kein Betrag** — und keine Frage an irgendwen.
  if (roh === '') {
    return { art: 'eindeutig', kanonisch: '' }
  }

  const ort = ortslesart(roh, sprache)
  const kanon =
    kanonisch.test(roh) && nachkomma(roh) <= stellen ? roh : undefined

  if (ort === undefined) {
    return kanon === undefined
      ? { art: 'fremd' }
      : { art: 'eindeutig', kanonisch: kanon }
  }

  // **Zwei Lesarten mit demselben Text sind eine** — »19.99« im Englischen
  // ist in beiden Formen dasselbe.
  if (kanon === undefined || kanon === ort) {
    return { art: 'eindeutig', kanonisch: ort }
  }

  return { art: 'mehrdeutig', lesarten: [ort, kanon] }
}

// alsZahl schreibt eine kanonische Zahl in dieser Sprache — mit genau
// `stellen` Nachkommastellen.
//
// **Für die Erklärung einer Mehrdeutigkeit**: wer »1.234« getippt hat, soll
// beide Lesarten in **seiner** Schreibweise sehen — »1.234,000« und
// »1,234« —, nicht in der kanonischen, die ihm gerade fremd war.
export function alsZahl(
  kanonischerText: string,
  sprache: string,
  stellen: number,
): string {
  return new Intl.NumberFormat(sprache, {
    minimumFractionDigits: stellen,
    maximumFractionDigits: stellen,
  }).format(Number(kanonischerText))
}

// ortslesart liest den Text so, wie diese Sprache ihn schriebe — oder gar
// nicht.
function ortslesart(roh: string, sprache: string): string | undefined {
  const { gruppe, dezimal } = trennzeichen(sprache)

  // **Erst die Gruppen, dann das Komma** — die beiden Zeichen sind in jeder
  // gemessenen Sprache verschieden, und die Reihenfolge macht, dass ein
  // Punkt im Deutschen nie zum Dezimalzeichen wird.
  const kandidat = roh.replaceAll(gruppe, '').replace(dezimal, '.')

  // **Mehr als zwanzig Stellen hinter dem Komma trägt keine Zahl** — und
  // `Intl` wiese sie ab. Dann bleibt nur die kanonische Lesart.
  if (!kanonisch.test(kandidat) || nachkomma(kandidat) > 20) {
    return undefined
  }

  const stellen = nachkomma(kandidat)

  // **Mit und ohne Gruppen** — wer »1234,56« tippt, hat so richtig
  // geschrieben wie wer »1.234,56« tippt.
  for (const useGrouping of [true, false]) {
    const geschrieben = new Intl.NumberFormat(sprache, {
      useGrouping,
      minimumFractionDigits: stellen,
      maximumFractionDigits: stellen,
    }).format(Number(kandidat))

    if (raumgleich(geschrieben) === roh) {
      return kandidat
    }
  }

  return undefined
}

// trennzeichen gibt Gruppen- und Dezimalzeichen einer Sprache.
//
// **Gemessen und nicht aufgezählt**: `Intl` weiß es für jede Sprache, und
// eine eigene Liste daneben liefe bei der ersten neuen auseinander (§10).
export function trennzeichen(sprache: string): {
  gruppe: string
  dezimal: string
} {
  let gruppe = ''
  let dezimal = '.'

  for (const teil of new Intl.NumberFormat(sprache).formatToParts(1234567.89)) {
    if (teil.type === 'group') {
      gruppe = raumgleich(teil.value)
    }

    if (teil.type === 'decimal') {
      dezimal = teil.value
    }
  }

  return { gruppe, dezimal }
}

// alsProzent schreibt einen Anteil als ganze Prozent.
//
// **Das Zeichen gehört zur Zahl und nicht zum Satz.** Ein »{prozent} %« im
// Katalog wäre im Englischen falsch — gemessen am 2026-09-22: `en-US`
// schreibt »42%« ohne Abstand, `de-DE` »42 %« mit einem **geschützten**,
// damit das Zeichen nicht allein in die nächste Zeile rutscht.
//
// **Auf ganze Zahlen**: „37,4 %" behauptet eine Genauigkeit, die bei zwölf
// Karten niemand hat (`Balken.tsx`).
export function alsProzent(anteil: number, sprache: string): string {
  return new Intl.NumberFormat(sprache, {
    style: 'percent',
    maximumFractionDigits: 0,
  }).format(anteil)
}

// nachkomma zählt die Stellen hinter dem Punkt.
function nachkomma(kanonischerText: string): number {
  const punkt = kanonischerText.indexOf('.')

  return punkt < 0 ? 0 : kanonischerText.length - punkt - 1
}

// raumgleich macht jedes Leerzeichen zu einem gewöhnlichen.
//
// **Weil niemand ein schmales geschütztes tippt.** Französisch gruppiert
// mit U+202F, Österreichisch mit U+00A0 — gemessen am 2026-09-22 —, und
// wer »1 234,56« eingibt, meint dasselbe.
function raumgleich(text: string): string {
  return text.replace(/\s/gu, ' ')
}
