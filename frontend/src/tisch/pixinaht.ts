// **Die Grafiknaht** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Die einzige Datei, die PixiJS kennt
//
// **Und sie holt es erst, wenn sie es braucht.** Ein `import` am Kopf zöge
// PixiJS in jeden Lauf, der irgendetwas aus `tisch/` anfasst — anderthalb
// Megabyte, die unter jsdom nichts zu tun haben und beim Laden nach WebGL
// greifen. `await import(…)` hält es draußen und teilt zugleich das Bündel:
// wer nie spielt, lädt nie einen Renderer.
//
// # Was hier nicht steht
//
// **Keine Entscheidung.** Wo etwas liegt, rechnet `tischplan`; wie groß es
// erscheint, `weltlage`; was auf einer Karte zu sehen ist, `kartenbilder`;
// was sich geändert hat, `abgeglichen`. Hier wird gesetzt, was jene sagen —
// und deshalb steht diese Datei in `.covignore`, während alles davor gemessen
// wird. Dieselbe Gestalt wie `internal/app/fensternaht.go`.
import { abgeglichen } from './szenenabgleich'
import { lohntBewegung } from './bewegung'
import { neueFluege } from './pixibewegung'
import { neuerHintergrund } from './pixihintergrund'
import { neuerTexturholer } from './pixitextur'
import { neuesBeiwerk } from './pixibeiwerk'
import { tischoptionen } from './pixioptionen'
import {
  ausDemDokument,
  dunklesBild,
  mitTischbild,
  tischfarben,
} from './tischfarben'
import { tischbild } from './tischbilder'
import { kartenbilder, weltlage } from './weltbild'
import { tischmalerei } from './maler'
import { zeichenflaeche } from './zeichenflaeche'
import { zonenrahmen } from './zonenfeld'

import type { Tischmaler, Tischmalerei, Zeichenwunsch } from './maler'
import type { Kartenbild } from './weltbild'
import type { Tischfarben } from './tischfarben'
import type { Tischplan } from './tischplan'

// tischmaler ist der Weg, den das Fenster nimmt.
//
// **Das Tischbild geht mit**, und die Farben werden bei jedem Anlauf frisch
// gelesen — damit ein Wechsel zwischen hell und dunkel ankommt.
//
// **Und er steht hier und nicht in der Seite.** Der Ruf an `pixinaht` wird
// unter jsdom nie erreicht — `grafikprobe` sagt dort nein, und die Naht bleibt
// ungerufen. In einer Seitendatei wäre das eine Zeile, die keine Abdeckung je
// betritt; hier ist sie Teil derselben Ausnahme, die es für die Naht ohnehin
// gibt. Was **entscheidet**, liegt weiterhin daneben und wird gemessen:
// `tischmalerei` in `maler.ts` fragt über ein Doppel.
//
// **Die Farben werden bei jedem Anlauf frisch gelesen**, damit ein Wechsel
// zwischen hell und dunkel ankommt.
export function tischmaler(
  ort: HTMLElement,
  bild = 'holz',
): Promise<Tischmalerei> {
  return tischmalerei(ort, (eine) => {
    const lesen = ausDemDokument(eine.ownerDocument)

    return pixinaht(
      eine,
      mitTischbild(tischfarben(lesen), tischbild(bild), dunklesBild(lesen)),
    )
  })
}

// pixinaht legt einen Renderer auf der Leinwand an.
//
// **Sie wirft, wenn es nicht geht** — `tischmalerei` fängt das und macht
// daraus die Auskunft `kein-renderer`.
export async function pixinaht(
  leinwand: HTMLCanvasElement,
  anfangsfarben: Tischfarben,
): Promise<Tischmaler> {
  // **Die Farben wandern, der Renderer nicht.** Ein Tischbild zu wechseln
  // hieße sonst, einen WebGL-Kontext wegzuwerfen und einen neuen aufzubauen —
  // gemessen kostet das erste Bild 17,70 ms, das vierundvierzigfache eines
  // gewöhnlichen (`docs/TISCHLEISTUNG.md`).
  let farben = anfangsfarben
  const pixi = await import('pixi.js')
  // **Ohne `eval`**: Pixis Polyfills statt `new Function` (SEC-17, Vermerk 0179).
  await import('pixi.js/unsafe-eval')

  // **Pixis Systemtakt läuft nicht von selbst an** (PLAY-06.AK-04, #1714).
  // Jeder Renderer hängt seinen Scheduler an `Ticker.system`
  // (`SchedulerSystem.mjs`, 8.20.1) — gemessen am 2026-09-25: sechzig Bilder
  // je Sekunde, am ruhenden Tisch wie nach einem verlorenen Kontext. Er räumt
  // alle dreißig Sekunden auf; unten treibt ihn ein Takt je Sekunde.
  pixi.Ticker.system.autoStart = false

  const anwendung = new pixi.Application()
  await anwendung.init(tischoptionen(leinwand, farben.grund))

  // **Kein Ereignissystem von PixiJS.** Es hängt sich ebenfalls an
  // `Ticker.system`, um Hover-Zustände nachzustellen; dieser Tisch benutzt
  // keine Pixi-Ereignisse — gezogen und geklickt wird über React und die
  // Knöpfe über der Leinwand. `destroy` nimmt Horcher und Takt ab und ist
  // beim Abräumen des Renderers wiederholbar (`EventSystem.mjs`, 8.20.1).
  anwendung.renderer.events.destroy()

  pixi.Ticker.system.stop()
  const aufraeumtakt = setInterval(() => {
    pixi.Ticker.system.update()
  }, 1000)

  const welt = new pixi.Container()
  // **Die Ebenen in ihrer Reihenfolge von unten nach oben.**
  //
  // `beziehungen` liegt **über** den Karten und **unter** den Marken: eine
  // Linie, die unter einer Karte durchliefe, wäre an der wichtigsten Stelle
  // unsichtbar — dort, wo sie ankommt. Und eine Marke, die eine Linie
  // verdeckt, ist das kleinere Übel als umgekehrt.
  const ebenen = {
    grund: new pixi.Container(),
    karten: new pixi.Container(),
    beziehungen: new pixi.Container(),
    marken: new pixi.Container(),
    auswahl: new pixi.Container(),
  }
  welt.addChild(
    ebenen.grund,
    ebenen.karten,
    ebenen.beziehungen,
    ebenen.marken,
    ebenen.auswahl,
  )
  ebenen.karten.sortableChildren = true
  anwendung.stage.addChild(welt)

  const sprites = new Map<string, InstanceType<typeof pixi.Sprite>>()

  // **Das Bild einer Karte auf ihr Sprite** — und ein Zeichnen, wenn es
  // ankommt (`pixitextur.ts`).
  const texturen = neuerTexturholer(pixi, anwendung, () => farben.wartet)

  // grund zeichnet die Sitzflächen und Zonenrahmen einmal je Plan neu.
  //
  // **Sie sind wenige und ändern sich selten** — ein Abgleich lohnte hier
  // nicht, und ein `Graphics` neu zu füllen ist billiger als Buch zu führen.
  const grund = new pixi.Graphics()
  ebenen.grund.addChild(grund)

  const hintergrundbild = neuerHintergrund(pixi, ebenen.grund)

  const zeichneGrund = (plan: Tischplan) => {
    grund.clear()

    // **Nur ohne eigenes Bild wird die Fläche gefüllt** — sonst läge der
    // Farbsatz über dem Hintergrund, und der Import wäre wirkungslos.
    if (!hintergrundbild.liegt())
      grund.rect(0, 0, plan.breite, plan.hoehe).fill({ color: farben.grund })

    // **Jede Fläche bekommt ihren Umriss.** Zwei Holztöne nebeneinander
    // tragen keine 3:1; die Linie tut es, und WCAG 1.4.11 lässt sie gelten.
    for (const flaeche of plan.sitze)
      grund
        .rect(flaeche.x, flaeche.y, flaeche.breite, flaeche.hoehe)
        .fill({ color: flaeche.eigen ? farben.eigenerSitz : farben.sitz })
        .stroke({ color: farben.kante, width: 1.5 })

    // **Ein Zonenfeld ist eine Vertiefung und kein Strich** — Befund (b) des
    // Entwurfs, PLAY-03.AK-06
    // (#1711).
    //
    // Ein Strich allein gäbe dünne Linien auf einer Sitzfläche, und niemand
    // sähe, welche die Bibliothek ist (Auditaufnahme
    // `048-nativer-spieltisch.png`).
    //
    // **Gefüllt wird mit dem Tischgrund und nicht mit einer neuen Farbe.**
    // Eine Zone ist eine Stelle, an der die Sitzfläche die Tischplatte
    // durchscheinen lässt — das braucht keinen siebten Ton, und jeder neue
    // wäre eine Messung mehr. Dass Grund und Sitz gegeneinander keine 3:1
    // tragen, ist bekannt und in `KONTRASTE.md` beantwortet: **die Linie
    // darüber trägt die Trennung**, und sie wird hier nach der Füllung
    // gezogen.
    // **Samt Namensleiste** (`zonenrahmen`): Name und Anzahl stehen in der
    // Zone, über ihren Karten, und nicht auf dem Grund daneben.
    for (const feld of plan.felder) {
      const rahmen = zonenrahmen(feld)

      grund
        .rect(rahmen.x, rahmen.y, rahmen.breite, rahmen.hoehe)
        .fill({ color: farben.grund })
        .stroke({ color: farben.kante, width: 0.8 })
    }
  }

  const fluege = neueFluege()

  const setze = (
    sprite: InstanceType<typeof pixi.Sprite>,
    bild: Kartenbild,
    bewegt: boolean,
  ) => {
    const jetzt = { x: bild.x, y: bild.y, drehung: bild.drehung }
    const vorher = {
      x: sprite.position.x,
      y: sprite.position.y,
      drehung: sprite.rotation,
    }

    if (bewegt && lohntBewegung(vorher, jetzt)) {
      fluege.starte(bild.kennung, vorher, jetzt)
    } else {
      fluege.halte(bild.kennung)
      sprite.position.set(bild.x, bild.y)
      sprite.rotation = bild.drehung
    }

    sprite.zIndex = bild.hoehenrang
    sprite.width = bild.breite
    sprite.height = bild.hoehe

    // **Ein Rücken ist ein Bild** — PLAY-03.AK-07
    // (#1711).
    //
    // `Texture.WHITE`, getönt mit `farben.ruecken`, wäre ein einzelnes weißes
    // Rechteck, wo eine Bibliothekskarte liegt, während die Listenansicht
    // daneben über `bausteine/Kartenbild.tsx` das echte Rückenbild zeigt —
    // dieselbe Partie, zwei Wahrheiten.
    //
    // **`verdeckt` bleibt die getönte Fläche**, und das ist kein Widerspruch:
    // dort **gibt** es kein Rückenbild, weil das Spiel
    // keines führt oder es noch unterwegs ist. Ein Platzhalter, der sagt „hier
    // liegt etwas verdeckt", ist dann die ehrliche Antwort.
    if (bild.inhalt.art === 'bild' || bild.inhalt.art === 'ruecken') {
      texturen.hole(sprite, bild.inhalt.adresse, bild)
    } else {
      texturen.platzhalter(
        sprite,
        bild.inhalt.art === 'verdeckt' ? farben.ruecken : farben.wartet,
      )
    }
  }

  // umrisse zieht den Rand um jede Karte — dieselbe Linie, die die Flächen
  // trennt, und aus demselben Grund.
  const beiwerk = neuesBeiwerk(pixi, ebenen, farben)

  return {
    zeichne({
      plan,
      kamera,
      bild,
      gewaehlt,
      hintergrund = '',
      bewegt = false,
      beschriftung,
      sitzkopf,
      wirkungen,
      beschriftungsstufe,
    }: Zeichenwunsch) {
      hintergrundbild.setze(hintergrund, plan)
      const lage = weltlage(kamera, zeichenflaeche(anwendung.renderer))
      welt.position.set(lage.x, lage.y)
      welt.pivot.set(lage.pivotX, lage.pivotY)
      welt.scale.set(lage.skalierung)
      welt.rotation = lage.drehung

      zeichneGrund(plan)

      const bilder = kartenbilder(plan, bild)
      const lauf = abgeglichen(new Set(sprites.keys()), bilder)

      for (const kennung of lauf.weg) {
        sprites.get(kennung)?.destroy()
        sprites.delete(kennung)
      }

      for (const eines of lauf.neu) {
        const sprite = new pixi.Sprite(pixi.Texture.WHITE)
        sprite.anchor.set(0.5)
        sprites.set(eines.kennung, sprite)
        ebenen.karten.addChild(sprite)
        setze(sprite, eines, false)
      }

      for (const eines of lauf.bleibt) {
        const sprite = sprites.get(eines.kennung)
        if (sprite) setze(sprite, eines, bewegt)
      }

      beiwerk.zeichne(
        plan,
        bilder,
        gewaehlt,
        beschriftung,
        sitzkopf,
        wirkungen,
        beschriftungsstufe ?? 'normal',
      )

      // **Auf Zuruf und nicht im Takt** — siehe `autoStart` oben.
      anwendung.render()

      // **Und mit Ticker, solange etwas fliegt.** Er wird für die Dauer einer
      // Bewegung angeworfen und hält danach wieder an; sechzig Bilder je
      // Sekunde für einen stillstehenden Tisch wären der Preis für nichts.
      fluege.treibe(anwendung.ticker, sprites)
    },

    masse: () => zeichenflaeche(anwendung.renderer),

    passeAn() {
      anwendung.resize()

      return zeichenflaeche(anwendung.renderer)
    },

    faerbe(neue: Tischfarben) {
      farben = neue
      anwendung.renderer.background.color = neue.grund
      beiwerk.faerbe(neue)
    },

    // **Der Abzug** — PLAY-05.AK-03 und AK-04
    // (#1713).
    //
    // **`target: welt` und ein `frame` in Millimetern.** Der Weltcontainer
    // trägt die Kameratransformation; seine **Kinder** liegen in
    // Millimetern, und `frame` wird im lokalen Raum des Ziels gelesen. Damit
    // ist der Abzug unabhängig davon, wo die Kamera gerade steht — nur der
    // gewählte Ausschnitt zählt (nachgelesen in `ExtractSystem.ts` von
    // PixiJS 8, `BaseExtractOptions`).
    //
    // **`resolution` ist Punkte je Millimeter**, also `dpi / 25,4`; die
    // Rechnung dazu steht in `tischausschnitt.ts`.
    //
    // **`antialias: true`**, weil ein Abzug gedruckt werden kann und eine
    // Treppe auf Papier bleibt.
    //
    // **Und ohne halbe Bewegung**: was noch fliegt, landet vorher an seinem
    // Ziel (Befund 3 des Zweitblicks über Phase 6.9).
    async abzug(ausschnitt, punkteJeMillimeter) {
      fluege.lande(sprites)

      return anwendung.renderer.extract.base64({
        target: welt,
        frame: new pixi.Rectangle(
          ausschnitt.x,
          ausschnitt.y,
          ausschnitt.breite,
          ausschnitt.hoehe,
        ),
        resolution: punkteJeMillimeter,
        antialias: true,

        // **Der Tischgrund kommt mit.** Ohne ihn stünde der Abzug auf
        // durchsichtigem Grund, und wer ihn in ein Dokument legt, bekäme
        // Karten auf Weiß statt auf dem Tisch, den er sieht.
        clearColor: farben.grund,
      })
    },

    beiVerlust(melde: () => void) {
      leinwand.addEventListener('webglcontextlost', () => {
        // **Auf einem verlorenen Kontext gibt es nichts aufzuräumen** — und
        // ein Takt, der danach weiterliefe, wäre genau die Schleife, die
        // PLAY-06.AK-04 ausschließt.
        clearInterval(aufraeumtakt)
        melde()
      })
    },

    raeume() {
      clearInterval(aufraeumtakt)
      sprites.clear()
      texturen.raeume()
      fluege.raeume()
      hintergrundbild.raeume()
      beiwerk.raeume()
      anwendung.destroy(true, { children: true })
      // **Erst jetzt die Texturen** — kein Sprite zeigt sie mehr (B04).
      texturen.freigeben()
    },
  }
}
