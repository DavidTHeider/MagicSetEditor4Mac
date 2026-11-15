// **Aus den Fällen eines Laufs werden Zahlen und ein Bericht** — QA-04.AK-03
// (#1735).
//
// Reine Funktionen; geprüft in `scripts/leistungsbericht.test.mjs`. Der
// Bericht ist eine **Lesehilfe** über dem Beleg — die Zahlen stehen im JSON,
// und wer eine davon zitiert, zitiert sie von dort.
//
// # Die erste Wiederholung steht gesondert
//
// **Den Seitenspeicher des Systems kann dieser Lauf nicht leeren** — `purge`
// braucht Verwalterrechte. Die erste Wiederholung einer Aufgabe ist deshalb die
// einzige, die eine Datei womöglich noch von der Platte liest; sie steht als
// `ersteWandMs` neben der Verteilung über alle.

import { reihenfolgeEffekt, verteilung } from './leistungsstatistik.mjs'

/** schluessel fasst Aufgabe und Satz zusammen — ohne Satz bleibt die Aufgabe. */
function schluessel(fall) {
  return fall.satz ? `${fall.aufgabe} · ${fall.satz}` : fall.aufgabe
}

/**
 * unvollstaendig sagt, ob ein Fall nicht getan hat, was er sollte: ein Fehler
 * oder ein Befund am Fall.
 */
const unvollstaendig = (fall) => Boolean(fall.fehler || fall.befund)

/**
 * zusammenfassen gibt je Aufgabe (samt Satz) und Programm die Kennzahlen.
 *
 * **Gerechnet wird nur über die Fälle ohne Fehler und ohne Befund.** Ein Fall,
 * der unvollständig war, hat keine Zeit, die man neben die eines vollständigen
 * stellt: im Beleg vom 2026-10-03 zeichnete MSE4Mac im Batchfall 24 von 70
 * Bildern, und seine 2,4 s standen neben den 3,0 s von MSE2 für 70. `faelle`
 * nennt, wie viele es waren, `fehler`, wie viele davon ausgeschlossen sind —
 * `wand.n` ist, über wie viele die Zeit gerechnet ist.
 */
export function zusammenfassen(faelle) {
  const ergebnis = {}
  for (const fall of faelle) {
    const zeile = (ergebnis[schluessel(fall)] ??= {})
    ;(zeile[fall.programm] ??= []).push(fall)
  }
  for (const zeile of Object.values(ergebnis)) {
    for (const [programm, eigene] of Object.entries(zeile)) {
      const heil = eigene.filter((f) => !unvollstaendig(f))
      const erste = heil.length > 0 ? heil.reduce((a, b) => (b.wiederholung < a.wiederholung ? b : a)) : null
      zeile[programm] = {
        faelle: eigene.length,
        wand: verteilung(heil.map((f) => f.wandMs)),
        ersteWandMs: erste?.wandMs ?? null,
        cpu: verteilung(heil.map((f) => f.cpuS).filter((w) => w !== null && w !== undefined)),
        rss: verteilung(heil.map((f) => f.rssByte).filter((w) => w !== null && w !== undefined)),
        fehler: eigene.filter(unvollstaendig).length,
        abstuerze: eigene.filter((f) => f.absturz).length,
        reihenfolge: reihenfolgeEffekt(heil.map((f) => ({ programm, platz: f.platz, wert: f.wandMs })))[programm]
          ?? { vorn: null, dahinter: null, nVorn: 0, nDahinter: 0 },
      }
    }
  }
  return ergebnis
}

/**
 * bilanz zählt, was der Lauf gefunden hat — **Befunde** und **Fehler getrennt**.
 *
 * Ein Befund ist, was an der Messung nicht stimmt: eine Gegenprobe zwischen den
 * Programmen, die nicht aufgeht, oder ein Fall, dessen Ergebnis nicht dem
 * entspricht, was verlangt war (zu wenige Bilder, andere Größe). Ein Fehler ist,
 * was ein Programm selbst meldet und der Beleg zählt, statt es zu verschlucken —
 * ein Rückgabewert ungleich null, ein Schritt mit Fehlersatz. **Hier stand
 * „null Befunde“ unter einem Lauf, dessen Batchfall 24 statt 70 Bilder
 * zeichnete**: gezählt war nur, was der Lauf als Ganzes anmerkte, nicht, was
 * seine Fälle trugen.
 */
export function bilanz({ faelle, schritte, befunde }) {
  const ausFaellen = faelle.filter((f) => f.befund)
  const beispiele = [
    ...befunde,
    ...ausFaellen.map((f) => `${f.aufgabe}${f.satz ? ` ${f.satz}` : ''} ${f.programm} #${f.wiederholung}: ${f.befund}`),
  ]
  return {
    faelle: faelle.length, schritte: schritte.length,
    befunde: beispiele.length, uebergreifend: befunde.length, ausFaellen: ausFaellen.length,
    faelleMitFehler: faelle.filter((f) => f.fehler && !f.befund).length,
    schritteMitFehler: schritte.filter((st) => st.ok === false).length,
    beispiele: beispiele.slice(0, 10),
  }
}

/** bilanzZeile schreibt die Bilanz als einen Satz — für den Lauf und den Bericht. */
export function bilanzZeile(b) {
  const fehler = []
  if (b.faelleMitFehler > 0) fehler.push(`${b.faelleMitFehler} ${b.faelleMitFehler === 1 ? 'Fall' : 'Fälle'}`)
  if (b.schritteMitFehler > 0) fehler.push(`${b.schritteMitFehler} ${b.schritteMitFehler === 1 ? 'Schritt' : 'Schritte'}`)
  return `${b.faelle} Fälle, ${b.schritte} Schritte, ${b.befunde} ${b.befunde === 1 ? 'Befund' : 'Befunde'}` +
    (fehler.length > 0 ? `, mit Fehler: ${fehler.join(' und ')}` : '')
}

/** schritteZusammenfassen gibt je Programm, Zustand und Schicht die Dauer der Schritte. */
export function schritteZusammenfassen(schritte) {
  const gruppen = {}
  for (const schritt of schritte) {
    ;(gruppen[`${schritt.programm} · ${schritt.zustand} · ${schritt.schicht}`] ??= []).push(schritt)
  }
  const ergebnis = {}
  for (const [name, eigene] of Object.entries(gruppen)) {
    ergebnis[name] = {
      dauer: verteilung(eigene.map((s) => s.dauerMs)),
      fehler: eigene.filter((s) => !s.ok).length,
    }
  }
  return ergebnis
}

/** zahl schreibt eine Zahl deutsch, mit höchstens so vielen Stellen. */
function zahl(wert, stellen = 1, mindestens = 0) {
  if (wert === null || wert === undefined || Number.isNaN(wert)) return '—'
  return wert.toLocaleString('de-DE', { maximumFractionDigits: stellen, minimumFractionDigits: mindestens })
}

/** berichtMarkdown schreibt die Tabellen des Berichts. */
export function berichtMarkdown(beleg) {
  const zeilen = [
    '# Leistungsvergleich — Bericht',
    '',
    `**Stand \`${beleg.commit}\`**, Lauf ab ${beleg.anfang}${beleg.maschine ? `, ${beleg.maschine.chip}, ${String(beleg.maschine.kerne)} Kerne, ${beleg.maschine.system}` : ''}. Erzeugt von`,
    '`scripts/leistungsvergleich.mjs`; die Zahlen stehen in',
    '`leistungsvergleich.json`, die Bewertung in `docs/LEISTUNGSVERGLEICH.md`.',
    '',
    '## Ganze Prozesse',
    '',
    'Wandzeit in Millisekunden, CPU (`user` + `sys`) in Sekunden, residenter',
    'Höchststand in MB (10⁶ Byte). *vorn / dahinter*: Median der Läufe, in denen',
    'das Programm zuerst lief, gegen den der übrigen.',
    '',
    '| Aufgabe | Programm | n | Median | p95 | p99 | erste | CPU | RSS | Fehler | Abstürze | vorn / dahinter |',
    '| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |',
  ]
  for (const [aufgabe, programme] of Object.entries(beleg.zusammenfassung)) {
    for (const [programm, z] of Object.entries(programme)) {
      const r = z.reihenfolge
      zeilen.push(`| ${aufgabe} | ${programm} | ${z.wand.n} | ${zahl(z.wand.median)} | ${zahl(z.wand.p95)} | ` +
        `${zahl(z.wand.p99)} | ${zahl(z.ersteWandMs)} | ${zahl(z.cpu.median, 2)} | ${zahl(z.rss.median / 1e6, 1, 1)} | ` +
        `${z.fehler} | ${z.abstuerze} | ${zahl(r.vorn)} / ${zahl(r.dahinter)} |`)
    }
  }
  zeilen.push('', '## Schritte in einer laufenden Sitzung', '',
    'Je Schritt von der geschriebenen Zeile bis zur Antwort, in Millisekunden.', '',
    '| Programm · Zustand · Schicht | n | Median | p95 | p99 | max | Fehler |',
    '| --- | ---: | ---: | ---: | ---: | ---: | ---: |')
  for (const [name, z] of Object.entries(beleg.schritte)) {
    zeilen.push(`| ${name} | ${z.dauer.n} | ${zahl(z.dauer.median)} | ${zahl(z.dauer.p95)} | ${zahl(z.dauer.p99)} | ` +
      `${zahl(z.dauer.max)} | ${z.fehler} |`)
  }
  if (beleg.langlauf) zeilen.push(...dauerbetrieb(beleg.langlauf))
  if (beleg.bilanz) zeilen.push('', '## Bilanz', '', bilanzZeile(beleg.bilanz), ...beleg.bilanz.beispiele.map((b) => `- ${b}`))
  return zeilen.join('\n') + '\n'
}

/**
 * dauerbetrieb schreibt die Tabelle des Langlaufs. **Versuche, gelungene Bilder
 * und Fehler stehen getrennt**, und ebenso die Zeit: das Bild und die
 * Ablehnung haben je ihre Verteilung.
 */
function dauerbetrieb(langlauf) {
  const mb = (byte) => zahl(byte / 1e6, 0)
  const zeilen = ['', '## Dauerbetrieb', '',
    'Eine Sitzung je Programm zeichnet reihum Karten; **Versuche** sind alle Rufe, **gelungen** und',
    '**Fehler** ihre Teile. Die Zeit in Millisekunden gilt für die entstandenen Bilder, die der',
    'Ablehnungen steht daneben. *Anfang → Ende*: Median im ersten und im letzten Fünftel der Bilder.',
    '', '| Programm | Versuche | gelungen | Fehler | Bild Median | Bild p95 | Ablehnung Median | Anfang → Ende | RSS MB Anfang → Höchststand → Ende |',
    '| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |']
  for (const [programm, l] of Object.entries(langlauf)) {
    const rss = (l.speicher ?? []).map((m) => m.byte).filter((b) => b !== null && b !== undefined)
    const verlauf = rss.length > 0 ? `${mb(rss[0])} → ${mb(Math.max(...rss))} → ${mb(rss.at(-1))}` : '—'
    zeilen.push(`| ${programm} | ${zahl(l.versuche, 0)} | ${zahl(l.gelungen, 0)} | ${zahl(l.fehler, 0)} | ` +
      `${zahl(l.dauer?.median)} | ${zahl(l.dauer?.p95)} | ${zahl(l.ablehnung?.median)} | ` +
      `${zahl(l.anfang?.median)} → ${zahl(l.schluss?.median)} | ${verlauf} |`)
  }
  return zeilen
}

/**
 * belegJson schreibt den Beleg mit **einer Zeile je Eintrag** auf zweiter
 * Ebene — ein Fall, ein Schritt, eine Zeile der Zusammenfassung. Eingerückt
 * bis ins Letzte wären es über zehntausend Zeilen, und der Umfang im README
 * zählte sie als Quelltext (`internal/werkzeuge/kennzahlen_test.go`).
 */
export function belegJson(beleg) {
  const wert = (v) => {
    if (Array.isArray(v)) {
      return v.length === 0 ? '[]' : `[\n${v.map((e) => `    ${JSON.stringify(e)}`).join(',\n')}\n  ]`
    }
    if (v !== null && typeof v === 'object') {
      const eintraege = Object.entries(v)
      return eintraege.length === 0 ? '{}'
        : `{\n${eintraege.map(([s, e]) => `    ${JSON.stringify(s)}: ${JSON.stringify(e)}`).join(',\n')}\n  }`
    }
    return JSON.stringify(v)
  }
  return `{\n${Object.entries(beleg).map(([s, v]) => `  ${JSON.stringify(s)}: ${wert(v)}`).join(',\n')}\n}\n`
}
