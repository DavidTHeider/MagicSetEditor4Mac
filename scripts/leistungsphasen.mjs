// **Die Teile des Leistungsvergleichs** — QA-04.AK-01 bis AK-04
// (#1735).
//
// Jeder Teil fährt seine Aufgabe für jeden Teilnehmer, je Wiederholung in
// gedrehter Reihenfolge (`reihenfolge`), und legt Fälle oder Schritte in den
// Beleg. **Geprüft wird beim Messen**: ein Fall, dessen Ergebnis nicht stimmt,
// trägt seinen Befund und zählt als Fehler.

import { createHash } from 'node:crypto'
import { existsSync, readFileSync, readdirSync, rmSync, statSync } from 'node:fs'
import { join } from 'node:path'

import { pngPruefen } from './pngpruefung.mjs'
import { reihenfolge } from './leistungsstatistik.mjs'

/** kartenImSet zählt die Karten einer gesicherten Setdatei — ohne sie zu deuten. */
export function kartenImSet(lesen, pfad) {
  if (!existsSync(pfad)) return null
  const text = lesen(pfad)
  return text === null ? null : (text.match(/^card:/gm) ?? []).length
}

/**
 * bildpruefung sagt, ob eine Datei ein **vollständiges** PNG ist, und nennt
 * Größe, Bytes und Hash. Geprüft wird die ganze Datei (`pngPruefen`), nicht der
 * Anfang: ein abgeschnittenes Bild hat dieselben ersten 24 Byte. **Maße gibt es
 * nur für ein heiles Bild** — wer sie vergleicht, vergleicht keine aus einer
 * Datei, die nicht stimmt.
 */
export function bildpruefung(pfad) {
  if (!existsSync(pfad)) return { da: false }
  const daten = readFileSync(pfad)
  const p = pngPruefen(daten)
  return {
    da: true, png: p.ok, ...(p.ok ? { breite: p.breite, hoehe: p.hoehe } : {}), fehler: p.ok ? null : p.fehler,
    bytes: daten.length, sha256: createHash('sha256').update(daten).digest('hex'),
  }
}

/**
 * bilderImVerzeichnis prüft jedes PNG eines Verzeichnisses und zählt die
 * Größen. `dateien` gibt die Maße je Datei, damit sich Karte für Karte
 * vergleichen lässt (`leistungsmasse.mjs`); `kaputteDateien` nennt die ersten
 * fünf, die nicht heil sind, samt Grund.
 */
export function bilderImVerzeichnis(verzeichnis) {
  const groessen = {}
  const dateien = {}
  const kaputteDateien = []
  let anzahl = 0
  let kaputt = 0
  for (const name of existsSync(verzeichnis) ? readdirSync(verzeichnis).sort() : []) {
    if (!name.endsWith('.png')) continue
    anzahl++
    const p = bildpruefung(join(verzeichnis, name))
    if (!p.png) {
      kaputt++
      if (kaputteDateien.length < 5) kaputteDateien.push({ datei: name, fehler: p.fehler })
    } else {
      groessen[`${p.breite}x${p.hoehe}`] = (groessen[`${p.breite}x${p.hoehe}`] ?? 0) + 1
      dateien[name] = { breite: p.breite, hoehe: p.hoehe }
    }
  }
  return { anzahl, kaputt, groessen, dateien, kaputteDateien }
}

/**
 * wiederhole fährt `einmal(teilnehmer, i)` für jede Wiederholung und jeden
 * Teilnehmer in gedrehter Reihenfolge und gibt die Fälle mit Platz und
 * Wiederholung zurück.
 *
 * **`vorher` wartet, bis der Rechner frei ist**, und gibt die Sekunden, die
 * es gewartet hat — sie stehen als `gewartetS` am Fall. Eine Pipeline braucht
 * die Maschine allein (`docs/PIPELINE.md`), und ein Fall, der neben ihr
 * gemessen wird, misst sie mit.
 */
export async function wiederhole(teilnehmer, anzahl, aufgabe, satz, einmal, melde, vorher = async () => 0) {
  const faelle = []
  for (let i = 0; i < anzahl; i++) {
    const folge = reihenfolge(i, teilnehmer)
    for (const [platz, t] of folge.entries()) {
      const gewartetS = await vorher()
      const m = await einmal(t, i)
      const { aus, ...ohneAusgabe } = m
      const fall = { aufgabe, satz, programm: t.name, wiederholung: i, platz, gewartetS, ...ohneAusgabe }
      faelle.push(fall)
      melde?.(fall)
    }
  }
  return faelle
}

/**
 * sitzungsschritte zeichnet die Auswahl eines Satzes in einer Sitzung: die
 * erste Karte **kalt**, jede weitere **warm**, und die erste am Ende noch
 * einmal als **Wiederholung**. Jedes Bild wird geprüft und danach gelöscht.
 */
export async function sitzungsschritte(sitzung, karten, verzeichnis, kennung) {
  const schritte = []
  const folge = [...karten.map((k, i) => ({ k, zustand: i === 0 ? 'kalt' : 'warm' })),
    { k: karten[0], zustand: 'wiederholung' }]
  for (const [n, { k, zustand }] of folge.entries()) {
    const ziel = join(verzeichnis, `${kennung}-${n}-${k.stelle}.png`)
    const antwort = await sitzung.zeichne(k.stelle, ziel)
    const bild = bildpruefung(ziel)
    rmSync(ziel, { force: true })
    schritte.push({
      zustand, schicht: k.schicht, satz: k.satz, stelle: k.stelle, vorlage: k.vorlage,
      dauerMs: antwort.dauerMs, ok: antwort.ok && bild.png === true,
      breite: bild.breite ?? null, hoehe: bild.hoehe ?? null, bytes: bild.bytes ?? null,
      befund: antwort.ok ? (bild.png ? null : bild.da ? `kein vollständiges PNG am Ziel: ${bild.fehler}` : 'kein PNG am Ziel') : antwort.text.slice(0, 300),
    })
  }
  return schritte
}

/** groesseInMb gibt eine Dateigröße in MB — für den Beleg. */
export function groesseInMb(pfad) {
  return existsSync(pfad) ? Math.round(statSync(pfad).size / 1e5) / 10 : null
}

/**
 * quelltextwert liest eine Zeichenkette aus MSE2s Antwort im Rohmodus — dort
 * steht `toCode()`, also Quelltext: in Anführungszeichen, mit `\"` und `\\`
 * (`src/cli/cli_main.cpp`, `handleCommand`: `result->toCode()`, Fassung
 * f9d9356 vom 2021-07-06). **Die
 * äußeren Zeichen gehören zur Schreibweise, die inneren zum Namen** — *"Ach!
 * Hans, Run!"* heißt wirklich so, mit Anführungszeichen.
 */
export function quelltextwert(text) {
  if (!text.startsWith('"') || !text.endsWith('"') || text.length < 2) return text
  return text.slice(1, -1).replace(/\\(["\\])/g, '$1')
}
