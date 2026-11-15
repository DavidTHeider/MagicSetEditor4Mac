// **Ruhe in Einstellungen, Hilfe, Suche und Profilblock, am nativen Fenster**
// — AP-11 (#2368),
// Befunde B-17, B-18, B-27, B-29, B-31, Entwurf `Ruhe09`.
//
// # Was es prüft
//
// 1. **Einstellungen**: die Rolltiefe jeder Gruppe bei 1.440 × 838 und
//    1.024 × 700, mit derselben Messung wie AP-13 (`messung` aus
//    `uirevue.mjs`). Mit `MSE_AP11_VORHER=1` schreibt der Lauf den Vorher-Satz
//    nach `docs/belege/ap11ruhe09-vorher.json` und vergleicht nicht; sonst
//    verlangt er, dass **keine Gruppe tiefer** rollt als vorher, **die
//    Summe sinkt** und **jede R4 hält** — höchstens 2,5 (#2465). Dazu: „Wirkt sofort“ steht nirgends; ein abweichender
//    Wert trägt eine Marke, ein Wert der Vorgabe keine; was erst nach dem
//    Neustart wirkt, sagt es in seiner Zeile; ein Schalter, den „nur lokal“
//    hält, sagt, was gilt und dass eine Richtlinie es hält.
// 2. **Hilfe**: ein Inhaltsverzeichnis links (eine senkrechte Liste mit
//    ↑/↓) statt der Laschen, die Suche darüber — und **jedes Ziel, das
//    vorher eine Lasche hatte**, steht darin (der Vorher-Satz nennt sie).
// 3. **Suche**: im offenen Set steht „In“ auf diesem Set, „Alle Sets“
//    daneben; „Recherche …“ steht nach dem Suchfeld in der Tabfolge, die
//    Recherche selbst nicht in der Seite.
// 4. **Profilblock**: eine Zeile mit dem Namen; als Leser im fremden Profil
//    stehen „nur lesen“ und der Eigner sichtbar darin, im eigenen keiner.
// 5. **Du**: die Seite „Nutzer und Profile“ siezt nicht.
//
// Hell und dunkel, künstlicher Bestand, eigenes Zuhause.
//
// # Was es nicht belegt
//
// **VoiceOver** hört ein Mensch. **Die sieben anderen Kataloge** hält
// `frontend/src/sprache/anrede.test.ts`; das Fenster zeigt hier Deutsch.
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { messung } from './uirevue.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const vorherdatei = join(wurzel, 'docs/belege/ap11ruhe09-vorher.json')
const vorherlauf = process.env.MSE_AP11_VORHER === '1'

const gruppen = ['Allgemein', 'Bereiche und Funktionen', 'Erscheinungsbild', 'Nutzer und Profil', 'Fachfunktionen', 'KI', 'Verbindungen', 'Dienste', 'Daten und Sicherung', 'Diagnose']
const groessen = [[1440, 838], [1024, 700]]

const lasche = (name) => `[...document.querySelectorAll('main [role=tab]')].find((t) => t.checkVisibility() && t.innerText.split('\\n')[0].trim() === ${JSON.stringify(name)})`
// **Die ganze Zeile** — Name, Feld, Hinweis und Herkunft; `data-einstellung`
// steht nur am Feld.
const zeile = (schluessel) => `document.querySelector('[data-einstellungszeile="${schluessel}"]')`

async function zuSeite(p, name, merkmal) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"), 10000)
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld(name)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim().split('\\n')[0].trim() === ${JSON.stringify(name)})`, `»${name}« in der Palette`)
  await p.wartetext(merkmal, 20000)
}

async function einstellungen(p, bild, satz) {
  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', () => p.seite(`return !!${lasche('Allgemein')}`), 20000)
  let sofort = false
  for (const [breite, hoehe] of groessen) {
    await p.groesse(breite, hoehe)
    for (const name of gruppen) {
      await p.klickeAuf(lasche(name), `die Gruppe »${name}«`)
      await p.warte(`»${name}« ist offen`, () => p.seite(`return ${lasche(name)}?.getAttribute('aria-selected') === 'true'`), 10000)
      await new Promise((weiter) => { setTimeout(weiter, 300) })
      const wert = await p.seite(messung)
      satz.einstellungen[`${String(breite)}x${String(hoehe)}/${name}`] = Math.round(wert.rolltiefe * 10) / 10
      sofort ||= await p.seite("return document.querySelector('main').innerText.includes('Wirkt sofort')")
    }
  }
  await p.groesse(1440, 838)
  p.pruefe(`${bild}: „Wirkt sofort“ steht in keiner Gruppe`, !sofort)

  await p.klickeAuf(lasche('Erscheinungsbild'), 'die Gruppe »Erscheinungsbild«')
  await new Promise((weiter) => { setTimeout(weiter, 500) })
  const farbwelt = await p.seite(`const z = ${zeile('farbwelt')}; return z ? { marke: z.querySelector('[data-herkunftsmarke]')?.innerText.trim() ?? '', text: z.innerText.replaceAll('\\n', ' · ') } : null`)
  p.pruefe(`${bild}: der abweichende Wert „Farbwelt“ trägt eine Marke`, !!farbwelt && farbwelt.marke !== '', JSON.stringify(farbwelt))
  const groesse = await p.seite(`const z = ${zeile('schriftgroesse')}; return z ? { marke: !!z.querySelector('[data-herkunftsmarke]'), text: z.innerText.replaceAll('\\n', ' · ') } : null`)
  p.pruefe(`${bild}: der Wert der Vorgabe „Schriftgröße“ trägt keine`, groesse !== null && !groesse.marke, JSON.stringify(groesse))
  await p.bild(`erscheinungsbild-${bild}`)

  await p.klickeAuf(lasche('Allgemein'), 'die Gruppe »Allgemein«')
  await new Promise((weiter) => { setTimeout(weiter, 500) })
  const neustart = await p.seite(`const z = ${zeile('aktualisierung')}; return z ? z.innerText.replaceAll('\\n', ' · ') : null`)
  p.pruefe(`${bild}: „Aktualisierung“ sagt in ihrer Zeile, dass sie erst nach dem Neustart wirkt`, /nach (einem|dem) Neustart/i.test(neustart ?? ''), String(neustart))
  await p.bild(`allgemein-${bild}`)
}

// **Ein Schalter, den eine Richtlinie hält, sagt es in seiner Zeile** —
// `Ruhe09`. „nur lokal“ beim Menschen schaltet Scryfall ab, dessen Vorgabe „an“
// ist (`domain.mitNurLokal`); der Schalter steht weiter auf an. Gesetzt wird
// erst **nach** den Messungen und danach zurückgenommen: die Sätze unter den
// gehaltenen Schaltern verlängerten sonst die Gruppe *Dienste*, und der
// Vergleich mit dem Vorher-Satz mäße zwei Lagen.
async function richtlinie(p, bild) {
  const setze = (wert) => p.seite(`try { const a = await window.go.app.Bindungen.NutzereinstellungSetzen('1', 'richtlinie.nurlokal', '${wert}'); return a.fehler ?? 'gesetzt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 200) }`)
  p.zeige(`nur lokal an: ${JSON.stringify(await setze('an'))}`)
  await neuLaden(p)
  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', () => p.seite(`return !!${lasche('Dienste')}`), 20000)
  await p.klickeAuf(lasche('Dienste'), 'die Gruppe »Dienste«')
  const gehalten = await p.warte('der Schalter von Scryfall steht da', () => p.seite(`const z = ${zeile('dienst.scryfall')}; return z && z.checkVisibility() ? { text: z.innerText.replaceAll('\\n', ' · '), eingeschaltet: z.querySelector('input[role=switch]')?.checked ?? null } : null`), 10000)
  p.pruefe(`${bild}: der Schalter, den „nur lokal“ hält, sagt in seiner Zeile, was gilt und warum`, gehalten.eingeschaltet === true && /Es gilt „[^“]+“ — eine Richtlinie hält es\./.test(gehalten.text), JSON.stringify(gehalten))
  await p.bild(`dienste-richtlinie-${bild}`)
  p.zeige(`nur lokal aus: ${JSON.stringify(await setze('aus'))}`)
  await neuLaden(p)
}

async function hilfe(p, bild, satz) {
  await zuSeite(p, 'Hilfe', 'Begriffe')
  const ziele = await p.seite("return [...document.querySelectorAll('main [role=tab]')].filter((t) => t.checkVisibility()).map((t) => t.innerText.split('\\n')[0].trim())")
  const senkrecht = await p.seite("const l = document.querySelector('main [role=tablist]'); return l ? l.getAttribute('aria-orientation') : null")
  if (vorherlauf) {
    satz.hilfeziele = ziele
    return
  }
  p.pruefe(`${bild}: die Hilfe führt ein Inhaltsverzeichnis links (senkrecht, ↑/↓)`, senkrecht === 'vertical', String(senkrecht))
  const vorher = JSON.parse(readFileSync(vorherdatei, 'utf8')).hilfeziele
  const fehlt = vorher.filter((name) => !ziele.includes(name))
  p.pruefe(`${bild}: kein Hilfeziel ging verloren (${String(vorher.length)} vorher)`, fehlt.length === 0, fehlt.length === 0 ? ziele.join(' · ') : `fehlt: ${fehlt.join(', ')}`)
  const sucheDavor = await p.seite("const f = [...document.querySelectorAll('main input[type=search], main input[type=text]')].find((e) => e.checkVisibility()); const l = document.querySelector('main [role=tablist]'); return !!f && !!l && !!(f.compareDocumentPosition(l) & Node.DOCUMENT_POSITION_FOLLOWING)")
  p.pruefe(`${bild}: die Suche steht über dem Inhaltsverzeichnis`, sucheDavor)
  await p.seite("[...document.querySelectorAll('main [role=tab]')].find((t) => t.getAttribute('aria-selected') === 'true')?.focus(); return true")
  const vorDem = await p.seite("return document.activeElement?.innerText.split('\\n')[0].trim() ?? ''")
  await p.taste('ArrowDown')
  const danach = await p.warte('↓ wählt das nächste Ziel', async () => { const n = await p.seite("return document.activeElement?.getAttribute('role') === 'tab' ? document.activeElement.innerText.split('\\n')[0].trim() : ''"); return n !== '' && n !== vorDem ? n : null }, 5000)
  p.pruefe(`${bild}: ↓ geht im Inhaltsverzeichnis weiter`, !!danach, `${vorDem} → ${String(danach)}`)
  await p.bild(`hilfe-${bild}`)
}

async function suche(p, bild) {
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.wartetext('Karte anlegen', 20000)
  const reiter = "[...document.querySelectorAll('[role=tab]')].find((t) => t.checkVisibility() && t.innerText.trim() === 'Suche')"
  await p.klickeAuf(reiter, 'die Lasche »Suche«')
  const umfang = await p.warte('die Suche steht da', () => p.seite("const s = [...document.querySelectorAll('main select, main [role=combobox]')].filter((e) => e.checkVisibility()).map((e) => ({ text: (e.selectedOptions?.[0]?.text ?? e.value ?? e.innerText ?? '').trim(), alle: [...(e.options ?? [])].map((o) => o.text) })); return s.length ? s : null"), 15000)
  const setwahl = umfang.find((s) => s.alle.includes('Alle Sets'))
  p.pruefe(`${bild}: im offenen Set steht „In“ auf diesem Set, „Alle Sets“ daneben`, setwahl?.text === 'Werkstatt-Testkarten', JSON.stringify(umfang))
  const recherche = await p.seite("const feld = [...document.querySelectorAll('main input')].find((e) => e.checkVisibility() && (e.type === 'search' || e.type === 'text')); const knopf = [...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.trim().startsWith('Recherche')); return { feld: !!feld, knopf: !!knopf, danach: !!feld && !!knopf && !!(feld.compareDocumentPosition(knopf) & Node.DOCUMENT_POSITION_FOLLOWING), offen: [...document.querySelectorAll('main h3')].some((h) => h.checkVisibility() && h.innerText.trim() === 'Im Wissensbestand nachschlagen') }")
  p.pruefe(`${bild}: „Recherche …“ steht nach dem Suchfeld, die Recherche selbst nicht in der Seite`, recherche.knopf && recherche.danach && !recherche.offen, JSON.stringify(recherche))
  await p.bild(`suche-${bild}`)
}

const block = "document.querySelector('section[aria-label=\"Arbeitskontext\"]')"

async function profilblock(p, bild, rolle) {
  const stand = await p.warte('der Profilblock steht', () => p.seite(`const b = ${block}; if (!b) return null; const r = b.getBoundingClientRect(); const ganz = (e) => e.checkVisibility() && !e.closest('.sr-only') && e.getBoundingClientRect().width >= 8 && e.scrollWidth <= e.clientWidth + 1; const sichtbar = (pruefe) => [...b.querySelectorAll('*')].some((e) => e.children.length === 0 && pruefe(e.textContent.trim()) && ganz(e)); const name = [...b.querySelectorAll('.truncate')].find((e) => e.textContent.includes(' · ')); const knopf = name?.closest('button'); return { hoehe: Math.round(r.height), text: b.innerText.replaceAll('\\n', ' · '), lesen: sichtbar((t) => t === 'nur lesen'), eigner: sichtbar((t) => t.startsWith('von ')), name: name ? { text: name.textContent, ganz: ganz(name) } : null, beschreibung: knopf ? document.getElementById(knopf.getAttribute('aria-describedby') ?? '')?.textContent ?? '' : '' }`), 15000)
  p.pruefe(`${bild}, ${rolle}: der Profilblock ist eine Zeile (höchstens 64 Punkte)`, stand.hoehe <= 64, JSON.stringify(stand))
  // **Ganz zu lesen heißt: Breite und nicht gekürzt** — `checkVisibility`
  // allein hielt einen auf null Punkte geschrumpften Eigner für sichtbar,
  // und neben einem eigenen Knopf stand vom Namen nur „Bo · M…“.
  p.pruefe(`${bild}, ${rolle}: „Name · Profil“ steht ungekürzt da`, stand.name?.ganz === true, JSON.stringify(stand.name))
  p.pruefe(`${bild}, ${rolle}: die Zeile führt zu den Nutzern („Profil wechseln …“)`, stand.beschreibung === 'Profil wechseln …', stand.beschreibung)
  if (rolle === 'Leser') {
    p.pruefe(`${bild}: „nur lesen“ steht sichtbar im Block`, stand.lesen, JSON.stringify(stand))
    p.pruefe(`${bild}: wessen Profil es ist, steht sichtbar im Block`, stand.eigner, JSON.stringify(stand))
  } else {
    p.pruefe(`${bild}: im eigenen Profil steht kein Eigner`, !stand.eigner && !stand.lesen, JSON.stringify(stand))
  }
}

async function neuLaden(p) {
  await p.seite('setTimeout(() => location.reload(), 30); return true')
  await new Promise((weiter) => { setTimeout(weiter, 1500) })
  await p.warte('die Leiste steht wieder', () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3"), 30000)
}

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'einstellungen', 'setzen', 'farbwelt', 'klar')
  p.mse(ort, 'einstellungen', 'setzen', 'aktualisierung', 'melden')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'nutzer-neu', 'Bo')
  p.mse(ort, 'freigabe-setzen', '1', '2', 'leser')
  const satz = { erhoben: new Date().toISOString(), einstellungen: {}, hilfeziele: [] }
  await p.starte({ ort })
  await p.groesse(1440, 838)

  p.notiz(`${bild}: Einstellungen`)
  await einstellungen(p, bild, satz)
  if (vorherlauf && bild === 'hell') writeFileSync(vorherdatei, `${JSON.stringify(satz, null, 1)}\n`)
  if (!vorherlauf) {
    p.notiz(`${bild}: eine Richtlinie hält einen Schalter`)
    await richtlinie(p, bild)
  }
  p.notiz(`${bild}: Hilfe`)
  await hilfe(p, bild, satz)
  // **Der Vorher-Satz gleich nach den Messungen** — eine Prüfung danach, die
  // am alten Stand scheitert, soll ihn nicht mitnehmen.
  if (vorherlauf && bild === 'hell') {
    writeFileSync(vorherdatei, `${JSON.stringify(satz, null, 1)}\n`)
    p.zeige(`Vorher-Satz geschrieben: ${JSON.stringify(satz.einstellungen)}`)
  }
  p.notiz(`${bild}: Suche`)
  await suche(p, bild)
  p.notiz(`${bild}: Profilblock und Du`)
  await profilblock(p, bild, 'Eigner')
  await zuSeite(p, 'Nutzer und Profile', 'Profil')
  p.pruefe(`${bild}: „Nutzer und Profile“ siezt nicht`, !(await p.seite("return /\\b(Sie|Ihnen|Ihre?[mnrs]?)\\b/.test(document.querySelector('main').innerText.replace(/(^|[.!?:„“]\\s*)Sie\\b/g, ''))")))
  p.zeige(`Nutzerwechsel: ${JSON.stringify(await p.seite("try { const a = await window.go.app.Bindungen.NutzerWechseln('2', ''); return a.fehler ?? 'gewechselt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 200) }"))}`)
  await neuLaden(p)
  // **Und in das Profil, das Bo nur lesen darf** — wie in `g10menue`.
  p.zeige(`Profilwechsel: ${JSON.stringify(await p.seite("try { const a = await window.go.app.Bindungen.ProfilWechseln('1', ''); return a.fehler ?? 'gewechselt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 200) }"))}`)
  await neuLaden(p)
  await profilblock(p, bild, 'Leser')
  await p.bild(`profilblock-leser-${bild}`)
  p.pruefe(`${bild}: ⌘Q beendet`, await p.beende())

  return satz
}

export default async function ap11ruhe09(p) {
  const saetze = []
  // **Im Vorher-Lauf nur hell** — er misst, und das Dunkle misst dasselbe.
  for (const bild of vorherlauf ? ['hell'] : ['hell', 'dunkel']) saetze.push(await lauf(p, bild))
  const [hell] = saetze
  if (vorherlauf) return
  if (!existsSync(vorherdatei)) {
    p.pruefe('es gibt einen Vorher-Satz', false, vorherdatei)
    return
  }
  const vorher = JSON.parse(readFileSync(vorherdatei, 'utf8')).einstellungen
  const tiefer = Object.entries(hell.einstellungen).filter(([wo, tiefe]) => tiefe > (vorher[wo] ?? Infinity))
  const summe = (s) => Object.values(s).reduce((a, b) => a + b, 0)
  p.pruefe('keine Gruppe rollt tiefer als vorher', tiefer.length === 0, tiefer.map(([wo, tiefe]) => `${wo}: ${String(vorher[wo])} → ${String(tiefe)}`).join('; ') || 'keine')
  p.pruefe(`die Rolltiefe der Einstellungen sinkt (Summe ${summe(vorher).toFixed(1)} → ${summe(hell.einstellungen).toFixed(1)})`, summe(hell.einstellungen) < summe(vorher), JSON.stringify(hell.einstellungen))
  // **Und jede hält R4** — #2465: AP-11 senkte die Summe, vier Gruppen
  // rollten danach noch tiefer als 2,5 Bildschirme.
  const ueber = Object.entries(hell.einstellungen).filter(([, tiefe]) => tiefe > 2.5)
  p.pruefe('jede Gruppe hält R4 (höchstens 2,5 Bildschirme)', ueber.length === 0, ueber.map(([wo, tiefe]) => `${wo}: ${String(tiefe)}`).join('; ') || 'alle')
}
