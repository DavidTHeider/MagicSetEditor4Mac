// **Zeigereingaben am nativen Fenster** — A04
// (#2218):
// G3 (eine Karte mit dem Zeiger ziehen — Hand → Spielfeld → Friedhof,
// Escape-Abbruch, Loslassen außerhalb) und das Kontextmenü mit der rechten
// Taste und mit ⇧F10 (an den Plätzen eines Decks).
//
// # Wie die Maus hineinkommt
//
// **Ab AppKit**: die Abnahmeleitung legt `NSEvent`s in die
// Ereignisschlange — runter, gezogen, hoch, rechts —, und WebKit macht
// daraus seine Zeigerereignisse wie bei einer echten Maus
// (`internal/fenster/abnahme_zug.go`). **Kein `dispatchEvent` in der
// Seite.** Die Ziele kommen aus den Marken der Zonenfelder
// (`Tischzonenmarken.tsx`) und den Kartenknöpfen; die Wirkung wird im
// Bestand nachgesehen, über `mse`.
//
// # Was es nicht belegt
//
// Den Fensterserver: ein Zug aus einer anderen Anwendung oder mit der Hand
// eines Menschen bleibt eine Handprüfung (Computer Use war in dieser
// Sitzung nicht freigegeben, siehe #2218).
import { readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { loslassen, rechtsklick, ziehe } from '../huellenrufbuch.mjs'

const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))

// zonen zählt die Zonen eines Sitzes im Bestand.
function zonen(p, ort, partie, sitz = 'ada') {
  const stand = JSON.parse(p.mse(ort, 'partie-zeigen', partie, sitz, '--json')).partie
  const zaehlung = {}
  for (const objekt of stand.objekte) {
    if (objekt.eigner !== sitz) continue
    zaehlung[objekt.zone] = (zaehlung[objekt.zone] ?? 0) + 1
  }

  return zaehlung
}

// mitte gibt die Mitte der Marke eines Zonenfelds in Punkten der Seite.
const mitte = (p, zone, sitz = 'ada') => p.seite(`const r = document.querySelector('[data-zonenfeld="${zone}"][data-sitz="${sitz}"]')?.getBoundingClientRect(); return r ? { x: r.left + r.width / 2, y: r.top + r.height / 2 } : null`)

// karteIn gibt die Mitte eines Kartenknopfs, der im Feld einer Zone liegt.
const karteIn = (p, zone, sitz = 'ada') => p.seite(`
  const feld = document.querySelector('[data-zonenfeld="${zone}"][data-sitz="${sitz}"]')?.getBoundingClientRect()
  if (!feld) return null
  for (const li of document.querySelectorAll('[role=application] li')) {
    const r = li.getBoundingClientRect()
    const x = r.left + r.width / 2, y = r.top + r.height / 2
    if (x >= feld.left && x <= feld.right && y >= feld.top && y <= feld.bottom) return { x, y }
  }
  return null`)

const gleich = (a, b) => JSON.stringify(a) === JSON.stringify(b)

export default async function a04(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  // **M14**: jede Karte trägt ein offenes Schlüsselwort (CR 702) — welche
  // auch immer gezogen wird, auf dem Spielfeld nennt der Tisch es.
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'feld', '1', karte, 'text', 'Hexproof')
  p.mse(ort, 'deck', 'neu', 'A04-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
  const partie = JSON.parse(p.mse(ort, 'partien', '--json')).partien[0].id
  for (const sitz of ['ada', 'bo']) {
    p.mse(ort, 'partie-befehl', partie, sitz, `start-${sitz}`, 'spieler.starthand')
    p.mse(ort, 'partie-befehl', partie, sitz, `behalten-${sitz}`, 'spieler.behalten')
  }
  // **M9**: eine zweite Partie mit zwei wartenden Auslösern desselben Sitzes
  // (CR 603.3b) — über Aus- und Einfuhr, denn welche Karte beide zugleich
  // auslöste, ist hier nicht die Frage, sondern wer sie ordnet.
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'cy:Cy:1')
  const vorlage = JSON.parse(p.mse(ort, 'partien', '--json')).partien.find((eine) => eine.sitze.some((s) => s.kennung === 'cy')).id
  for (const sitz of ['ada', 'cy']) {
    p.mse(ort, 'partie-befehl', vorlage, sitz, `start-${sitz}`, 'spieler.starthand')
    p.mse(ort, 'partie-befehl', vorlage, sitz, `behalten-${sitz}`, 'spieler.behalten')
  }
  const ausfuhr = join(ort.ort, 'reihung.json')
  p.mse(ort, 'partie-ausfuhr', vorlage, ausfuhr)
  const roh = JSON.parse(readFileSync(ausfuhr, 'utf8'))
  roh.partie.Ausgeloest = [
    { Art: 'strahlung', Beherrscher: 'ada', Wen: '', Kennung: 'ada/strahlung/0' },
    { Art: 'monarch.zug', Beherrscher: 'ada', Wen: '', Kennung: 'ada/monarch.zug/0' },
  ]
  roh.partie.Ausloeserreihung = 'ada'
  // **Ein eigener Name für die eingelesene Partie** — die Vorlage bleibt
  // beendet in der Liste, und zwei Einträge »Ada · Cy« ließen die Probe je
  // nach Ordnung die falsche treffen (gesehen im vollen Lauf).
  for (const sitz of roh.partie.Sitze) if (sitz.Kennung === 'cy') sitz.Name = 'Dora'
  writeFileSync(ausfuhr, JSON.stringify(roh))
  p.mse(ort, 'partie-stand', vorlage, 'ada', 'beenden')
  const reihung = p.mse(ort, 'partie-einfuhr', ausfuhr).match(/als ([0-9a-f]{32})/)[1]
  await p.starte({ ort })

  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  const weiter = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '') === 'Weiterspielen: Ada · Bo')"
  await p.warte('»Weiterspielen« steht', () => p.seite(`return !!${weiter}`), 20000)
  await p.klickeAuf(weiter, '»Weiterspielen: Ada · Bo«')
  await p.klickeText('Weiterspielen als Ada')
  await p.warte('der Tisch trägt seine Zonenmarken', () => p.seite("return !!document.querySelector('[data-zonenfeld=\"hand\"][data-sitz=\"ada\"]')"), 30000)
  await p.ruhe()
  await p.bild('tisch-vorher')
  const anfang = zonen(p, ort, partie)
  p.zeige(`Zonen vorher: ${JSON.stringify(anfang)}`)

  p.notiz('Hand → Spielfeld, mit der Maus')
  const handkarte = await karteIn(p, 'hand')
  const spielfeld = await mitte(p, 'tisch')
  p.pruefe('eine Karte der Hand und das Spielfeld stehen da', !!handkarte && !!spielfeld, JSON.stringify({ handkarte, spielfeld }))
  await ziehe(p, [handkarte, { x: (handkarte.x + spielfeld.x) / 2, y: (handkarte.y + spielfeld.y) / 2 }, spielfeld])
  await p.warte('die Karte liegt auf dem Spielfeld', () => (zonen(p, ort, partie).tisch ?? 0) > (anfang.tisch ?? 0), 10000)
  const gelegt = zonen(p, ort, partie)
  p.pruefe('Hand → Spielfeld: eine Karte weniger auf der Hand', (gelegt.hand ?? 0) === (anfang.hand ?? 0) - 1, JSON.stringify(gelegt))
  await p.ruhe()
  await p.bild('tisch-nach-legen')
  // **M14**: die Befundliste über dem Tisch nennt das offene Schlüsselwort.
  await p.warte('der Tisch nennt »hexproof« als nicht durchgesetzt', () => p.seite("return [...document.querySelectorAll('[role=status]')].some((s) => s.checkVisibility() && s.innerText.includes('hexproof'))"), 10000)
  p.pruefe('M14: der Befund steht über dem Tisch', true, await p.seite("return [...document.querySelectorAll('[role=status]')].map((s) => s.innerText).find((t) => t.includes('hexproof')) ?? ''"))

  p.notiz('Spielfeld → Friedhof, mit der Maus')
  const tischkarte = await karteIn(p, 'tisch')
  const friedhof = await mitte(p, 'friedhof')
  await ziehe(p, [tischkarte, { x: (tischkarte.x + friedhof.x) / 2, y: (tischkarte.y + friedhof.y) / 2 }, friedhof])
  await p.warte('die Karte liegt im Friedhof', () => (zonen(p, ort, partie).friedhof ?? 0) > (anfang.friedhof ?? 0), 10000)
  const begraben = zonen(p, ort, partie)
  p.pruefe('Spielfeld → Friedhof: das Spielfeld ist wieder leer', (begraben.tisch ?? 0) === (anfang.tisch ?? 0), JSON.stringify(begraben))
  await p.ruhe()
  await p.bild('tisch-nach-friedhof')

  p.notiz('Ein Zug, den Escape abbricht')
  // **Erst belegen, dass dieser Zug die Karte gefasst hat** — Befund 6 des
  // Zweitblicks über Phase A: ohne diese Zwischenprüfung bestand der Abbruch
  // auch, wenn der Zug gar nicht begonnen hatte.
  const nochEine = await karteIn(p, 'hand')
  await p.seite(`for (const li of document.querySelectorAll('[role=application] li')) delete li.dataset.a04; const r = (li) => li.getBoundingClientRect(); const ziel = [...document.querySelectorAll('[role=application] li')].find((li) => Math.abs(r(li).left + r(li).width / 2 - ${nochEine.x}) < 1 && Math.abs(r(li).top + r(li).height / 2 - ${nochEine.y}) < 1); ziel.dataset.a04 = 'gefasst'; return true`)
  const gefasste = "(() => { const r = document.querySelector('[data-a04=gefasst]')?.getBoundingClientRect(); return r ? { x: r.left + r.width / 2, y: r.top + r.height / 2 } : null })()"
  const weg = (a, b) => Math.hypot(a.x - b.x, a.y - b.y)
  await ziehe(p, [nochEine, spielfeld], true)
  await p.warte('die gefasste Karte geht mit', async () => weg(await p.seite(`return ${gefasste}`), nochEine) > 40, 10000)
  await p.taste('Escape')
  await p.warte('Escape legt sie zurück', async () => weg(await p.seite(`return ${gefasste}`), nochEine) < 1, 10000)
  await loslassen(p, spielfeld.x, spielfeld.y)
  await schlafe(800)
  p.pruefe('Escape bricht ab: keine Zone hat sich verändert', gleich(zonen(p, ort, partie), begraben), JSON.stringify(zonen(p, ort, partie)))
  p.pruefe('Escape bricht ab: die Karte liegt wieder in der Hand', weg(await p.seite(`return ${gefasste}`), nochEine) < 1)

  p.notiz('Ein Zug, der außerhalb des Fensters endet')
  const fenster = await p.seite('return { b: window.innerWidth, h: window.innerHeight }')
  await ziehe(p, [nochEine, spielfeld], true)
  await p.warte('die gefasste Karte geht wieder mit', async () => weg(await p.seite(`return ${gefasste}`), nochEine) > 40, 10000)
  await loslassen(p, -60, Math.round(fenster.h / 2))
  await p.warte('Loslassen außerhalb legt sie zurück', async () => weg(await p.seite(`return ${gefasste}`), nochEine) < 1, 10000)
  // **Und danach ist nichts mehr gefasst**: eine Bewegung ohne Taste über
  // das Spielfeld nimmt die Karte nicht mit.
  await p.frage({ was: 'bewege', x: spielfeld.x, y: spielfeld.y })
  await schlafe(800)
  p.pruefe('Loslassen außerhalb: eine Bewegung ohne Taste zieht nichts mehr', weg(await p.seite(`return ${gefasste}`), nochEine) < 1)
  p.pruefe('Loslassen außerhalb: keine Zone hat sich verändert', gleich(zonen(p, ort, partie), begraben), JSON.stringify(zonen(p, ort, partie)))
  await p.bild('tisch-nach-abbruechen')

  p.notiz('Das Kontextmenü an den Plätzen eines Decks — rechte Taste und ⇧F10')
  // **⌘4 führt seit dem Umbau der Decks in 0.9.0 gleich zur Deckliste**, und
  // ein Deck öffnet sein Knopf »Öffnen: <Name>« — wie in `ap6decks09` und
  // `deckmenue09`; den Zwischenschritt »Decks« gibt es nicht mehr.
  const oeffnen = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim() === 'Öffnen: A04-Deck')"
  await p.taste('4', 'befehl')
  await p.warte('die Deckliste steht', () => p.seite(`return !!${oeffnen}`), 20000)
  await p.klickeAuf(oeffnen, '»Öffnen: A04-Deck«')
  await p.wartetext('Langtext', 20000)
  const zeile = "[...document.querySelectorAll('main li')].find((z) => z.checkVisibility() && !z.closest('nav') && z.onkeydown !== undefined && z.innerText.includes('Langtext') && z.querySelector('*'))"
  const stelle = await p.seite(`const z = ${zeile}; if (!z) return null; z.scrollIntoView({ block: 'center' }); const r = z.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }`)
  p.pruefe('der Platz »Langtext« steht da', !!stelle, JSON.stringify(stelle))
  const menueOffen = "return [...document.querySelectorAll('[role=menu]')].some((m) => m.checkVisibility())"
  await rechtsklick(p, stelle.x, stelle.y)
  await p.warte('das Kontextmenü steht da', () => p.seite(menueOffen), 10000)
  const eintraege = await p.seite("return [...document.querySelectorAll('[role=menu] [role=menuitem]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim())")
  p.pruefe('Rechtsklick: das Kontextmenü nennt seine Einträge', eintraege.length > 0, eintraege.join(' · '))
  await p.bild('kontextmenue-maus')
  await p.taste('Escape')
  await p.warte('Escape schließt das Kontextmenü', async () => !(await p.seite(menueOffen)), 10000)

  // **Den Fokus setzt die Seite**: ein Klick fokussiert in WebKit nicht
  // (`klick-fokussiert-in-webkit-nicht`); geprüft wird hier die Taste.
  await p.seite(`const z = ${zeile}; (z.querySelector('button, [tabindex]') ?? z).focus(); return true`)
  await p.taste('F10', 'umschalt')
  await p.warte('⇧F10 öffnet das Kontextmenü', () => p.seite(menueOffen), 10000)
  const fokus = await p.seite("return document.activeElement?.getAttribute('role') ?? document.activeElement?.tagName")
  p.pruefe('⇧F10: der Fokus steht im Menü', fokus === 'menuitem', fokus)
  await p.bild('kontextmenue-tastatur')
  await p.taste('Escape')

  // **M9**: die offene Wahl der Reihenfolge, mit einem Klick beantwortet.
  p.notiz('M9 — zwei Auslöser zugleich: Ada ordnet')
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  // **Der offene Tisch steht vor der Liste** — erst schließen (`abn01tisch`).
  const schliessen = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Partie schließen')"
  await p.warte('der Tisch oder die Liste steht', () => p.seite(`return !!${schliessen} || document.body.innerText.includes('Eine Partie anfangen')`), 20000)
  if (await p.seite(`return !!${schliessen}`)) await p.klickeAuf(schliessen, '»Partie schließen«')
  await p.wartetext('Eine Partie anfangen', 20000)
  const weiterCy = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '') === 'Weiterspielen: Ada · Dora')"
  await p.warte('»Weiterspielen« der eingelesenen Partie steht', () => p.seite(`return !!${weiterCy}`), 20000)
  await p.klickeAuf(weiterCy, '»Weiterspielen: Ada · Dora«')
  await p.klickeText('Weiterspielen als Ada')
  await p.wartetext('Offene Wahl: Reihenfolge der Fähigkeiten', 20000)
  await p.bild('reihung-vorher')
  const legen = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.includes('Als Nächste legen:') && b.innerText.includes('Radioaktivitätsmarken'))"
  await p.seite(`(${legen}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(legen, '»Als Nächste legen: Radioaktivitätsmarken mahlen«')
  await p.warte('die Wahl ist beantwortet', () => JSON.parse(p.mse(ort, 'partie-zeigen', reihung, 'ada', '--json')).partie.ausloeserreihung === '', 10000)
  const gelegteFaehigkeiten = JSON.parse(p.mse(ort, 'partie-zeigen', reihung, 'ada', '--json')).partie.objekte.filter((o) => o.zone === 'stapel').map((o) => o.name)
  p.pruefe('M9: beide liegen auf dem Stapel, keine wartet mehr', gelegteFaehigkeiten.length === 2, JSON.stringify(gelegteFaehigkeiten))
  await p.bild('reihung-nachher')

  // **⌘P führt zum Druckbogen und nicht in ein Druckpanel** — gedruckt wird
  // ein PDF (0013), und dessen Ort fragt der Sichern-Dialog von macOS.
  p.notiz('⌘P — der Druckbogen und sein Sichern-Dialog')
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.wartetext('Karte anlegen', 20000)
  await p.taste('p', 'befehl')
  await p.warte('der Druckbogen steht da', () => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() === 'Druckbogen'"), 20000)
  await p.wartetext('Druckbogen legen …', 20000)
  await p.klickeText('Druckbogen legen …')
  await p.dialogZeigtSich('sichern', '»Druckbogen legen …«')

  p.pruefe('⌘Q beendet', await p.beende())
}
