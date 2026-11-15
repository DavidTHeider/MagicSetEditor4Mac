// **Eine Karte zonengültig bewegen, am nativen Fenster** — FIX-F04
// (#1748)
// und die Spezifikation S05.
//
// - S05.NAT-04: nach dem Anfangen steht der Tisch da, und der Zug ist zu
//   sehen und mit der Tastatur zu erreichen,
// - NAT-01/S05.NAT-01: eine Karte geht aus der Bibliothek auf die eigene
//   Hand — über das Formular im Fenster,
// - NAT-02/S05.NAT-02: die geordnete Bibliothek bietet eine Stelle an, die
//   ungeordnete Hand nicht; wer von der Bibliothek zur Hand wechselt, schickt
//   keine Stelle mit,
// - S05.NAT-03: die Sicht des Gegenübers zeigt keine verdeckte Hand.
//
// **Der Beginn geht vorweg, wenn er verlangt ist** (#1851): Starthand und
// Behalten für beide Sitze, wie `play-beginn.probe`.

const sichtAuf = (p, wert) => p.seite(`const f = [...document.querySelectorAll('label')].find((l) => l.textContent.startsWith('Sicht für'))?.querySelector('select'); if (!f) return false; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(f, ${JSON.stringify(wert)}); f.dispatchEvent(new Event('change', { bubbles: true })); return true`)

const knoepfe = (p) => p.seite("return [...document.querySelectorAll('[role=application] li button')].map((b) => b.getAttribute('aria-label') || b.textContent)")

// haende zählt die Handkarten in der gewählten Sicht: die eigenen tragen
// ihren Namen, die fremden stehen verdeckt da.
async function haende(p) {
  const hand = (await knoepfe(p)).filter((t) => t.endsWith('· Hand'))
  const verdeckt = hand.filter((t) => t.startsWith('verdeckt')).length

  return { offen: hand.length - verdeckt, verdeckt }
}

// vorhangZu ist der Vorhang zwischen zwei Sitzen am selben Gerät (P7, #2375).
const vorhangZu = "return [...document.querySelectorAll('section')].some((s) => s.checkVisibility() && (document.getElementById(s.getAttribute('aria-labelledby') ?? '')?.textContent ?? '').trim().startsWith('Gerät an '))"

// sicht wechselt die Sicht und wartet, bis der Tisch sie zeigt — erkennbar
// an der Zahl der offenen Handkarten, die der neue Sitz sieht.
//
// **Seit P7 kann davor der Vorhang stehen** (#2375): hat der **andere** Sitz
// das Wort, fragt er, ob das Gerät an ihn geht. **Esc** bleibt bei der
// gewählten Sicht — ↩ gäbe es zurück. Bis dahin fand diese Probe hinter ihm
// keine Hand.
async function sicht(p, sitz, offen) {
  await sichtAuf(p, sitz)
  // **Abgefragt und nicht abgewartet** — ein Vorhang, der nicht kommt, ist
  // kein Befund: er steht nur, wenn der andere Sitz entscheidet.
  await new Promise((los) => setTimeout(los, 1500))
  if (await p.seite(vorhangZu)) await p.taste('Escape')
  await p.warte(`die Sicht für ${sitz} steht`, async () => (await haende(p)).offen === offen)
  await new Promise((los) => setTimeout(los, 500))
}

async function beginnFuer(p, sitz) {
  await sichtAuf(p, sitz)
  const offen = await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Starthand ziehen')")
  if (!offen) return false
  await p.klickeText('Starthand ziehen')
  await p.wartetext('Deine Starthand hat')
  await p.klickeText('Hand behalten')
  await p.warte('behalten', () => p.seite("return !document.body.innerText.includes('Deine Starthand hat')"))

  return true
}

export default async function fixF04(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')

  await p.starte({ ort })

  p.notiz('S05.NAT-04 — der Tisch steht da, und der Zug ist erreichbar')
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  // **Die Sitze stehen im Menü „Weiterspielen" der Zeile** — seit UXR-13
  // (#1921); vorher trug die Zeile je Sitz einen Knopf.
  await p.wartetext('Weiterspielen')
  await p.klickeText('Weiterspielen')
  await p.klickeText('Weiterspielen als Ada')
  await p.warte('die Leinwand zeichnet', () => p.seite("return !!document.querySelector('canvas') && document.querySelector('canvas').checkVisibility()"), 20000)
  const tisch = await p.seite("const a = document.querySelector('[role=application]'); const z = document.querySelector('[role=progressbar][aria-labelledby]'); return { tisch: !!a, fokussierbar: !!a && a.tabIndex >= 0, zug: z ? z.getAttribute('aria-valuenow') + '/' + z.getAttribute('aria-valuemax') + ' ' + (document.getElementById(z.getAttribute('aria-labelledby'))?.textContent ?? '') : '' }")
  p.pruefe('der Tisch steht da und lässt sich fokussieren', tisch.tisch && tisch.fokussierbar, JSON.stringify(tisch))
  p.pruefe('der Zug ist zu sehen', tisch.zug !== '', tisch.zug)
  await p.bild('f04-tisch')

  const beginn = [await beginnFuer(p, 'ada'), await beginnFuer(p, 'bo')]
  p.zeige(`der Beginn war ${beginn.some(Boolean) ? 'verlangt und ist gespielt' : 'nicht verlangt'}`)
  // **Nach dem Beginn hält jeder seine Starthand** — sieben Karten, offen
  // für den eigenen Sitz, verdeckt für den anderen.
  const starthand = beginn[0] ? 7 : 0
  await sicht(p, 'ada', starthand)

  p.notiz('NAT-01 und NAT-02 — Bibliothek zur Hand, ohne Stelle')
  const vorher = await haende(p)
  p.zeige(`Hände aus Adas Sicht: ${JSON.stringify(vorher)}`)
  // **Die eigenen Zonen liegen unten**, wie am Tisch — die unterste
  // Bibliothek ist Adas (dieselbe Wahl wie `play03-quer.probe`). **Und nur
  // eine Karte, die an ihrer Mitte obenauf liegt**: die Bibliothek ist ein
  // Stapel, und der Klick auf die Mitte einer verdeckten Karte darunter
  // trifft die darüber — `klickeAuf` prüft das seit WK03 und hat es hier
  // angestrichen.
  await p.klickeAuf("(() => { const obenauf = (x) => { const r = x.getBoundingClientRect(); return x.contains(document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2)) }; const b = [...document.querySelectorAll('[role=application] li button')].filter((x) => x.checkVisibility() && (x.getAttribute('aria-label') || x.textContent).includes('Bibliothek') && obenauf(x)); return b.reduce((a, c) => (c.getBoundingClientRect().top > a.getBoundingClientRect().top ? c : a)) })()", 'die oberste Karte aus Adas Bibliothek')
  await p.wartetext('Mit dieser Karte')
  await p.klickeText('Schieben nach…')
  await p.warte('das Formular steht', () => p.seite("return document.querySelector('[role=group][aria-label=\"Schieben nach\"]') !== null"))
  const setze = async (zone) => {
    await p.seite(`const g = document.querySelector('[role=group][aria-label="Schieben nach"]'); const ziel = g.querySelector('select'); Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(ziel, ${JSON.stringify(zone)}); ziel.dispatchEvent(new Event('change', { bubbles: true })); return true`)
    return p.warte(`das Formular zeigt ${zone}`, () => p.seite(`const g = document.querySelector('[role=group][aria-label="Schieben nach"]'); const s = [...g.querySelectorAll('select')]; return s[0].value === ${JSON.stringify(zone)} ? s.map((x) => [...x.options].map((o) => o.value)) : false`))
  }
  const bibliothek = await setze('bibliothek')
  p.pruefe('die Bibliothek bietet eine Stelle an', bibliothek.length === 2 && bibliothek[1].length >= 2, JSON.stringify(bibliothek[1] ?? []))
  const hand = await setze('hand')
  p.pruefe('die Hand bietet keine an', hand.length === 1, JSON.stringify(hand))
  await p.klickeText('Stellen')
  const nachher = await p.warte('die Karte liegt auf der Hand', async () => {
    const jetzt = await haende(p)
    return jetzt.offen > vorher.offen ? jetzt : false
  }).catch(async (grund) => {
    p.zeige(`nachher: ${JSON.stringify(await haende(p))}`)
    p.zeige(`Hinweise: ${JSON.stringify(await p.seite("return document.body.innerText.split('\\n').filter((z) => /nicht|kein|Stelle/i.test(z)).slice(0, 8)"))}`)
    throw grund
  })
  p.pruefe('eine Karte mehr auf Adas Hand, und kein Fehler', nachher.offen === vorher.offen + 1 &&
    !(await p.seite("return document.body.innerText.includes('gibt es keine Stelle')")), `${String(vorher.offen)} → ${String(nachher.offen)}`)
  await p.bild('f04-hand')

  p.notiz('S05.NAT-03 — Bo sieht Adas Hand nicht')
  await sicht(p, 'bo', nachher.verdeckt)
  const bo = await haende(p)
  p.pruefe('aus Bos Sicht liegt Adas ganze Hand verdeckt da, keine Karte mit Namen', bo.verdeckt === nachher.offen, JSON.stringify(bo))
  p.pruefe('⌘Q beendet', await p.beende())
}
