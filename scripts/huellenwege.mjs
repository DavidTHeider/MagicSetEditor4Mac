// **Die Wege, gezählt und als Datensatz** — P0
// (#2336),
// Befund C-22 aus dem Audit zur Produktionsreife 0.9.0.
//
// `huellenzaehler.mjs` zählt Handlungen und Maschinenzeit. Für die
// Ausgangswerte von Spielen und Handeln verlangt P0 mehr je Weg: den
// **Eingang**, die **Art** (Maus oder Tastatur), die **Sichtwechsel**, die
// **Tabstellen**, die **Fehlversuche** und die **Wirkung** — und alles als
// Zeile eines Datensatzes, damit der Nachherlauf **dieselben Rohdaten und
// dieselben Zählgrenzen** benutzt. Den Vertrag des Datensatzes hält
// `ausgangswege.mjs`.
//
// # Was ein Sichtwechsel ist
//
// **Ein anderer Blick nach einer Handlung**: eine andere Hauptüberschrift im
// Inhalt, eine Schublade, ein Dialog oder ein Menü, das auf- oder zugeht —
// oder am Tisch eine andere **Sicht für** einen Sitz, denn im Hot-Seat ist
// das der Wechsel, den Audit C zählt (S2, S4, S8). Ein Kasten, der aufklappt,
// ist keiner — er ändert den Blick nicht, nur seine Länge.
//
// # Was er nicht zählt
//
// **Lesen, Suchen, Zögern** — Maschinenzeit ist keine Menschenzeit. Und
// **Fehlversuche eines Menschen**: die Probe kennt den Weg; die Spalte steht
// da, damit die Tafel dieselbe Gestalt hat wie die eines Menschen.

import { zaehler } from './huellenzaehler.mjs'

// sichtschluessel beschreibt den Blick, den das Fenster gerade zeigt.
export const sichtschluessel = `/* sichtschluessel */
  const sichtbar = (e) => !!e && e.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true })
  const ueber = [...document.querySelectorAll('[role=dialog], [role=alertdialog], [role=menu]')].filter(sichtbar)
    .map((d) => d.getAttribute('aria-label') || d.querySelector('h1, h2, h3')?.innerText.trim() || d.getAttribute('role'))
  const titel = [...document.querySelectorAll('main h1')].find(sichtbar)?.innerText.trim() ?? ''
  const sicht = [...document.querySelectorAll('label')].find((l) => sichtbar(l) && l.textContent.startsWith('Sicht für'))?.querySelector('select')?.value ?? ''
  return [titel, sicht, ...ueber].join('|')
`

// tabBis drückt ⇥, bis der Fokus auf dem Ziel steht, und gibt die Zahl der
// Drücke — oder -1, wenn das Ziel in `hoechstens` Drücken nicht kommt. Das
// ist selbst ein Messwert: ein Ziel, das ⇥ nicht erreicht, hat keinen
// Tastaturweg.
//
// **Gefragt wird erst, wenn der Druck angekommen ist.** Die Taste geht über
// die Hülle, und die Seite sieht sie später als die Antwort darauf; am
// 2026-10-13 sah die Prüfung den alten Fokus, drückte noch einmal, und der
// Name landete in der Deckwahl. Bewegt ein Druck den Fokus gar nicht, wird
// höchstens `frist` Millisekunden gewartet. `druecke` lässt einen Zähler
// den Druck sehen.
export async function tabBis(p, ziel, hoechstens = 80, frist = 1000, druecke = () => p.taste('Tab')) {
  for (let druck = 1; druck <= hoechstens; druck += 1) {
    await p.seite('window.__vorTab = document.activeElement; return true')
    await druecke()
    const bis = Date.now() + frist
    while (!(await p.seite('return document.activeElement !== window.__vorTab')) && Date.now() < bis) {
      await new Promise((weiter) => setTimeout(weiter, 20))
    }

    if (await p.seite(`return document.activeElement === (${ziel})`)) return druck
  }

  return -1
}

// wegzaehler zählt einen Weg mit allen Spalten aus P0 und legt ihn in den
// Datensatz. `uhr` ist austauschbar, damit sich die Zählung ohne Fenster
// prüfen lässt.
//
// **Der Blick wird gelesen, bis er ruht** — höchstens `frist` Millisekunden,
// alle `takt` (Befund 5 des Zweitblicks über Phase 0): ein Klick kehrt
// zurück, wenn er angekommen ist, und eine Navigation zeichnet einen Takt
// später. **Ruhe heißt `ruhe` gleiche Abfragen hintereinander**, nicht zwei:
// zwei gleiche alte Blicke beendeten das Warten, bevor der neue kam (Befund 3
// des Nachlaufs). **Und vor jeder Handlung wird noch einmal gelesen** — ein
// Wechsel, der erst nach der Frist ankommt, zählt dann für die Handlung, die
// ihn ausgelöst hat, statt mit dem der nächsten in einen zu fallen. Und
// **`ende` nimmt den Abschlussblick** — sonst fiele der letzte Wechsel eines
// Wegs aus der Zählung, weil das Drehbuch erst danach auf die Wirkung wartet.
//
// **Die Maschinenzeit enthält dieses Warten** — mindestens `ruhe` × `takt`
// je Handlung. Sie ist keine Zielgröße; verglichen wird sie nur mit einem
// Lauf desselben Zählers.
export function wegzaehler(p, datensatz, uhr = Date.now, { takt = 100, ruhe = 3, frist = 1500 } = {}) {
  const grund = zaehler(p, uhr)
  let kopf = { weg: '', eingang: '', art: '' }
  let sichtwechsel = 0
  let tabstellen = 0
  let vorher = ''
  let griffe = []

  const blick = async () => String((await p.seite(sichtschluessel)) ?? '')
  const stabil = async () => {
    let letzter = await blick()
    let gleich = 0
    const bis = Date.now() + frist
    while (gleich < ruhe && Date.now() < bis) {
      await new Promise((weiter) => setTimeout(weiter, takt))
      const neu = await blick()
      gleich = neu === letzter ? gleich + 1 : 0
      letzter = neu
    }

    return letzter
  }
  const vergleiche = (jetzt) => {
    if (jetzt !== vorher) sichtwechsel += 1
    vorher = jetzt
  }
  const vorab = async () => vergleiche(await blick())
  const nachher = async () => vergleiche(await stabil())
  const gezaehlt = (ruf) => async (...werte) => {
    await vorab()
    await ruf(...werte)
    await nachher()
  }

  return {
    async beginne(weg, eingang, art) {
      kopf = { weg, eingang, art }
      sichtwechsel = 0
      tabstellen = 0
      griffe = []
      vorher = await blick()
      grund.beginne()
    },
    klick: gezaehlt(grund.klick),
    text: gezaehlt(grund.text),
    klickAuf: gezaehlt(grund.klickAuf),
    tippe: gezaehlt(grund.tippe),
    menue: gezaehlt(grund.menue),
    wahl: gezaehlt(grund.wahl),
    dialog: gezaehlt(grund.dialog),
    async taste(taste, ...mit) {
      if (taste === 'Tab') tabstellen += 1
      await vorab()
      await grund.taste(taste, ...mit)
      await nachher()
    },
    // ende schließt den Weg. `wirkung` ist die beobachtete Wirkung im
    // Bestand oder am Tisch — ohne sie ist der Weg gegangen, nicht gemessen.
    async ende({ wirkung }) {
      const { handlungen, dauer } = grund.ende(`${kopf.weg} (${kopf.eingang}, ${kopf.art})`)
      vergleiche(await blick())
      const zeile = {
        ...kopf, handlungen, sichtwechsel, tabstellen, fehlversuche: 0,
        maschinenzeitMs: dauer, wirkung: Boolean(wirkung), bedingt: griffe,
      }
      datensatz.wege.push(zeile)
      p.pruefe(`${kopf.weg} (${kopf.eingang}, ${kopf.art}): Wirkung beobachtet`, Boolean(wirkung))

      return zeile
    },
    // bedingt nennt einen Griff, der vom Stand abhängt — `gegriffen`, wenn
    // die Probe ihn tat, sonst ausgelassen. **Zwei Läufe desselben Wegs
    // dürfen sich nur hier unterscheiden** (Befund 3 des Nachlaufs über
    // Phase 0): eine Zeile, die einen Zusatzgriff still mitzählt, sieht wie
    // ein längerer Weg aus.
    bedingt(grundsatz, gegriffen) {
      griffe.push({ grund: grundsatz, gegriffen: Boolean(gegriffen) })
    },
    // ohneGang nennt einen Weg, den die Probe nicht gehen kann oder den es
    // nicht gibt — mit `NOT_RUN` oder `FEHLT` und Grund, statt ihn wegzulassen.
    ohneGang(weg, eingang, art, status, grundsatz) {
      datensatz.wege.push({ weg, eingang, art, status, grund: grundsatz })
      p.zeige(`${weg} (${eingang}, ${art}): ${status} — ${grundsatz}`)
    },
  }
}
