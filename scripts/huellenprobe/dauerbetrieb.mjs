// **Zwei Stunden gemischter nativer Betrieb** — B04
// (#2223).
//
// # Der Vertrag, vor dem Lauf festgelegt
//
// - **Eine** Sitzung, ein Prozess, **mindestens 120 Minuten** (`MSE_DAUER_MIN`
//   verkürzt sie für eine Probe — dann ist der Beleg keiner für B04).
// - Reihum: **Bearbeiten** (ein Kartenfeld über die Tastatur), **Navigation**
//   (sechs Seiten über die Befehlspalette), **Sammlung** (ein Exemplar über
//   die Bindung anlegen und wieder löschen), **Tisch** (die Partie am Tisch
//   öffnen, einen Münzwurf befehlen, den Tisch schließen — Pixi wird je Runde
//   auf- und abgebaut), **Ausgabe** (ein Kartenbild über `BildSchreiben` in
//   ein Verzeichnis der Probe — **ohne** Dateidialog; der gehört dem System).
//   **Eine Karte ziehen tut die Runde nicht**: `karte.ziehen` ist kein Befehl
//   des Fensters, und ein Zug mit dem Zeiger steht in `a04`.
// - **Alle 30 Sekunden** eine Probe, **zwischen zwei Runden** (#2527) — bis
//   zum 2026-10-20 nahm sie eine Uhr, und weil eine Runde rund fünf Sekunden
//   dauert, traf sie über lange Strecken denselben Teil: am Tisch lag der
//   Median 217 MiB über dem beim Bearbeiten, und ein Abschnitt verglich
//   Teile statt Zeiten. `bewertung` erklärt Proben aus verschiedenen Teilen
//   seither für ungültig. Gemessen wird der **physische Fußabdruck**
//   (`footprint`) des Go-Prozesses und der WebKit-Prozesse, die nach dem
//   Start entstanden, dazu RSS und CPU, die Kennungen und die Aufgabe, die
//   gerade läuft. **Bewertet wird der Fußabdruck** (E06, #2278): RSS zählt
//   unter macOS Seiten mit, die Go mit `MADV_FREE_REUSABLE` schon
//   zurückgegeben hat, und den Grafikspeicher des Webinhalts nicht — die
//   Anwendung lag im Diagnoselauf vom 2026-10-12 bei 87–89 MB Fußabdruck und
//   215–228 MB RSS. Der Fußabdruck ist das, wonach macOS den Speicherdruck
//   misst und was die Aktivitätsanzeige zeigt.
// - **Am Ende zwei Minuten Ruhe und danach eine Speicherwarnung an WebKit**
//   (`notifyutil -p org.WebKit.lowMemory`) — seit Vermerk 0204 Teil der
//   Regel; sie erreicht jede WebKit-Anwendung dieses Rechners, und
//   `MSE_DAUER_WARNUNG=0` lässt sie weg (dann ist der Beleg keiner für B04).
//   **Gesendet ist nicht empfangen**: die Probe sieht danach nur den
//   Fußabdruck, nicht, ob WebKit etwas freigegeben hat oder wovon.
// - **Bewertet**: kein Absturz (dieselbe Kennung am Ende), keine verlorene
//   Änderung (jede getippte Kennzeile steht danach im Bestand, gelesen über
//   `mse`), **nach der Aufwärmstunde** ein Wachstum von höchstens 10 % je
//   Stunde **und** kein Anstieg ab 1 % über drei aufeinanderfolgende
//   20-Minuten-Abschnitte (Vermerk 0204, bis zum 2026-10-20: 20 Minuten und
//   jeder Anstieg), und nach der Warnung liegt WebKit unter seinem Median im
//   ersten Abschnitt.
//
// **Was es nicht belegt**: einen Menschen, der zwei Stunden arbeitet — die
// Probe tippt schneller und liest nichts. Und **diese** Maschine.
import { execFileSync } from 'node:child_process'
import { join } from 'node:path'

import { abschnitte, bewertung, eindeutigEigen, freigegeben } from '../dauerbewertung.mjs'
import { MESSVERTRAG, tafel, urteil as vertragsurteil } from '../messvertrag.mjs'

// **Die Bewertung steht daneben** (`scripts/dauerbewertung.mjs`) und wird
// von hier weitergegeben — `dauerteile` und die Läufe der Bewertung lesen sie
// über dieses Drehbuch.
export { abschnitte, bewertung, eindeutigEigen, freigegeben }

const minuten = Number(process.env.MSE_DAUER_MIN ?? '125')

// **Der Go-Heap von außen** (E06, #2278): alle `MSE_DAUER_SPEICHER` Minuten
// ein Abzug über die Leitung — Stand nach einer Bereinigung und ein
// Heap-Profil im Bildordner des Laufs. **Über die Leitung und nicht aus der
// Seite**: die Heap-Probe aus der Seite wuchs selbst mit (LEISTUNGSSCHNITTE
// §9.1). Null heißt aus.
const speichertakt = Number(process.env.MSE_DAUER_SPEICHER ?? '0')
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))

// fussabdruckAus liest die Ausgabe von `footprint -f bytes --noCategories`:
// je Prozess eine Zeile mit `[Kennung]` und dem Fußabdruck in Bytes — hier
// in MB. Ein Prozess ohne Zeile fehlt, statt mit null zu zählen.
export function fussabdruckAus(ausgabe) {
  const aus = new Map()
  for (const zeile of ausgabe.split('\n')) {
    const fund = /\[(\d+)\]: .*Footprint: (\d+) B/.exec(zeile)
    if (fund) aus.set(fund[1], Math.round(Number(fund[2]) / 1048576))
  }

  return aus
}

export const fussabdruck = (kennungen) => {
  try {
    return fussabdruckAus(execFileSync('footprint', ['-f', 'bytes', '--noCategories', ...kennungen.map(String)], { stdio: ['ignore', 'pipe', 'ignore'] }).toString())
  } catch (fehler) {
    return fussabdruckAus(fehler.stdout?.toString() ?? '')
  }
}

// aufteilen trennt die WebKit-Prozesse in die der Anwendung — die mit dem
// Fenster entstanden — und fremde. **Fremde zählen nicht mit**: das System
// startete im Lauf vom 2026-10-12 etwa alle 35 Minuten einen
// `WebContent.EnhancedSecurity`, der die Anwendung überlebte; `ps` und
// `footprint` sagen nicht, wem ein WebKit-Prozess gehört.
export function aufteilen(webkit, eigene) {
  return {
    eigen: webkit.filter((x) => eigene.has(String(x.pid))),
    fremd: webkit.filter((x) => !eigene.has(String(x.pid))),
  }
}

const prozesse = () => execFileSync('ps', ['-Ao', 'pid=,rss=,%cpu=,comm=']).toString().split('\n')
  .map((z) => z.trim().split(/\s+/)).filter((t) => t.length >= 4)
  .map(([pid, rss, cpu, ...name]) => ({ pid, rss: Number(rss), cpu: Number(cpu), name: name.join(' ') }))


// zurSeite geht über die Befehlspalette — die Leiste führt Arbeitsbereiche.
async function zurSeite(p, schluessel) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht da', () => p.seite("return [...document.querySelectorAll('[data-palettenzeile]')].some((z) => z.checkVisibility())"), 15000)
  const zeile = `document.querySelector('[data-palettenzeile][data-seite=${JSON.stringify(schluessel)}]')`
  await p.seite(`(${zeile}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(zeile, `»${schluessel}« in der Palette`)
  await p.warte('die Palette ist zu', async () => !(await p.seite("return [...document.querySelectorAll('[data-palettenzeile]')].some((z) => z.checkVisibility())")), 10000)
}
const feld = "[...document.querySelectorAll('input, textarea, [contenteditable=true]')].find((e) => e.checkVisibility() && e.getAttribute('aria-label') === 'text')"
const knopf = (beschriftung) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && ((b.getAttribute('aria-label') ?? '') === ${JSON.stringify(beschriftung)} || b.innerText.trim() === ${JSON.stringify(beschriftung)}))`

// exemplar baut ein Exemplar wie das Fenster — die Druckkennung ist eine UUID
// (`domain.druckkennung`), jede Runde eine neue.
const exemplar = "({ id: 0, fassung: 0, printing: crypto.randomUUID(), druckID: '', vorlaeufig: false, menge: 1, zustand: 'nearMint', sprache: 'de', finish: 'nonfoil', erwerbsdatum: '2026-10-01', erwerbspreis: '1.00', erwerbswaehrung: 'EUR', notiz: '', name: 'Dauerkarte', setcode: 'dau', nummer: '1', lagerort: '', wunsch: false, bezeichnung: '' })"

// teileFuer baut die fünf Teile einer Runde. **Benannt**, damit `dauerteile`
// jeden für sich fahren kann (B04: welcher Teil wächst?).
// neuePartie legt eine Partie zwischen Ada und Bo an, beide mit Starthand
// und behalten, und gibt ihre Kennung.
function neuePartie(p, ort) {
  const vorher = new Set(JSON.parse(p.mse(ort, 'partien', '--json')).partien.map((x) => x.id))
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
  const partie = JSON.parse(p.mse(ort, 'partien', '--json')).partien.find((x) => !vorher.has(x.id)).id
  for (const sitz of ['ada', 'bo']) {
    p.mse(ort, 'partie-befehl', partie, sitz, `start-${sitz}`, 'spieler.starthand')
    p.mse(ort, 'partie-befehl', partie, sitz, `behalten-${sitz}`, 'spieler.behalten')
  }

  return partie
}

// textDerKarten liest den Text der Werkstattkarten über die Bindung.
const textDerKarten = "const B = window.go.app.Bindungen; const s = await B.Sets(); const set = s.sets.find((x) => x.name === 'Werkstatt-Testkarten'); const k = await B.Karten(set.id); return k.karten.map((c) => c.werte.map((w) => w.wert).join('\\n')).join('\\n')"

// enthaelt sagt, ob eine Marke als ganzes Wort im Text steht — `d1` steckt
// sonst auch in `d12`.
export const enthaelt = (text, marke) => new RegExp(`(^|\\W)${marke}(?!\\d)`).test(text)

// **Gleichbleibende Last** (E06, #2278): bis zum 2026-10-12 hängte jede Runde
// eine Marke an dasselbe Textfeld — nach 1.710 Runden rund 9.000 Zeichen, die
// jede Vorschau und jede Ausgabe neu setzten —, und jeder Münzwurf verlängerte
// das Protokoll derselben Partie. Ein Dauerlauf mit wachsender Last misst kein
// Leck. Jetzt **ersetzt** jede Runde das Feld, und was verloren ginge, sieht
// die nächste: sie liest die vorige Marke im Bestand. Die Partie wird alle
// `partienrunden` Runden durch eine frische ersetzt und die alte beendet.
const partienrunden = 50

function teileFuer(p, ort, halter, marken, fehler, verloren) {
  const ziele = ['start', 'decks', 'sammlung', 'sets', 'turniere', 'einstellungen']

  return {
    async bearbeiten(runde) {
      await p.taste('2', 'befehl')
      await p.wartetext('Werkstatt-Testkarten', 20000)
      await p.klickeText('Werkstatt-Testkarten')
      await p.warte('das Textfeld steht da', () => p.seite(`return !!(${feld})`), 20000)
      // **Erst wenn die Vorschau steht**: bis dahin lädt der Inspektor nach,
      // und unter Last rückte das Textfeld nach dem Messen noch — der Klick
      // traf am 2026-10-11 »Vorschlagen« unter dem Namen (E06, #2278).
      await p.warte('die Vorschau steht', () => p.seite("const b = document.querySelector('img[alt=\"Vorschau der Karte\"]'); return !!b && b.complete && b.naturalWidth > 0"), 20000)
      const vorige = marken.at(-1)
      if (vorige !== undefined) {
        const stand = await p.seite(textDerKarten)
        if (!enthaelt(stand, vorige)) verloren.push(`${vorige} (im Bestand: ${JSON.stringify(stand.slice(-24))})`)
      }
      await p.klickeAuf(feld, 'das Textfeld')
      // **Der Inhalt des Felds wird im Feld markiert**, nicht über ⌘A: das traf
      // im ersten Probelauf nicht immer das Feld, und die Marke hing dann ohne
      // Leerzeichen am alten Text.
      await p.seite(`const e = ${feld}; e.focus(); if (e.select) { e.select() } else { const r = document.createRange(); r.selectNodeContents(e); const s = getSelection(); s.removeAllRanges(); s.addRange(r) } return true`)
      const zeile = `d${String(runde)}`
      await p.tippe(zeile)
      marken.push(zeile)
    },
    async navigation() {
      for (const ziel of ziele) await zurSeite(p, ziel)
    },
    async sammlung(runde) {
      const antwort = await p.seite(`const B = window.go.app.Bindungen; const a = await B.ExemplarAnlegen(${exemplar}); if (a.fehler) return a.fehler.meldung; const e = (await B.Sammlung()).exemplare.find((x) => x.id === a.id); const l = await B.ExemplarLoeschen(e.id, e.fassung); return l?.meldung ?? 'ok'`)
      if (antwort !== 'ok') fehler.push(`Runde ${String(runde)}, Sammlung: ${String(antwort)}`)
    },
    async tisch(runde) {
      if (runde % partienrunden === 0) {
        const alt = halter.partie
        halter.partie = neuePartie(p, ort)
        // **Gelöscht, nicht beendet** (Nachlauf über die Folgen vor 0.8.0,
        // Befund 5): eine beendete Partie bleibt nachvollziehbar im Bestand
        // (`domain.Loeschbar`), und nach zwei Stunden las jede Runde 35 davon
        // mit. Eine laufende ohne Einwilligung darf gehen.
        p.mse(ort, 'partie-loeschen', alt)
      }
      const partie = halter.partie
      await p.taste('4', 'befehl')
      await p.wartetext('Partie anfangen', 20000)
      await p.klickeText('Partie anfangen')
      await p.warte('»Weiterspielen« steht', () => p.seite(`return !!${knopf('Weiterspielen: Ada · Bo')}`), 20000)
      await p.klickeAuf(knopf('Weiterspielen: Ada · Bo'), '»Weiterspielen: Ada · Bo«')
      await p.klickeText('Weiterspielen als Ada')
      await p.warte('der Tisch steht', () => p.seite("return !!document.querySelector('[data-zonenfeld=\"hand\"][data-sitz=\"ada\"]')"), 30000)
      // **Jede Leinwand schwach gemerkt** (B04): lebt sie nach dem Schließen
      // weiter, hält jemand sie fest — samt ihrem WebGL-Kontext.
      // **Nur die lebenden bleiben in der Liste** — sonst wüchse sie selbst.
      await p.seite("const c = document.querySelector('canvas'); window.__leinwaende = (window.__leinwaende ?? []).filter((r) => r.deref()); if (c) window.__leinwaende.push(new WeakRef(c)); return true")
      const wurf = await p.seite(`const a = await window.go.app.Bindungen.PartieBefehl(${JSON.stringify(partie)}, 'dauer-wurf-${String(runde)}', 'ada', 'spieler.muenzwurf', {}, ''); return a.fehler?.meldung ?? 'ok'`)
      if (wurf !== 'ok') fehler.push(`Runde ${String(runde)}, Münzwurf: ${String(wurf)}`)
      await p.klickeAuf(knopf('Partie schließen'), '»Partie schließen«')
      await p.wartetext('Eine Partie anfangen', 20000)
    },
    async ausgabe(runde) {
      const bild = await p.seite(`const s = await window.go.app.Bindungen.Sets(); const k = await window.go.app.Bindungen.Karten(s.sets[0].id); const a = await window.go.app.Bindungen.BildSchreiben(k.karten[0].id, ${JSON.stringify(join(ort.ort, 'ausgabe'))}, 'karte-' + ${String(runde)}, 'png', 0, 'ersetzen', ''); return a.fehler?.meldung ?? 'ok'`)
      if (bild !== 'ok') fehler.push(`Runde ${String(runde)}, Ausgabe: ${String(bild)}`)
    },
  }
}

// fahreDauer richtet die Probe ein, startet das Fenster und fährt Phasen —
// je Phase eine Liste von Teilen, reihum, bis ihre Minuten um sind. Proben
// zwischen zwei Runden, sobald `takt` Millisekunden um sind. Gibt zurück,
// was die Bewertung braucht.
//
// **`heap`** fragt nach jeder Runde den Go-Heap über die Bindung `Metriken`
// ab — damit ist zu unterscheiden, ob der Anwendungsprozess im Go-Heap
// wächst oder daneben (WebKit, Cocoa). Ein Ruf mehr je Runde; der
// Zwei-Stunden-Lauf fragt ihn deshalb nicht.
// **`warnung`** endet mit Ruhe und Speicherwarnung — der Zwei-Stunden-Lauf
// verlangt es (Vermerk 0204), `dauerteile` nicht.
export async function fahreDauer(p, phasen, takt = 30000, heap = false, warnung = false) {
  const vorher = new Set(prozesse().map((x) => x.pid))
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Dauerdeck', 'probe')
  for (const karte of ['1', '2', '3']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  const halter = { partie: neuePartie(p, ort) }
  // `MSE_DAUER_STAPEL`: `MallocStackLogging` nur für diesen Prozess (#2312).
  await p.starte({ ort, umgebung: process.env.MSE_DAUER_STAPEL ? { MallocStackLogging: process.env.MSE_DAUER_STAPEL } : {} })
  // **Jede geladene Adresse zählt** (B04) — ohne größeren Puffer verlöre der
  // Browser nach 250 Einträgen die ältesten.
  if (heap) await p.seite('performance.setResourceTimingBufferSize(1000000); return true')
  const kennung = p.pids.at(-1)
  const beginn = Date.now()
  let aufgabe = 'Start'
  const proben = []
  const haufen = []
  const marken = []
  const verloren = []
  // **Was eine Runde fachlich ablehnt, zählt** — bis zum ersten Probelauf
  // verschluckte die Runde jede Antwort, auch einen Fehler.
  const fehler = []
  const teile = teileFuer(p, ort, halter, marken, fehler, verloren)
  const staende = []
  const abzug = async (anlass) => {
    const minute = (Date.now() - beginn) / 60000
    const datei = join(p.bilder, `heap-${String(Math.round(minute)).padStart(3, '0')}.pprof`)
    const antwort = await p.leitung.frage({ was: 'speicher', datei })
    staende.push({ minute, anlass, datei, ...(antwort.ok ? antwort.daten : { fehler: antwort.fehler }) })
  }
  let naechsterAbzug = speichertakt > 0 ? beginn + 3 * 60000 : Infinity

  // eigene sind die WebKit-Prozesse, die mit dem Fenster entstanden — die
  // erste Probe legt sie fest.
  let eigene
  const probe = () => {
    const alle = prozesse()
    const app = alle.find((x) => String(x.pid) === String(kennung))
    const neu = alle.filter((x) => !vorher.has(x.pid) && x.name.includes('WebKit.'))
    eigene ??= new Set(neu.map((x) => String(x.pid)))
    const { eigen: web, fremd } = aufteilen(neu, eigene)
    const fuss = fussabdruck([...(app ? [app.pid] : []), ...web.map((x) => x.pid)])
    const webFuss = web.reduce((s, x) => s + (fuss.get(String(x.pid)) ?? 0), 0)
    proben.push({
      minute: (Date.now() - beginn) / 60000, aufgabe,
      app: app ? fuss.get(String(app.pid)) ?? null : null, appCpu: app?.cpu ?? null,
      web: webFuss, webPids: web.map((x) => x.pid), webArten: web.map((x) => x.name.split('.').at(-1)),
      fremdePids: fremd.map((x) => x.pid),
      webJe: Object.fromEntries(web.map((x) => [x.name.split('.').at(-1), fuss.get(String(x.pid)) ?? null])),
      // **Fehlt ein Wert, fehlt die Summe** — `bewertung` erklärt das Urteil
      // dann für ungültig, statt eine Null mitzuzählen.
      gesamt: app && [app, ...web].every((x) => fuss.has(String(x.pid))) ? fuss.get(String(app.pid)) + webFuss : null,
      rss: Math.round(((app?.rss ?? 0) + web.reduce((s, x) => s + x.rss, 0)) / 1024),
    })
  }
  probe()
  let naechsteProbe = beginn + takt

  let runde = 0
  try {
    for (const phase of phasen) {
      const ende = Date.now() + phase.minuten * 60000
      while (Date.now() < ende) {
        runde++
        for (const teil of phase.teile) {
          aufgabe = `${phase.name} — Runde ${String(runde)}: ${teil}`
          await teile[teil](runde)
        }
        if (heap) {
          const m = await p.seite('const a = await window.go.app.Bindungen.Metriken(); return a.prozess')
          // **Und was im Dokument steht** — Leinwände, Bilder, Knoten. Eine
          // Fläche je Runde, die nicht freikommt, steht hier als Zahl.
          const dom = await p.seite("const w = (window.__leinwaende ?? []).map((r) => r.deref()).filter(Boolean); const adressen = (teil) => new Set(performance.getEntriesByType('resource').map((e) => e.name).filter((n) => n.includes(teil))).size; return { spielbilder: adressen('/spielbild/'), vorschauen: adressen('/vorschau/'), leinwaende: document.querySelectorAll('canvas').length, bilder: document.querySelectorAll('img').length, knoten: document.getElementsByTagName('*').length, lebend: w.length, mitKontext: w.filter((c) => { const g = c.getContext('webgl2') ?? c.getContext('webgl'); return g && !g.isContextLost() }).length }")
          haufen.push({ minute: (Date.now() - beginn) / 60000, phase: phase.name, heap: Math.round(m.heapBytes / 1048576), sys: Math.round(m.systemBytes / 1048576), goroutinen: m.goroutinen, ...dom })
        }
        if (Date.now() >= naechsterAbzug) {
          await abzug(aufgabe)
          naechsterAbzug += speichertakt * 60000
        }
        await schlafe(2000)
        if (Date.now() >= naechsteProbe) {
          aufgabe = `${phase.name} — Runde ${String(runde)}: zwischen den Runden`
          probe()
          while (naechsteProbe <= Date.now()) naechsteProbe += takt
        }
      }
    }
    if (speichertakt > 0) await abzug('Ende')
    if (warnung) {
      aufgabe = 'Ruhe'
      await schlafe(120000)
      probe()
      execFileSync('notifyutil', ['-p', 'org.WebKit.lowMemory'])
      aufgabe = 'nach der Speicherwarnung'
      await schlafe(30000)
      probe()
    }
  } finally {
    aufgabe = 'Ende'
    probe()
  }

  return { ort, kennung, runde, proben, haufen, marken, verloren, fehler, staende }
}

export default async function dauerbetrieb(p) {
  const { ort, kennung, runde, proben, marken, verloren, fehler, staende } = await fahreDauer(p, [
    { name: 'Dauer', minuten, teile: ['bearbeiten', 'navigation', 'sammlung', 'tisch', 'ausgabe'] },
  ], 30000, false, process.env.MSE_DAUER_WARNUNG !== '0')

  p.notiz('Ergebnis')
  const lebt = prozesse().some((x) => String(x.pid) === String(kennung))
  p.pruefe(`dieselbe Kennung nach ${minuten.toFixed(0)} Minuten — kein Absturz`, lebt, `Kennung ${String(kennung)}, ${String(runde)} Runden`)
  // **Die Werte stehen nur unter `--json`** — ohne nennt `karten` Kennung und
  // Namen, und der erste Probelauf meldete jede Marke als verloren. Jede
  // Runde prüfte die vorige; hier steht die letzte.
  const text = JSON.parse(p.mse(ort, 'karten', '1', '--json')).karten.map((k) => Object.values(k.werte).join('\n')).join('\n')
  if (marken.length > 0 && !enthaelt(text, marken.at(-1))) verloren.push(marken.at(-1))
  p.pruefe('jede Runde gelang fachlich — Sammlung, Münzwurf, Ausgabe', fehler.length === 0, `${String(fehler.length)}: ${fehler.slice(0, 5).join('; ')}`)
  p.pruefe('keine getippte Änderung ging verloren', verloren.length === 0, `${String(marken.length)} getippt, verloren: ${verloren.slice(0, 5).join(', ')}`)
  const urteil = bewertung(proben)
  p.pruefe('Speicher nach der Aufwärmstunde: höchstens 10 % je Stunde und kein Anstieg ab 1 % über drei Abschnitte',
    urteil.gueltig && urteil.haelt, JSON.stringify(urteil))
  const frei = freigegeben(proben)
  p.pruefe('nach Ruhe und Speicherwarnung liegt WebKit unter seinem ersten Abschnitt — was wuchs, ist freigebbar',
    frei.gueltig && frei.haelt, JSON.stringify(frei))
  // **PERF01: der Fußabdruck nach 120 Minuten** (#2351) — der höchste Wert
  // von Anwendung und Webansicht in den letzten zehn Minuten der Arbeit,
  // nicht nach der Ruhe; ein kürzerer Lauf misst ihn nicht.
  const vertrag = MESSVERTRAG.find((e) => e.beleg === 'huellenprobe/dauerbetrieb.md')
  const spaet = minuten >= 120 ? proben.filter((x) => x.minute >= minuten - 10 && x.minute <= minuten && x.gesamt !== null && !/Ruhe|Speicherwarnung/.test(x.aufgabe)).map((x) => x.gesamt) : []
  p.zeilen.push('', '| Reihe | n | min | Median | P95 | max | Grenze | Urteil |', '| --- | --- | --- | --- | --- | --- | --- | --- |', tafel(vertrag, spaet))
  p.pruefe(`${vertrag.reihe}: ${vertrag.mass} ≤ ${String(vertrag.grenze)}`, vertragsurteil(vertrag, spaet) === 'gehalten', vertragsurteil(vertrag, spaet))
  p.zeilen.push('', `Proben (${String(proben.length)}, alle 30 s): ${JSON.stringify(proben)}`)
  if (staende.length > 0) {
    const mb = (b) => Math.round(b / 1048576)
    const erster = staende[0]
    const letzter = staende.at(-1)
    p.zeilen.push('', `Go-Heap nach Bereinigung (${String(staende.length)} Abzüge, Profile im Bildordner): HeapAlloc ${String(mb(erster.heapAlloc ?? 0))} → ${String(mb(letzter.heapAlloc ?? 0))} MB, ungenutzt und nicht zurückgegeben ${String(mb((erster.heapIdle ?? 0) - (erster.heapReleased ?? 0)))} → ${String(mb((letzter.heapIdle ?? 0) - (letzter.heapReleased ?? 0)))} MB — ${JSON.stringify(staende)}`)
  }
  p.pruefe('⌘Q beendet', await p.beende())
}
