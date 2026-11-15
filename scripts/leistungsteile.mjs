// **Die Teile des Leistungsvergleichs** — QA-04.AK-01 bis AK-04
// (#1735).
//
// Je Teil eine Funktion über dem Kontext des Läufers (`c`). Fälle sind ganze
// Prozesse, Schritte sind Befehle in einer laufenden Sitzung. **Geprüft wird
// beim Messen**: wo das Ergebnis nicht stimmt, steht ein Befund am Fall, und
// er zählt als Fehler.

import { createHash } from 'node:crypto'
import { existsSync, readFileSync, rmSync } from 'node:fs'
import { join } from 'node:path'

import { schrittmasseVergleichen, vergleicheBilderpaare } from './leistungsmasse.mjs'
import { reihenfolge } from './leistungsstatistik.mjs'

const mitBefund = (m, befund) => ({ ...m, befund, fehler: m.fehler ?? befund })

/**
 * ergaenzeBefund hängt einen Befund an einen Fall, der schon gemessen ist — er
 * zählt dann als Fehler und fällt aus dem Vergleich der Zeiten
 * (`zusammenfassen`).
 */
function ergaenzeBefund(fall, text) {
  fall.befund = fall.befund ? `${fall.befund}; ${text}` : text
  fall.fehler = fall.fehler ?? fall.befund
}

/** verkuerzt nennt die ersten Texte und zählt den Rest. */
const verkuerzt = (texte, zeigen = 3) =>
  texte.slice(0, zeigen).join('; ') + (texte.length > zeigen ? `; … und ${texte.length - zeigen} weitere` : '')
const sha = (pfad) => (existsSync(pfad) ? createHash('sha256').update(readFileSync(pfad)).digest('hex') : null)

/** start: Prozess bis zur ersten Antwort — MSE4Mac an einem leeren, schon angelegten Bestand. */
export async function start(c) {
  c.faelle.push(...await c.wiederhole(c.teilnehmer, c.k.wiederholungen, 'start', '', (t) => t.start(t.leer), c.melde))
}

/** oeffnenSatz öffnet einen Satz je Wiederholung — MSE4Mac in einen frischen Bestand. */
async function oeffnenSatz(c, name, anzahl, aufgabe) {
  const erwartet = c.kartenzahl[name]
  return c.wiederhole(c.teilnehmer, anzahl, aufgabe, name, async (t, i) => {
    const bestand = t.name === 'mse2' ? null : c.bestandsort(t, `${aufgabe}-${name}-${i}`)
    const m = await t.oeffnen(c.satz(name), bestand)
    if (bestand) t[`letzter-${name}`] = { bestand, id: m.setID }
    return mitBefund(m, m.karten === erwartet ? null : `${m.karten} statt ${erwartet} Karten`)
  }, c.melde)
}

/** oeffnen: ein kleiner Satz; dazu Wiederöffnen aus dem Bestand, das es nur bei MSE4Mac gibt. */
export async function oeffnen(c) {
  c.faelle.push(...await oeffnenSatz(c, 'Archenemy', c.k.wiederholungen, 'oeffnen'))
  c.faelle.push(...await wiederoeffnen(c, 'Archenemy', c.k.wiederholungen, (t) => ({ bestand: t.vorrat, id: t.ids.Archenemy })))
}

/** wiederoeffnen liest die Karten eines eingefahrenen Satzes erneut. */
async function wiederoeffnen(c, name, anzahl, ort) {
  return c.wiederhole(c.mse4macs, anzahl, 'wiederoeffnen', name, async (t) => {
    const { bestand, id } = ort(t)
    const m = await t.wiederoeffnen({ ...c.satz(name), id }, bestand)
    return mitBefund(m, m.karten === c.kartenzahl[name] ? null : `${m.karten} statt ${c.kartenzahl[name]} Karten`)
  }, c.melde)
}

/** sichernSatz schreibt einen Satz je Wiederholung und zählt die Karten im Ergebnis. */
async function sichernSatz(c, name, anzahl, ort) {
  const aus = c.ausgabe(`sichern-${name}`)
  const behalten = {}
  const faelle = await c.wiederhole(c.teilnehmer, anzahl, 'sichern', name, async (t, i) => {
    const ziel = join(aus, `${t.name}-${i}.mse-set`)
    const { bestand, id } = ort(t)
    const m = await t.sichern({ ...c.satz(name), id }, ziel, bestand)
    const karten = c.kartenImSet(c.setLesen, ziel)
    if (behalten[t.name]) rmSync(behalten[t.name], { recursive: true, force: true })
    behalten[t.name] = ziel
    return mitBefund(m, karten === c.kartenzahl[name] ? null : `${karten} statt ${c.kartenzahl[name]} Karten in der Datei`)
  }, c.melde)
  return { faelle, behalten }
}

/**
 * kreuzprobe öffnet, was das eine Programm gesichert hat, im anderen — die
 * Unversehrtheit einer Datei zeigt sich daran, dass **der andere** sie liest.
 */
async function kreuzprobe(c, name, behalten) {
  const proben = []
  const mse2 = c.teilnehmer.find((t) => t.name === 'mse2')
  for (const t of c.mse4macs) {
    if (behalten[t.name]) {
      const m = await mse2.oeffnen({ name, pfad: behalten[t.name] })
      proben.push({ geschrieben: t.name, gelesen: 'mse2', satz: name, karten: m.karten, fehler: m.fehler })
    }
    if (behalten.mse2) {
      const m = await t.oeffnen({ name, pfad: behalten.mse2 }, c.bestandsort(t, `kreuz-${name}`))
      proben.push({ geschrieben: 'mse2', gelesen: t.name, satz: name, karten: m.karten, fehler: m.fehler })
    }
  }
  for (const p of proben) {
    if (p.karten !== c.kartenzahl[name]) c.befunde.push(`Kreuzprobe ${name}: ${p.gelesen} liest aus ${p.geschrieben} ${p.karten} Karten`)
  }
  ;(c.extra.kreuzproben ??= []).push(...proben)
}

/** sichern: ein kleiner Satz, und das Ergebnis im jeweils anderen Programm geöffnet. */
export async function sichern(c) {
  const { faelle, behalten } = await sichernSatz(c, 'Archenemy', c.k.wiederholungen, (t) => ({ bestand: t.vorrat, id: t.ids?.Archenemy }))
  c.faelle.push(...faelle)
  await kreuzprobe(c, 'Archenemy', behalten)
}

/** gross: Öffnen, Wiederöffnen und Sichern des größten Satzes — mit eigenem Budget. */
export async function gross(c) {
  c.faelle.push(...await oeffnenSatz(c, 'Black', c.k.gross, 'oeffnen'))
  c.faelle.push(...await wiederoeffnen(c, 'Black', c.k.gross, (t) => t['letzter-Black']))
  const { faelle, behalten } = await sichernSatz(c, 'Black', c.k.gross, (t) => t['letzter-Black'] ?? {})
  c.faelle.push(...faelle)
  await kreuzprobe(c, 'Black', behalten)
}

/** sitzungFuer startet die Sitzung eines Teilnehmers an einem Satz. */
const sitzungFuer = (t, c, name) => (t.name === 'mse2' ? t.sitzung(c.satz(name)) : t.sitzung({ ...c.satz(name), id: t.ids[name] }, t.vorrat))

/**
 * navigation: die Auswahl aus QA-02, je Satz in einer laufenden Sitzung —
 * kalt, warm und die erste Karte noch einmal. In der ersten Wiederholung wird
 * je Stelle der Kartenname beider Seiten gegen die Setdatei gehalten.
 */
export async function navigation(c) {
  const saetze = [...new Set(c.auswahl.map((k) => k.satz))]
  const aus = c.ausgabe('navigation')
  for (let i = 0; i < c.k.navigation; i++) {
    for (const [platz, t] of reihenfolge(i, c.teilnehmer).entries()) {
      for (const name of saetze) {
        const karten = c.auswahl.filter((k) => k.satz === name)
        const gewartetS = await c.warte()
        const s = await sitzungFuer(t, c, name)
        if (i === 0) {
          for (const k of karten) {
            const n = await s.name(k.stelle)
            const gelesen = t.name === 'mse2' ? c.quelltextwert(n.text ?? '') : n.text
            if (gelesen !== k.name) c.befunde.push(`${t.name}: ${name}[${k.stelle}] heißt ${JSON.stringify(n.text)}, die Setdatei sagt ${JSON.stringify(k.name)}`)
          }
        }
        const schritte = [{ zustand: 'bereit', schicht: 'Sitzung', satz: name, dauerMs: s.bereitMs, ok: true },
          ...await c.sitzungsschritte(s, karten, aus, `${t.name}-${i}`)]
        const ende = await s.lauf.schliessen()
        c.schritte.push(...schritte.map((x) => ({ ...x, programm: t.name, wiederholung: i, platz })))
        c.faelle.push({ aufgabe: 'navigation-sitzung', satz: name, programm: t.name, wiederholung: i, platz, gewartetS, pipelineDanach: c.pipelineLaeuft(), ...ohneAusgabe(ende) })
        console.log(`  navigation ${name} ${t.name} #${i}: ${schritte.map((x) => Math.round(x.dauerMs ?? -1)).join(' ')}`)
      }
    }
  }
  masseInSitzungen(c)
}

/**
 * masseInSitzungen hält die Bildmaße der Schritte gegen MSE2 — je Satz und
 * Stelle — und trägt eine Abweichung an die Sitzungen des Programms ein, in
 * denen sie vorkam (Z2). Eine Sitzung mit Befund fällt aus dem Vergleich der
 * Zeiten: sie hat nicht dasselbe gezeichnet.
 */
export function masseInSitzungen(c) {
  const je = new Map()
  for (const a of schrittmasseVergleichen(c.schritte)) {
    const schluessel = `${a.programm}\u0000${a.satz}`
    je.set(schluessel, [...(je.get(schluessel) ?? []), a])
  }
  for (const abweichungen of je.values()) {
    const [{ programm, satz }] = abweichungen
    const text = `Bildgröße gegen MSE2: ${verkuerzt(abweichungen.map((a) => a.text))}`
    for (const fall of c.faelle.filter((f) => f.aufgabe === 'navigation-sitzung' && f.programm === programm && f.satz === satz)) {
      ergaenzeBefund(fall, text)
    }
  }
}

const ohneAusgabe = ({ aus, ...rest }) => rest

/** skript: drei Ausdrücke je fünfmal in einer laufenden Sitzung; das Ergebnis muss auf beiden Seiten gleich sein. */
export async function skript(c) {
  const stelle = 3
  const ausdruecke = [
    { name: 'Rechnung', mse2: '1 + 1', mse4mac: '1 + 1', erwartet: '2' },
    { name: 'Zeichenkette', mse2: 'to_upper("abc")', mse4mac: 'to_upper("abc")', erwartet: '"ABC"' },
    { name: 'Kartenfeld', mse2: `set.cards[${stelle}].name`, mse4mac: 'card.name', erwartet: JSON.stringify(c.vorlagen.Archenemy[stelle].name) },
  ]
  for (let i = 0; i < c.k.navigation; i++) {
    for (const [platz, t] of reihenfolge(i, c.teilnehmer).entries()) {
      const gewartetS = await c.warte()
      const s = await sitzungFuer(t, c, 'Archenemy')
      for (const a of ausdruecke) {
        for (let n = 0; n < 5; n++) {
          const r = t.name === 'mse2' ? await s.skript(a.mse2) : await s.skript(a.mse4mac, stelle)
          c.schritte.push({ programm: t.name, zustand: n === 0 && a === ausdruecke[0] ? 'kalt' : 'warm', schicht: `Skript: ${a.name}`, dauerMs: r.dauerMs, ok: r.ok && r.text === a.erwartet, wiederholung: i, platz, befund: r.text === a.erwartet ? null : r.text })
        }
      }
      c.faelle.push({ aufgabe: 'skript-sitzung', satz: 'Archenemy', programm: t.name, wiederholung: i, platz, gewartetS, pipelineDanach: c.pipelineLaeuft(), ...ohneAusgabe(await s.lauf.schliessen()) })
    }
  }
}

/**
 * zeichneAlle lässt jede Karte der Sitzung zeichnen und gibt die Rufe, die
 * gescheitert sind. **Jeder Ruf wird abgewartet und geprüft**: im Beleg vom
 * 2026-10-03 wies das Guthaben von `mse-mcp` nach 24 Karten alles ab, und die
 * Schleife ignorierte jede Antwort.
 */
async function zeichneAlle(s, ziel) {
  const gescheitert = []
  for (let stelle = 0; stelle < s.karten.length; stelle++) {
    const r = await s.zeichne(stelle, join(ziel, `${stelle}.png`))
    if (!r.ok) gescheitert.push({ stelle, text: r.text ?? '' })
  }
  return gescheitert
}

/** bilderBefund sagt, was an den Bildern eines Falls nicht stimmt — `null`, wenn nichts. */
function bilderBefund({ pruefung, erwartet, gescheitert }) {
  const teile = []
  if (gescheitert.length > 0) {
    teile.push(`${gescheitert.length} von ${erwartet} Zeichenrufen gescheitert (Karte ${gescheitert[0].stelle}: ${gescheitert[0].text.slice(-200)})`)
  }
  if (pruefung.anzahl !== erwartet || pruefung.kaputt > 0) {
    const kaputt = pruefung.kaputteDateien[0]
    teile.push(`${pruefung.anzahl} Bilder, ${pruefung.kaputt} kaputt, bei ${erwartet} Karten` +
      (kaputt ? ` (${kaputt.datei}: ${kaputt.fehler})` : ''))
  }
  return teile.length > 0 ? teile.join('; ') : null
}

/**
 * bilder: jede Karte eines Satzes als PNG — MSE2 in einem Aufruf, MSE4Mac Karte
 * für Karte über mse-mcp. **Ein Fall ist nur vollständig, wenn jeder Ruf
 * gelang, jedes Bild heil ist und jede Karte dieselbe Breite und Höhe hat wie
 * bei MSE2** (Z1, Z2); sonst trägt er einen Befund und fällt aus dem Vergleich
 * der Zeiten.
 */
export async function bilder(c) {
  const name = 'Archenemy'
  const groessen = {}
  const dateien = {}
  const faelle = await c.wiederhole(c.teilnehmer, c.k.bilder, 'bilder', name, async (t, i) => {
    const ziel = c.ausgabe(`bilder-${t.name}-${i}`)
    let m
    let gescheitert = []
    if (t.name === 'mse2') {
      m = await t.bilder(c.satz(name), ziel)
    } else {
      const s = await sitzungFuer(t, c, name)
      gescheitert = await zeichneAlle(s, ziel)
      m = await s.lauf.schliessen()
    }
    const pruefung = c.bilderImVerzeichnis(ziel)
    groessen[t.name] ??= pruefung.groessen
    ;(dateien[t.name] ??= {})[i] = pruefung.dateien
    rmSync(ziel, { recursive: true, force: true })
    const befund = bilderBefund({ pruefung, erwartet: c.kartenzahl[name], gescheitert })
    return mitBefund({ ...m, bilder: pruefung.anzahl, durchsatz: pruefung.anzahl / (m.wandMs / 1000) }, befund)
  }, c.melde)
  c.faelle.push(...faelle)
  c.extra.bildgroessen = groessen

  // **Karte für Karte gegen MSE2**, je Wiederholung. Was bei den anderen fehlt,
  // steht schon in der Zahl ihrer Bilder; was bei MSE2 fehlt, ist ein Befund
  // über den Vergleich selbst (der Dateiname wird abgeleitet).
  const namen = c.vorlagen[name].map((k) => k.name)
  for (const t of c.mse4macs) {
    for (const [i, andere] of Object.entries(dateien[t.name] ?? {})) {
      const referenz = dateien.mse2?.[i]
      if (!referenz) continue
      const v = vergleicheBilderpaare({ namen, referenz, andere, andererName: (stelle) => `${stelle}.png` })
      const texte = [...v.abweichungen, ...v.ohneReferenz]
      const fall = faelle.find((f) => f.programm === t.name && f.wiederholung === Number(i))
      if (texte.length > 0 && fall) ergaenzeBefund(fall, `Maße gegen MSE2: ${verkuerzt(texte)}`)
    }
  }
}

/**
 * Die gepaarten Exportvorlagen — **jede, deren Ausgabe beide Programme
 * bytegleich schreiben sollen**; eine andere Ausgabe ist ein Befund.
 * `magic-text` und `magic-mws` ordnen ihre Karten mit `sort_list` nach der
 * Kartennummer und stehen hier seit #2006 (`docs/LEISTUNGSVERGLEICH.md` §6).
 */
export const ausfuhrvorlagen = ['magic-simple-exporter.mse-export-template', 'magic-namecheck.mse-export-template',
  'magic-text.mse-export-template', 'magic-mws.mse-export-template']

/** ausfuhr: die Exportvorlagen aus `ausfuhrvorlagen`, je mit Hashvergleich gegen MSE2. */
export async function ausfuhr(c) {
  const name = 'Archenemy'
  for (const vorlage of ausfuhrvorlagen) {
    const hashes = {}
    c.faelle.push(...await c.wiederhole(c.teilnehmer, c.k.wiederholungen, `ausfuhr ${vorlage.replace('.mse-export-template', '')}`, name, async (t, i) => {
      const ziel = join(c.ausgabe(`ausfuhr-${t.name}-${i}`), 'ausgabe.txt')
      const m = await t.ausfuhr({ ...c.satz(name), id: t.ids?.[name] }, vorlage, ziel, t.vorrat)
      const hash = sha(ziel)
      hashes[t.name] ??= hash
      return mitBefund(m, hash === null ? 'keine Ausgabe' : null)
    }, c.melde))
    const verschieden = Object.entries(hashes).filter(([, h]) => h !== hashes.mse2)
    for (const [programm] of verschieden) c.befunde.push(`${vorlage}: ${programm} schreibt eine andere Ausgabe als MSE2`)
    ;(c.extra.ausfuhrhashes ??= {})[vorlage] = hashes
  }
}

export { langlauf, last } from './leistungsdauer.mjs'
