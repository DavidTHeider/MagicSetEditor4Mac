// **Das eigene Tischbild auf der Leinwand** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Eine eigene Datei
//
// **Der Schnitt folgt der Frage** (§2): `pixinaht.ts` beantwortet, *was* auf
// die Leinwand kommt — hier steht, *was darunter liegt*. Der Anlass war die
// Zeilengrenze; dass beides zusammenfiel, ist der übliche Fall.
//
// **Und sie steht in `.covignore`**, wie die beiden Nähte daneben: hier wird
// nichts entschieden, hier wird ein Sprite gesetzt.

// **`import type` und kein Import.** Er wird beim Übersetzen entfernt und
// zieht zur Laufzeit nichts herein — die Regel „nur `pixinaht.ts` kennt
// PixiJS" gilt für das **Laden**, und hier werden nur Namen genannt.
import { fuerPixi } from './pixiadresse'
import type * as Pixi from 'pixi.js'
import type { Tischplan } from './tischplan'

// Hintergrund hält das eigene Tischbild.
export interface Hintergrund {
  // setze legt das Bild unter den Tisch — oder nimmt es weg.
  setze: (adresse: string, plan: Tischplan) => void

  // liegt sagt, ob eines da ist. **Nur dann** darf die Grundfläche nicht
  // gefüllt werden; sonst läge der Farbsatz über dem Bild, und der Import
  // wäre wirkungslos.
  liegt: () => boolean

  raeume: () => void
}

// Pixiteile ist, was diese Datei von PixiJS braucht — und nicht mehr.
type Pixiteile = Pick<typeof Pixi, 'Sprite' | 'Texture' | 'Assets'>

export function neuerHintergrund(
  pixi: Pixiteile,
  grund: Pixi.Container,
): Hintergrund {
  let bild: Pixi.Sprite | undefined
  let adresse = ''

  return {
    liegt: () => bild !== undefined,

    raeume() {
      bild = undefined
      adresse = ''
    },

    setze(neueAdresse, plan) {
      if (neueAdresse === '') {
        bild?.destroy()
        bild = undefined
        adresse = ''

        return
      }

      if (neueAdresse !== adresse) {
        adresse = neueAdresse
        bild?.destroy()

        // **Erst durchsichtig, dann sichtbar.** Ein weißes Rechteck, das
        // aufblitzt und dann zum Bild wird, sieht aus wie ein Fehler.
        const neues = new pixi.Sprite(pixi.Texture.WHITE)
        neues.alpha = 0
        bild = neues
        grund.addChildAt(neues, 0)

        // **Absolut, sonst landet sie unter `wails://` am falschen Ort** (#2066).
        void pixi.Assets.load(fuerPixi(neueAdresse, document.baseURI))
          .then((textur: Pixi.Texture) => {
            if (neues.destroyed) return

            neues.texture = textur
            neues.alpha = 1
          })
          .catch(() => {
            // **Ein Hintergrund, der nicht kommt, lässt den Farbsatz stehen**
            // — er ist Gestaltung und kein Grund, den Tisch anzuhalten (§16).
            adresse = ''
          })
      }

      if (bild && !bild.destroyed) {
        bild.position.set(0, 0)
        bild.width = plan.breite
        bild.height = plan.hoehe
      }
    },
  }
}
