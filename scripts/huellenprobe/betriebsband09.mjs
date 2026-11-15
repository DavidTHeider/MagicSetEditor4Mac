// **Das Betriebsband, am nativen Fenster** — OPS02
// (#2348),
// Zielartboards `Start09` und `Betrieb09`.
//
// # Was hier geschieht
//
// 1. **Vorher**: ein frischer Bestand auf einem eigenen Abbild (`hdiutil`,
//    3 GB, dünn) — „Was Aufmerksamkeit braucht" sagt „Nichts".
// 2. **Eine Sicherung von vor zwanzig Tagen**: der Hinweis steht da, und
//    **mit der Tastatur** führt er zu *Einstellungen › Daten und Sicherung*.
// 3. **Der Platz, an zwei Schwellen**: das Abbild wird gefüllt, bis unter
//    2 GiB frei sind — der Hinweis kommt; dann frei bis 2,25 GiB — er bleibt,
//    denn wieder gut ist der Platz erst über 2,5 GiB; dann alles frei — er
//    geht.
// 4. **Der Webprozess endet**: der WebContent-Prozess dieses Fensters wird
//    beendet, Wails lädt die Seite neu, und der Hinweis nennt es.
// 5. **Was ein Abruf kostet**: derselbe Ruf, den das Fenster alle 30
//    Sekunden schickt, hundertmal über die Naht — p95 unter 20 ms.
// 6. **Der Speicher, getrennt**: der Hinweis auf den Webprozess führt zu
//    *Einstellungen › Diagnose*, und dort steht der Fußabdruck der
//    Anwendung und der ihrer Webprozesse.
// 7. **Dunkel**.
//
// # Was es nicht belegt
//
// **Den Dienst, der endet** — dafür braucht es einen geladenen LaunchAgent,
// und den gibt es nur im Gast (`betriebsband09gast`). **VoiceOver** — die
// Sätze sind gelesen, gesprochen hat sie hier niemand.
import { execFileSync } from 'node:child_process'
import { closeSync, existsSync, ftruncateSync, mkdirSync, openSync, readFileSync, rmSync, statSync, writeSync } from 'node:fs'
import { join } from 'node:path'

const bereich = "[...document.querySelectorAll('section')].find((s) => s.getAttribute('aria-label') === 'Was Aufmerksamkeit braucht')"
const bereichstext = (p) => p.seite(`return (${bereich})?.innerText ?? ''`)
const hinweis = (anfang) => `[...(${bereich})?.querySelectorAll('button, p') ?? []].find((e) => e.checkVisibility() && e.innerText.trim().startsWith(${JSON.stringify(anfang)}))`

// webinhalte sind die WebContent-Prozesse dieses Rechners.
function webinhalte() {
  return execFileSync('ps', ['-axo', 'pid=,comm=']).toString().split('\n')
    .map((zeile) => zeile.trim().split(/\s+/, 2))
    .filter(([, name]) => name?.includes('com.apple.WebKit.WebContent'))
    .map(([pid]) => Number(pid))
}

async function zurStartseite(p) {
  await p.taste('1', 'befehl')
  await p.warte('»Was Aufmerksamkeit braucht« steht da', () => p.seite(`return !!(${bereich})`), 20000)
}

// steht wartet, bis der Bereich den Text zeigt — **das Fenster fragt alle 30
// Sekunden**, die Frist ist eine Runde und etwas mehr.
async function steht(p, beschreibung, text, frist = 40000) {
  return p.warte(beschreibung, async () => {
    const inhalt = await bereichstext(p)

    return inhalt.includes(text) ? inhalt : false
  }, frist, 500)
}

async function fehlt(p, beschreibung, text, frist = 40000) {
  return p.warte(beschreibung, async () => {
    const inhalt = await bereichstext(p)

    return inhalt.includes(text) ? false : inhalt || ' '
  }, frist, 500)
}

// freiAuf ist der freie Platz des Abbilds, wie `df` ihn nennt.
const freiAuf = (halt) => Number(execFileSync('df', ['-k', halt]).toString().trim().split('\n').at(-1).split(/\s+/)[3]) * 1024

// fuellenBis schreibt die Füllung so lang, dass `frei` Bytes übrig sind.
function fuellenBis(halt, frei) {
  const jetzt = freiAuf(halt)
  const datei = join(halt, 'fuellung')
  const da = existsSync(datei) ? statSync(datei).size : 0
  const ziel = Math.max(0, da + jetzt - frei)
  const griff = openSync(datei, 'a')
  try {
    ftruncateSync(griff, 0)
    const stueck = Buffer.alloc(64 << 20, 1)
    for (let geschrieben = 0; geschrieben < ziel; geschrieben += stueck.length) {
      writeSync(griff, stueck, 0, Math.min(stueck.length, ziel - geschrieben))
    }
  } finally {
    closeSync(griff)
  }
  execFileSync('sync')

  return freiAuf(halt)
}

const gib = 2 ** 30
const gb = (bytes) => `${(bytes / gib).toFixed(2)} GiB`

export default async function betriebsband09(p) {
  const grund = p.ort({ zuhause: true })
  const abbild = join(grund.ort, 'band.sparseimage')
  const halt = join(grund.ort, 'platte')
  mkdirSync(halt)
  execFileSync('hdiutil', ['create', '-size', '3g', '-type', 'SPARSE', '-fs', 'HFS+', '-volname', 'mse-band', join(grund.ort, 'band')], { stdio: 'ignore' })
  execFileSync('hdiutil', ['attach', '-nobrowse', '-mountpoint', halt, abbild], { stdio: 'ignore' })
  const ort = { ...grund, bestand: join(halt, 'bestand.db') }
  try {
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    p.zeige(`frei auf dem Abbild: ${gb(freiAuf(halt))}`)

    p.notiz('1 — vorher: nichts meldet sich')
    await p.starte({ ort })
    await zurStartseite(p)
    const vorher = await steht(p, 'der Bereich ist geladen', 'durchgesehen', 20000)
    p.pruefe('ohne Lage kein Hinweis — »Nichts«', vorher.includes('Nichts') && !vorher.includes('Letzte Sicherung vor'), vorher.replace(/\n/g, ' · '))
    await p.bild('vorher')
    p.pruefe('⌘Q beendet', await p.beende())

    p.notiz('2 — eine Sicherung von vor zwanzig Tagen, und mit der Tastatur zum Ziel')
    const damals = new Date(Date.now() - 20 * 24 * 3600 * 1000).toISOString().replace(/\.\d{3}Z$/, '.000Z')
    execFileSync('sqlite3', [ort.bestand, `INSERT INTO sicherungsvermerke (gesichert, anlass, flaeche, datei, groesse, pruefsumme) VALUES ('${damals}', 'hand', 'cli', 'damals.msesicherung', 1, '${'0'.repeat(64)}');`])
    const vorStart = new Set(webinhalte())
    await p.starte({ ort })
    await zurStartseite(p)
    const alt = await steht(p, 'der Hinweis auf die alte Sicherung steht da', 'Letzte Sicherung vor 20 Tagen')
    p.pruefe('mit Rat und Ziel', alt.includes('Älter als 14 Tage') && alt.includes('Daten und Sicherung'), alt.replace(/\n/g, ' · '))
    await p.bild('sicherung-alt')
    await p.seite(`(${hinweis('Letzte Sicherung vor 20 Tagen')}).focus(); return true`)
    await p.taste('Enter')
    await p.wartetext('Letzte Sicherung', 20000)
    const gruppe = await p.seite("return [...document.querySelectorAll('section')].some((s) => s.checkVisibility() && s.querySelector('h3')?.innerText.trim() === 'Letzte Sicherung')")
    p.pruefe('die Eingabetaste führt zu »Daten und Sicherung«', gruppe === true)
    await p.bild('ziel-daten')

    p.notiz('3 — der Platz, an zwei Schwellen')
    await zurStartseite(p)
    const knapp = fuellenBis(halt, gib + gib / 2)
    p.zeige(`frei: ${gb(knapp)}`)
    const eng = await steht(p, 'unter 2 GiB kommt der Hinweis', 'frei neben dem Bestand')
    p.pruefe('er nennt den Platz und führt zu »Daten und Sicherung«', /Nur noch 1,[0-9] GB frei/.test(eng), eng.replace(/\n/g, ' · '))
    await p.bild('platz-knapp')
    const dazwischen = fuellenBis(halt, 2 * gib + gib / 4)
    p.zeige(`frei: ${gb(dazwischen)} — zwischen den Schwellen`)
    await new Promise((weiter) => setTimeout(weiter, 35000))
    const bleibt = await bereichstext(p)
    p.pruefe('zwischen 2 und 2,5 GiB bleibt der Hinweis stehen — einmal, nicht neu', (bleibt.match(/frei neben dem Bestand/g) ?? []).length === 1, bleibt.replace(/\n/g, ' · '))
    rmSync(join(halt, 'fuellung'))
    execFileSync('sync')
    p.zeige(`frei: ${gb(freiAuf(halt))}`)
    await fehlt(p, 'über 2,5 GiB geht der Hinweis', 'frei neben dem Bestand')

    p.notiz('4 — der Webprozess endet, Wails lädt neu')
    const eigene = webinhalte().filter((pid) => !vorStart.has(pid))
    p.zeige(`WebContent dieses Fensters: ${eigene.join(', ')}`)
    p.pruefe('genau ein neuer WebContent-Prozess', eigene.length === 1, eigene.join(', '))
    process.kill(eigene[0], 'SIGKILL')
    await steht(p, 'der Hinweis nennt das Neuladen', 'Die Webansicht wurde 1-mal neu geladen', 30000)
    // **Nach dem Neuladen steht die Einführung wieder da** — gesehen in
    // diesem Drehbuch am 2026-10-15: „in dieser Sitzung geschlossen" lebt im
    // Webprozess und geht mit ihm. Festgehalten und geschlossen wie beim
    // Start, der Befund steht als eigenes Issue.
    const einfuehrung = "return !!document.querySelector('[data-einfuehrung]')?.checkVisibility()"
    const wiederDa = await p.seite(einfuehrung)
    p.zeige(`die Einführung nach dem Neuladen: ${wiederDa ? 'wieder da' : 'zu'}`)
    if (wiederDa) {
      await p.taste('Escape')
      await p.warte('die Einführung ist zu', async () => !(await p.seite(einfuehrung)), 10000)
    }
    const protokoll = join(halt, 'mse.log')
    p.pruefe('das Protokoll neben dem Bestand nennt es', existsSync(protokoll) && readFileSync(protokoll, 'utf8').includes('der Webprozess endete und wurde neu geladen'))
    await p.bild('webprozess-neu')

    p.notiz('5 — was ein Abruf kostet: derselbe Ruf, hundertmal über die Naht')
    await p.seite(`
      const echt = window.fetch.bind(window)
      window.__betriebsband = { ruf: undefined }
      window.fetch = (adresse, angaben) => {
        if (window.__betriebsband.ruf === undefined && typeof angaben?.body === 'string' && angaben.body.includes('Betriebsband')) {
          window.__betriebsband.ruf = { adresse: String(adresse), angaben }
        }
        return echt(adresse, angaben)
      }
      window.__betriebsband.echt = echt
      return true`)
    await p.warte('das Fenster fragt von selbst', () => p.seite('return !!window.__betriebsband?.ruf'), 40000, 500)
    const messung = await p.seite(`
      const { ruf, echt } = window.__betriebsband
      const dauer = []
      const aufwand = []
      for (let i = 0; i < 100; i++) {
        const koerper = JSON.parse(ruf.angaben.body)
        koerper['call-id'] = 'messung-' + String(i)
        const beginn = performance.now()
        const antwort = await echt(ruf.adresse, { ...ruf.angaben, body: JSON.stringify(koerper) })
        const inhalt = await antwort.json()
        dauer.push(performance.now() - beginn)
        aufwand.push(inhalt.aufwandMikros / 1000)
      }
      const stelle = (liste, q) => [...liste].sort((a, b) => a - b)[Math.ceil(q * liste.length) - 1]
      return { p50: stelle(dauer, 0.5), p95: stelle(dauer, 0.95), max: Math.max(...dauer), aufwand95: stelle(aufwand, 0.95) }`)
    p.zeige(`über die Naht: p50 ${messung.p50.toFixed(2)} ms, p95 ${messung.p95.toFixed(2)} ms, höchstens ${messung.max.toFixed(2)} ms; in Go p95 ${messung.aufwand95.toFixed(2)} ms`)
    p.pruefe('p95 unter 20 ms', messung.p95 < 20, JSON.stringify(messung))

    p.notiz('6 — der Hinweis führt zur Diagnose, dort der Speicher, getrennt')
    await p.seite(`(${hinweis('Die Webansicht wurde')}).scrollIntoView({ block: 'center' }); return true`)
    await p.klickeAuf(hinweis('Die Webansicht wurde'), 'den Hinweis auf den Webprozess')
    await p.wartetext('Laufzeiten und Verbrauch', 20000)
    const zahlen = await p.warte('der Speicher ist gemessen', async () => {
      const text = await p.seite("return [...document.querySelectorAll('dl')].map((d) => d.innerText).join(' · ')")

      return /Fußabdruck/.test(text) && /Webinhalt/.test(text) ? text : false
    }, 20000)
    p.pruefe('Anwendung und Webprozess stehen getrennt', /MB Fußabdruck — davon .* Go-Heap/.test(zahlen) && /MB zusammen — .*Webinhalt/.test(zahlen), zahlen.replace(/\n/g, ' · '))
    await p.seite("[...document.querySelectorAll('h3')].find((h) => h.innerText.trim() === 'Laufzeiten und Verbrauch')?.scrollIntoView({ block: 'start' }); return true")
    await p.bild('speicher')
    p.pruefe('⌘Q beendet', await p.beende())

    p.notiz('7 — dunkel')
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'dunkel')
    await p.starte({ ort })
    await zurStartseite(p)
    await steht(p, 'der Hinweis steht auch dunkel da', 'Letzte Sicherung vor 20 Tagen')
    p.pruefe('Mantine steht dunkel', (await p.seite('return document.documentElement.dataset.mantineColorScheme')) === 'dark')
    await p.bild('dunkel')
    p.pruefe('⌘Q beendet', await p.beende())
  } finally {
    try {
      execFileSync('hdiutil', ['detach', halt], { stdio: 'ignore' })
    } catch {
      execFileSync('hdiutil', ['detach', '-force', halt], { stdio: 'ignore' })
    }
  }
}
