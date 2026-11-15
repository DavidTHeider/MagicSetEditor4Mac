// **Der Entwickler: Diagnose, API, Bau** — G01
// (#2248),
// am nativen Fenster und **in einem Zug**.
//
// Drei Wege, je mit dem Fehler, der einem Entwickler dort begegnet:
//
//   1. **Diagnose** — ein Diagnosepaket zusammenstellen und sichern;
//      vorher ein Ort ohne Schreibrecht,
//   2. **API** — den Dienst für Skripte aus dem Fenster starten und ihn
//      fragen; vorher eine Anfrage ohne Zugang,
//   3. **Bau** — welcher Commit läuft, aus der Hilfe gelesen und gegen den
//      Baum gehalten.
//
// **Was er nicht drückt**: »Alles einrichten« und die Ziele, die außerhalb
// des Testzuhauses schreiben (Vermerk 0159) — die Einrichtung fährt `g08`
// in einem eigenen Zuhause.
import { execFileSync } from 'node:child_process'
import { chmodSync, existsSync, mkdirSync } from 'node:fs'
import { request } from 'node:http'
import { join, resolve } from 'node:path'

import { zaehler } from '../huellenzaehler.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const meldungen = "return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean).join(' | ')"
const dienst = "[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && s.querySelector('h3')?.innerText === 'Die Schnittstelle für Skripte')"

// frage fragt den Dienst nach seinen Sets — mit oder ohne Zugang.
function frage(adresse, marke) {
  return new Promise((fertig) => {
    const ziel = new URL(adresse)
    const anfrage = request({
      host: ziel.hostname.replace(/^\[|\]$/g, ''), port: ziel.port, path: '/sets',
      headers: marke === '' ? {} : { authorization: `Bearer ${marke}` },
    }, (antwort) => { antwort.resume(); fertig(antwort.statusCode) })
    anfrage.on('error', (fehler) => { fertig(String(fehler)) })
    anfrage.end()
  })
}

async function abschnitt(p, z, tafel, { name, beginn, entscheidungen, fehler, ziel }, weg) {
  p.notiz(`${name} — Beginn: ${beginn}`)
  z.beginne()
  const erreicht = await weg()
  const { handlungen, dauer } = z.ende(name)
  p.pruefe(`${name}: Ziel erreicht — ${ziel}`, erreicht === true, typeof erreicht === 'string' ? erreicht : '')
  tafel.push({ name, beginn, entscheidungen, fehler, ziel, handlungen, dauer })
}

export default async function g01entwickler(p) {
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  const gesperrt = join(ort.ort, 'gesperrt')
  mkdirSync(gesperrt)
  chmodSync(gesperrt, 0o500)
  const paket = join(ort.ort, 'bericht.msediagnose')
  const tafel = []

  await p.starte({ ort })
  const z = zaehler(p)

  await abschnitt(p, z, tafel, {
    name: 'Diagnose',
    beginn: 'Einstellungen › Diagnose (⌘,)',
    entscheidungen: 'welche Teile ins Paket gehen; wohin',
    fehler: 'ein Ort ohne Schreibrecht',
    ziel: 'ein Diagnosepaket auf der Platte',
  }, async () => {
    await z.taste(',', 'befehl')
    await p.wartetext('Diagnose', 20000)
    await z.text('Diagnose')
    await p.wartetext('Zusammenstellen', 20000)
    await z.text('Zusammenstellen')
    await p.warte('die Teile stehen', () => p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim().endsWith('Paket sichern'))"), 20000)
    const sichern = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith('Paket sichern'))"
    await z.dialog(join(gesperrt, 'bericht.msediagnose'))
    await z.klickAuf(sichern, '»Paket sichern«')
    const verweigert = await p.warte('der gesperrte Ort wird abgewiesen', async () => (await p.seite(meldungen)) || false, 20000).catch(() => '(keine Meldung)')
    p.zeige(`gesperrter Ort: ${verweigert}`)
    await p.bild('g01d-gesperrt')
    await z.dialog(paket)
    await z.klickAuf(sichern, '»Paket sichern«')
    await p.warte('das Paket liegt da', () => Promise.resolve(existsSync(paket)), 20000)

    return existsSync(paket)
  })

  await abschnitt(p, z, tafel, {
    name: 'API',
    beginn: 'Einstellungen › Verbindungen',
    entscheidungen: 'wo der Dienst lauscht — hier die Vorgabe »Nur dieser Mac«',
    fehler: 'eine Anfrage ohne Zugang',
    ziel: 'der Dienst antwortet mit 200 auf /sets',
  }, async () => {
    await z.text('Verbindungen')
    await z.text('Was Fachleute hier sonst noch einstellen')
    await p.wartetext('Die Schnittstelle für Skripte', 20000)
    await z.klickAuf(`${dienst}?.querySelector('button:not([disabled])')`, 'Starten an der Skriptschnittstelle')
    const posten = await p.warte('der Dienst läuft', () => p.seite(`const t = (${dienst})?.innerText ?? ''; return /Zugang:\\s*[0-9a-f]{32,}/.test(t) ? t : null`), 30000)
    const adresse = [...posten.matchAll(/https?:\/\/[^\s,]+/g)].map((t) => t[0])[0] ?? ''
    const marke = posten.match(/Zugang:\s*([0-9a-f]{32,})/)?.[1] ?? ''
    p.zeige(`Dienst: ${adresse}`)
    await p.bild('g01d-dienst')
    const ohne = await frage(adresse, '')
    p.zeige(`ohne Zugang: ${String(ohne)}`)
    const mit = await frage(adresse, marke)
    p.zeige(`mit Zugang: ${String(mit)}`)
    p.pruefe('ohne Zugang weist er ab', ohne === 401, String(ohne))

    return mit === 200 ? true : `mit Zugang: ${String(mit)}`
  })

  await abschnitt(p, z, tafel, {
    name: 'Bau',
    beginn: 'das Menü Hilfe',
    entscheidungen: 'keine — gelesen wird',
    fehler: 'ein Bau mit ungesicherten Änderungen sagt es',
    ziel: 'die Hilfe nennt den Commit, aus dem das Bündel gebaut ist',
  }, async () => {
    await z.menue('Hilfe', 'MagicSetEditor4Mac-Hilfe')
    await p.wartetext('Betrieb und Vertrauen', 20000)
    await z.text('Betrieb und Vertrauen')
    const satz = await p.warte('die Fassung steht', () => p.seite("return document.body.innerText.split('\\n').find((z) => z.includes('gebaut aus')) ?? false"), 20000)
    const veraendert = await p.seite("return document.body.innerText.includes('Der Bau enthielt Änderungen, die nicht eingecheckt waren.')")
    p.zeige(`Fassung: ${satz}${veraendert ? ' — mit ungesicherten Änderungen' : ''}`)
    await p.bild('g01d-fassung')
    const commit = execFileSync('git', ['rev-parse', '--short=7', 'HEAD'], { cwd: wurzel, encoding: 'utf8' }).trim()

    return satz.includes(commit) ? true : `${satz} — erwartet ${commit}`
  })

  p.notiz('Die Tafel')
  for (const zeile of tafel) {
    p.zeige(`| ${zeile.name} | ${zeile.beginn} | ${zeile.entscheidungen} | ${zeile.fehler} | ${zeile.ziel} | ${String(zeile.handlungen)} | ${(zeile.dauer / 1000).toFixed(1).replace('.', ',')} s |`)
  }
  chmodSync(gesperrt, 0o700)
  p.pruefe('⌘Q beendet', await p.beende())
}
