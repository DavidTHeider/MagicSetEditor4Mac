// **Ein Knopf richtet alles ein, und die Schnittstellen sind wählbar** — G08
// (#2272),
// am nativen Fenster:
//
//   1. *Einstellungen › Verbindungen › Werkzeuge ohne Fenster*: über den Zielen
//      steht „Alles auf einmal“ mit **einer** Vorschau über alles, was fehlt,
//   2. „Wo die Dienste lauschen“: Vorgabe „Nur dieser Mac“; gewählt werden
//      `127.0.0.1` und `::1`, und eine Adresse nach außen zeigt den Hinweis
//      auf TLS und Token — gespeichert wird nur die Schleife,
//   3. *Verbindungen › Was Fachleute hier sonst noch einstellen*: die
//      Skriptschnittstelle startet **ohne Adresse** und lauscht damit auf
//      beiden; über jede kommt eine echte Anfrage mit der Marke durch.
//
// **Nicht gedrückt wird „Alles einrichten“** — es meldete einen Dienst beim
// launchd dieses Rechners an und riefe das echte `claude` (Vermerk 0159); das
// gehört in einen Wegwerf-Gast. Die Wirkung je Ziel hält
// `internal/app/schnittstellenalle_intern_test.go`.
import { request } from 'node:http'

const istSchleife = (adresse) => adresse.startsWith('127.') || adresse === '::1'

// frage holt `/sets` über eine Adresse mit der Marke und gibt den Status.
function frage(adresse, marke) {
  return new Promise((fertig) => {
    const ziel = new URL(adresse)
    const anfrage = request({
      host: ziel.hostname.replace(/^\[|\]$/g, ''), port: ziel.port, path: '/sets',
      headers: { authorization: `Bearer ${marke}` },
    }, (antwort) => { antwort.resume(); fertig(antwort.statusCode) })
    anfrage.on('error', (fehler) => { fertig(String(fehler)) })
    anfrage.end()
  })
}

export default async function g08(p) {
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  await p.starte({ ort })

  p.notiz('1 — Alles auf einmal: eine Vorschau über alles, was fehlt')
  await p.taste(',', 'befehl')
  await p.wartetext('Netz, Gegenüber, CLI, MCP, REST', 20000)
  // **Seit AP-10 (#2367) unter Verbindungen** und nicht mehr unter Diagnose.
  await p.klickeText('Verbindungen')
  await p.klickeText('Werkzeuge ohne Fenster')
  await p.wartetext('Alles auf einmal', 20000)
  const vorschau = await p.warte('die Vorschau steht', () => p.seite("return [...document.querySelectorAll('summary')].find((s) => s.checkVisibility() && s.innerText.startsWith('Was dabei geschrieben wird'))?.innerText ?? null"), 20000)
  p.zeige(vorschau)
  p.pruefe('die Vorschau zählt die Ziele, die fehlen', /\(\d+ Ziele\)/.test(vorschau))
  await p.seite("[...document.querySelectorAll('summary')].find((s) => s.innerText.startsWith('Was dabei geschrieben wird'))?.click(); return true")
  await p.bild('g08-alles-auf-einmal')

  p.notiz('2 — Wo die Dienste lauschen: Vorgabe, Auswahl, Hinweis nach außen')
  await p.seite("document.querySelector('h4')?.closest('section') && [...document.querySelectorAll('h4')].find((h) => h.innerText === 'Wo die Dienste lauschen')?.scrollIntoView({ block: 'center' }); return true")
  const vorgabe = await p.warte('die Wahl ist gelesen', () => p.seite("return [...document.querySelectorAll('input[type=radio]')].find((r) => r.checked)?.parentElement?.innerText.trim() ?? null"), 20000)
  p.pruefe('vorgegeben ist »Nur dieser Mac«', vorgabe.startsWith('Nur dieser Mac'), vorgabe)
  await p.klickeText('Ausgewählte Adressen')
  const adressen = await p.warte('die Schnittstellen stehen', async () => {
    const liste = await p.seite("return [...document.querySelectorAll('input[type=checkbox]')].map((k) => k.parentElement?.innerText.trim() ?? '')")
    return liste.length > 0 ? liste : null
  }, 20000)
  p.zeige(`zur Wahl: ${adressen.join(', ')}`)
  const aussen = adressen.find((adresse) => !istSchleife(adresse))
  if (aussen !== undefined) {
    await p.klickeAuf(`[...document.querySelectorAll('input[type=checkbox]')].find((k) => k.parentElement?.innerText.trim() === ${JSON.stringify(aussen)})`, `die Adresse ${aussen}`)
    await p.wartetext('Eine gewählte Adresse verlässt diesen Mac', 10000)
    await p.bild('g08-hinweis-nach-aussen')
    await p.klickeAuf(`[...document.querySelectorAll('input[type=checkbox]')].find((k) => k.parentElement?.innerText.trim() === ${JSON.stringify(aussen)})`, `die Adresse ${aussen} wieder`)
  } else {
    p.zeige('keine Adresse nach außen auf diesem Rechner — der Hinweis bleibt ungezeigt')
  }
  for (const adresse of ['127.0.0.1', '::1']) {
    await p.klickeAuf(`[...document.querySelectorAll('input[type=checkbox]')].find((k) => k.parentElement?.innerText.trim() === ${JSON.stringify(adresse)})`, `die Adresse ${adresse}`)
  }
  await p.klickeText('Wahl übernehmen')
  await p.wartetext('Übernommen — wirkt beim nächsten Start eines Dienstes.', 10000)
  await p.bild('g08-lauschwahl')
  // **Über die Liste**, denn `mse einstellungen` kennt kein `lesen` — der
  // erste native Lauf (2026-10-10) brach genau hier ab.
  const lauschen = p.mse(ort, 'einstellungen').split('\n').find((zeile) => zeile.includes('netz.lauschen')) ?? ''
  p.pruefe('die Wahl steht in der Einstellung', lauschen.includes('127.0.0.1,::1'), lauschen)

  p.notiz('3 — der Dienst aus dem Fenster lauscht auf beiden')
  await p.taste('Escape')
  await p.taste(',', 'befehl')
  await p.klickeText('Verbindungen')
  await p.klickeText('Was Fachleute hier sonst noch einstellen')
  await p.wartetext('Die Schnittstelle für Skripte', 20000)
  await p.klickeAuf("[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && s.querySelector('h3')?.innerText === 'Die Schnittstelle für Skripte')?.querySelector('button:not([disabled])')", 'Starten an der Skriptschnittstelle')
  const posten = await p.warte('der Dienst läuft', () => p.seite("const s = [...document.querySelectorAll('section')].find((s) => s.querySelector('h3')?.innerText === 'Die Schnittstelle für Skripte'); const t = s?.innerText ?? ''; return t.includes('Lauscht außerdem auf') ? t : null"), 30000)
  p.zeige(posten.replace(/\s+/g, ' ').slice(0, 400))
  const lauscher = [...posten.matchAll(/https?:\/\/[^\s,]+/g)].map((treffer) => treffer[0])
  // **Die Marke steht als „Zugang"** — der erste Anlauf suchte „Marke" und
  // schickte die Anfragen ohne sie (401, nativer Lauf 2026-10-10).
  const marke = posten.match(/Zugang:\s*([0-9a-f]{32,})/)?.[1] ?? ''
  p.pruefe('die Seite nennt den Zugang', marke !== '', marke === '' ? 'keiner' : `${marke.slice(0, 8)}…`)
  p.pruefe('zwei Lauscher', new Set(lauscher).size >= 2, lauscher.join(' '))
  for (const adresse of new Set(lauscher)) {
    const status = await frage(adresse, marke)
    p.pruefe(`über ${adresse} kommt eine Anfrage durch`, status === 200, String(status))
  }
  await p.bild('g08-dienst-auf-beiden')
  p.pruefe('⌘Q beendet', await p.beende())
}
