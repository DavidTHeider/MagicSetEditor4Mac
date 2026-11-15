// **Das Bild einer Karte auf ihr Sprite** — herausgelöst aus `pixinaht.ts`,
// als dort der Nachzeichner dazukam und die Datei an der Zeilengrenze stand
// (§2).
//
// # Was hier entschieden wird und was nicht
//
// **Nichts.** Ob geladen werden muss und ob eine Antwort noch gilt, sagt der
// `Texturlader`; wann neu gezeichnet wird, der `Nachzeichner`. Beide werden
// gemessen. Hier steht der Ruf an PixiJS (`Assets.load`), und den gibt es
// unter jsdom nicht — dieselbe Gestalt wie `pixibeiwerk.ts`.
import { neuerNachzeichner } from './nachzeichner'
import { fuerPixi } from './pixiadresse'
import { neuesTexturbuch } from './texturbuch'
import { neuerTexturlader } from './texturlader'

import type * as Pixi from 'pixi.js'
import type { Kartenbild } from './weltbild'

type Sprite = InstanceType<typeof Pixi.Sprite>

// Texturholer setzt Bilder auf Sprites — oder den Platzhalter.
export interface Texturholer {
  // hole lädt das Bild einer Adresse und setzt es, sobald es da ist. **Der
  // Platzhalter bleibt so lange stehen**: ein Sprite ohne Textur wäre ein
  // Loch, eines mit der falschen wäre schlimmer.
  hole: (sprite: Sprite, adresse: string, bild: Kartenbild) => void

  // platzhalter zeigt eine getönte Fläche statt eines Bildes.
  platzhalter: (sprite: Sprite, farbe: number) => void

  // raeume hält das Nachzeichnen an — **vor** dem Abbau des Renderers.
  raeume: () => void

  // freigeben gibt die Texturen dieses Tisches an PixiJS zurück — **nach**
  // dem Abbau, wenn kein Sprite sie mehr zeigt (`texturbuch.ts`).
  freigeben: () => void
}

export function neuerTexturholer(
  pixi: typeof Pixi,
  anwendung: InstanceType<typeof Pixi.Application>,
  wartefarbe: () => number,
): Texturholer {
  const lader = neuerTexturlader<Sprite>()
  const buch = neuesTexturbuch()

  // **Ein angekommenes Bild wird gezeichnet** — der Tisch zeichnet auf
  // Zuruf, und die Antwort kommt nach dem Zeichnen, das sie angefordert hat.
  const nach = neuerNachzeichner(
    () => {
      anwendung.render()
    },
    (lauf) => {
      requestAnimationFrame(lauf)
    },
  )

  const platzhalter = (sprite: Sprite, farbe: number) => {
    lader.vergiss(sprite)
    sprite.texture = pixi.Texture.WHITE
    sprite.tint = farbe
  }

  return {
    hole(sprite, adresse, bild) {
      if (!lader.soll(sprite, adresse)) return

      sprite.texture = pixi.Texture.WHITE
      sprite.tint = wartefarbe()

      // **Absolut, sonst landet sie unter `wails://` am falschen Ort** (#2066).
      const absolut = fuerPixi(adresse, document.baseURI)
      void pixi.Assets.load(absolut)
        .then((textur: InstanceType<typeof Pixi.Texture>) => {
          if (buch.geladen(absolut) === 'freigeben') {
            void pixi.Assets.unload(absolut)

            return
          }
          if (sprite.destroyed || !lader.gilt(sprite, adresse)) return
          sprite.texture = textur
          sprite.tint = 0xffffff
          sprite.width = bild.breite
          sprite.height = bild.hoehe
          nach.bitte()
        })
        .catch(() => {
          // **Ein Bild, das nicht kommt, lässt den Platzhalter stehen.** Es
          // ist kein Grund, den Tisch anzuhalten (§16: ohne Netz arbeitet die
          // Anwendung weiter) — und der nächste Anlauf holt es erneut.
          if (lader.gilt(sprite, adresse)) lader.vergiss(sprite)
        })
    },

    platzhalter,

    raeume() {
      nach.halte()
    },

    freigeben() {
      const frei = buch.raeume()
      if (frei.length > 0) void pixi.Assets.unload(frei)
    },
  }
}
