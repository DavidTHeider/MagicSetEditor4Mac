// **Der Hinweis nach einem Abbruch, am nativen Fenster** — F04
// (#2343),
// Befund A-04.
//
// Die Abnahme: ein künstlicher Absturz und ein Abschuss (SIGKILL) zeigen beim
// Folgestart einen Hinweis, ein sauberer Schluss zeigt keinen. Der Absturz
// kommt über die Naht des Abnahmebaus (`MSE_ABNAHME_ABSTURZ`,
// `internal/fenster/probepanik_abnahme_darwin.go`): eine Panik in einer
// Goroutine, die niemand fängt — der Prozess endet so, wie ein echter Fehler
// außerhalb jedes Riegels ihn enden ließe.
//
// # Was es prüft
//
// Sieben Starts im selben Bestand: sauber und noch einmal sauber (kein
// Hinweis), Absturz (Marke und Spur bleiben), Folgestart (Hinweis mit Spur,
// der Weg in die Diagnose über die Tastatur, das Diagnosepaket trägt die
// Spur mit Stapel und ohne den freien Text der Panik), sauber (kein Hinweis),
// Abschuss, Folgestart dunkel (Hinweis ohne Spur, Ausblenden über die
// Tastatur), sauber (kein Hinweis). Neben jeder Marke liegt die Sperre ihres
// Anspruchs; sie geht mit der Marke.
//
// # Was es nicht belegt
//
// VoiceOver — das Band ist `role="status"`, ob es vorgelesen wird, hört nur
// ein Mensch. Und einen Absturz in Wails' Objective-C-Teil: dort schreibt Go
// keine Spur, und der Hinweis käme ohne.
import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs'
import { dirname, join } from 'node:path'

const band = "document.querySelector('[role=status][aria-label=\"Letzte Sitzung\"]')"

export default async function wiederanlauf09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const sitzungen = join(dirname(ort.bestand), 'Sitzungen')
  const liste = (muster) => (existsSync(sitzungen) ? readdirSync(sitzungen) : []).filter((name) => muster.test(name))
  const spuren = () => liste(/^absturz-.*\.txt$/).filter((name) => statSync(join(sitzungen, name)).size > 0)

  p.notiz('Zweimal sauber: kein Hinweis')
  await p.starte({ ort })
  p.pruefe('während der Sitzung liegt eine Marke', liste(/^fenster-\d+\.json$/).length === 1, liste(/./).join(', '))
  p.pruefe('⌘Q beendet', await p.beende())
  p.pruefe('nach dem sauberen Schluss liegt keine Marke', liste(/^fenster-/).length === 0, liste(/./).join(', '))
  await p.starte({ ort })
  p.pruefe('nach einem sauberen Schluss: kein Hinweis', (await hinweis(p)) === '')
  await p.bild('f04-sauber-ohne-hinweis')
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Absturz: eine Panik, die niemand fängt')
  await p.starte({ ort, umgebung: { MSE_ABNAHME_ABSTURZ: '3s' } })
  const abgestuerzt = await Promise.race([p.beendet.then(() => true), schlafe(20000).then(() => false)])
  p.leitung.schliesse()
  p.pruefe('der Prozess endet von selbst', abgestuerzt)
  p.pruefe('die Marke bleibt liegen', liste(/^fenster-\d+\.json$/).length === 1, liste(/./).join(', '))
  const spur = spuren().map((name) => readFileSync(join(sitzungen, name), 'utf8')).join('\n')
  p.zeige(spur.split('\n').slice(0, 3).join(' ⏎ '))
  p.pruefe('die Absturzspur nennt die Panik', spur.includes('panic: Probeabsturz nach 3s'))

  p.notiz('Folgestart: Hinweis mit Spur, Weg in die Diagnose')
  await p.starte({ ort })
  const mitSpur = await hinweis(p)
  p.zeige(mitSpur.replace(/\s+/g, ' '))
  p.pruefe('der Hinweis nennt den Abbruch', mitSpur.includes('Die letzte Sitzung, begonnen am'))
  p.pruefe('der Hinweis sagt, dass eine Spur vorliegt', mitSpur.includes('Eine Absturzspur liegt vor'))
  p.pruefe('die gelesene Marke ist weg', liste(/^fenster-\d+\.json$/).length === 1, liste(/./).join(', '))
  // **Und ihre Sperre mit ihr** (Zweitblick über Phase 1, Befund 11): es
  // liegt nur noch die der laufenden Sitzung.
  p.pruefe('nur die laufende Sitzung hält eine Sperre', liste(/^fenster-\d+\.sperre$/).length === 1, liste(/./).join(', '))
  await p.bild('f04-hinweis-mit-spur-hell')
  await mitTaste(p, 'Diagnose öffnen')
  await p.warte('die Gruppe Diagnose ist offen', () => p.seite("return [...document.querySelectorAll('h3')].some((h) => h.innerText.trim() === 'Diagnosepaket')"), 20000)
  await p.bild('f04-diagnose-offen')
  const paket = join(ort.ort, 'f04.msediagnose')
  await knopfIn(p, 'Diagnosepaket', 'Zusammenstellen')
  await p.wartetext('Paket sichern', 20000)
  await p.dialogantwort(paket)
  await knopfIn(p, 'Diagnosepaket', 'Paket sichern')
  await p.warte('das Paket liegt da', () => Promise.resolve(existsSync(paket)), 20000)
  const inhalt = execFileSync('unzip', ['-p', paket], { encoding: 'utf8', maxBuffer: 64 << 20 })
  // **Mit Stapel, ohne den freien Text der Panik** (Zweitblick über Phase 1,
  // Befund 7): was der Code mit `panic` warf, bleibt in der Spur auf dem
  // Rechner; ins Paket geht, was das Laufzeitsystem schreibt.
  p.pruefe('das Diagnosepaket trägt die Absturzspur', inhalt.includes('--- Absturzspur, geschrieben ') && inhalt.includes('goroutine '))
  p.pruefe('der freie Text der Panik bleibt draußen', !inhalt.includes('Probeabsturz nach 3s') && inhalt.includes('panic: «geschwärzt»'))
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Wieder sauber: kein Hinweis mehr')
  await p.starte({ ort })
  p.pruefe('ein gelesener Abbruch kommt nicht wieder', (await hinweis(p)) === '')

  p.notiz('Abschuss: SIGKILL')
  const pid = p.pids.at(-1)
  process.kill(pid, 'SIGKILL')
  const abgeschossen = await Promise.race([p.beendet.then(() => true), schlafe(15000).then(() => false)])
  p.leitung.schliesse()
  p.pruefe('der Prozess ist weg', abgeschossen)
  p.pruefe('die Marke bleibt liegen', liste(/^fenster-\d+\.json$/).length === 1, liste(/./).join(', '))

  p.notiz('Folgestart dunkel: Hinweis ohne Spur')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'dunkel')
  await p.starte({ ort })
  p.pruefe('das Fenster ist dunkel', (await p.seite('return document.documentElement.dataset.erscheinungsbild')) === 'dunkel')
  const ohneSpur = await hinweis(p)
  p.zeige(ohneSpur.replace(/\s+/g, ' '))
  p.pruefe('der Hinweis nennt den Abbruch', ohneSpur.includes('Die letzte Sitzung, begonnen am'))
  p.pruefe('der Hinweis sagt, dass es keine Spur gibt', ohneSpur.includes('Eine Absturzspur gibt es nicht'))
  await p.bild('f04-hinweis-ohne-spur-dunkel')
  await mitTaste(p, 'Ausblenden')
  await p.warte('der Hinweis ist weg', async () => (await p.seite(`return ${band}?.innerText ?? ''`)) === '')
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Zum Schluss sauber: kein Hinweis')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  await p.starte({ ort })
  p.pruefe('nach dem Ausblenden und einem sauberen Schluss: kein Hinweis', (await hinweis(p)) === '')
  p.pruefe('⌘Q beendet', await p.beende())
  p.pruefe('höchstens eine Spur mit Inhalt liegt', spuren().length === 1, spuren().join(', '))
}

// hinweis wartet, bis das Band seine Antwort hat, und gibt seinen Text —
// **die Wirkung und nicht den Ruf** (§3): vorher ist es auch dann leer, wenn
// es gleich etwas zeigt.
async function hinweis(p) {
  await p.warte('das Band hat seine Antwort', () => p.seite(`return ${band}?.getAttribute('aria-busy') === 'false'`), 20000)

  return p.seite(`return ${band}.innerText`)
}

// mitTaste legt den Fokus auf den Knopf im Band und drückt ↩ — der Weg über
// die Tastatur, und **kein nativer Klick**, der nach einem Start ein `DIV`
// treffen kann (F01, 2026-10-14).
async function mitTaste(p, text) {
  const gefunden = await p.seite(`const k = [...${band}.querySelectorAll('button')].find((b) => b.innerText.trim() === ${JSON.stringify(text)}); k?.focus(); return document.activeElement === k`)
  p.pruefe(`»${text}« nimmt den Fokus`, gefunden)
  await p.taste('Enter')
}

// knopfIn drückt einen Knopf im Abschnitt mit der genannten Überschrift —
// **über `includes`**: ein `Aktionsknopf` trägt sein Zeichen vor dem Text.
async function knopfIn(p, ueberschrift, text) {
  const gefunden = await p.seite(`const h = [...document.querySelectorAll('h3')].find((e) => e.innerText.trim() === ${JSON.stringify(ueberschrift)}); const k = [...(h?.parentElement?.querySelectorAll('button') ?? [])].find((b) => b.innerText.includes(${JSON.stringify(text)})); k?.focus(); return document.activeElement === k`)
  p.pruefe(`»${text}« im Abschnitt »${ueberschrift}« nimmt den Fokus`, gefunden)
  await p.taste('Enter')
}

function schlafe(ms) {
  return new Promise((los) => setTimeout(los, ms))
}
