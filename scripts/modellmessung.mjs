#!/usr/bin/env node
// **Die fünf kleinen Modelle an der Laufzeit gemessen** — AI-02.AK-01, AK-02
// und der Bildteil von AK-03
// (#1743).
//
//     ollama serve
//     node scripts/modellmessung.mjs
//
// # Was es misst
//
// Je Modell, **eines nach dem anderen** und mit allen anderen entladen
// (AK-04): die Pinnung aus `/api/tags` und `/api/show` — Digest, Größe,
// Quantisierung, Lizenz, Parameter —, drei Kaltstarts mit der Ladezeit, die
// die Laufzeit selbst nennt (`load_duration`), eine feste Arbeit, und danach
// den **Spitzenwert des physischen Speichers** aller Ollama-Prozesse
// (`vmmap --summary`, »Physical footprint (peak)«) des Prozesses, in dem das
// Modell rechnet. **Für das Bildmodell** fragt es, ob die Laufzeit es
// überhaupt ausführt — und hält die Antwort fest, wie sie kommt.
//
// # Was es nicht misst
//
// **Die Güte.** Ob ein Name passt oder eine Suche das Richtige findet, sagen
// `internal/app/modellvergleich_test.go` und der Retrievallauf — über die
// Klienten der Anwendung und nicht an der Laufzeit vorbei. **Und es zieht
// nichts**: die Modelle werden von Hand geladen, wie bei jedem Lauf hier.
//
// Der Beleg geht nach `docs/belege/modellmessung.md`, die Rohwerte nach
// `docs/belege/modellmessung.json`.
import { execFileSync } from 'node:child_process'
import { readFileSync, writeFileSync } from 'node:fs'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const wurzel = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const adresse = process.env.OLLAMA_HOST
  ? `http://${process.env.OLLAMA_HOST.replace(/^https?:\/\//, '')}`
  : 'http://127.0.0.1:11434'

// modelle sind die fünf aus AI-02.AK-01, je mit ihrem Zweck.
const modelle = [
  { name: 'qwen3:0.6b', zweck: 'Text' },
  { name: 'gemma3:1b', zweck: 'Text' },
  { name: 'all-minilm:22m', zweck: 'Einbettung' },
  { name: 'embeddinggemma:300m', zweck: 'Einbettung' },
  { name: 'x/flux2-klein:4b', zweck: 'Bild' },
]

// kaltstarts ist, wie oft ein Modell aus dem entladenen Zustand geladen wird.
const kaltstarts = 3

const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))

// frage schickt einen Rumpf und gibt die Antwort als JSON.
async function frage(weg, rumpf, signal) {
  const antwort = await fetch(adresse + weg, {
    method: rumpf === undefined ? 'GET' : 'POST',
    body: rumpf === undefined ? undefined : JSON.stringify(rumpf),
    signal,
  })
  const text = await antwort.text()
  if (!antwort.ok) throw new Error(`${weg}: ${antwort.status} ${text.slice(0, 200)}`)
  return JSON.parse(text)
}

// geladen nennt die Modelle, die die Laufzeit gerade hält.
async function geladen() {
  return (await frage('/api/ps')).models ?? []
}

// entlade wirft jedes geladene Modell aus dem Speicher und wartet, bis es
// fort ist — **und bis kein Laufprozess mehr steht**, sonst misst der
// nächste Spitzenwert den vorigen mit.
async function entladeAlle() {
  for (const m of await geladen()) {
    const einbettend = modelle.find((eines) => eines.name === m.name)?.zweck === 'Einbettung'
    await frage(einbettend ? '/api/embed' : '/api/generate', {
      model: m.name, keep_alive: 0, ...(einbettend ? { input: [] } : {}),
    }).catch(() => {})
  }
  for (let warten = 0; warten < 120; warten++) {
    if ((await geladen()).length === 0 && laufprozesse().length === 0) return
    await schlafe(500)
  }
  throw new Error(`entladen: ${JSON.stringify(await geladen())}`)
}

// laufprozesse sind die Prozesse, in denen ein Modell rechnet — **nicht**
// der Dienst selbst. Bei Ollama 0.35.1 aus Homebrew ist das
// `libexec/lib/ollama/llama-server`, je geladenem Modell einer (gemessen am
// 2026-10-04); ältere Fassungen hießen `ollama runner`.
function laufprozesse() {
  const zeilen = execFileSync('/bin/ps', ['-axo', 'pid=,command='], { encoding: 'utf8' }).split('\n')
  return zeilen
    .map((zeile) => zeile.trim().match(/^(\d+)\s+(.*)$/))
    .filter((treffer) => treffer && /\/ollama\/llama-server\b|\bollama runner\b/.test(treffer[2]))
    .map((treffer) => ({ pid: Number(treffer[1]), befehl: treffer[2] }))
}

// speicher liest zwei Werte eines Prozesses aus `vmmap --summary`, in Bytes:
//
//   - **eigen** — »Physical footprint (peak)«, was der Kern dem Prozess
//     anrechnet, samt GPU-Puffern (`owned unmapped (graphics)`);
//   - **gemappt** — die residente Größe der Region `mapped file`: die
//     Gewichte, die Ollama aus der Datei einblendet. Sie zählt **nicht** zum
//     Fußabdruck — gemessen am 2026-10-04 an gemma3:1b: 314 MB eigen, 787 MB
//     gemappt, bei 815 MB Gewichten. Ohne sie sähe ein Modell kleiner aus als
//     seine eigene Datei.
function speicher(pid) {
  const text = execFileSync('/usr/bin/vmmap', ['--summary', String(pid)], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] })
  const bytes = (zahl, einheit) => Number(zahl) * ({ K: 1024, M: 1024 ** 2, G: 1024 ** 3 }[einheit] ?? 1)
  const eigen = text.match(/Physical footprint \(peak\):\s+([\d.]+)([KMG])/)
  const gemappt = text.match(/^mapped file\s+[\d.]+[KMG]?\s+([\d.]+)([KMG]?)/m)
  return {
    eigen: eigen ? bytes(eigen[1], eigen[2]) : 0,
    gemappt: gemappt ? bytes(gemappt[1], gemappt[2]) : 0,
  }
}

// pinnung liest, was ein Modell ist — aus `/api/tags` und `/api/show`.
async function pinnung(name) {
  const eintrag = (await frage('/api/tags')).models.find((m) => m.name === name)
  if (!eintrag) throw new Error(`BLOCKED: ${name} ist nicht gezogen`)
  const gezeigt = await frage('/api/show', { model: name })
  const info = gezeigt.model_info ?? {}
  const wert = (endung) => Object.entries(info).find(([schluessel]) => schluessel.endsWith(endung))?.[1]
  return {
    digest: eintrag.digest,
    bytes: eintrag.size,
    format: eintrag.details?.format,
    familie: eintrag.details?.family || undefined,
    parameter: wert('general.parameter_count') ?? eintrag.details?.parameter_size,
    quantisierung: eintrag.details?.quantization_level || undefined,
    lizenz: (gezeigt.license ?? '').trim().split('\n')[0].trim(),
    faehigkeiten: gezeigt.capabilities ?? [],
    kontext: wert('.context_length'),
    stellen: wert('.embedding_length'),
    vorgaben: (gezeigt.parameters ?? '').split('\n').map((zeile) => zeile.trim().replace(/\s+/g, ' ')).filter(Boolean),
    verlangt: gezeigt.requires,
  }
}

// lade lädt ein Modell aus dem entladenen Zustand und gibt die Ladezeit,
// die die Laufzeit nennt, und die Wanduhr.
//
// **Mit einer Frage von einem Token und nicht mit einer leeren**: auf einen
// leeren Auftrag lädt Ollama zwar, nennt aber keine `load_duration` (gemessen
// am 2026-10-04 — der erste Lauf zeigte 0,00 s bei 0,78 s Wanduhr).
async function lade(modell) {
  await entladeAlle()
  const anfang = performance.now()
  const antwort =
    modell.zweck === 'Einbettung'
      ? await frage('/api/embed', { model: modell.name, input: ['x'] })
      : await frage('/api/generate', { model: modell.name, prompt: 'x', stream: false, think: false, options: { num_predict: 1 } })
  return { laufzeit: (antwort.load_duration ?? 0) / 1e9, wand: (performance.now() - anfang) / 1000 }
}

// textarbeit fragt zehnmal dasselbe mit festem Startwert.
async function textarbeit(name) {
  const laeufe = []
  for (let runde = 0; runde < 10; runde++) {
    const antwort = await frage('/api/generate', {
      model: name, prompt: 'Nenne drei Farben, durch Kommas getrennt.', stream: false, think: false,
      options: { seed: 7, num_predict: 64 },
    })
    laeufe.push({ text: antwort.response, tokenProSekunde: antwort.eval_count / (antwort.eval_duration / 1e9) })
  }
  return {
    laeufe: laeufe.length,
    gleich: laeufe.every((lauf) => lauf.text === laeufe[0].text),
    tokenProSekunde: median(laeufe.map((lauf) => lauf.tokenProSekunde)),
  }
}

// einbettarbeit bettet den Probekorpus ein, in Bündeln zu sechzehn.
async function einbettarbeit(name) {
  const korpus = JSON.parse(readFileSync(join(wurzel, 'testdaten/retrieval/korpus.json'), 'utf8'))
  const texte = korpus.dokumente.map((dokument) => `${dokument.titel}\n${dokument.text}`)
  const anfang = performance.now()
  let stellen = 0
  for (let ab = 0; ab < texte.length; ab += 16) {
    const antwort = await frage('/api/embed', { model: name, input: texte.slice(ab, ab + 16) })
    stellen = antwort.embeddings[0].length
  }
  return { texte: texte.length, sekunden: (performance.now() - anfang) / 1000, stellen }
}

function median(werte) {
  const sortiert = [...werte].sort((a, b) => a - b)
  return sortiert[Math.floor(sortiert.length / 2)]
}


// anweisung liest `KIKartenanweisung` aus dem Quelltext der Anwendung —
// **keine Abschrift**: eine Messung mit einer anderen Systemzeile wäre eine
// Messung von etwas anderem.
function anweisung() {
  const quelle = readFileSync(join(wurzel, 'internal/app/kilaufzeit.go'), 'utf8')
  const treffer = quelle.match(/const KIKartenanweisung = ((?:"[^"]*"\s*\+?\s*)+)/)
  if (!treffer) throw new Error('KIKartenanweisung nicht gefunden')
  return [...treffer[1].matchAll(/"([^"]*)"/g)].map((teil) => teil[1]).join('')
}

// denkprobe stellt einem Modell, das denken kann, die 32 Fragen des
// Testsatzes so, wie `internal/ollama` sie stellt — einmal **ohne** das Feld
// `think`, wie bis AI-02, und einmal mit `false`, wie seitdem. Je Frage
// höchstens 30 s; wer dann nicht fertig ist, zählt als nicht angekommen.
async function denkprobe(name) {
  const satz = JSON.parse(readFileSync(join(wurzel, 'testdaten/ki/testsatz.json'), 'utf8'))
  const system = anweisung()
  const fragen = satz.karten.flatMap((karte) => ['de', 'en'].flatMap((sprache) => ['name', 'flavor text'].map((feld) => ({
    feld, text: `${satz.auftraege[sprache][feld]}\n\n${satz.kartebisher[sprache]}\ntype: ${karte.typ}\nrule text: ${karte.text}`,
  }))))
  const ergebnis = {}
  for (const [art, denken] of [['ohne think', undefined], ['think: false', false]]) {
    const laeufe = []
    for (const eine of fragen) {
      const anfang = performance.now()
      try {
        const antwort = await frage('/api/chat', {
          model: name, stream: false, ...(denken === undefined ? {} : { think: denken }),
          messages: [{ role: 'system', content: system }, { role: 'user', content: eine.text }],
          format: { type: 'object', required: [eine.feld], properties: { [eine.feld]: { type: 'string' } } },
          options: { temperature: 0 },
        }, AbortSignal.timeout(30000))
        laeufe.push({ sekunden: (performance.now() - anfang) / 1000, token: antwort.eval_count })
      } catch {
        laeufe.push({ sekunden: (performance.now() - anfang) / 1000 })
      }
    }
    const angekommen = laeufe.filter((lauf) => lauf.token !== undefined)
    ergebnis[art] = {
      angekommen: angekommen.length, von: laeufe.length,
      median: median(laeufe.map((lauf) => lauf.sekunden)), langsamste: Math.max(...laeufe.map((lauf) => lauf.sekunden)),
      token: angekommen.length ? median(angekommen.map((lauf) => lauf.token)) : undefined,
    }
  }
  return ergebnis
}

// bildprobe fragt, ob die Laufzeit ein Bildmodell ausführt — über
// `/api/generate`, den Weg, den ihr eigener Befehl `ollama run` nimmt, und
// über die zwei Wege, die es bei anderen Anbietern gibt.
async function bildprobe(name) {
  const wege = {}
  for (const weg of ['/api/images', '/v1/images/generations']) {
    const antwort = await fetch(adresse + weg, { method: 'POST', body: '{}' })
    wege[weg] = `${antwort.status} ${(await antwort.text()).trim().slice(0, 80)}`
  }
  const antwort = await fetch(adresse + '/api/generate', {
    method: 'POST',
    body: JSON.stringify({ model: name, prompt: 'a red apple on a wooden table', stream: false }),
  })
  wege['/api/generate'] = `${antwort.status} ${(await antwort.text()).trim().slice(0, 120)}`
  return { erzeugt: antwort.ok, wege }
}

// speicherJetzt liest den Speicher aller rechnenden Prozesse.
function speicherJetzt() {
  return laufprozesse().map((prozess) => speicher(prozess.pid)).reduce(
    (summe, eines) => ({ eigen: summe.eigen + eines.eigen, gemappt: summe.gemappt + eines.gemappt }),
    { eigen: 0, gemappt: 0 },
  )
}

// miss misst ein Modell, mit allen anderen entladen.
async function miss(modell) {
  const ergebnis = { ...modell, pinnung: await pinnung(modell.name) }
  if (modell.zweck === 'Bild') {
    await entladeAlle()
    ergebnis.bild = await bildprobe(modell.name)
    return ergebnis
  }

  ergebnis.kaltstarts = []
  for (let lauf = 0; lauf < kaltstarts; lauf++) ergebnis.kaltstarts.push(await lade(modell))

  ergebnis.arbeit = modell.zweck === 'Text' ? await textarbeit(modell.name) : await einbettarbeit(modell.name)

  const gehalten = (await geladen()).find((m) => m.name === modell.name)
  ergebnis.gehalten = gehalten && { bytes: gehalten.size, gpu: gehalten.size_vram, fenster: gehalten.context_length }
  ergebnis.speicher = speicherJetzt()

  // **Die Denkprobe erst nach dem Lesen** — und danach noch einmal gelesen:
  // die Spitze ist eine des Prozesses und wächst mit der Arbeit. Gemessen am
  // 2026-10-04: bei qwen3:0.6b 3,9 GB eigen nach der festen Arbeit, 5,7 GB,
  // als die Denkprobe davor lief.
  if (ergebnis.pinnung.faehigkeiten.includes('thinking')) {
    ergebnis.denken = await denkprobe(modell.name)
    ergebnis.speicherNachDenken = speicherJetzt()
  }
  await entladeAlle()
  return ergebnis
}

const mb = (bytes) => (bytes / 1e6).toLocaleString('de-DE', { maximumFractionDigits: 0 })
const s = (sekunden) => sekunden.toLocaleString('de-DE', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

// beleg schreibt den Beleg als Markdown.
function beleg(lauf) {
  const zeilen = [
    '# Die fünf kleinen Modelle an der Laufzeit',
    '',
    `**${lauf.datum}** — AI-02 (#1743), Commit ${lauf.commit},`,
    `${lauf.rechner}, ${lauf.speicher} GB, macOS ${lauf.system}, Ollama ${lauf.ollama}.`,
    '',
    '    ollama serve',
    '    node scripts/modellmessung.mjs',
    '',
    '## Gepinnt',
    '',
    '| Modell | Zweck | Digest | Größe | Format | Quantisierung | Parameter | Kontext | Stellen | Lizenz | Fähigkeiten |',
    '| --- | --- | --- | ---: | --- | --- | ---: | ---: | ---: | --- | --- |',
    ...lauf.modelle.map((m) => {
      const p = m.pinnung
      return `| ${m.name} | ${m.zweck} | \`${p.digest.slice(0, 12)}\` | ${mb(p.bytes)} MB | ${p.format ?? '—'} | ${p.quantisierung ?? '—'} | ${typeof p.parameter === 'number' ? (p.parameter / 1e6).toLocaleString('de-DE', { maximumFractionDigits: 1 }) + ' Mio.' : (p.parameter || '—')} | ${p.kontext ?? '—'} | ${p.stellen ?? '—'} | ${p.lizenz || '—'} | ${p.faehigkeiten.join(', ')} |`
    }),
    '',
    '## Laden, Speicher, Stabilität',
    '',
    `Je Modell ${kaltstarts} Kaltstarts aus dem entladenen Zustand; **Ladezeit** ist \`load_duration\` der Laufzeit, **Wanduhr** die ganze erste Anfrage. Der Speicher stammt aus \`vmmap --summary\` des rechnenden Prozesses nach Laden und Arbeit: **eigen** ist »Physical footprint (peak)« samt GPU-Puffern, **Gewichte** die residente Größe der eingeblendeten Modelldatei, die nicht zum Fußabdruck zählt. **Ollama plant** nennt, was \`/api/ps\` angibt — eine Schätzung der Laufzeit, keine Messung.`,
    '',
    '| Modell | Ladezeit (Median) | Wanduhr (Median) | eigen (Spitze) | Gewichte | zusammen | Ollama plant | Fenster | Arbeit | Ergebnis |',
    '| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |',
    ...lauf.modelle.filter((m) => m.zweck !== 'Bild').map((m) => {
      const arbeit = m.zweck === 'Text'
        ? [`${m.arbeit.laeufe} Antworten, Startwert 7`, `${m.arbeit.gleich ? 'alle gleich' : 'nicht alle gleich'}, ${m.arbeit.tokenProSekunde.toLocaleString('de-DE', { maximumFractionDigits: 0 })} Token/s`]
        : [`${m.arbeit.texte} Texte eingebettet`, `${s(m.arbeit.sekunden)} s, ${m.arbeit.stellen} Stellen`]
      return `| ${m.name} | ${s(median(m.kaltstarts.map((k) => k.laufzeit)))} s | ${s(median(m.kaltstarts.map((k) => k.wand)))} s | ${mb(m.speicher.eigen)} MB | ${mb(m.speicher.gemappt)} MB | ${mb(m.speicher.eigen + m.speicher.gemappt)} MB | ${mb(m.gehalten.bytes)} MB | ${m.gehalten.fenster} | ${arbeit[0]} | ${arbeit[1]} |`
    }),
    '',
    '## Ein Modell, das denken kann',
    '',
    'Die 32 Fragen aus `testdaten/ki/testsatz.json`, gestellt wie `internal/ollama` sie stellt — Systemzeile `KIKartenanweisung` aus dem Quelltext, Schema, Temperatur 0 —, einmal **ohne** das Feld `think` (bis AI-02) und einmal mit `false` (seitdem). Je Frage höchstens 30 s.',
    '',
    '| Modell | Frage | angekommen | Dauer (Median) | langsamste | Token (Median) |',
    '| --- | --- | ---: | ---: | ---: | ---: |',
    ...lauf.modelle.filter((m) => m.denken).flatMap((m) => Object.entries(m.denken).map(([art, w]) =>
      `| ${m.name} | ${art} | ${w.angekommen} von ${w.von} | ${s(w.median)} s | ${s(w.langsamste)} s | ${w.token ?? '—'} |`)),
    '',
    ...lauf.modelle.filter((m) => m.denken).map((m) =>
      `**Und der Speicher wächst mit**: ${m.name} nach der festen Arbeit ${mb(m.speicher.eigen)} MB eigen, nach der Denkprobe ${mb(m.speicherNachDenken.eigen)} MB — die Spitze ist eine des Prozesses.`),
    '',
    '## Das Bildmodell',
    '',
    ...lauf.modelle.filter((m) => m.zweck === 'Bild').flatMap((m) => [
      `**${m.name}**: ${m.bild.erzeugt ? 'erzeugt' : '**nicht ausgeführt**'} — die Antworten der Laufzeit, wörtlich:`,
      '',
      '| Weg | Antwort |',
      '| --- | --- |',
      ...Object.entries(m.bild.wege).map(([weg, antwort]) => `| \`${weg}\` | ${antwort.replaceAll('|', '/')} |`),
    ]),
    '',
  ]
  return zeilen.join('\n')
}

async function hauptlauf() {
  const befehl = (programm, ...argumente) => execFileSync(programm, argumente, { encoding: 'utf8' }).trim()
  const lauf = {
    datum: new Date().toLocaleString('sv-SE', { timeZone: 'Europe/Berlin' }).slice(0, 16),
    commit: befehl('git', '-C', wurzel, 'rev-parse', '--short', 'HEAD'),
    rechner: befehl('sysctl', '-n', 'machdep.cpu.brand_string'),
    speicher: Number(befehl('sysctl', '-n', 'hw.memsize')) / 1024 ** 3,
    system: befehl('sw_vers', '-productVersion'),
    ollama: (await frage('/api/version')).version,
    modelle: [],
  }
  for (const modell of modelle) {
    process.stdout.write(`${modell.name} … `)
    lauf.modelle.push(await miss(modell))
    process.stdout.write('gemessen\n')
  }
  writeFileSync(join(wurzel, 'docs/belege/modellmessung.json'), JSON.stringify(lauf, null, 2) + '\n')
  writeFileSync(join(wurzel, 'docs/belege/modellmessung.md'), beleg(lauf))
  process.stdout.write(beleg(lauf))
}

await hauptlauf()
