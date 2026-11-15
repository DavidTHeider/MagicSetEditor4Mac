// **Der Vertragsrundgang** — B01
// (#2220).
//
// Ruft jede Bindung, für die eine Gruppe unter `vertrag/` einen Eintrag hat,
// **unmittelbar** über die Wails-IPC des nativen Fensters
// (`window.go.app.Bindungen`, `frontend/src/anbindung.ts`) — mit echten
// Argumenten über einem echten Bestand, in einer Reihenfolge, die die nötigen
// Kennungen erst anlegt und dann benutzt.
//
// # Wozu, wenn es die Bedienwege gibt
//
// **Die Naht zwischen den Prüfebenen** (§3): Frontendtests laufen nie durch
// die JSON-Serialisierung, Go-Tests sehen die Struct-Marken nicht. Hier geht
// jeder Ruf durch beides — Rumpf, Marken, Ausgang, Fehlerübersetzung — am
// gebauten Programm. Was er **nicht** sagt: ob ein Mensch die Bindung über die
// Oberfläche erreicht. Deshalb zählt `bindungsreichweite.mjs` ihn getrennt
// und nennt die Quote der Bedienwege daneben.
//
// # Wie ein Eintrag aussieht
//
// `{ name, argumente?(k), nimm?, nach?(wert, k), erwartet?, grund?, vorher?(p, k) }`
// — `argumente` rechnet die Argumente aus dem Kontext `k`; `nimm` ist ein
// Ausdruck über die Antwort `a`, **in der Seite** ausgewertet, damit große
// Antworten nicht über die Leitung müssen; `nach` legt Kennungen in `k` ab.
// `erwartet` ist `gelungen` (Vorgabe) oder `fachfehler` — dann mit `grund`,
// warum dieser Ruf fachlich abgelehnt werden **muss** —, und für die
// Bindungen ohne eigenen Eintrag `erreicht` (`vertrag/ersatz.mjs`).
//
// **Nicht gerufen** wird, was in `scripts/bindungsreichweite.ausnahmen.json`
// mit Grund steht: Systemdialoge, fremde Dienste, LaunchAgents.
import { execFileSync } from 'node:child_process'
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { pruefeFormQuelle } from '../antwortform.mjs'
import { antwortformen } from '../antwortformen.mjs'
import { listeNach } from '../bindungsreichweite.mjs'
import { ersatz, sitzungsende, sitzungsreihe, zerstoerend } from './vertrag/ersatz.mjs'
import { gruppen, kohorten, schluss } from './vertrag/gruppen.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')

// signaturen liest die gebundenen Methoden aus dem heutigen Quelltext.
const signaturen = () => JSON.parse(execFileSync('go', ['run', 'scripts/bindungssignaturen.go', 'internal/app'], { cwd: wurzel, encoding: 'utf8' }))

// uebrige baut Einträge für jede Bindung ohne eigenen — **erreicht** genügt:
// gelungen oder fachlich abgelehnt, nur nicht am Tor abgewiesen.
function uebrige(alle, bekannt, ausnahmen) {
  const offen = alle.filter((m) => !bekannt.has(m.name) && !(m.name in ausnahmen))
  const eintrag = (m) => ({ name: m.name, argumente: (k) => m.parameter.map((p) => ersatz(p, k)), erwartet: 'erreicht' })

  return [
    { name: 'Übrige, lesend und schreibend', rufe: offen.filter((m) => !zerstoerend(m.name)).map(eintrag) },
    { name: 'Übrige, zuletzt: was wegnimmt oder beendet', rufe: offen.filter((m) => zerstoerend(m.name) && !sitzungsende(m.name)).map(eintrag) },
    // **Was die Sitzung selbst beendet, ganz zuletzt** — danach weist das
    // Tor alles ab, und kein Ruf hinter ihm sagte noch etwas.
    { name: 'Übrige, ganz zuletzt: was die Sitzung beendet', rufe: offen.filter((m) => sitzungsende(m.name)).sort((a, b) => sitzungsreihe.indexOf(a.name) - sitzungsreihe.indexOf(b.name)).map(eintrag) },
  ]
}

// ruf ruft eine Bindung und gibt Ausgang, Fehler und den genommenen Wert.
//
// **Ein Dateidialog wird abgebrochen, bevor er aufgeht** — eine leere Antwort
// der Abnahmeleitung (`dialogantwort`); sonst stünde der echte Dialog da, und
// niemand beantwortete ihn. **Ein Blatt, das danach offen steht, ebenso.**
async function ruf(p, eintrag, k) {
  const argumente = eintrag.argumente ? await eintrag.argumente(k) : []
  if (k.dialoge.has(eintrag.name)) await p.frage({ was: 'dialogantwort', datei: '' })
  if (eintrag.vorher) await eintrag.vorher(p, k)
  try {
    return await rufen(p, eintrag, argumente, k.formen[eintrag.name])
  } finally {
    const blatt = await p.frage({ was: 'blatt' }).catch(() => null)
    if (blatt?.art) await p.frage({ was: 'blattantwort', abbrechen: true }).catch(() => undefined)
  }
}

async function rufen(p, eintrag, argumente, form) {
  const nimm = eintrag.nimm ?? 'null'

  // **Die Form wird in der Seite geprüft** (C01, #2226) — gegen die Typen des
  // Fensters; zurück kommen nur die Abweichungen, nicht die ganze Antwort.
  return p.seite(`${pruefeFormQuelle}
  try {
    const a = await window.go.app.Bindungen[${JSON.stringify(eintrag.name)}](...${JSON.stringify(argumente)})
    // **Eine Bindung, die nur einen Fehler zurückgibt** (\`*Fehler\`), gibt ihn
    // ohne Umschlag — dann ist die Antwort selbst der Fehler.
    const fehler = a && typeof a === 'object' ? (a.fehler ?? ('code' in a && 'schluessel' in a ? a : null)) : null
    const formbefund = ${form ? `pruefeForm(a, ${JSON.stringify(form)})` : 'null'}
    return { ausgang: fehler ? 'fachfehler' : 'gelungen', fehler, wert: fehler ? null : (${nimm}), form: formbefund && { abweichungen: formbefund.abweichungen.slice(0, 5), zusaetzlich: formbefund.zusaetzlich.slice(0, 5), unbekannt: formbefund.unbekannt } }
  } catch (grund) {
    return { ausgang: 'abgewiesen', fehler: { meldung: String(grund?.message ?? grund).slice(0, 300) } }
  }`)
}

export default async function vertragsrundgang(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  await p.starte({ ort })
  const alle = signaturen()
  // **Die Dialoge aus dem Aufrufgraphen** (`scripts/bindungssignaturen.go`).
  const k = { ort, dialoge: new Set(alle.filter((m) => m.dialog).map((m) => m.name)), formen: antwortformen() }
  const formbuch = {}
  mkdirSync(join(ort.ort, 'vertrag'), { recursive: true })
  const zeilen = []
  const ausnahmen = JSON.parse(readFileSync(join(wurzel, 'scripts/bindungsreichweite.ausnahmen.json'), 'utf8'))
  // **Kohorten zählen nicht als bekannt** — ihre Bindungen laufen danach
  // noch einmal mit Daten.
  const bekannt = new Set([...gruppen, ...schluss].flatMap((g) => g.rufe.map((r) => r.name)))

  const [lesendSchreibend, wegnehmend, sitzungsenden] = uebrige(alle, bekannt, ausnahmen)
  for (const gruppe of [...kohorten, ...gruppen, lesendSchreibend, wegnehmend, ...schluss, sitzungsenden]) {
    p.notiz(`Gruppe ${gruppe.name}`)
    const abweichend = []
    for (const eintrag of gruppe.rufe) {
      const erwartet = eintrag.erwartet ?? 'gelungen'
      let antwort
      try {
        antwort = await ruf(p, eintrag, k)
      } catch (grund) {
        antwort = { ausgang: 'abgebrochen', fehler: { meldung: String(grund.message).slice(0, 200) } }
      }
      if (antwort.ausgang === 'gelungen' && eintrag.nach) {
        try {
          eintrag.nach(antwort.wert, k)
        } catch (grund) {
          antwort = { ausgang: 'unerwartet', fehler: { meldung: `die Antwort trägt nicht, was die nächsten brauchen: ${String(grund.message)}` } }
        }
      }
      const passt = erwartet === 'erreicht' ? ['gelungen', 'fachfehler'].includes(antwort.ausgang) : antwort.ausgang === erwartet
      const was = antwort.fehler ? `${antwort.fehler.schluessel ?? ''} ${antwort.fehler.meldung ?? ''}`.trim().slice(0, 160) : ''
      // **Die Form zählt bei jedem Ruf, der antwortet** (C01) — gelungen
      // **und** fachlich abgelehnt: auch die Antwort mit `fehler` ist eine
      // Antwort des Typs, den das Fenster erwartet, und eine Liste darin ist
      // `[]` und nicht `null`. Getrennt gezählt, weil die Quote die
      // gelungenen meint.
      if (['gelungen', 'fachfehler'].includes(antwort.ausgang) && antwort.form) {
        const bisher = formbuch[eintrag.name] ?? { geprueft: 0, abweichungen: [], zusaetzlich: [], imFehler: { geprueft: 0, abweichungen: [] } }
        const teil = antwort.ausgang === 'gelungen' ? bisher : bisher.imFehler
        teil.geprueft++
        teil.abweichungen.push(...antwort.form.abweichungen.map((a) => (eintrag.kohorte ? `${a} (${eintrag.kohorte})` : a)))
        if (antwort.ausgang === 'gelungen') bisher.zusaetzlich.push(...antwort.form.zusaetzlich)
        formbuch[eintrag.name] = bisher
      }
      const formzelle = antwort.form ? (antwort.form.abweichungen.length ? `✗ ${antwort.form.abweichungen.join(', ')}` : '✓') : ''
      zeilen.push(`| ${gruppe.name} | \`${eintrag.name}\` | ${erwartet} | ${antwort.ausgang} | ${formzelle.replaceAll('|', '/')} | ${was.replaceAll('|', '/')} |`)
      if (!passt) abweichend.push(`${eintrag.name}: ${antwort.ausgang}${was ? ` (${was})` : ''}`)
    }
    p.pruefe(`${gruppe.name}: ${String(gruppe.rufe.length)} Rufe antworten wie erwartet`, abweichend.length === 0, abweichend.slice(0, 6).join('; '))
  }

  const mitAbweichung = Object.entries(formbuch).filter(([, b]) => b.abweichungen.length > 0)
  const imFehlerAbweichend = Object.entries(formbuch).filter(([, b]) => b.imFehler.abweichungen.length > 0)
  p.pruefe(`die Antwortformen bei fachlicher Ablehnung: ${String(Object.values(formbuch).filter((b) => b.imFehler.geprueft > 0).length)} gerufene Bindungen geprüft, ${String(imFehlerAbweichend.length)} davon weichen ab`, imFehlerAbweichend.length === 0,
    imFehlerAbweichend.slice(0, 8).map(([n, b]) => `${n}: ${[...new Set(b.imFehler.abweichungen)].slice(0, 2).join(', ')}`).join('; '))

  // **Die Quote über die lesende Registry** (C01): der Nenner ist **jede**
  // lesende Bindung aus `docs/SCHREIBWEGE.md`, auch die, die eine Ausnahme
  // nicht ruft; validiert ist, was mindestens eine gelungene Antwort ohne
  // Abweichung hatte.
  const lesend = listeNach(readFileSync(join(wurzel, 'docs/SCHREIBWEGE.md'), 'utf8'), 'lesenden Bindungen')
  const validiert = lesend.filter((n) => formbuch[n]?.geprueft > 0 && formbuch[n].abweichungen.length === 0)
  const quote = validiert.length / lesend.length
  p.pruefe(`lesende Registry: ${String(validiert.length)} von ${String(lesend.length)} mit validierter Antwortform (${(quote * 100).toFixed(1).replace('.', ',')} %, Schwelle 90 %)`, quote >= 0.9,
    lesend.filter((n) => !validiert.includes(n)).join(', '))
  p.pruefe(`die Antwortformen: ${String(Object.keys(formbuch).length)} gerufene Bindungen mit gelungener Antwort geprüft, ${String(mitAbweichung.length)} davon weichen von den Typen des Fensters ab`, mitAbweichung.length === 0,
    mitAbweichung.slice(0, 8).map(([n, b]) => `${n}: ${[...new Set(b.abweichungen)].slice(0, 2).join(', ')}`).join('; '))
  writeFileSync(join(wurzel, 'docs/belege/huellenprobe/vertrag-antwortformen.json'), `${JSON.stringify(formbuch, null, 1)}\n`)
  p.notiz('Jeder Ruf')
  p.zeilen.push('| Gruppe | Bindung | erwartet | Ausgang | Form | Fehler |', '| --- | --- | --- | --- | --- | --- |', ...zeilen)
}
