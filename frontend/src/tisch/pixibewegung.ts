// **Was fliegt, und wie lange der Ticker dafür läuft** — Werkstück G3 von
// Phase 5.5
// (#1460).
//
// # Eine eigene Datei
//
// **Der Schnitt folgt der Frage** (§2): `pixinaht.ts` beantwortet, *was* auf
// die Leinwand kommt — hier steht, *wie lange es dorthin unterwegs ist*. Der
// Anlass war die Zeilengrenze; dass beides zusammenfiel, ist der übliche Fall.
//
// # Und sie steht in `.covignore`, wie die Naht daneben
//
// **Weil hier nichts entschieden wird.** Die Beschleunigung, der kürzere
// Drehweg, die Frage, ob sich das Bewegen überhaupt lohnt — all das rechnet
// `bewegung.ts` und wird gemessen. Was übrig bleibt, ist ein Ticker anwerfen
// und Positionen setzen.
import { Bewegungsdauer, zwischen } from './bewegung'

import type { Beweglich } from './bewegung'

// Beweger ist ein Sprite, wie diese Datei es braucht — und nicht mehr.
export interface Beweger {
  position: { set: (x: number, y: number) => void }
  rotation: number
}

// Takter ist der Ticker von PixiJS, auf das Nötige verengt.
export interface Takter {
  started: boolean
  add: (schritt: () => void) => unknown
  start: () => void
  stop: () => void
}

// Fluege hält, was gerade unterwegs ist.
export interface Fluege {
  starte: (kennung: string, von: Beweglich, nach: Beweglich) => void
  halte: (kennung: string) => void
  raeume: () => void

  // lande setzt jede fliegende Karte an ihr Ziel und beendet ihren Flug —
  // **vor einem Abzug**, damit das Bild den bestätigten Stand zeigt und
  // keinen Zwischenstand einer Bewegung (Befund 3 des Zweitblicks über
  // Phase 6.9, PLAY-05.AK-04).
  lande: (sprites: Map<string, Beweger>) => void

  // treibe hängt den Schritt ein und wirft den Ticker an, sobald etwas
  // fliegt — **einmal eingehängt und danach nur noch gestartet**.
  treibe: (ticker: Takter, sprites: Map<string, Beweger>) => void
}

export function neueFluege(): Fluege {
  const laufend = new Map<
    string,
    { von: Beweglich; nach: Beweglich; seit: number }
  >()
  let eingehaengt = false

  const setze = (sprite: Beweger, ort: Beweglich) => {
    sprite.position.set(ort.x, ort.y)
    sprite.rotation = ort.drehung
  }

  return {
    starte(kennung, von, nach) {
      laufend.set(kennung, { von, nach, seit: performance.now() })
    },

    halte(kennung) {
      laufend.delete(kennung)
    },

    raeume() {
      laufend.clear()
    },

    lande(sprites) {
      for (const [kennung, flug] of laufend) {
        const sprite = sprites.get(kennung)
        if (sprite) setze(sprite, flug.nach)
      }

      laufend.clear()
    },

    treibe(ticker, sprites) {
      if (!eingehaengt) {
        eingehaengt = true
        ticker.add(() => {
          const jetzt = performance.now()

          for (const [kennung, flug] of laufend) {
            const sprite = sprites.get(kennung)
            const anteil = (jetzt - flug.seit) / Bewegungsdauer

            if (!sprite || anteil >= 1) {
              if (sprite) setze(sprite, flug.nach)
              laufend.delete(kennung)
              continue
            }

            setze(sprite, zwischen(flug.von, flug.nach, anteil))
          }

          if (laufend.size === 0) ticker.stop()
        })
      }

      if (laufend.size > 0 && !ticker.started) ticker.start()
    },
  }
}
