// **25 Seiten in sechs Zuständen, die Setseiten in einem siebten** — B03
// (#2222).
//
// Jede registrierte Seite (`frontend/src/navigation.ts`, gelesen und nicht
// abgeschrieben) in **leer, befüllt, laden, Fehler, offline und nur lesen** —
// dazu die 13 Setseiten mit einem **leeren Set** (G02) — 163 Zellen, jede
// mit Bild und Urteil oder einem ausdrücklichen „nicht anwendbar" samt
// Grund. Keine Zelle übernimmt einen alten Wert.
//
// # Wie die Zustände entstehen
//
// - **befüllt**: Saatset, Deck, Partie und Exemplar über `mse`; dabei lernt
//   das Drehbuch aus dem Rufbuch, **welche Bindungen eine Seite beim Öffnen
//   ruft** — die Grundlage für die beiden nächsten.
// - **laden**: diese Bindungen um 4 s verzögert (`stoere`,
//   `internal/fenster/rufstoerung.go`); geprüft wird, dass die Seite sagt,
//   dass sie lädt (`aria-busy` oder ein Ladehinweis), statt leer zu stehen —
//   **oder dass sie wie befüllt steht** („steht wie befüllt"): dann zeigt
//   sie nichts von diesen Bindungen (G11, `urteil`).
// - **Fehler**: dieselben Bindungen fachlich gescheitert; geprüft wird eine
//   Meldung (`role=alert`) — oder wieder der befüllte Inhalt („steht wie
//   befüllt").
// - **leer**: ein frischer Bestand, die Setseiten ohne Set; geprüft wird ein
//   Leerzustand, der den nächsten Schritt nennt (`data-leerzustand`). Dazu
//   ein **leeres Set**: wer Karten braucht, nennt den Weg zu ihnen (`lasche`).
// - **offline**: dieselbe Anwendung mit einem Proxy, der nicht antwortet
//   (`HTTP_PROXY`/`HTTPS_PROXY` auf einen geschlossenen Port); geprüft wird,
//   dass die Seite steht.
// - **nur lesen**: ein zweiter Nutzer mit der Rolle Leser am Profil.
//
// **Was es nicht belegt**: ob ein Leerzustand **gut** erklärt (G02) und ob
// eine Meldung den richtigen Rat gibt — das liest ein Mensch an den Bildern.
import { readFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { rufbuchHolen, stoere } from '../huellenrufbuch.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))

// **Im leeren Set zählt die Lasche, nicht die Karten dahinter** (G02, #2249):
// sieben Laschen stehen als Schublade über der Kartenliste, und deren „Noch
// keine Karte" gehörte sonst jeder. Verlangt wird ein Leerzustand **mit Weg**.
const lasche = `return (() => {
  const ort = [...document.querySelectorAll('[role=dialog]')].find((d) => d.checkVisibility()) ?? document.querySelector('main')
  const leer = [...(ort?.querySelectorAll('[data-leerzustand]') ?? [])].filter((e) => e.checkVisibility())
  const weg = leer.some((e) => [...e.querySelectorAll('button')].some((b) => b.checkVisibility()))
  return { leer: leer.length > 0, weg, text: (ort?.innerText ?? '').replace(/\\s+/g, ' ').slice(0, 120) }
})()`

// brauchtKarten sind die Setseiten, die ohne Karte nichts zu tun haben.
const brauchtKarten = ['karten', 'suche', 'ersetzen', 'statistik', 'stil', 'pack', 'bilder', 'druckbogen', 'ausgeben']

// seitenliste liest die registrierten Schlüssel aus `navigation.ts`.
export function seitenliste(quelle) {
  const liste = (name) => [...quelle.match(new RegExp(`export const ${name} = \\[([\\s\\S]*?)\\] as const`))[1].matchAll(/^\s+'([a-z]+)',/gm)].map((t) => t[1])

  return { anwendung: liste('anwendungsschluessel'), set: liste('setschluessel') }
}

// **Beobachtet wird nur der Inhalt der Seite** (`main`) — ein Alarm im
// Meldungsband, ein Ladehinweis der Leiste oder ein Leerzustand einer Schublade
// gehörten sonst der Zelle, auf der die Probe gerade steht (Zweitblick über
// Phase B, Befund 5). `seite` ist der Schlüssel, den der Rahmen an `main`
// schreibt — der Beleg, auf welcher Seite die Probe steht.
const lage = `return (() => {
  const sichtbar = (e) => e && e.checkVisibility()
  const main = document.querySelector('main')
  const darin = (wahl) => [...(main?.querySelectorAll(wahl) ?? [])]
  return {
    seite: main?.dataset.seite ?? '',
    // **Ein Leerzustand ist ein Baustein oder ein Satz, der sagt, was fehlt**
    // — „Noch kein Set geöffnet. Öffne …" ist einer, auch ohne Marke.
    leer: darin('[data-leerzustand]').some(sichtbar) || /\\b(noch kein|noch keine|kein .{1,40} (aufgeschlagen|angelegt|gewählt|geöffnet))/i.test(main?.innerText ?? ''),
    laedt: darin('[aria-busy=true]').some(sichtbar) || /lädt|wird geladen|Augenblick/i.test(main?.innerText ?? ''),
    alarm: darin('[role=alert]').filter(sichtbar).map((e) => e.innerText.trim()).filter(Boolean).slice(0, 3),
    titel: document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? '',
    text: (main?.innerText ?? '').replace(/\\s+/g, ' ').slice(0, 200),
    ganz: (main?.innerText ?? '').replace(/\\s+/g, ' '),
    // **Und was der Text nicht zeigt** — Befund 6 des Zweitblicks über Phase
    // G: Feldwerte, Auswahl, gesperrte Knöpfe und Bildquellen. Eine andere
    // Wahl oder ein gesperrter Knopf gab sonst denselben Text.
    bedienung: [
      ...darin('input, select, textarea').filter(sichtbar).map((e) => [e.tagName, e.type, e.value, e.checked, e.disabled].join(':')),
      ...darin('button').filter(sichtbar).map((e) => [e.innerText.trim(), e.disabled, e.getAttribute('aria-disabled') ?? ''].join(':')),
      ...darin('img').filter(sichtbar).map((e) => e.currentSrc),
    ].join('|'),
  }
})()`

// zurSeite geht über die Befehlspalette zu einer Seite. **Die Palette wird an
// ihren eigenen Zeilen erkannt** — einige Setseiten sind selbst eine
// Schublade (`role=dialog`), und ein Warten auf „kein Dialog" hinge dort.
const paletteOffen = "return [...document.querySelectorAll('[data-palettenzeile]')].some((z) => z.checkVisibility())"

//
// **Angekommen ist, wer die Seite sieht**: die Palette zu schließen bestätigte
// nur, dass sie zu ist (Zweitblick über Phase B, Befund 5) — jetzt muss
// `main` den Schlüssel der Seite tragen.
async function zurSeite(p, schluessel) {
  await p.taste('k', 'befehl')
  await p.warte('die Befehlspalette steht da', () => p.seite(paletteOffen), 15000)
  const zeile = `document.querySelector('[data-palettenzeile][data-seite=${JSON.stringify(schluessel)}]')`
  if (!(await p.seite(`return !!(${zeile})`))) {
    await p.taste('Escape')
    return false
  }
  await p.seite(`(${zeile}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(zeile, `»${schluessel}« in der Palette`)
  await p.warte('die Palette ist zu', async () => !(await p.seite(paletteOffen)), 10000)

  return (await bis(p, `return document.querySelector('main')?.dataset.seite === ${JSON.stringify(schluessel)}`, 10000)) === true
}

// bis fragt die Seite, bis die Antwort wahr ist oder die Frist um — **ohne
// Befund**: hier ist „trat nicht ein" eine Antwort und kein Fehler der Probe
// (anders als `p.warte`).
async function bis(p, frage, frist) {
  const ende = Date.now() + frist
  let zuletzt
  do {
    zuletzt = await p.seite(frage).catch(() => undefined)
    if (zuletzt) return zuletzt
    await schlafe(150)
  } while (Date.now() < ende)

  return zuletzt
}

// beruhigt wartet, bis die Seite nicht mehr lädt — statt einer festen Pause,
// die zu kurz oder zu lang sein konnte —, höchstens fünf Sekunden.
async function beruhigt(p) {
  await bis(p, `return !(${lage.slice('return '.length)}).laedt`, 5000)
  await schlafe(300)

  return p.seite(lage)
}

// setOeffnen öffnet das Saatset über ⌘2 — wie `a02`. **Geöffnet ist es,
// wenn eine Setseite dasteht**, nicht nach einer festen Pause (Nachlauf des
// Zweitblicks über Phase B); sonst wirft es, und der Weg ist nicht erreicht.
async function setOeffnen(p, setseiten, name = 'Werkstatt-Testkarten') {
  await p.taste('2', 'befehl')
  await p.wartetext(name, 20000)
  await p.klickeText(name)
  if (!(await bis(p, `return ${JSON.stringify(setseiten)}.includes(document.querySelector('main')?.dataset.seite)`, 10000))) {
    throw new Error('das Set öffnete nicht')
  }
}

// neuLaden lädt die Webansicht neu und wartet, bis die Leiste wieder steht —
// **ein frischer Ladevorgang**: eine Seite, die ihre Daten im Zustand der
// Anwendung hält, holt sie beim bloßen Zurückkehren nicht neu (erster Lauf).
async function neuLaden(p) {
  await p.seite('setTimeout(() => location.reload(), 30); return true')
  await schlafe(1500)
  await p.warte('die Leiste steht wieder', () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3"), 30000)
}

// zaehlung und delta sagen, welche Bindungen **zwischen zwei Zeitpunkten**
// gerufen wurden — das Rufbuch der Probe sammelt über das ganze Drehbuch
// (erster Lauf: die Listen je Seite wuchsen mit jeder Seite).
const zaehlung = async (p) => structuredClone(await rufbuchHolen(p))
const delta = (vor, nach) => Object.keys(nach).filter((n) => nach[n].anzahl > (vor[n]?.anzahl ?? 0))

// ruhe wartet, bis das Rufbuch anderthalb Sekunden lang nichts Neues
// einträgt — höchstens zehn. **Das Rufbuch trägt nach der Antwort ein**, und
// eine späte Antwort der vorigen Seite landete sonst in der Liste der
// nächsten (Nachlauf des Zweitblicks über Phase B).
async function ruhe(p) {
  const summe = (z) => Object.values(z).reduce((s, e) => s + e.anzahl, 0)
  const ende = Date.now() + 10000
  let vorher = summe(await zaehlung(p))
  while (Date.now() < ende) {
    await schlafe(1500)
    const jetzt = summe(await zaehlung(p))
    if (jetzt === vorher) return
    vorher = jetzt
  }
}
// gestoert besucht eine Seite mit gestörten Bindungen nach einem frischen Laden.
//
// **Beobachtet wird, solange die Störung wirkt** (Zweitblick über Phase B,
// Befund 5) — beim Laden innerhalb der Verzögerung von vier Sekunden, beim
// Fehler, bis eine Meldung auf der Seite steht; eine feste Pause sah sonst
// einen Zustand, der schon vorbei oder noch gar nicht da war. Wer das Set
// nicht öffnen oder die Seite nicht erreichen konnte, bekommt keine Zelle mit
// Urteil, sondern „nicht erreicht".
//
// **Und belegt ist eine Zelle erst, wenn die gestörte Bindung wirklich lief**
// — sonst urteilte sie über eine Seite, die die Störung nie sah.
async function gestoert(p, seite, setseiten, stoerungen) {
  await stoere(p, stoerungen)
  try {
    const vor = await zaehlung(p)
    await neuLaden(p)
    if (setseiten.includes(seite) && !(await setOeffnen(p, setseiten).then(() => true, () => false))) return { nichtErreicht: 'das Set ließ sich nicht öffnen' }
    if (!(await zurSeite(p, seite))) return { nichtErreicht: 'die Seite kam nicht' }
    const fehler = stoerungen[0]?.fehler === true
    await bis(p, `return (${lage.slice('return '.length)}).${fehler ? 'alarm.length > 0' : 'laedt'}`, fehler ? 4000 : 3000)
    const l = await p.seite(lage)
    await p.bild(`${fehler ? 'fehler' : 'laden'}-${seite}`)
    const gestoerte = stoerungen.map((s) => s.bindung)
    const lief = async () => delta(vor, await zaehlung(p)).some((b) => gestoerte.includes(b))
    const ende = Date.now() + 6000
    while (!(await lief()) && Date.now() < ende) await schlafe(300)
    if (!(await lief())) return { nichtErreicht: 'keine der gestörten Bindungen lief' }

    return l
  } finally {
    await stoere(p, [])
  }
}

// urteil gibt das Urteil einer Zelle „laden" oder „Fehler" — G11 (#2277).
//
// **Ob eine Seite von ihren Bindungen abhängt, wird gemessen**: steht sie bei
// gestörten Bindungen wie im befüllten Zustand, zeigt sie nichts davon — dann
// ist „steht wie befüllt" das Urteil und kein Befund. Bei jeder Setseite sind
// auch die Bindungen des Setöffnens gestört (zweiter Lauf von B03), und
// Ersetzen und Ausgeben zeigen davon nichts. **Ohne befüllten Vergleich**
// bleibt es beim Befund.
//
// **„Wie befüllt" heißt: derselbe Text von `main` und derselbe
// Bedienzustand** — Feldwerte, Auswahl, gesperrte Knöpfe, Bildquellen
// (Befund 6 des Zweitblicks über Phase G; bis dahin nur der Text, und das
// Urteil hieß „vollständig"). **Was es nicht vergleicht**: gezeichnete
// Flächen (`canvas`), Stilzustände und alles außerhalb von `main`. Es ist
// deshalb kein Beleg für funktionale Vollständigkeit, sondern dafür, dass
// das Sichtbare und Bedienbare gleich steht.
export function urteil(zustand, l, befuellt) {
  const wieBefuellt = befuellt !== undefined && l.ganz === befuellt.ganz && l.bedienung === befuellt.bedienung
  if (zustand === 'laden') return l.laedt ? 'sagt, dass sie lädt' : wieBefuellt ? 'steht wie befüllt' : 'ohne Ladehinweis'

  return l.alarm.length > 0 ? 'meldet' : wieBefuellt ? 'steht wie befüllt' : 'ohne Meldung'
}

export default async function zustaende(p) {
  const { anwendung, set } = seitenliste(readFileSync(join(wurzel, 'frontend/src/navigation.ts'), 'utf8'))
  const alle = [...anwendung, ...set]
  p.pruefe(`die Liste nennt ${String(alle.length)} Seiten (registriert sind 25)`, alle.length === 25, alle.join(', '))
  const zellen = []
  const zelle = (seite, zustand, urteil, wie) => zellen.push({ seite, zustand, urteil, wie })

  // **befüllt** — und dabei die Bindungen je Seite lernen.
  const voll = p.ort()
  p.mse(voll, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(voll, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(voll, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(voll, 'deck', 'neu', 'Zustandsdeck', 'probe')
  for (const karte of ['1', '2', '3']) p.mse(voll, 'deck', 'legen', '1', `--karte=${karte}`, '4')
  p.mse(voll, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
  // **Der zweite Nutzer für „nur lesen"** — mit Leserfreigabe am Profil 1,
  // angelegt, bevor das Fenster startet.
  p.mse(voll, 'nutzer-neu', 'Zweite')
  p.mse(voll, 'freigabe-setzen', '1', '2', 'leser')
  await p.starte({ ort: voll })
  // **Was im Hintergrund ohnehin läuft**, drei Sekunden Ruhe lang — das stört
  // die Probe nicht, es träfe die ganze Anwendung.
  let vor = await zaehlung(p)
  await schlafe(3000)
  const hintergrund = new Set(delta(vor, await zaehlung(p)))
  // **Was die Palette selbst ruft**, gehört keiner Seite (Zweitblick über
  // Phase B, Befund 5): einmal öffnen und schließen, ohne zu wählen.
  vor = await zaehlung(p)
  await p.taste('k', 'befehl')
  await p.warte('die Befehlspalette steht da', () => p.seite(paletteOffen), 15000)
  await p.taste('Escape')
  await p.warte('die Palette ist zu', async () => !(await p.seite(paletteOffen)), 10000)
  for (const bindung of delta(vor, await zaehlung(p))) hintergrund.add(bindung)
  // **Was das Öffnen des Sets ruft** — die Daten der Setseiten kommen von
  // dort und nicht von der Seite (zweiter Lauf).
  vor = await zaehlung(p)
  await setOeffnen(p, set)
  // **Ohne die Setliste und das Profil** — sie zu stören hieße, das Set gar
  // nicht erst öffnen zu können (dritter Lauf).
  const setbindungen = delta(vor, await zaehlung(p)).filter((b) => !hintergrund.has(b) && !['Sets', 'GezeigtesProfil'].includes(b))
  const bindungen = {}
  const befuellt = {}
  p.notiz('befüllt')
  for (const seite of alle) {
    await ruhe(p)
    vor = await zaehlung(p)
    if (!(await zurSeite(p, seite))) { zelle(seite, 'befüllt', 'nicht erreicht', 'die Seite kam nicht'); continue }
    const l = await beruhigt(p)
    await ruhe(p)
    bindungen[seite] = delta(vor, await zaehlung(p)).filter((b) => !hintergrund.has(b))
    await p.bild(`befuellt-${seite}`)
    befuellt[seite] = { ganz: l.ganz, bedienung: l.bedienung }
    zelle(seite, 'befüllt', l.alarm.length === 0 ? 'steht' : 'Meldung', `»${l.titel}«${l.leer ? ', leer' : ''}${l.alarm.length ? `, ${l.alarm.join(' / ')}` : ''}`)
  }

  // **Ohne die Setliste und das Profil, auch aus der Liste der Seite** — sie
  // zu stören hieße, das Set gar nicht erst öffnen zu können (vierter Lauf).
  const noetig = ['Sets', 'GezeigtesProfil']
  const seiteneigene = Object.fromEntries(alle.map((seite) => [seite, [...new Set([...(bindungen[seite] ?? []), ...(set.includes(seite) ? setbindungen : [])])].filter((b) => !(set.includes(seite) && noetig.includes(b)))]))
  p.zeige(`im Hintergrund, nicht gestört: ${[...hintergrund].join(', ')}; beim Öffnen des Sets: ${setbindungen.join(', ')}`)

  p.notiz('laden')
  for (const seite of alle) {
    if (!seiteneigene[seite]?.length) { zelle(seite, 'laden', 'nicht anwendbar', 'die Seite ruft beim Öffnen keine eigene Bindung'); continue }
    const l = await gestoert(p, seite, set, seiteneigene[seite].map((bindung) => ({ bindung, verzoegerungMs: 4000 })))
    if (l.nichtErreicht) { zelle(seite, 'laden', 'nicht erreicht', l.nichtErreicht); continue }
    zelle(seite, 'laden', urteil('laden', l, befuellt[seite]), `${seiteneigene[seite].join(', ')} um 4 s verzögert`)
  }

  p.notiz('Fehler')
  for (const seite of alle) {
    if (!seiteneigene[seite]?.length) { zelle(seite, 'Fehler', 'nicht anwendbar', 'die Seite ruft beim Öffnen keine eigene Bindung'); continue }
    const l = await gestoert(p, seite, set, seiteneigene[seite].map((bindung) => ({ bindung, fehler: true })))
    if (l.nichtErreicht) { zelle(seite, 'Fehler', 'nicht erreicht', l.nichtErreicht); continue }
    zelle(seite, 'Fehler', urteil('Fehler', l, befuellt[seite]), `${seiteneigene[seite].join(', ')} gescheitert${l.alarm.length ? ` — ${l.alarm.join(' / ').slice(0, 100)}` : ''}`)
  }
  await neuLaden(p)

  p.pruefe('befüllt: ⌘Q beendet', await p.beende())

  p.notiz('offline')
  await p.starte({ ort: voll, umgebung: { HTTP_PROXY: 'http://127.0.0.1:9', HTTPS_PROXY: 'http://127.0.0.1:9', NO_PROXY: '' } })
  const offlineSet = await setOeffnen(p, set).then(() => true, () => false)
  for (const seite of alle) {
    if (set.includes(seite) && !offlineSet) { zelle(seite, 'offline', 'nicht erreicht', 'das Set ließ sich nicht öffnen'); continue }
    if (!(await zurSeite(p, seite))) { zelle(seite, 'offline', 'nicht erreicht', 'die Seite kam nicht'); continue }
    const l = await beruhigt(p)
    await p.bild(`offline-${seite}`)
    // **Das Urteil folgt der Seite** und nicht dem Plan: bis zum Zweitblick
    // stand hier für jede Seite „steht".
    zelle(seite, 'offline', l.alarm.length ? 'meldet' : 'steht', l.alarm.join(' / ').slice(0, 120) || 'keine Meldung')
  }
  p.pruefe('offline: ⌘Q beendet', await p.beende())

  p.notiz('leer')
  const frisch = p.ort()
  p.mse(frisch, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(frisch, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  await p.starte({ ort: frisch })
  // **Auch die Setseiten** — ohne Set stehen sie in der Palette und sagen,
  // dass eines fehlt (`uxr23`). Hier stand „ohne Set gibt es keine
  // Setseite", und dreizehn Zellen hießen „nicht anwendbar", ohne gemessen zu
  // sein (Bilanz der Arbeitswege, `belege/arbeitswege-phase-b.md`).
  for (const seite of alle) {
    if (!(await zurSeite(p, seite))) { zelle(seite, 'leer', 'nicht erreicht', 'die Seite kam nicht'); continue }
    const l = await beruhigt(p)
    await p.bild(`leer-${seite}`)
    zelle(seite, 'leer', l.leer ? 'Leerzustand' : 'ohne Leerzustand', l.text.slice(0, 120))
  }
  p.pruefe('leer: ⌘Q beendet', await p.beende())

  // **Ein leeres Set** (G02, #2249) — ohne Set sagen die Setseiten nur, dass
  // eines fehlt; mit einem ohne Karte zeigt sich, ob sie den Weg nennen.
  p.notiz('leeres Set')
  const leeresSet = p.ort()
  p.mse(leeresSet, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(leeresSet, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(leeresSet, 'set-neu', 'Leeres Probeset', 'probe', 'probe-schlicht')
  await p.starte({ ort: leeresSet })
  await setOeffnen(p, set, 'Leeres Probeset')
  const ohneWeg = []
  for (const seite of set) {
    if (!(await zurSeite(p, seite))) { zelle(seite, 'leeres Set', 'nicht erreicht', 'die Seite kam nicht'); continue }
    await beruhigt(p)
    const l = await p.seite(lasche)
    await p.bild(`leeresset-${seite}`)
    if (brauchtKarten.includes(seite) && !l.weg) ohneWeg.push(seite)
    zelle(seite, 'leeres Set', l.weg ? 'Leerzustand mit Weg' : l.leer ? 'Leerzustand' : 'Inhalt ohne Karten', l.text)
  }
  p.pruefe('leeres Set: jede Seite, die Karten braucht, nennt den Weg', ohneWeg.length === 0, ohneWeg.join(', ') || brauchtKarten.join(', '))
  p.pruefe('leeres Set: ⌘Q beendet', await p.beende())

  p.notiz('nur lesen')
  await p.starte({ ort: voll })
  p.zeige(`Nutzerwechsel: ${JSON.stringify(await p.seite("try { const a = await window.go.app.Bindungen.NutzerWechseln('2', ''); return a.fehler ?? 'gewechselt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 200) }"))}`)
  await neuLaden(p)
  p.zeige(`Profilwechsel: ${JSON.stringify(await p.seite("try { const a = await window.go.app.Bindungen.ProfilWechseln('1', ''); return a.fehler ?? 'gewechselt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 200) }"))}`)
  await neuLaden(p)
  // **Die Rolle steht unter dem Profil in der Leiste** — wie „Verwalter".
  const rolle = await p.seite("return document.body.innerText.match(/\\bLeser · nur lesen\\b|\\bLeser\\b/)?.[0] ?? ''")
  p.pruefe('nur lesen: das Fenster nennt die Leserrolle', rolle !== '', rolle)
  const leseSet = await setOeffnen(p, set).then(() => true, () => false)
  for (const seite of alle) {
    if (set.includes(seite) && !leseSet) { zelle(seite, 'nur lesen', 'nicht erreicht', 'das Set ließ sich nicht öffnen'); continue }
    if (!(await zurSeite(p, seite))) { zelle(seite, 'nur lesen', 'nicht erreicht', 'die Seite kam nicht'); continue }
    const l = await beruhigt(p)
    // **Ein Schreibknopf ist ein Knopf mit einer schreibenden Handlung** —
    // keine Lasche: „Ersetzen" ist eine Seite, kein Schreiben (vierter Lauf).
    const aktiv = await p.seite("return [...document.querySelectorAll('main button:not([role=tab])')].filter((b) => b.checkVisibility() && !b.disabled && b.getAttribute('aria-disabled') !== 'true' && /\\b(anlegen|löschen|speichern|übernehmen|einlesen|hinzufügen)\\b|^\\+/i.test(b.innerText.trim())).map((b) => b.innerText.trim()).slice(0, 4)")
    await p.bild(`lesen-${seite}`)
    zelle(seite, 'nur lesen', aktiv.length ? 'Schreibknöpfe aktiv' : 'ohne aktiven Schreibknopf', aktiv.length ? aktiv.join(', ') : (l.alarm.join(' / ') || l.titel))
  }
  p.pruefe('nur lesen: ⌘Q beendet', await p.beende())


  p.notiz('Die Zellen')
  p.zeilen.push('| Seite | Zustand | Urteil | wie |', '| --- | --- | --- | --- |', ...zellen.map((z) => `| ${z.seite} | ${z.zustand} | ${z.urteil} | ${z.wie.replaceAll('|', '/')} |`))
  const gezaehlt = zellen.length
  p.pruefe(`${String(gezaehlt)} Zellen — jede mit Urteil oder Grund`, gezaehlt === 163)
}
