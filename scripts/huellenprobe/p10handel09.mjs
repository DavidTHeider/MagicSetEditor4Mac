// **Handelszeilen mit Karte, Bild und Preisen am nativen Fenster** — P10
// (#2378),
// Befunde C-13 und C-17 aus dem Audit zur Produktionsreife 0.9.0.
//
// # Was es prüft
//
// **H1 über den Namen**: Anna legt einen Tausch mit Bert an — je Seite eine
// Zeile über die Suche „Karte“, ↩, „+ Zeile hinzufügen“ —, eröffnet und
// bietet an; gezählt nach `huellenwege.mjs` wie in `ausgangswerte09handeln`.
// Die Druckkennung steht dabei nirgends im Bild. **Danach die Zeilen**: Name,
// Set und Nummer, Zustand und das Bild — aus dem, was die Suche gezeigt hat
// (der erste Lauf fand hier „Unbekannter Druck“ mit der Kennung, bei grünen
// Einheitstests). **„Preise holen“ an der Bilanz** gegen eine Gegenstelle
// auf 127.0.0.1 (`MSE_SCRYFALL_ADRESSE`), wie bei `ausgangswerte09handeln`
// und `ap10exemplar09`: ein 429, eine Gegenstelle, die aus ist, eine, die
// mitten in der Anfrage auflegt (#2487), ein Abbruch — und nach jedem steht
// der Handel unverändert angeboten mit zwei Zeilen.
// **Ein eigener Druck** (eine Druckzeile ohne Scryfall-Kennung, angelegt über
// die Kommandozeile) steht mit gestricheltem Platzhalter und ohne Bild. Hell
// und dunkel.
//
// # Die Eingaben
//
// **Erfunden**: zwei Drucke mit Namen, die es nicht gibt, und ein Bild als
// `data:`-Adresse — die Inhaltsrichtlinie des Fensters lässt außer `self`
// und `data:` nur Scryfalls Bildserver zu, und ins Netz geht diese Probe nicht.
//
// # Was es nicht belegt
//
// **Scryfall selbst**, **VoiceOver**, zwei Geräte (#1998). **H1 ≤ 7**: die
// Zahl steht im Beleg; ob sie die Schwelle hält, sagt sie selbst.
import { execFileSync } from 'node:child_process'
import { writeFileSync } from 'node:fs'
import { createServer } from 'node:http'
import { join, resolve } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { gekoppelt, handel, knopf, vorgangsstand } from '../huellenhandel.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = join(wurzel, 'docs/belege/p10handel09-wege.json')
const bild = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=='
const blitz = '00000000-0000-4000-8000-0000000a1001'
const sturm = '00000000-0000-4000-8000-0000000b1002'
const karten = {
  [blitz]: { name: 'Blitzschlag', nummer: '146' },
  [sturm]: { name: 'Sturmrufer', nummer: '3' },
}
const karte = (id) => ({
  object: 'card', id, name: karten[id].name, lang: 'de', layout: 'normal', mana_cost: '{R}',
  type_line: 'Instant', rarity: 'common', set: 'p10', set_name: 'Probesatz Zehn',
  collector_number: karten[id].nummer, artist: 'Selbst Gemalt', scryfall_uri: 'https://beispiel.invalid/karte',
  image_uris: { small: bild }, prices: { eur: id === blitz ? '1.50' : '0.40', usd: null },
})

const mitEingabe = (ort, eingabe, ...argumente) =>
  execFileSync(join(wurzel, '.go/bin/mse'), argumente, { cwd: wurzel, input: eingabe, env: kindumgebung(ort) }).toString()
const bertWahl = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /^Bert aufg?e?schlagen$/i.test((b.getAttribute('aria-label') ?? '').trim()))"
const suchfeld = "[...document.querySelectorAll('input[type=search]')].find((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').trim().startsWith('Karte'))"
const spalte = (titel) => `[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && s.querySelector('p')?.innerText.trim() === ${JSON.stringify(titel)})`
const spaltentext = (titel) => `return (${spalte(titel)})?.innerText.replace(/\\s+/g, ' ').trim() ?? ''`
const bilder = (titel) => `return [...((${spalte(titel)})?.querySelectorAll('img') ?? [])].map((i) => i.getAttribute('src') ?? '')`
const meldung = "return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).join(' | ')"
const stand = "return [...document.querySelectorAll('p[aria-live=polite]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).join(' | ')"
const eintrag = (name) => `[...document.querySelectorAll('main li button')].find((b) => b.checkVisibility() && b.innerText.trim().startsWith(${JSON.stringify(name)}))`

export default async function p10handel09(p) {
  const gefragt = []
  const modus = { karte: 'gut' }
  const gegenstelle = createServer((anfrage, antwort) => {
    const pfad = decodeURIComponent(anfrage.url)
    gefragt.push(pfad)
    if (pfad.startsWith('/cards/search')) {
      const treffer = Object.keys(karten).filter((id) => pfad.includes(karten[id].name)).map(karte)
      antwort.writeHead(200, { 'content-type': 'application/json' })
      antwort.end(JSON.stringify({ object: 'list', total_cards: treffer.length, has_more: false, data: treffer }))

      return
    }
    const id = pfad.split('?')[0].split('/').pop()
    // **Eine Gegenstelle, die auflegt** (#2487): angenommen, gelesen und
    // ohne Antwort geschlossen.
    if (modus.karte === 'auflegen' && id === blitz) {
      anfrage.socket.destroy()

      return
    }
    if (modus.karte === '429' && id === blitz) {
      antwort.writeHead(429, { 'content-type': 'application/json', 'retry-after': '1' })
      antwort.end(JSON.stringify({ object: 'error', status: 429, details: 'Zu viele Anfragen' }))

      return
    }
    const senden = () => {
      antwort.writeHead(200, { 'content-type': 'application/json' })
      antwort.end(JSON.stringify(karte(id)))
    }
    if (modus.karte === 'langsam') setTimeout(senden, 2500)
    else senden()
  })
  await new Promise((fertig) => { gegenstelle.listen(0, '127.0.0.1', fertig) })
  const port = gegenstelle.address().port
  const adresse = `http://127.0.0.1:${String(port)}`
  // **Offline ist ein Port, an dem niemand annimmt** — so sieht es der
  // Transport auch ohne Netz, und ins Netz geht diese Probe nicht.
  modus.aus = () => new Promise((fertig) => { gegenstelle.close(fertig); gegenstelle.closeAllConnections() })
  modus.an = () => new Promise((fertig) => { gegenstelle.listen(port, '127.0.0.1', fertig) })
  try {
    await lauf(p, adresse, modus, gefragt)
  } finally {
    gegenstelle.close()
  }
}

async function lauf(p, adresse, modus, gefragt) {
  const annaOrt = p.ort()
  const bertOrt = p.ort()
  const cleoOrt = p.ort()
  const wer = gekoppelt(p, { Anna: annaOrt, Bert: bertOrt, Cleo: cleoOrt })
  p.mse(annaOrt, 'einstellungen', 'setzen', 'ersteschritte', 'aus')

  // **Ein eigener Druck bei Cleo**: eine Druckzeile ohne Scryfall-Kennung.
  const eigen = mitEingabe(annaOrt, JSON.stringify([
    { von: wer.Anna.kennung, printing: '', druckId: 4242, menge: 1, guete: 'nearMint', sprache: 'de', finish: 'nonfoil' },
    { von: wer.Cleo.kennung, printing: sturm, menge: 1, guete: 'nearMint', sprache: 'de', finish: 'nonfoil' },
  ]), 'handel', 'eroeffnen', 'v-p10-eigen', 'tausch', wer.Cleo.kennung)
  p.zeige(`Tausch mit Cleo: ${eigen.split('\n')[0]}`)

  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const z = wegzaehler(p, datensatz)
  await p.starte({ ort: annaOrt, umgebung: { MSE_SCRYFALL_ADRESSE: adresse } })
  await p.groesse(1440, 900)

  p.notiz('H1 — einen Tausch mit je einer Karte pro Seite anbieten, über den Namen')
  await z.beginne('H1', 'bereich', 'maus')
  await z.taste('5', 'befehl')
  await p.wartetext('Neuer Tausch', 20000)
  await z.klickAuf(knopf('Neuer Tausch …'), '»Neuer Tausch …«')
  await p.warte('Bert steht zur Wahl', () => p.seite(`return !!${bertWahl}`), 10000)
  await z.klickAuf(bertWahl, 'Bert in der Schublade')
  for (const [seite, id] of [['Von dir', blitz], ['Vom Gegenüber', sturm]]) {
    await p.warte(`»Zeile hinzufügen: ${seite}« steht`, () => p.seite(`return !!${knopf(`Zeile hinzufügen: ${seite}`)}`), 10000)
    await z.klickAuf(knopf(`Zeile hinzufügen: ${seite}`), `»Zeile hinzufügen: ${seite}«`)
    const fokus = await p.warte('der Fokus steht in der Suche', () => p.seite(`return document.activeElement === ${suchfeld}`), 5000).catch(() => false)
    p.pruefe(`${seite}: der Fokus steht beim Öffnen in „Karte“`, Boolean(fokus))
    p.pruefe(`${seite}: kein Feld „Druck“ im Hauptweg`, !(await p.seite("return [...document.querySelectorAll('input')].some((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').trim() === 'Druck')")))
    await z.tippe(karten[id].name)
    await z.taste('Enter')
    await p.warte('ein Druck steht gewählt', () => p.seite("return [...document.querySelectorAll('input[type=radio]')].some((r) => r.checkVisibility() && r.checked)"), 15000)
    await z.klickAuf(knopf('+ Zeile hinzufügen'), '»+ Zeile hinzufügen«')
  }
  await z.klickAuf(knopf('Eröffnen'), '»Eröffnen«')
  // **Seit P13 ein Zug** (#2381): „Anbieten und zustimmen" — Bert horcht
  // hier nicht, die Zustellung bleibt im Ausgang, angeboten ist trotzdem.
  await p.warte('»Anbieten und zustimmen« steht', () => p.seite(`return !!${knopf('Anbieten und zustimmen')}`), 10000)
  await z.klickAuf(knopf('Anbieten und zustimmen'), '»Anbieten und zustimmen«')
  const angeboten = await p.warte('angeboten', () => handel(p, annaOrt).zeilen.find((x) => x.vorgang.zustand === 'angeboten')?.vorgang.vorgang ?? null, 10000).catch(() => '')
  const h1 = await z.ende({ wirkung: angeboten !== '' })
  p.zeige(`H1: ${String(h1.handlungen)} Handlungen (Ausgang 13, Schwelle 7)`)
  // **Bert hat keine Adresse** — der Zug sagt, wo sie fehlt (#2491); bis
  // dahin stand hier „lade die Liste neu".
  const ohneAdresse = await p.warte('der Zug sagt, warum Bert es noch nicht hat', async () => { const m = await p.seite(meldung); return m.includes('Bert hat es noch nicht') ? m : null }, 10000).catch(() => '')
  p.pruefe('ohne Adresse sagt der Zug, wo sie fehlt, statt „lade die Liste neu"', ohneAdresse.includes('keine Adresse eingetragen') && !ohneAdresse.includes('lade die Liste neu'), ohneAdresse)
  p.pruefe('gefragt wurde nach Drucken, über den Namen', gefragt.filter((u) => u.startsWith('/cards/search') && u.includes('unique=prints')).length === 2, gefragt.join(', '))

  p.notiz('Die Zeilen nennen ihre Karte, mit Bild aus der Suche')
  const meine = await p.warte('die Zeile nennt Blitzschlag', async () => { const t = await p.seite(spaltentext('Von dir')); return t.includes('1 × Blitzschlag') ? t : null }, 10000)
  const seine = await p.seite(spaltentext('Vom Gegenüber'))
  p.pruefe('Name, Set, Nummer und Zustand je Zeile', meine.includes('P10 #146 · Near Mint') && seine.includes('1 × Sturmrufer') && seine.includes('P10 #3 · Near Mint'), `${meine} || ${seine}`)
  p.pruefe('keine Druckkennung im Bild', !(await p.seite("return document.body.innerText")).includes(blitz))
  p.pruefe('jede Zeile trägt das Bild ihres Drucks', (await p.seite(bilder('Von dir')))[0] === bild && (await p.seite(bilder('Vom Gegenüber')))[0] === bild)
  p.pruefe('noch kein Preis gefragt', !gefragt.some((u) => u.startsWith('/cards/0')), gefragt.join(', '))
  await p.bild('zeilen')

  const zweiZeilen = () => {
    const st = vorgangsstand(p, annaOrt, angeboten)
    return st.zustand === 'angeboten' && (st.positionen?.length ?? 0) === 2
  }

  p.notiz('Preise holen: ein 429 für Blitzschlag')
  modus.karte = '429'
  await p.klickeAuf(knopf('Preise holen'), '»Preise holen«')
  const nach429 = await p.warte('der Grund steht am Knopf', async () => { const m = await p.seite(meldung); return m.includes('1 von 2') ? m : null }, 20000).catch(() => '')
  p.pruefe('ein 429 nennt seinen Grund, der andere Druck kam', nach429 !== '', nach429)
  p.pruefe('der Handel steht unverändert angeboten mit zwei Zeilen', zweiZeilen())

  p.notiz('Preise holen: die Gegenstelle ist aus')
  modus.karte = 'gut'
  await modus.aus()
  await new Promise((weiter) => setTimeout(weiter, 1200))
  await p.klickeAuf(knopf('Preise holen'), '»Preise holen«')
  const offline = await p.warte('der Grund steht am Knopf', async () => { const m = await p.seite(meldung); return m.includes('1 von 2') && m !== nach429 ? m : null }, 30000).catch(() => '')
  p.pruefe('ohne Gegenstelle steht der Grund, und nichts ist verloren', offline !== '' && !offline.includes('unerwarteter Fehler') && zweiZeilen(), offline)
  await p.bild('preise-offline')
  await modus.an()

  p.notiz('Preise holen: die Gegenstelle legt mitten in der Anfrage auf (#2487)')
  modus.karte = 'auflegen'
  await p.klickeAuf(knopf('Preise holen'), '»Preise holen«')
  const aufgelegt = await p.warte('der Grund steht am Knopf', async () => { const m = await p.seite(meldung); return m.includes('brach ab') ? m : null }, 30000).catch(() => '')
  p.pruefe('eine Gegenstelle, die auflegt, heißt so, und nichts ist verloren', aufgelegt !== '' && !aufgelegt.includes('unerwarteter Fehler') && zweiZeilen(), aufgelegt || (await p.seite(meldung)))
  await p.bild('preise-aufgelegt')

  p.notiz('Preise holen: abgebrochen, während Blitzschlag noch läuft')
  modus.karte = 'langsam'
  const vorher = gefragt.length
  await p.klickeAuf(knopf('Preise holen'), '»Preise holen«')
  await p.warte('der Abbruch steht', () => p.seite(`return !!${knopf('Preisabruf abbrechen')}`), 5000)
  await p.klickeAuf(knopf('Preisabruf abbrechen'), '»Preisabruf abbrechen«')
  const abgebrochen = await p.warte('der Abbruch ist vermerkt', async () => { const s = await p.seite(stand); return s.includes('Abgebrochen nach 1 von 2') ? s : null }, 15000).catch(() => '')
  p.pruefe('ein Abbruch hält vor dem nächsten Druck', abgebrochen !== '' && gefragt.length - vorher === 1 && zweiZeilen(), `${abgebrochen} · ${String(gefragt.length - vorher)} Anfrage(n)`)

  p.notiz('Preise holen: alles da — die Bilanz nennt Währung, Alter und Bilder')
  modus.karte = 'gut'
  await p.klickeAuf(knopf('Preise holen'), '»Preise holen«')
  await p.warte('geholt', async () => (await p.seite(stand)).includes('Preise für 2 Drucke geholt.'), 15000).catch(() => false)
  const bilanz = await p.warte('die Bilanz nennt Preise', async () => { const t = await p.seite("return document.querySelector('main')?.innerText ?? ''"); return t.includes('Preise vom') ? t : null }, 15000).catch(() => '')
  p.pruefe('die Bilanz nennt Betrag, Richtung und Herkunft', /⇆ 1,10\s€/.test(bilanz) && bilanz.includes('Preise vom'), bilanz.split('\n').filter((z) => z.includes('⇆') || z.includes('Preise vom')).join(' | '))
  p.pruefe('beide Zeilen tragen jetzt ihr Bild', (await p.seite(bilder('Von dir')))[0] === bild && (await p.seite(bilder('Vom Gegenüber')))[0] === bild)
  // **Und ihren Namen** — der erste grüne Lauf prüfte ihn nur vor dem
  // Abruf, und danach stand „Unbekannter Druck“ da (die Druckzeile, die der
  // Preisabruf anlegt, hat keinen Namen).
  const nachher = `${await p.seite(spaltentext('Von dir'))} || ${await p.seite(spaltentext('Vom Gegenüber'))}`
  p.pruefe('nach dem Abruf nennen die Zeilen weiter ihre Karte', nachher.includes('1 × Blitzschlag') && nachher.includes('1 × Sturmrufer') && !nachher.includes('Unbekannter Druck'), nachher)
  await p.bild('preise-hell')

  p.notiz('Ein eigener Druck ohne Scryfall-Kennung')
  await p.klickeAuf(eintrag('Cleo'), 'den Handel mit Cleo')
  const eigeneZeile = await p.warte('die eigene Zeile steht', async () => { const t = await p.seite(spaltentext('Von dir')); return t.includes('eigener Druck') ? t : null }, 10000).catch(() => '')
  p.pruefe('daneben nennt der Druck des Gegenübers seine Karte', (await p.seite(spaltentext('Vom Gegenüber'))).includes('1 × Sturmrufer'))
  p.pruefe('ein eigener Druck steht mit eigenem Platzhalter und ohne Bild', eigeneZeile.includes('eigener Druck · ohne Bild') && (await p.seite(bilder('Von dir'))).length === 0
    && (await p.seite(`return !!(${spalte('Von dir')})?.querySelector('span.border-dashed')`)), eigeneZeile)
  await p.bild('eigener-druck')
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Dasselbe im dunklen Erscheinungsbild')
  p.mse(annaOrt, 'einstellungen', 'setzen', 'erscheinungsbild', 'dunkel')
  await p.starte({ ort: annaOrt, umgebung: { MSE_SCRYFALL_ADRESSE: adresse } })
  await p.groesse(1440, 900)
  await p.taste('5', 'befehl')
  await p.warte('der Handel mit Bert steht', () => p.seite(`return !!${eintrag('Bert')}`), 20000)
  await p.klickeAuf(eintrag('Bert'), 'den Handel mit Bert')
  await p.warte('die Zeilen stehen mit Bild', async () => (await p.seite(bilder('Von dir'))).length === 1, 10000).catch(() => false)
  await p.bild('preise-dunkel')
  p.pruefe('⌘Q beendet', await p.beende())

  writeFileSync(ziel, `${JSON.stringify(datensatz, null, 2)}\n`)
}
