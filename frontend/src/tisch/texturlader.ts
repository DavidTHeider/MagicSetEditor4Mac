// **Welches Bild ein Sprite trägt oder erwartet** — die Entscheidung, ob
// `pixinaht.ts` ein Bild holt.
//
// # Warum am Sprite und nicht an der Karte
//
// **Weil der Vermerk beschreiben muss, was er beschreibt.** Bis zum
// 2026-09-24 hing „schon geladen" an Kennung und Adresse und überlebte zwei
// Dinge: dass ein Sprite auf den Platzhalter zurückfiel, und dass für
// dieselbe Karte ein neues Sprite entstand. Fragte danach dieselbe Adresse
// wieder, lud niemand — und zwei Karten in Adas Hand blieben Platzhalter,
// bis die Leinwand neu entstand (WebKit, Set „Green").
//
// **Und warum er hier steht**: `pixinaht.ts` ruft PixiJS und wird nicht
// gemessen (`.covignore`); was dort entscheidet, gehört daneben.

// Texturlader merkt sich je Sprite die Adresse, die es trägt oder erwartet.
export interface Texturlader<S extends object> {
  // soll sagt, ob für dieses Sprite geladen werden muss — und merkt sich die
  // Adresse, wenn ja.
  soll: (sprite: S, adresse: string) => boolean

  // gilt sagt, ob eine angekommene Textur noch die ist, die das Sprite
  // erwartet. **Eine überholte Antwort gilt nicht** — sonst setzte sich ein
  // langsames altes Bild über ein schnelles neues.
  gilt: (sprite: S, adresse: string) => boolean

  // vergiss löst das Sprite von seinem Bild: es zeigt wieder den Platzhalter,
  // und die nächste Adresse wird geholt, auch wenn es dieselbe ist.
  vergiss: (sprite: S) => void
}

export function neuerTexturlader<S extends object>(): Texturlader<S> {
  // **Eine `WeakMap`**: ein zerstörtes Sprite nimmt seinen Vermerk mit, und
  // niemand muss aufräumen.
  const erwartet = new WeakMap<S, string>()

  return {
    soll(sprite, adresse) {
      if (erwartet.get(sprite) === adresse) return false

      erwartet.set(sprite, adresse)

      return true
    },

    gilt: (sprite, adresse) => erwartet.get(sprite) === adresse,

    vergiss(sprite) {
      erwartet.delete(sprite)
    },
  }
}
