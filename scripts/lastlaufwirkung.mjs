// **Hat der Ruf getan, was er zusagte?** — Z3 des Zweitblicks über Phase 6.11
// (QA-08, #1739).
//
// # Ein 2xx ist nicht die Wirkung
//
// `istGelungen` (`lastlaufbericht.mjs`) liest Status und MCP-Fehler. Antwortet
// der Dienst 201 und schreibt kein Bild, steht der Ruf im Durchsatz — und ein
// Dienst, der nach dem ersten Ruf nur noch „schon offen“ sagt, hat in einer
// Minute Zehntausende Einfuhren „geschafft“. Hier wird nachgeprüft, was ein
// Profil zusagt, **an der Stichprobe und am Ausgang**: das Bild am Ziel, das
// Feld im Bestand, das Set nach der Einfuhr.
//
// # Wie
//
// Ein Profil hängt an jede Probe `wirkung` — **was** sie zugesagt hat (`art`
// und die Angaben dazu). Nach dem Profil ruft `pruefeWirkungen` je Gruppe:
//
//   - **Dateien** (`bild`, `bogen`): `stichprobe` Rufe, gleichmäßig verteilt,
//     erster und letzter eingeschlossen; das Bild muss ein **vollständiges**
//     PNG sein (`pngpruefung.mjs`), der Bogen ein PDF mit Anfang und `%%EOF`;
//   - **Antworten** (`karten`, `sets`, `mcp-karten`, `json`): dieselbe Stichprobe,
//     am Text der Antwort — eine Kartenliste mit der Kartenzahl der Saat;
//   - **Bestand** (`feld`, `einfuhr`): **eine** Abfrage nach dem Lauf. Ein Feld
//     ist geschrieben, wenn die Karte danach einen der Werte trägt, die für sie
//     gelungen sind — ein späterer Schreiber darf ihn überschrieben haben. Eine
//     Einfuhr hat gewirkt, wenn ein Set des Namens dasteht, das vorher nicht
//     da war, mit den Karten der Datei.
//
// **Eine Probe, deren Wirkung fehlt, ist ein Fehler** (`wirkungFehlt`; `istGelungen`
// liest es) — nicht eine Drosselung und nicht ein Erfolg.
//
// # Was das nicht ist
//
// **Eine Stichprobe, keine Vollprüfung.** Gezählt wird, was geprüft wurde; ein
// Dienst, der jeden hundertsten Ruf verschluckt, fällt womöglich durch das Netz.
// Und **bei der Einfuhr** gilt die Wirkung dem Lauf, nicht dem einzelnen Ruf:
// derselbe Pfad, zum zweiten Mal gelesen, meldet `schonOffen` und legt nichts an
// — das ist richtig, und die Zahl der Rufe, die ein Set anlegten, sagt der
// Beleg daneben.

import { existsSync, readFileSync } from 'node:fs'
import { basename } from 'node:path'
import { setTimeout as pause } from 'node:timers/promises'

import { istGelungen } from './lastlaufbericht.mjs'
import { pngPruefen } from './pngpruefung.mjs'

/**
 * hoeflich fragt, bis keine Drosselung mehr antwortet — höchstens achtmal, und
 * mindestens eine Sekunde zwischen den Versuchen. Gedrosselt ist ein 429 über
 * REST und, über MCP, eine Antwort mit `gedrosselt`. **Die Auskunft selbst ist
 * keine Last**, die der Lauf messen wollte: ein leeres Guthaben nach einer
 * Überlast soll sie nicht verhindern.
 */
export async function hoeflich(abfrage, schlafe = pause) {
  let antwort
  for (let versuch = 0; versuch < 8; versuch++) {
    antwort = await abfrage()
    if (antwort.status !== 429 && antwort.gedrosselt !== true) return antwort
    await schlafe(Math.max(antwort.warten ?? 0, 1) * 1000)
  }
  return antwort
}

/** stichprobeAus wählt `anzahl` Elemente gleichmäßig, das erste und das letzte eingeschlossen. */
function stichprobeAus(liste, anzahl) {
  if (liste.length <= anzahl) return liste
  if (anzahl === 1) return [liste[0]]
  const stellen = new Set(Array.from({ length: anzahl }, (_, i) => Math.round((i * (liste.length - 1)) / (anzahl - 1))))
  return [...stellen].map((stelle) => liste[stelle])
}

const json = (text) => {
  try {
    return JSON.parse(text ?? '')
  } catch {
    return undefined
  }
}

/** Dateien: der Grund, warum die Wirkung fehlt, oder `null`. */
const dateiPruefer = {
  bild(w) {
    if (!existsSync(w.ziel)) return `Bild ${w.ziel} nicht da`
    const p = pngPruefen(readFileSync(w.ziel))
    return p.ok ? null : `${basename(w.ziel)}: kein vollständiges PNG (${p.fehler})`
  },
  bogen(w) {
    if (!existsSync(w.ziel)) return `Bogen ${w.ziel} nicht da`
    const daten = readFileSync(w.ziel)
    if (!daten.subarray(0, 5).equals(Buffer.from('%PDF-'))) return `${basename(w.ziel)}: kein PDF`
    return daten.subarray(-1024).includes('%%EOF') ? null : `${basename(w.ziel)}: ohne %%EOF, die Datei ist abgeschnitten`
  },
}

/** zahlKarten gibt die Zahl der Karten in einer Liste oder Hülle — `undefined`, wenn es keine ist. */
const zahlKarten = (gelesen) => (Array.isArray(gelesen) ? gelesen : gelesen?.karten)?.length

/** Antworten: der Grund oder `null`, gelesen am Text der Probe. */
const antwortPruefer = {
  karten(p, w) {
    const n = zahlKarten(json(p.text))
    if (n === undefined) return 'keine Kartenliste in der Antwort'
    return n === w.anzahl ? null : `${n} statt ${w.anzahl} Karten`
  },
  // **Nur eine Liste, nicht ein bestimmtes Set darin**: was `GET /sets` einem
  // Principal zeigt, richtet sich nach seinem Profil, und die Fairness fragt mit
  // zwei Tokens eines anderen Profils.
  sets(p) {
    return Array.isArray(json(p.text)?.sets) ? null : 'keine Setliste in der Antwort'
  },
  'mcp-karten'(p, w) {
    const n = zahlKarten(json(p.inhalt))
    if (n === undefined) return 'keine Kartenliste im Ergebnis'
    return n === w.anzahl ? null : `${n} statt ${w.anzahl} Karten`
  },
  json(p) {
    const gelesen = json(p.text)
    return gelesen !== null && typeof gelesen === 'object' ? null : 'kein JSON in der Antwort'
  },
}

/**
 * feldWirkung prüft die Schreibrufe einer Gruppe an einer Abfrage des Bestands je Set.
 *
 * **Der Endstand belegt den Ruf, dessen Wert dasteht** — N3 des Nachlaufs über
 * die QA-Arbeit. Schrieben mehrere dasselbe Feld, sind die übrigen
 * **verdeckt**: ihr Wert ist überschrieben, und ob sie je gewirkt haben, sagt
 * der Endstand nicht. Gibt `{ gründe, verdeckt }` zurück.
 */
async function feldWirkung(kandidaten, abfrage, schlafe) {
  const gründe = new Map()
  const verdeckt = new Map()
  const jeSet = Map.groupBy(kandidaten, (p) => p.wirkung.setID)
  for (const [setID, proben] of jeSet) {
    const antwort = await hoeflich(() => abfrage(`/sets/${setID}/karten`), schlafe)
    const karten = json(antwort.text)?.karten
    if (antwort.status !== 200 || !Array.isArray(karten)) {
      for (const p of proben) gründe.set(p, `Bestand nicht lesbar (${antwort.status})`)
      continue
    }
    const stand = new Map(karten.map((k) => [String(k.id), k.werte]))
    const geschrieben = new Map()
    for (const p of proben) {
      const schluessel = `${p.wirkung.karte}\u0000${p.wirkung.feld}`
      geschrieben.set(schluessel, (geschrieben.get(schluessel) ?? new Set()).add(p.wirkung.wert))
    }
    for (const p of proben) {
      const werte = stand.get(String(p.wirkung.karte))
      if (werte === undefined) {
        gründe.set(p, `Karte ${p.wirkung.karte} fehlt im Bestand`)
        continue
      }
      const heute = werte[p.wirkung.feld]
      if (!geschrieben.get(`${p.wirkung.karte}\u0000${p.wirkung.feld}`).has(heute)) {
        gründe.set(p, `Karte ${p.wirkung.karte}: der Bestand hat „${heute}“, keiner der geschriebenen Werte`)
      } else if (heute !== p.wirkung.wert) {
        verdeckt.set(p, `Karte ${p.wirkung.karte}: überschrieben — der Bestand hat „${heute}“ von einem anderen Ruf`)
      }
    }
  }
  return { gründe, verdeckt }
}

/** einfuhrWirkung prüft, ob die Einfuhr ein Set angelegt hat, das vorher nicht da war. */
async function einfuhrWirkung(kandidaten, abfrage, schlafe) {
  const w = kandidaten[0].wirkung
  const grund = await (async () => {
    const antwort = await hoeflich(() => abfrage('/sets'), schlafe)
    const sets = json(antwort.text)?.sets
    if (antwort.status !== 200 || !Array.isArray(sets)) return `Bestand nicht lesbar (${antwort.status})`
    const neu = sets.find((s) => s.name === w.name && !w.vorher.includes(String(s.id)))
    if (!neu) return `kein neues Set „${w.name}“ im Bestand`
    const karten = await hoeflich(() => abfrage(`/sets/${neu.id}/karten`), schlafe)
    const n = json(karten.text)?.karten?.length ?? 0
    return n >= w.karten ? null : `${n} von ${w.karten} Karten im neuen Set`
  })()
  return new Map(grund === null ? [] : kandidaten.map((p) => [p, grund]))
}

/**
 * pruefeWirkungen prüft je Gruppe, ob die gelungenen Rufe wirkten, und gibt
 * `{ gruppe: { geprueft, fehlend, beispiele } }`. Jede Probe, deren Wirkung
 * fehlt, trägt `wirkungFehlt` mit dem Grund und zählt danach als Fehler.
 *
 * `abfrage(weg)` liest den Dienst (GET) und gibt eine Probe; ohne sie gibt es
 * keine Prüfung des Bestands, und Proben, die ihn bräuchten, scheitern.
 */
export async function pruefeWirkungen(gruppen, { abfrage, stichprobe = 25, schlafe = pause } = {}) {
  const ergebnis = {}
  for (const [gruppe, proben] of Object.entries(gruppen)) {
    const kandidaten = proben.filter((p) => p.wirkung && istGelungen(p))
    const gruende = new Map()
    const verdeckt = new Map()
    let geprueft = 0

    const nachArt = Map.groupBy(kandidaten, (p) => p.wirkung.art)
    for (const [art, liste] of nachArt) {
      if (art === 'feld') {
        geprueft += liste.length
        const feld = await feldWirkung(liste, abfrage, schlafe)
        for (const [p, grund] of feld.gründe) gruende.set(p, grund)
        for (const [p, grund] of feld.verdeckt) verdeckt.set(p, grund)
      } else if (art === 'einfuhr') {
        const probe = stichprobeAus(liste, stichprobe)
        geprueft += probe.length
        for (const [p, grund] of await einfuhrWirkung(probe, abfrage, schlafe)) gruende.set(p, grund)
      } else {
        // **Eine Art, die niemand prüft, ist ein Fehler im Profil** — kein
        // stilles Bestätigen.
        if (!dateiPruefer[art] && !antwortPruefer[art]) throw new Error(`unbekannte Art einer Wirkung: ${art}`)
        for (const p of stichprobeAus(liste, stichprobe)) {
          geprueft++
          const grund = dateiPruefer[art] ? dateiPruefer[art](p.wirkung) : antwortPruefer[art](p, p.wirkung)
          if (grund !== null) gruende.set(p, grund)
        }
      }
    }

    for (const [p, grund] of gruende) p.wirkungFehlt = grund
    for (const [p, grund] of verdeckt) p.wirkungVerdeckt = grund

    // **Null geprüft ist nicht grün** — QA-LAST01 (#2092). Sagt eine Gruppe
    // Wirkung zu und bringt keinen Ruf durch, steht sie als ungeprüft da, mit
    // den Status, an denen es lag.
    const zugesagt = proben.filter((p) => p.wirkung)
    const status = Object.entries(Object.groupBy(zugesagt, (p) => String(p.status)))
      .map(([wert, liste]) => `${wert} × ${String(liste.length)}`).join(', ')
    ergebnis[gruppe] = {
      zugesagt: zugesagt.length, geprueft, fehlend: gruende.size, verdeckt: verdeckt.size,
      beispiele: [...new Set(gruende.values())].slice(0, 3),
      ...(zugesagt.length > 0 && geprueft === 0
        ? {
            ungeprueft: `keiner von ${String(zugesagt.length)} Rufen mit zugesagter Wirkung gelang — ${status}`,
            nurGedrosselt: zugesagt.every((p) => p.status === 429),
          }
        : {}),
    }
  }
  return ergebnis
}
