// **Eine Zahl und ihr Wort** —
// #1218.
//
// # Der Anlass
//
// Die Warnung vor dem Löschen sagte „mit allen 1 Karten". Gefunden bei der
// Sichtprüfung im laufenden Fenster; ein Test konnte es nicht sehen, denn die
// Zeichenfolge stand ja da.
//
// **Es ist nicht die einzige Stelle** — die Setliste, „Zuletzt benutzt" und
// der Druckbogen setzen Zahl und Wort ebenso aneinander. Sie sind ein eigenes
// Werkstück (#1219);
// dieser Helfer steht hier, damit es dort keinen zweiten gibt.
//
// # Warum kein `Intl.PluralRules`
//
// Weil es für zwei Sprachen mit **derselben** Regel — eins gegen alles andere
// — eine Abhängigkeit einführte, deren Kategorien (`one`, `few`, `many`,
// `other`) beide Kataloge tragen müssten. Käme eine Sprache dazu, deren Regel
// anders ist — Polnisch, Russisch —, wäre das der Anlass, es zu tun. Bis
// dahin wäre es eine Allgemeinheit ohne Fall.

// Zaehlform sind die Formen eines Worts — zwei, oder drei, wo die Sprache
// sie braucht.
//
// **`wenige` ist die Form für 2 bis 4** (und 22 bis 24 …) im Russischen —
// LOC-01 (#1687).
// Hier stand „Käme eine Sprache dazu, deren Regel anders ist — Polnisch,
// Russisch —, wäre das der Anlass": sie ist gekommen. Die Regel kommt aus
// `Intl.PluralRules`, und nur ein Katalog, der die dritte Form führt, zählt
// danach; die übrigen bleiben bei eins gegen alles andere.
export interface Zaehlform {
  eine: string
  wenige?: string
  viele: string

  // regel ist die Sprache nach BCP 47, nach der die drei Formen zählen —
  // gesetzt von `ueberlagere` an jeder Form, deren Wörter eine eigene
  // Übersetzung setzt, und **nicht aufzählbar**, damit sie in keiner flachen
  // Form als Satz erscheint. Ohne sie zählt die Form russisch: der einzige
  // eingebaute Katalog mit dritter Form.
  //
  // **An der Form und nicht im Modul** — Befund N5 des Nachlaufs über Phase
  // 6.5: die Regel stand in einem Zustand, den jeder Sprachraum beim
  // Zeichnen setzte, und der zuletzt gezeichnete galt für alle.
  regel?: string
}

// regelwerk hält je Sprache ihre Regel — ein Zwischenspeicher, kein
// Zustand: dieselbe Sprache gibt immer dieselbe.
const regelwerk = new Map<string, Intl.PluralRules>()

// regelFuer gibt die Pluralkategorien einer Sprache. **Kennt `Intl` sie
// nicht, gilt die russische.**
function regelFuer(sprache: string): Intl.PluralRules {
  let regel = regelwerk.get(sprache)

  if (regel === undefined) {
    regel = new Intl.PluralRules(
      Intl.PluralRules.supportedLocalesOf([sprache]).length > 0
        ? sprache
        : 'ru',
    )
    regelwerk.set(sprache, regel)
  }

  return regel
}

// zaehlwort setzt eine Zahl vor ihr Wort — in der Form, die zu ihr passt.
//
// **Genau eins, nicht Betrag eins.** Eine negative Anzahl kommt in keiner
// Kartenzahl vor; stünde `Math.abs` hier, wäre das eine Regel für einen Fall,
// den es nicht gibt — und eine, die niemand mehr hinterfragt.
export function zaehlwort(anzahl: number, form: Zaehlform): string {
  return `${String(anzahl)} ${zaehlform(anzahl, form)}`
}

// zaehlform ist das Wort allein, in der Form, die zur Zahl passt — für eine
// Kennzahl, deren Zahl groß und deren Wort klein darunter steht (UXR-03,
// #1911).
// **Dieselbe Regel wie mit der Zahl davor**, denn `zaehlwort` ruft sie.
export function zaehlform(anzahl: number, form: Zaehlform): string {
  if (form.wenige !== undefined) {
    const kategorie = regelFuer(form.regel ?? 'ru').select(anzahl)

    return kategorie === 'one'
      ? form.eine
      : kategorie === 'few'
        ? form.wenige
        : form.viele
  }

  return anzahl === 1 ? form.eine : form.viele
}
