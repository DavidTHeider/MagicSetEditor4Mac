// **SEC-17 am nativen Fenster: die WebView läuft unter einer CSP** —
// #2164.
//
// Drei Läufe über dasselbe Bündel:
//
// 1. **Beobachtend** (`MSE_CSP_BEOBACHTEN=1`, nur im Abnahmebau): die Policy
//    steht als `Content-Security-Policy-Report-Only` da. Ein künstliches
//    Inline-Skript **läuft**, und das Protokoll nennt, was die Policy
//    abgewiesen hätte — das ist das Inventar der Ressourcen.
// 2. **Durchgesetzt, hell**: dasselbe Skript, `eval` und ein fremdes Skript
//    werden abgewiesen und gemeldet; die Kartenseite zeigt ihr Bild, der
//    Tisch zeichnet seine Leinwand samt Karten.
// 3. **Durchgesetzt, dunkel**: dieselben Positivkontrollen.
//
// **Am Ende steht im Protokoll nichts außer den künstlichen Verstößen** — ein
// weiterer wäre eine Ressource, die die Policy der Anwendung wegnimmt.
//
// **Was sie nicht belegt**: Dateidialoge und Ablegen aus dem Finder (sie
// laufen nicht durch die WebView), den Servermodus (`make dev-browser` —
// er trägt die Policy nicht) und die Wege, die keine Seite dieser Probe
// betritt; die übrigen Drehbücher unter der durchgesetzten Policy sind der
// Funktionszugang.
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

import { zurSeite } from './uirevue.mjs'

// eindringen schleust dreierlei ein und sagt, was davon lief.
const eindringen = `
  const vorher = window.__sec17 ?? 0
  const verstoesse = []
  const horch = (e) => verstoesse.push(e.violatedDirective + ' ' + (e.blockedURI || 'inline'))
  document.addEventListener('securitypolicyviolation', horch)
  const eigen = document.createElement('script')
  eigen.textContent = 'window.__sec17 = (window.__sec17 ?? 0) + 1'
  document.head.appendChild(eigen)
  let ausgewertet = 'erlaubt'
  try { (0, eval)('1') } catch { ausgewertet = 'abgewiesen' }
  const fremd = document.createElement('script')
  fremd.src = 'https://sec17.invalid/fremd.js'
  document.head.appendChild(fremd)
  await new Promise((los) => setTimeout(los, 800))
  document.removeEventListener('securitypolicyviolation', horch)
  return { inline: (window.__sec17 ?? 0) > vorher ? 'lief' : 'abgewiesen', eval: ausgewertet, verstoesse }
`

// kuenstlich trifft genau die drei Verstöße, die dieses Drehbuch selbst
// auslöst — das Inline-Skript, `eval` und das fremde Skript; ein Stil
// `inline` ist keiner davon. `huellenlauf.mjs` streicht jeden anderen an.
const kuenstlich = /anweisung=script-src-elem abgewiesen=inline|anweisung=script-src abgewiesen=eval|abgewiesen=https:\/\/sec17\.invalid/

const meldungen = (p) => p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean)")

// cspZeilen gibt die Verstöße, die das Protokoll seit `ab` Zeichen nennt.
function cspZeilen(ort, ab = 0) {
  const protokoll = readFileSync(join(dirname(ort.bestand), 'mse.log'), 'utf8')

  return { zeilen: protokoll.slice(ab).split('\n').filter((z) => z.includes('CSP')), laenge: protokoll.length }
}

async function kartenbild(p, was) {
  await zurSeite(p, 'Karten')
  if (await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Navigator einblenden')")) {
    await p.klickeText('Navigator einblenden')
  }
  await p.klickeAuf("[...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.startsWith('Sonderlayout'))", 'die Karte »Sonderlayout«')
  const bild = await p.warte(`${was}: die Vorschau zeigt ein Bild`, () => p.seite("const b = [...document.querySelectorAll('img')].find((i) => i.checkVisibility() && i.complete && i.naturalWidth > 0 && i.src.includes('/vorschau')); return b ? b.naturalWidth : 0"), 20000)
  p.pruefe(`${was}: das Kartenbild ist geladen`, bild > 0, `${String(bild)} px breit`)
}

async function tisch(p, was) {
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  await p.wartetext('Weiterspielen')
  await p.klickeText('Weiterspielen')
  await p.klickeText('Weiterspielen als Ada')
  await p.warte(`${was}: die Leinwand des Tisches zeichnet`, () => p.seite("return !!document.querySelector('canvas')?.checkVisibility()"), 20000)
  p.pruefe(`${was}: am Tisch steht keine Fehlermeldung`, (await meldungen(p)).length === 0, JSON.stringify(await meldungen(p)))
  await p.bild(`sec17-tisch-${was}`)
}

// starteMitSet startet das Bündel und öffnet das erste Set — dieselben
// Schritte wie `uirevue.oeffnen`, nur mit eigener Umgebung.
async function starteMitSet(p, ort, umgebung = {}) {
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '1280')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '800')
  await p.starte({ ort, umgebung })
  await p.taste('2', 'befehl')
  await p.warte('die Sets stehen', () => p.seite("return !!document.querySelector('main article')"), 20000)
  await p.klickeAuf("[...document.querySelectorAll('main article :is(h2, h3) button, main article :is(h2, h3) a')].find((b) => b.checkVisibility())", 'das erste Set')
  await new Promise((los) => setTimeout(los, 2000))
}

function partieAnlegen(p, ort) {
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  const angelegt = p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
  const partie = /Partie ([0-9a-f]+)/.exec(angelegt)?.[1] ?? ''
  p.pruefe('die Partie ist angelegt', partie !== '', angelegt.split('\n')[0])
  for (const sitz of ['ada', 'bo']) {
    for (const art of ['spieler.starthand', 'spieler.behalten']) {
      p.mse(ort, 'partie-befehl', partie, sitz, `${art}/${sitz}`, art)
    }
  }
}

export default async function sec17(p) {
  p.erwarteteVerstoesse = kuenstlich
  const ort = p.ort()
  partieAnlegen(p, ort)

  p.notiz('1 — beobachtend: die Policy meldet, hält aber nichts auf')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  await starteMitSet(p, ort, { MSE_CSP_BEOBACHTEN: '1' })
  const beobachtet = await p.seite(eindringen)
  p.zeige(`beobachtend: ${JSON.stringify(beobachtet)}`)
  p.pruefe('beobachtend läuft das künstliche Skript', beobachtet.inline === 'lief')
  await kartenbild(p, 'beobachtend')
  await tisch(p, 'beobachtend')
  const inventar = cspZeilen(ort)
  p.zeige(`beobachtend gemeldet (${String(inventar.zeilen.length)}): ${inventar.zeilen.map((z) => z.slice(0, 160)).join(' | ')}`)
  p.pruefe('⌘Q beendet', await p.beende())

  let ab = cspZeilen(ort).laenge
  for (const bild of ['hell', 'dunkel']) {
    p.notiz(`${bild === 'hell' ? '2' : '3'} — durchgesetzt, ${bild}`)
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    await starteMitSet(p, ort)
    const durchgesetzt = await p.seite(eindringen)
    p.zeige(`${bild}: ${JSON.stringify(durchgesetzt)}`)
    p.pruefe(`${bild}: das künstliche Inline-Skript ist abgewiesen`, durchgesetzt.inline === 'abgewiesen')
    p.pruefe(`${bild}: eval ist abgewiesen`, durchgesetzt.eval === 'abgewiesen')
    p.pruefe(`${bild}: das fremde Skript ist gemeldet`, durchgesetzt.verstoesse.some((v) => v.includes('sec17.invalid')), durchgesetzt.verstoesse.join(' · '))
    await kartenbild(p, bild)
    await tisch(p, bild)

    const gemeldet = cspZeilen(ort, ab)
    ab = gemeldet.laenge
    const fremde = gemeldet.zeilen.filter((z) => !kuenstlich.test(z))
    p.zeige(`${bild}: das Protokoll nennt ${String(gemeldet.zeilen.length)} Verstöße`)
    p.pruefe(`${bild}: das Protokoll nennt die künstlichen Verstöße`, gemeldet.zeilen.filter((z) => kuenstlich.test(z)).length === 3, String(gemeldet.zeilen.length))
    p.pruefe(`${bild}: und keinen anderen`, fremde.length === 0, fremde.map((z) => z.slice(0, 200)).join(' | '))
    p.pruefe('⌘Q beendet', await p.beende())
  }
}
