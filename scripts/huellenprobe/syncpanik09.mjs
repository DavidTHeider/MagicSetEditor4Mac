// **Eine Panik in einem synchronen Ruf auf dem Hauptfaden, am nativen
// Fenster** — #2398,
// Befund 3 des Zweitblicks über Phase 1 von 0.9.0 (F02).
//
// Die Panik kommt über die Naht des Abnahmebaus (`MSE_ABNAHME_SYNCPANIK`,
// `internal/fenster/probepanik_abnahme_darwin.go`): drei Sekunden nach dem
// Start wirft der Rückruf eines `application.InvokeSync`. Wails fängt sie in
// `handlePanic`. **Bis v3.0.0-beta.27 rief es `wg.Done()` danach nicht mehr**
// — wer wartete, wartete für immer, und deshalb endete die Anwendung. Seit
// beta.30 steht es im `defer` (`mainthread.go`), und es gilt, was F02 für jede
// andere gefangene Panik sagt: melden, nicht beenden.
//
// # Was es prüft
//
// 1. **Der Wartende kommt frei** — die Naht schreibt eine Zeile, sobald der
//    Ruf zurückkehrt.
// 2. **Die Anwendung läuft weiter und antwortet**, und das Protokoll nennt
//    die Panik mit Vorgangskennung.
// 3. **⌘Q beendet**, und der Folgestart nennt keine unerwartet beendete
//    Sitzung.
//
// # Was es nicht belegt
//
// **Eine Panik in Wails' eigenem Code** — eingespeist wird sie im Rückruf,
// den `InvokeSync` ausführt; der Weg danach ist derselbe. **VoiceOver** — das
// Band ist `role="status"`, ob es vorgelesen wird, hört nur ein Mensch.
import { existsSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

const band = "document.querySelector('[role=status][aria-label=\"Letzte Sitzung\"]')"

export default async function syncpanik09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const protokoll = join(dirname(ort.bestand), 'mse.log')
  const zeilen = () => (existsSync(protokoll) ? readFileSync(protokoll, 'utf8').split('\n') : [])

  p.notiz('1 — eine Panik im Rückruf eines InvokeSync, drei Sekunden nach dem Start')
  await p.starte({ ort, umgebung: { MSE_ABNAHME_SYNCPANIK: '3s' } })
  await p.warte('der synchrone Ruf kam nach der Panik zurück', () => zeilen().some((zeile) => zeile.includes('der synchrone Ruf kam nach der Panik zurück')), 20000)
  const zurueck = zeilen().find((zeile) => zeile.includes('der synchrone Ruf kam nach der Panik zurück')) ?? ''
  p.zeige(zurueck)
  p.pruefe('der Wartende kam nach der Panik frei', zurueck !== '', zurueck)

  p.notiz('2 — die Anwendung läuft weiter, und das Protokoll nennt die Panik')
  const panik = zeilen().find((zeile) => zeile.includes('Wails fing eine Panik')) ?? ''
  p.zeige(panik)
  p.pruefe('die Panik steht mit Vorgangskennung da', /kennung=\S+/.test(panik) && panik.includes('Probepanik im Hauptfaden'), panik)
  p.pruefe('kein Ende ist angesagt', !zeilen().some((zeile) => zeile.includes('die Anwendung endet')))
  p.pruefe('die Seite antwortet nach der Panik', (await p.seite('return document.readyState')) === 'complete')
  await p.bild('nach-der-panik')

  p.notiz('3 — ⌘Q beendet, und der Folgestart nennt keinen Abbruch')
  p.pruefe('⌘Q beendet', await p.beende())
  await p.starte({ ort })
  await p.warte('das Band hat seine Antwort', () => p.seite(`return ${band} === null || ${band}.getAttribute('aria-busy') === 'false'`), 20000)
  const hinweis = await p.seite(`return ${band}?.innerText ?? ''`)
  p.pruefe('kein Hinweis auf einen Abbruch', !hinweis.includes('Die letzte Sitzung, begonnen am'), hinweis)
  p.pruefe('⌘Q beendet', await p.beende())
}
