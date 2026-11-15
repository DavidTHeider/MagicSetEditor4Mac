// **Die Griffe am Spieltisch für die Hüllenprobe** — gezogen aus
// `huellenprobe/ausgangswerte09spielen.mjs` (P0,
// #2336),
// damit das Drehbuch unter der Zeilengrenze aus §2 bleibt und die Messungen
// der Phase 7 dieselben Griffe benutzen statt eigener.
//
// **Nicht unter `huellenprobe/`**: dort ist jede Datei ein Drehbuch.

import { tabBis } from './huellenwege.mjs'

// namensfeldLeeren fokussiert und leert ein Namensfeld — **ungezählt**: seit
// P6 (#2374) steht dort ein Vorschlag, und ein Mensch markiert und tippt
// darüber (wie beim Lebensfeld in `ausgangswerte09spielen`).
export const namensfeldLeeren = (n) => `const e = ${namensfeld(n)}; e.focus(); Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(e, ''); e.dispatchEvent(new Event('input', { bubbles: true })); return true`
export const namensfeld = (n) => `[...document.querySelectorAll('input')].find((x) => x.checkVisibility() && (x.labels?.[0]?.textContent ?? '').trim() === 'Name am Tisch ${String(n)}')`
export const deckwahl = (stelle, deck) => `const e = [...document.querySelectorAll('section[aria-label="Wer sitzt am Tisch"] select')][${String(stelle)}]; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(e, ${JSON.stringify(deck)}); e.dispatchEvent(new Event('change', { bubbles: true })); return true`
export const anfangen = "[...document.querySelectorAll('main button')].filter((b) => b.checkVisibility() && b.innerText.trim() === 'Partie anfangen').at(-1)"
export const sichtwert = "return [...document.querySelectorAll('label')].find((l) => l.textContent.startsWith('Sicht für'))?.querySelector('select')?.value"
export const sichtwahl = (sitz) => `const f = [...document.querySelectorAll('label')].find((l) => l.textContent.startsWith('Sicht für'))?.querySelector('select'); if (!f) return false; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(f, ${JSON.stringify(sitz)}); f.dispatchEvent(new Event('change', { bubbles: true })); return true`
export const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.replace(/\\s+/g, ' ').trim().replace(/^[^\\p{L}]+/u, '').startsWith(${JSON.stringify(text)}))`

// kartenknopf findet am gezeichneten Tisch den Knopf einer Karte, der an
// seiner Mitte obenauf liegt — `klickeAuf` prüft das, und ein Stapel
// verdeckter Karten trüge sonst die darüber.
export const kartenknopf = (name, zone) => `(() => { const obenauf = (x) => { const r = x.getBoundingClientRect(); return x.contains(document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2)) }; return [...document.querySelectorAll('[role=application] li button')].find((x) => x.checkVisibility() && (x.textContent ?? '').startsWith(${JSON.stringify(`${name} · ${zone}`)}) && obenauf(x)) })()`

// waehlen ist der Klick, der eine Karte am Tisch wählt — **⌘-Klick** seit P4
// (#2372); ein schlichter Klick löst die Primärhandlung aus. Die Wege über
// die Handlungsleiste bleiben damit, was sie in P0 waren.
export const waehlen = { mit: ['befehl'] }

// waehlenMitTasten wählt die Karte mit dem Fokus über ihr Menü: ⇧F10, Ende
// („Weitere Handlungen …“), ↩ — der Tastaturweg in die Leiste seit P4.
export const waehlenMitTasten = [{ taste: ['F10', 'umschalt'] }, { taste: ['End'] }, { taste: ['Enter'] }]

// mitteVon gibt die Mitte eines Elements. **Gerollt wird nur auf Wunsch**:
// wer zwei Punkte für ein Ziehen misst, rollt einmal — ein zweites Rollen
// machte den ersten Punkt ungültig, und die Geste ging ins Leere.
export const mitteVon = (p, ausdruck, rollen = true) => p.seite(`const e = (${ausdruck}); if (!e) return null; if (${String(rollen)}) e.scrollIntoView({ block: 'center', behavior: 'instant' }); const r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }`)

// zugpunkte misst Start und Ziel eines Ziehens nach **einem** Rollen.
export async function zugpunkte(p, von, nach) {
  await p.seite("document.querySelector('[role=application]')?.scrollIntoView({ block: 'start', behavior: 'instant' }); return true")

  return [await mitteVon(p, von, false), await mitteVon(p, nach, false)]
}
export const zonenfeld = (zone, sitz) => `document.querySelector('[data-zonenfeld="${zone}"][data-sitz="${sitz}"]')`

// stand liest eine Partie so, wie Ada sie sieht.
export const stand = (p, ort, partie, sitz = 'ada') => JSON.parse(p.mse(ort, 'partie-zeigen', partie, sitz, '--json')).partie
// heisst sagt, ob ein Objekt die Karte `name` ist — der Name steht unter
// `eigenschaften`; `name` bleibt bei einer Karte aus dem Bestand leer.
export const heisst = (o, name) => (o.eigenschaften?.name || o.name) === name
export const zaehle = (p, ort, partie, name, zone) => stand(p, ort, partie).objekte.filter((o) => heisst(o, name) && o.zone === zone).length
// kennungen nennt die Partien des Bestands; `hinzu` die eine, die seit
// `vorher` dazukam. **Nicht die erste der Liste**: sie ist nach dem letzten
// Zug geordnet, und eine eben begonnene ältere Partie stünde davor.
export const kennungen = (p, ort) => JSON.parse(p.mse(ort, 'partien', '--json')).partien.map((x) => x.id)
export const hinzu = (p, ort, vorher) => kennungen(p, ort).find((id) => !vorher.includes(id))

// tabZu drückt ⇥ über den Zähler, bis das Ziel den Fokus hat — oder gibt
// nach `hoechstens` Drücken `false`. Gewartet wird wie in `tabBis`, bis ein
// Druck angekommen ist.
export async function tabZu(p, z, ausdruck, hoechstens = 200) {
  return (await tabBis(p, ausdruck, hoechstens, 1000, () => z.taste('Tab'))) > 0
}

// tastenweg geht einen Weg nur mit der Tastatur. Je Schritt: `[was, ziel]`
// heißt ⇥ bis zum Ziel und ↩, `[was, ziel, text]` ⇥ bis zum Feld und den
// Text tippen, `{ taste }` ein Kürzel, `{ tab, was }` nur ⇥ bis zum Ziel
// (ohne ↩ — eine Karte löst seit P4 mit ↩ ihre Primärhandlung aus), und
// `{ wahl, wo }` eine Wahl in einer
// Liste, die die Probe setzt, **sobald `wo` dasteht** — ohne das Warten lief
// sie einmal gegen ein Formular, das noch nicht offen war. Ein Ziel, das ⇥
// nicht erreicht, macht den Weg zu `FEHLT` mit seinem Namen.
export async function tastenweg(p, z, weg, schritte, wirkung, eingang = 'tisch') {
  await z.beginne(weg, eingang, 'tastatur')
  for (const schritt of schritte) {
    if (schritt.wahl !== undefined) {
      await p.warte('die Liste steht', () => p.seite(`return !!${schritt.wo}`), 10000)
      await z.wahl(schritt.wahl)
      continue
    }
    if (schritt.tab !== undefined) {
      if (!(await tabZu(p, z, schritt.tab))) {
        z.ohneGang(weg, eingang, 'tastatur', 'FEHLT', `${schritt.was} ist in 200 ⇥ nicht zu erreichen`)

        return
      }
      continue
    }
    if (!Array.isArray(schritt)) {
      await z.taste(...schritt.taste)
      continue
    }
    const [was, ziel, text] = schritt
    await p.warte(`${was} steht`, () => p.seite(`return !!${ziel}`), 10000).catch(() => false)
    if (!(await tabZu(p, z, ziel))) {
      z.ohneGang(weg, eingang, 'tastatur', 'FEHLT', `${was} ist in 200 ⇥ nicht zu erreichen`)

      return
    }
    if (text === undefined) await z.taste('Enter')
    else await z.tippe(text)
  }
  await p.warte(`${weg} wirkt`, wirkung, 10000).catch(() => false)
  await z.ende({ wirkung: await wirkung() })
}

// zumTisch öffnet eine Partie über die Liste — Weg, kein Messwert. `als`
// ist der Name des ersten Sitzes; die Probe legt Partien mit eigenen Namen
// an, denn »Weiterspielen: Ada · Bo« träfe die erste von mehreren.
export async function zumTisch(p, partie, als = 'Ada') {
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  const weiter = `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '') === ${JSON.stringify(`Weiterspielen: ${partie}`)})`
  await p.warte(`»Weiterspielen: ${partie}« steht`, () => p.seite(`return !!${weiter}`), 20000)
  await p.klickeAuf(weiter, `»Weiterspielen: ${partie}«`)
  await p.klickeText(`Weiterspielen als ${als}`)
  await p.warte('der Tisch trägt seine Zonenmarken', () => p.seite(`return !!${zonenfeld('hand', 'ada')}`), 30000)
  await p.ruhe()
}

// schliessen legt den offenen Tisch beiseite — Weg, kein Messwert.
export async function schliessen(p) {
  if (await p.seite(`return !!${knopf('Partie schließen')}`)) {
    await p.klickeAuf(knopf('Partie schließen'), '»Partie schließen«')
    await p.warte('der Tisch ist zu', () => p.seite("return !document.querySelector('[role=application]')"), 10000)
  }
}

// begonnen legt eine Partie über `mse` an und spielt ihren Beginn.
export function begonnen(p, ort, ada, bo, ohneBeginn = false, namen = ['Ada', 'Bo']) {
  const vorher = kennungen(p, ort)
  p.mse(ort, 'partie-anlegen', 'magic', `ada:${namen[0]}:${ada}`, `bo:${namen[1]}:${bo}`)
  const partie = hinzu(p, ort, vorher)
  if (!ohneBeginn) {
    for (const sitz of ['ada', 'bo']) {
      p.mse(ort, 'partie-befehl', partie, sitz, `start-${sitz}`, 'spieler.starthand')
      p.mse(ort, 'partie-befehl', partie, sitz, `behalten-${sitz}`, 'spieler.behalten')
    }
  }

  return partie
}

// bisZug passt über `mse`, bis `ziel(zug)` gilt, und sagt, ob es gelang —
// Vorbereitung, kein Messwert: der Weg in die Hauptphase gehört nicht zu
// dem, was S3 bis S9 zählen. Ein Land geht nur dort (CR 116.2a, Fassung
// 2026-09-25), und dieses Programm lehnt es davor ab.
export function bisZug(p, ort, partie, ziel, hoechstens = 40) {
  for (let schritt = 0; schritt < hoechstens; schritt += 1) {
    const { zug } = stand(p, ort, partie)
    if (ziel(zug)) return true
    const kennung = `vor-${String(Date.now())}-${String(schritt)}`
    // **Im Aufräumschritt hat niemand Priorität**, und wer mehr als sieben
    // Karten hält, wirft ab (CR 514.1, Fassung 2026-09-25) — sonst endet
    // der Zug nicht.
    if (zug.abwurf > 0 && !zug.prioritaet) {
      const karte = stand(p, ort, partie, zug.aktiv).objekte.find((o) => o.zone === 'hand' && o.eigner === zug.aktiv)
      p.mse(ort, 'partie-befehl', partie, zug.aktiv, kennung, 'zug.abwerfen', JSON.stringify({ objekt: karte?.kennung ?? '' }))
      continue
    }
    p.mse(ort, 'partie-befehl', partie, zug.prioritaet || zug.aktiv, kennung, 'zug.passen')
  }

  return false
}

// meldungen nennt, was das Fenster als Absage zeigt — ein Weg ohne Wirkung
// steht im Beleg mit dem Satz, den ein Mensch gesehen hätte.
export const meldungen = (p) => p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter((t) => t.startsWith('Das ging nicht') || t.includes('CR '))")
