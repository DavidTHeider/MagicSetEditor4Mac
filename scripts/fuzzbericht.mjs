// **Was eine Fuzzkampagne aus ihren Läufen macht** — Z5 des Zweitblicks über
// Phase 6.11 (QA-06, #1737).
//
// Reine Auswertung; `scripts/fuzzkampagne.mjs` fährt, hier wird gelesen und
// geurteilt, und geprüft ist es in `scripts/fuzzbericht.test.mjs`.
//
// # Drei Teile, und jeder ist erforderlich
//
//   1. der **Fuzzlauf** (`go test -fuzz`),
//   2. die **Übernahme des erarbeiteten Korpus** in den Baum, auf Zeit,
//   3. der **Abdeckungslauf** über das ganze Korpus.
//
// Bis dahin zählte nur der erste: der Rückgabestatus der beiden anderen wurde
// nicht gelesen, ein fehlender Prozentwert wurde zu „–“, und der Exitcode
// hing allein von Funden ab. **Scheiterte die Instrumentierung nach einem
// grünen Fuzzlauf, stand „0 mit Fund“ im Bericht, der Anteil war leer, und der
// Lauf endete mit 0.** Ein Teil, der fehlt, macht das Ziel `unvollständig` und
// die Kampagne rot — mit dem Grund beim Ziel.
//
// # Ein Fund ist etwas anderes als ein Fehler
//
// **Ein Fund** ist eine Eingabe, die das Ziel zum Scheitern bringt: Go
// speichert sie und meldet `Failing input written to …`. **Ein gescheiterter
// Teil** ist ein Lauf, der nicht tat, was er sollte — ein Übersetzungsfehler,
// ein Signal, ein Abdeckungslauf ohne Zahl. Beides macht die Kampagne rot,
// aber nicht gleich: Exitcode 1 für einen Fund, 3 für einen gescheiterten Teil,
// und der Fund wiegt schwerer.
//
// **Mit Fund darf der Abdeckungslauf rot enden**: die gespeicherte Eingabe
// liegt unter `testdata/fuzz/<Ziel>` und lässt `-run ^<Ziel>$` scheitern, das
// ist ihr Zweck. Er zählt dann nicht als zweiter Fehler.

import { copyFileSync, existsSync, mkdirSync, readdirSync, rmSync } from 'node:fs'
import { join } from 'node:path'

/** ende gibt das Ende einer Ausgabe — dort steht, woran es lag. */
const ende = (text, zeichen = 300) => text.trim().slice(-zeichen)

/**
 * fuzzlaufAuswerten liest das Ergebnis von `go test -fuzz` (ein `spawnSync`-
 * Ergebnis): `ohne Fund`, `Fund` oder `gescheitert` — mit den Zahlen der
 * letzten Fortschrittszeile (`fuzz: elapsed: 1m0s, execs: 123 (2/sec), new
 * interesting: 4 (total: 9)`).
 */
export function fuzzlaufAuswerten(lauf) {
  const ausgabe = `${lauf.stdout ?? ''}${lauf.stderr ?? ''}`
  const zeilen = [...ausgabe.matchAll(/elapsed: (\S+?), execs: (\d+) .*?new interesting: (\d+) \(total: (\d+)\)/g)]
  const letzte = zeilen.at(-1)
  // **Der Verlauf, nicht nur die letzte Zeile** — Zweitblick über Phase C,
  // Befund 4: wann zuletzt eine neue Eingabe kam, lässt sich aus Endzahlen
  // nicht nachholen. Je Zeile Sekunden, Ausführungen, neue Eingaben.
  const verlauf = zeilen.map((z) => [dauerSekunden(z[1]), Number(z[2]), Number(z[3])])
  // zuletztNeu ist die Sekunde, in der die Zahl der neuen Eingaben ihren
  // Endstand erreichte — 0, wenn nie eine kam, `null` ohne Verlauf.
  const endstand = verlauf.at(-1)?.[2]
  const zuletztNeu = endstand === undefined ? null : endstand === 0 ? 0 : verlauf.find((v) => v[2] === endstand)[0]
  const fund = (ausgabe.match(/Failing input written to (\S+)/) ?? [])[1] ?? ''
  const grund = (ausgabe.match(/--- FAIL: \S+ \([^)]*\)\n\s+(?:--- FAIL: \S+ \([^)]*\)\n\s+)?(.+)/) ?? [])[1] ?? ''
  const zahlen = {
    ausfuehrungen: letzte ? Number(letzte[2]) : 0, neu: letzte ? Number(letzte[3]) : 0, korpus: letzte ? Number(letzte[4]) : 0,
    fund, grund: grund.trim(), verlauf, zuletztNeu,
  }
  const gescheitert = (fehler) => ({ ausgang: 'gescheitert', fehler, ...zahlen })

  if (lauf.error) return gescheitert(`Fuzzlauf nicht gestartet: ${lauf.error.message}`)
  if (lauf.status === null || lauf.status === undefined) return gescheitert(`Fuzzlauf von ${lauf.signal ?? 'einem Signal'} beendet`)
  if (lauf.status === 0) {
    if (/no fuzz tests to fuzz/.test(ausgabe)) return gescheitert('kein Fuzzziel gefunden: go test meldet „no fuzz tests to fuzz“')
    return { ausgang: 'ohne Fund', fehler: null, ...zahlen }
  }
  if (fund) return { ausgang: 'Fund', fehler: null, ...zahlen }
  return gescheitert(`Fuzzlauf endete mit Status ${lauf.status} ohne gespeicherte Eingabe: ${ende(ausgabe)}`)
}

/**
 * abdeckungAuswerten liest den Lauf über das ganze Korpus. **Ohne Fund ist ein
 * Status ungleich null ein Fehler, und eine fehlende Zahl ist es auch** — nie
 * ein stiller Strich.
 */
export function abdeckungAuswerten(lauf, fund) {
  const ausgabe = `${lauf.stdout ?? ''}${lauf.stderr ?? ''}`
  const prozent = (`${lauf.stdout ?? ''}`.match(/coverage: ([\d.]+)% of statements/) ?? [])[1] ?? null
  if (!fund) {
    if (lauf.error) return { prozent: null, fehler: `Abdeckungslauf nicht gestartet: ${lauf.error.message}` }
    if (lauf.status !== 0) return { prozent: null, fehler: `Abdeckungslauf endete mit Status ${lauf.status ?? lauf.signal}: ${ende(ausgabe)}` }
    if (prozent === null) return { prozent: null, fehler: `keine Abdeckungszahl in der Ausgabe des Abdeckungslaufs: ${ende(ausgabe)}` }
  }
  return { prozent, fehler: null }
}

/**
 * anteilMessen misst, welchen Anteil der Anweisungen seines Pakets das Korpus
 * eines Ziels betritt: das erarbeitete Korpus (`$(go env GOCACHE)/fuzz/…`) wird
 * auf Zeit in den Baum kopiert, `goTest` fährt den Lauf, und die Kopien gehen
 * danach wieder — **auch bei einem Fehler**. Gibt `{ prozent, korpusdateien,
 * fehler }`; jeder gescheiterte Teil steht in `fehler`, der Lauf geschieht
 * trotzdem, damit der Rest sichtbar bleibt.
 */
export function anteilMessen(ziel, { wurzel, zwischen, modul, fund, goTest }) {
  const erarbeitet = join(zwischen, 'fuzz', `${modul}/${ziel.paket}`, ziel.name)
  const eingecheckt = join(wurzel, ziel.paket, 'testdata', 'fuzz', ziel.name)
  const kopiert = []
  const fehler = []

  try {
    if (existsSync(erarbeitet)) {
      mkdirSync(eingecheckt, { recursive: true })
      try {
        for (const datei of readdirSync(erarbeitet)) {
          const nach = join(eingecheckt, `kampagne-${datei}`)
          kopiert.push(nach)
          copyFileSync(join(erarbeitet, datei), nach)
        }
      } catch (grund) {
        fehler.push(`Korpusübernahme gescheitert: ${grund.message}`)
      }
    }

    const profil = join(zwischen, `kampagne-${ziel.name}.out`)
    const lauf = goTest(['test', `./${ziel.paket}`, '-run', `^${ziel.name}$`, '-count', '1',
      '-covermode', 'set', '-coverprofile', profil, '-coverpkg', `./${ziel.paket}`])
    const gemessen = abdeckungAuswerten(lauf, fund)
    if (gemessen.fehler) fehler.push(gemessen.fehler)

    return { prozent: gemessen.prozent, korpusdateien: kopiert.length, fehler }
  } finally {
    for (const datei of kopiert) rmSync(datei, { force: true })
  }
}

/**
 * zielErgebnis fügt die Teile eines Ziels zusammen. `ohne Fund` gibt es nur,
 * wenn **jeder** Teil gelang; ein Fund bleibt ein Fund.
 */
export function zielErgebnis(ziel, fuzz, anteil, sekunden) {
  const fehler = [...(fuzz.fehler ? [fuzz.fehler] : []), ...anteil.fehler]
  const ausgang = fuzz.ausgang === 'Fund' ? 'Fund' : fehler.length > 0 ? 'unvollständig' : 'ohne Fund'
  return {
    ...ziel, sekunden, ausgang, gruen: ausgang === 'ohne Fund', anteil: anteil.prozent ?? '–', korpusdateien: anteil.korpusdateien,
    ausfuehrungen: fuzz.ausfuehrungen, neu: fuzz.neu, korpus: fuzz.korpus, fund: fuzz.fund, grund: fuzz.grund, fehler,
    verlauf: fuzz.verlauf, zuletztNeu: fuzz.zuletztNeu,
  }
}

/** dauerSekunden liest eine Dauer, wie Go sie schreibt (`3s`, `1m0s`, `1h2m3s`, `10m`). */
export function dauerSekunden(text) {
  const [, h = '0', m = '0', sek = '0'] = /^(?:(\d+)h)?(?:(\d+)m)?(?:(\d+(?:\.\d+)?)s)?$/.exec(text) ?? []
  return Number(h) * 3600 + Number(m) * 60 + Number(sek)
}

/**
 * saettigung urteilt über ein Ziel — **gesättigt, wenn im letzten Viertel
 * seines Budgets keine neue Eingabe kam**. Festgelegt vor der Kampagne (C02,
 * #2227); ein Maß für „hier kommt in dieser Zeit wenig mehr", kein Beweis,
 * dass nichts mehr zu finden wäre.
 */
//
// **Nur über ein voll beobachtetes Budget** (Nachlauf über Phase C): ein Lauf
// mit Fund oder einer, dessen letzte Meldung früher kam, ist „nicht
// bewertbar“. Die Meldungen kommen alle drei Sekunden; fünf Sekunden
// Spielraum lassen die letzte vor dem Ende gelten.
export function saettigung(e, budget) {
  if (e.zuletztNeu === null || e.zuletztNeu === undefined || e.verlauf?.length === 0) return 'ohne Verlauf'
  const beobachtet = e.verlauf?.at(-1)?.[0] ?? 0
  if (e.ausgang === 'Fund' || beobachtet + 5 < dauerSekunden(budget)) return 'nicht bewertbar'
  return e.zuletztNeu <= 0.75 * dauerSekunden(budget) ? 'gesättigt' : 'wächst noch'
}

/**
 * standPruefen vergleicht den Stand vor und nach dem Lauf — Zweitblick über
 * Phase C, Befund 4: der Commit wurde erst am Ende gelesen, und ungesicherte
 * Änderungen zählten nicht. `offen` ist die größere Zahl der beiden.
 *
 * **Und der Inhalt** (Nachlauf über Phase C): `inhalt` ist ein Abdruck der
 * ungesicherten Änderungen samt neuer Dateien, ohne `testdata/fuzz`. `saat`
 * ist der Abdruck des Korpus dort — er steht im Kopf und wird **nicht**
 * verglichen: ein Fund legt seine Eingabe selbst dort ab.
 */
export function standPruefen(anfang, ende) {
  return {
    commit: anfang.commit, geaendert: anfang.commit !== ende.commit || anfang.inhalt !== ende.inhalt,
    offen: Math.max(anfang.offen, ende.offen), inhalt: anfang.inhalt, saat: anfang.saat,
  }
}

/** kampagnenstatus zählt Funde und unvollständige Ziele und gibt den Exitcode: 0, 1 (Fund) oder 3 (Teil gescheitert). */
export function kampagnenstatus(ergebnisse) {
  const funde = ergebnisse.filter((e) => e.ausgang === 'Fund')
  const unvollstaendig = ergebnisse.filter((e) => e.ausgang === 'unvollständig')
  return { funde, unvollstaendig, exitcode: funde.length > 0 ? 1 : unvollstaendig.length > 0 ? 3 : 0 }
}

const tausender = (n) => n.toLocaleString('de-DE')
const zelle = (text) => String(text).replaceAll('|', '\\|').replaceAll('\n', ' ')

/** berichtText schreibt den Bericht als Markdown. */
export function berichtText(ergebnisse, kopf) {
  const { funde, unvollstaendig } = kampagnenstatus(ergebnisse)
  const ergebnisZelle = (e) => {
    if (e.ausgang === 'ohne Fund') return 'ohne Fund'
    if (e.ausgang === 'Fund') return `**Fund**: ${zelle(e.grund)} — \`${e.fund}\``
    return `**unvollständig**: ${zelle(e.fehler.join('; '))}`
  }
  return `# Fuzzkampagne

**${ergebnisse.length} Ziele, ${funde.length} mit Fund, ${unvollstaendig.length} unvollständig** — Commit \`${kopf.commit}\`${kopf.offen > 0 ? ` mit **${kopf.offen} ungesicherte Änderung(en)**` : ''}${kopf.inhalt ? ` (${kopf.offen > 0 ? `Inhalt \`${kopf.inhalt}\`` : 'sauber gegen den Commit'}, Saat \`${kopf.saat}\`)` : ''}${kopf.geaendert ? ', **der Stand änderte sich während des Laufs**' : ''},
${kopf.zeit} UTC, ${kopf.go}.
Budget je Ziel: **${kopf.budget}**, **${kopf.parallel}** Arbeiter, \`GOMEMLIMIT=${kopf.speicher}\`,
Minimierung höchstens 5 s. Gefahren mit \`scripts/fuzzkampagne.mjs\`; was
der Bericht misst und was nicht, steht in seinem Kopf. **Anteil** ist der
Anteil der Anweisungen des Pakets, den Saat und erarbeitetes Korpus betreten —
keine Kantenabdeckung, und der Nenner ist das ganze Paket. **Unvollständig**
heißt: ein Teil der Kampagne (Fuzzlauf, Korpusübernahme, Abdeckungslauf) ist
gescheitert; der Strich beim Anteil ist dann keine Messung. **Zuletzt neu**
ist die Sekunde, in der die letzte neue Eingabe kam; **gesättigt** heißt, im
letzten Viertel des Budgets kam keine mehr — ein festgelegtes Maß, kein
Beweis, dass nichts mehr zu finden wäre. **Nicht bewertbar** heißt, der Lauf
endete vor dem Budget (ein Fund) oder seine letzte Meldung kam früher.

| Ziel | Paket | Dauer | Ausführungen | neue Eingaben | zuletzt neu | Sättigung | Korpus | Anteil | Ergebnis |
| --- | --- | ---: | ---: | ---: | ---: | --- | ---: | ---: | --- |
${ergebnisse.map((e) => `| \`${e.name}\` | \`${e.paket}\` | ${e.sekunden} s | ${tausender(e.ausfuehrungen)} | ${tausender(e.neu)} | ${e.zuletztNeu === null || e.zuletztNeu === undefined ? '–' : `${e.zuletztNeu} s`} | ${saettigung(e, kopf.budget)} | ${tausender(e.korpus)} | ${e.anteil === '–' ? '–' : `${e.anteil} %`} | ${ergebnisZelle(e)} |`).join('\n')}
`
}

// zieleAus liest die Ziele aus `RELEASE.md` — Paket und Name, in ihrer
// Reihenfolge. **Derselbe Leser wie der Wächter** (`internal/werkzeuge/
// fuzzziele_test.go`), und damit kein zweites Verzeichnis, das
// auseinanderläuft. Hier und nicht im Skript, damit ein Test ihn laden kann,
// ohne die Kampagne zu starten.
export function zieleAus(text) {
  return [...text.matchAll(/go test \.\/(internal\/[\w/]+) -run Fuzz -fuzz (Fuzz[A-Z]\w*)/g)].map((t) => ({ paket: t[1], name: t[2] }))
}

// zieleWaehlen nimmt die angeforderten Ziele aus allen — ohne Anforderung
// alle — und **nennt jeden Namen, den es nicht gibt** (N4 des Nachlaufs über
// die QA-Arbeit). Bis dahin wurde ein vertippter Name nur gemeldet, wenn gar
// kein Ziel übrig blieb; neben einem bekannten fiel er still heraus.
export function zieleWaehlen(alle, gewaehlt) {
  if (gewaehlt.length === 0) return { ziele: alle, unbekannt: [] }
  const namen = new Set(alle.map((z) => z.name))
  return {
    ziele: alle.filter((z) => gewaehlt.includes(z.name)),
    unbekannt: gewaehlt.filter((name) => !namen.has(name)),
  }
}


// zusammenfuehren macht aus den Hälften einer Kampagne eine — C02 (#2227).
// **Nur vom selben Stand und mit demselben Budget**: ein Bericht, dessen Ziele
// an verschiedenen Commits oder mit verschiedener Zeit gefuzzt wurden, sähe
// aus wie einer und wäre keiner. Die Reihenfolge ist die der Ziele in
// `RELEASE.md`, die Zeit die der zuletzt fertigen Hälfte.
//
// **Vom selben Inhalt, nicht nur vom selben Commit** (J05, #2269): `inhalt`
// ist der Abdruck der ungesicherten Änderungen, und er gilt für alle Hälften.
//
// **Ein Ziel ist Paket und Name** (Zweitblick über Phase C, Befund 1): 60
// Ziele haben 55 Namen. **Und es kommt jedes, und keines sonst** (Befund 2):
// eine fehlende Hälfte ergab bis dahin einen grünen Bericht über die andere.
export function zusammenfuehren(teile, ziele) {
  const [erster] = teile
  for (const teil of teile) {
    if (teil.kopf.commit !== erster.kopf.commit) throw new Error(`die Hälften kommen von verschiedenen Ständen: ${erster.kopf.commit} und ${teil.kopf.commit}`)
    if (teil.kopf.geaendert) throw new Error(`der Stand änderte sich während des Laufs einer Hälfte (${teil.kopf.commit})`)
    // **Derselbe Commit heißt nicht derselbe Quelltext** (J05, #2269): `inhalt`
    // ist der Abdruck der ungesicherten Änderungen. Zwei je für sich stabile
    // Läufe mit verschiedenen Änderungen am selben Commit wären sonst ein
    // Bericht über einen Stand, den es nie gab — und der Kopf trüge nur den
    // Abdruck der ersten. `saat` bleibt ausgenommen (siehe `standPruefen`).
    if (teil.kopf.inhalt !== erster.kopf.inhalt) throw new Error(`die Hälften haben verschiedenen Inhalt (ungesicherte Änderungen am selben Commit ${erster.kopf.commit}): ${erster.kopf.inhalt ?? 'ohne Angabe'} und ${teil.kopf.inhalt ?? 'ohne Angabe'}`)
    for (const feld of ['budget', 'parallel', 'speicher', 'go']) {
      if (teil.kopf[feld] !== erster.kopf[feld]) throw new Error(`die Hälften haben verschiedenes Budget (${feld}): ${erster.kopf[feld]} und ${teil.kopf[feld]}`)
    }
  }
  const schluessel = (z) => `${z.paket} ${z.name}`
  const alle = teile.flatMap((teil) => teil.ergebnisse)
  const geliefert = alle.map(schluessel)
  const doppelt = geliefert.filter((s, stelle) => geliefert.indexOf(s) !== stelle)
  if (doppelt.length > 0) throw new Error(`ein Ziel steht doppelt: ${doppelt.join(', ')}`)
  const erwartet = ziele.map(schluessel)
  const fehlt = erwartet.filter((s) => !geliefert.includes(s))
  if (fehlt.length > 0) throw new Error(`es fehlt: ${fehlt.join(', ')}`)
  const unbekannt = geliefert.filter((s) => !erwartet.includes(s))
  if (unbekannt.length > 0) throw new Error(`unbekannt: ${unbekannt.join(', ')}`)

  return {
    ergebnisse: [...alle].sort((a, b) => erwartet.indexOf(schluessel(a)) - erwartet.indexOf(schluessel(b))),
    kopf: { ...erster.kopf, zeit: teile.map((teil) => teil.kopf.zeit).sort().at(-1), offen: Math.max(...teile.map((teil) => teil.kopf.offen ?? 0)) },
  }
}
