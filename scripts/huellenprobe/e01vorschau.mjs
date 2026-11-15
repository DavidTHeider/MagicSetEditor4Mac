// **Feld → sichtbare Vorschau, fünfmal, kurz und lang** — E01
// (#2238).
//
// B02 bestätigte den Median aus einem Lauf über eine Karte. E01 verlangt
// mehr, und dieses Drehbuch fährt es am nativen Fenster:
//
//   - **fünf Läufe**, je ein frischer Prozess über demselben Bestand,
//   - **ein langer und ein kurzer Kartentext** — „Langtext" aus der Saat und
//     eine Karte, deren Text auf ein Wort gesetzt ist,
//   - je Karte zwölf Eingaben, gemessen wie in `budgets.mjs` (`vorschauzeit`):
//     vom `keydown` bis das Vorschaubild eine neue Quelle hat und dekodiert ist
//     — **je ein anderes Zeichen**, das Löschen dazwischen ungemessen: ein
//     Wechsel aus Zeichen und Löschen bringt jedes zweite Mal einen schon
//     gerenderten Stand zurück, und der kommt aus dem Zwischenspeicher
//     (erster Lauf: 9 ms im Median über der kurzen Karte),
//   - **keine alte Vorschau nach neuem Entwurf**: ein schneller Schub von
//     sechs Zeichen, dann Ruhe — das stehende Bild muss Pixel für Pixel das
//     sein, das dieselbe Karte nach dem Speichern frisch gerendert zeigt,
//   - **der Fehlerweg**: `Probevorschau` scheitert (`stoere`, B03) — die Fläche
//     sagt es —, und nach der Störung kommt die Vorschau wieder.
//
// **Was er nicht belegt**: wie es ein Mensch empfindet, und andere Karten
// als die zwei des Prüfbestands.
import { stoere } from '../huellenrufbuch.mjs'
import { kennzahlen, vorschauzeit } from './budgets.mjs'

const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))
const setname = 'Werkstatt-Testkarten'
const feld = "[...document.querySelectorAll('input, textarea')].find((e) => e.checkVisibility() && e.getAttribute('aria-label') === 'name')"
const zeile = (name) => `[...document.querySelectorAll('[role=row]')].find((e) => (e.getAttribute('aria-label') ?? '').startsWith(${JSON.stringify(name)}))`
const quelle = "return document.querySelector('img[alt=\"Vorschau der Karte\"]')?.src ?? ''"

// abdruck liest die Pixel des Vorschaubilds und gibt einen Abdruck — die
// Quelle allein sagt nichts, sie ist bei jedem Rendern eine neue.
const abdruck = `return (async () => {
  const b = document.querySelector('img[alt="Vorschau der Karte"]')
  if (!b) return null
  await b.decode()
  const c = document.createElement('canvas')
  c.width = b.naturalWidth; c.height = b.naturalHeight
  const x = c.getContext('2d'); x.drawImage(b, 0, 0)
  const d = x.getImageData(0, 0, c.width, c.height).data
  let h = 2166136261
  for (let i = 0; i < d.length; i++) h = Math.imul(h ^ d[i], 16777619) >>> 0
  return h.toString(16) + ' ' + c.width + 'x' + c.height
})()`

// ruhig wartet, bis die Quelle des Vorschaubilds anderthalb Sekunden lang
// dieselbe bleibt.
async function ruhig(p) {
  let letzte = await p.seite(quelle)
  let seit = Date.now()
  const bis = Date.now() + 20000
  while (Date.now() < bis) {
    await schlafe(100)
    const jetzt = await p.seite(quelle)
    if (jetzt !== letzte) { letzte = jetzt; seit = Date.now() }
    if (jetzt && Date.now() - seit >= 1500) return jetzt
  }
  throw new Error('die Vorschau kam nicht zur Ruhe')
}

async function karteWaehlen(p, name) {
  await p.klickeAuf(zeile(name), `die Zeile »${name}«`)
  await p.warte(`das Namensfeld zeigt »${name}«`, () => p.seite(`return ((${feld})?.value ?? '').startsWith(${JSON.stringify(name)})`), 15000)
  await ruhig(p)
}

// messen tippt zwölf verschiedene Zeichen, jedes gemessen, und löscht es
// danach ungemessen wieder.
async function messen(p) {
  const werte = []
  const zeichen = 'bcdfghjklmnpqrstvw'
  for (let i = 0; werte.length < 12 && i < zeichen.length; i++) {
    if (!(await p.seite(`return document.activeElement === (${feld})`))) await p.klickeAuf(feld, 'das Namensfeld')
    const vorher = await p.seite(`return (${feld}).value`)
    await p.seite(vorschauzeit)
    await p.taste(zeichen[i])
    const { ms } = await p.seite('return window.__vorschaumessung')
    if ((await p.seite(`return (${feld}).value`)) !== vorher && ms >= 0) werte.push(ms)
    await p.taste('Backspace')
    await ruhig(p)
  }
  return werte
}

export default async function e01vorschau(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  // **Ein kurzer Text**: die vierte Karte bekommt ein Wort und ihren Namen.
  p.mse(ort, 'feld', '1', '4', 'text', 'Flug.')
  p.mse(ort, 'feld', '1', '4', 'name', 'Kurz')

  const reihen = { lang: [], kurz: [] }
  const runden = Number(process.env.MSE_E01_RUNDEN ?? 5)

  for (let runde = 1; runde <= runden; runde++) {
    p.notiz(`Lauf ${String(runde)}`)
    await p.starte({ ort })
    await p.taste('2', 'befehl')
    await p.wartetext(setname, 30000)
    await p.klickeText(setname)
    await p.warte('die Kartenliste steht', () => p.seite(`return !!(${zeile('Langtext')})`), 30000)

    for (const [art, name] of [['lang', 'Langtext'], ['kurz', 'Kurz']]) {
      await karteWaehlen(p, name)
      reihen[art].push(...(await messen(p)))
    }

    // **Keine alte Vorschau nach neuem Entwurf** — ein schneller Schub, Ruhe,
    // und das Bild gegen das der gespeicherten Karte.
    await karteWaehlen(p, 'Kurz')
    await p.klickeAuf(feld, 'das Namensfeld')
    await p.tippe(`s${String(runde)}abcd`)
    await ruhig(p)
    const nachSchub = await p.seite(abdruck)
    const entwurf = await p.seite(`return (${feld}).value`)
    await p.taste('Tab')
    await p.warte('der Entwurf steht im Bestand', () => p.mse(ort, 'karten', '1').includes(entwurf), 15000)
    await karteWaehlen(p, 'Langtext')
    await karteWaehlen(p, 'Kurz')
    const frisch = await p.seite(abdruck)
    p.pruefe(`Lauf ${String(runde)}: nach dem Schub steht die Vorschau des letzten Entwurfs`, nachSchub !== null && nachSchub === frisch, `${String(nachSchub)} / ${String(frisch)}`)

    // **Der Fehlerweg** — die Vorschau scheitert, die Fläche sagt es, und
    // nach der Störung kommt sie wieder.
    await stoere(p, [{ bindung: 'Probevorschau', fehler: true }])
    await p.klickeAuf(feld, 'das Namensfeld')
    await p.taste('x')
    const meldung = await p.warte('die Vorschau meldet den Fehler', () => p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).join(' / ')"), 15000).catch(() => '')
    await stoere(p, [])
    const vorErholung = await p.seite(quelle)
    // **Ein neues Zeichen**, kein Löschen: das führte in einen schon
    // gerenderten Stand zurück, dessen Quelle dieselbe sein kann.
    await p.taste('z')
    const erholt = await p.warte('die Vorschau kommt wieder', async () => (await p.seite(quelle)) !== vorErholung && !(await p.seite("return [...document.querySelectorAll('[role=alert]')].some((e) => e.checkVisibility())")), 15000).then(() => true, () => false)
    p.pruefe(`Lauf ${String(runde)}: der Fehler der Vorschau steht da und vergeht`, meldung !== '' && erholt, meldung.slice(0, 120))
    p.pruefe(`Lauf ${String(runde)}: ⌘Q beendet`, await p.beende())
  }

  p.notiz('Feld → dekodiertes Vorschaubild')
  p.zeilen.push('| Text | Eingaben | Median | P95 | max |', '| --- | ---: | ---: | ---: | ---: |')
  for (const [art, werte] of Object.entries(reihen)) {
    const k = kennzahlen(werte)
    p.zeilen.push(`| ${art} | ${String(k.n)} | ${String(k.median)} ms | ${String(k.p95 ?? '—')} ms | ${String(k.max)} ms |`)
  }
  p.zeilen.push('', `Rohwerte — lang: ${JSON.stringify(reihen.lang.map(Math.round))}; kurz: ${JSON.stringify(reihen.kurz.map(Math.round))}`)
  const alle = [...reihen.lang, ...reihen.kurz]
  p.pruefe('Median ≤ 150 ms, lang wie kurz', kennzahlen(reihen.lang).median <= 150 && kennzahlen(reihen.kurz).median <= 150, `${String(kennzahlen(reihen.lang).median)} / ${String(kennzahlen(reihen.kurz).median)}`)
  p.pruefe(`jede Eingabe gab eine Vorschau (${String(alle.length)} von ${String(runden * 24)})`, alle.length === runden * 24)
}
