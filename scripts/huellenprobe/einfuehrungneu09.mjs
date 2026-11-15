// **Die geschlossene Einführung bleibt zu, wenn der Webprozess neu lädt** —
// #2428,
// gefunden in der Hüllenprobe `betriebsband09` (OPS02, #2348).
//
// # Was hier geschieht
//
// 1. **Die Einführung steht beim Start** und wird mit Escape geschlossen.
// 2. **Der WebContent-Prozess dieses Fensters wird beendet**; Wails lädt die
//    Seite neu (`webViewWebContentProcessDidTerminate:`).
// 3. **Die Einführung bleibt zu** — und die Startseite steht.
// 4. **Ein neuer Start** zeigt sie wieder: „in dieser Sitzung geschlossen“
//    gilt für die Sitzung, nicht für immer.
//
// # Vorher
//
// `betriebsband09` am Commit `5cdf6bc0f`: „die Einführung nach dem Neuladen:
// wieder da“.
import { execFileSync } from 'node:child_process'

const einfuehrung = "return !!document.querySelector('[data-einfuehrung]')?.checkVisibility()"

function webinhalte() {
  return execFileSync('ps', ['-axo', 'pid=,comm=']).toString().split('\n')
    .map((zeile) => zeile.trim().split(/\s+/, 2))
    .filter(([, name]) => name?.includes('com.apple.WebKit.WebContent'))
    .map(([pid]) => Number(pid))
}

export default async function einfuehrungneu09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  p.notiz('1 — die Einführung steht und wird geschlossen')
  const vorStart = new Set(webinhalte())
  await p.starte({ ort, einfuehrung: true })
  await p.warte('die Einführung steht', () => p.seite(einfuehrung), 20000)
  await p.taste('Escape')
  await p.warte('die Einführung ist zu', async () => !(await p.seite(einfuehrung)), 10000)

  p.notiz('2 — der Webprozess endet, Wails lädt neu')
  const eigene = webinhalte().filter((pid) => !vorStart.has(pid))
  p.pruefe('genau ein neuer WebContent-Prozess', eigene.length === 1, eigene.join(', '))
  await p.seite('window.__vorDemNeuladen = true; return true')
  process.kill(eigene[0], 'SIGKILL')
  await p.warte('die Seite ist neu geladen', async () => {
    try {
      return await p.seite("return window.__vorDemNeuladen === undefined && !!document.querySelector('main')")
    } catch {
      return false
    }
  }, 30000, 500)

  p.notiz('3 — die Einführung bleibt zu')
  await new Promise((weiter) => setTimeout(weiter, 3000))
  p.pruefe('nach dem Neuladen steht die Einführung nicht wieder da', !(await p.seite(einfuehrung)))
  await p.bild('nach-dem-neuladen')
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('4 — ein neuer Start zeigt sie wieder')
  await p.starte({ ort, einfuehrung: true })
  await p.warte('die Einführung steht nach einem neuen Start', () => p.seite(einfuehrung), 20000)
  await p.bild('neuer-start')
  await p.taste('Escape')
  p.pruefe('⌘Q beendet', await p.beende())
}
