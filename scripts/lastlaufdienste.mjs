// **Die Prozesse des Lastlaufs** — Z4 des Zweitblicks über Phase 6.11
// (QA-08, #1739).
//
// Eine eigene Datei aus demselben Grund wie `lastlaufnetz.mjs`: `lastlauf.mjs`
// baut und fährt, hier steht, **wie Prozesse gehalten und beendet werden**. Und
// wie dort wird kein Prozess von dieser Datei selbst erzeugt: der Läufer reicht
// die Erzeugung herein (`Kinderregister`), damit eine Prüfung ihre eigenen
// Programme nehmen kann (`internal/werkzeuge/nahtlauf_test.go`).
//
// # Was schiefging
//
//   - Der Lauf startete `mse-rest`, dann `mse-mcp`. Scheiterte der zweite,
//     blieb der erste liegen, denn aufgeräumt wurde nur nach einem
//     **gelungenen** Profil. Dasselbe galt für eine Frist beim Start und für
//     jeden Abbruch von außen: der Handler beim Verlassen räumte den
//     Arbeitsordner und sonst nichts.
//   - Ein Ruf über stdio wartete auf eine Antwort, die ein gestorbener Prozess
//     nie schickte — für immer.
//
// # Wie es jetzt ist
//
// **Jeder Prozess steht in einem Register**, vom Start bis zu seinem Ende.
// Der Läufer beendet das Register in `finally`, bei einem Signal und — ohne zu
// warten — beim Verlassen des Prozesses. **Ein Ruf, dessen Prozess endet, wird
// abgewiesen**, mit dem Ende als Grund; ebenso einer, dessen Antwort ausbleibt.
// Ein Prozess, der eine Ansage schuldet und sie nicht macht, wird beendet, nicht
// nur aufgegeben.
//
// **Was das nicht kann**: ein `SIGKILL` gegen den Läufer selbst. Dann bleiben
// die Dienste, wo sie sind — sie lauschen auf der Schleife, mit einer Marke,
// die nur im Speicher des toten Läufers stand, und sind unter `TMPDIR` zu
// finden.

import http from 'node:http'
import { performance } from 'node:perf_hooks'

import { mcpErgebnis } from './lastlaufnetz.mjs'

/** lebt sagt, ob ein Prozess noch läuft — weder mit einem Rückgabewert noch mit einem Signal beendet. */
const lebt = (kind) => kind.exitCode === null && kind.signalCode === null

/**
 * beendeKind schickt SIGTERM und wartet auf das Ende — höchstens `frist`
 * Millisekunden, danach folgt SIGKILL.
 */
export async function beendeKind(kind, frist = 5000) {
  if (!lebt(kind)) return
  await new Promise((fertig) => {
    const zwang = setTimeout(() => kind.kill('SIGKILL'), frist)
    kind.once('exit', () => { clearTimeout(zwang); fertig() })
    kind.kill('SIGTERM')
  })
}

/** Das Register aller Prozesse eines Laufs. */
export class Kinderregister {
  #erzeuge
  #kinder = new Set()

  constructor(erzeuge) {
    this.#erzeuge = erzeuge
  }

  /** zahl gibt, wie viele Prozesse noch im Register stehen. */
  get zahl() {
    return this.#kinder.size
  }

  /** starte erzeugt einen Prozess und trägt ihn ein, bis er endet. */
  starte(programm, argumente, optionen) {
    const kind = this.#erzeuge(programm, argumente, optionen)
    this.#kinder.add(kind)
    kind.once('exit', () => this.#kinder.delete(kind))
    // **Ein Prozess, der nie anlief**, endet nicht mit `exit`, sondern mit einem
    // Fehler; ohne Behandler risse er den Läufer mit.
    kind.on('error', () => this.#kinder.delete(kind))
    return kind
  }

  /** beendeAlle beendet jeden Prozess im Register und wartet auf alle. */
  async beendeAlle(frist = 5000) {
    await Promise.all([...this.#kinder].map((kind) => beendeKind(kind, frist)))
  }

  /**
   * beendeSofort beendet ohne zu warten — für den Handler beim Verlassen des
   * Prozesses, in dem nichts mehr asynchron geschehen kann.
   */
  beendeSofort() {
    for (const kind of this.#kinder) if (lebt(kind)) kind.kill('SIGKILL')
  }
}

/**
 * starteDienst startet `mse-rest` oder `mse-mcp` über HTTP und wartet auf seine
 * Ansage: Adresse und Marke stehen in den ersten Zeilen.
 *
 * **Der Klient hält höchstens `verbindungen` offen** — beim Dienst die
 * sechzehn, die er annimmt (`internal/rest/gedraenge.go`). Wer mehr Klienten
 * fährt, wartet beim Klienten statt im Rückstau des Betriebssystems: dort
 * bekäme eine Anfrage keinen Platz, solange die sechzehn ruhenden
 * Verbindungen der anderen offen bleiben. Was das kostet, misst `langsam`.
 *
 * **Bleibt die Ansage aus, wird der Prozess beendet**, nicht nur aufgegeben.
 */
export function starteDienst(register, programm, argumente, umgebung, verbindungen, { frist = 30_000 } = {}) {
  const kind = register.starte(programm, argumente, { env: umgebung, stdio: ['pipe', 'pipe', 'pipe'] })
  let mitschrift = ''

  return new Promise((gut, schlecht) => {
    const scheitern = (fehler) => {
      clearTimeout(zeitgeber)
      if (lebt(kind)) kind.kill('SIGKILL')
      schlecht(fehler)
    }
    const zeitgeber = setTimeout(() => scheitern(new Error(`${programm} hat nichts angesagt:\n${mitschrift}`)), frist)

    const lesen = (stueck) => {
      mitschrift += stueck
      const adresse = mitschrift.match(/^Dienst auf (https?:\/\/[^\s/]+)(\S*)$/m)
      const marke = mitschrift.match(/^Marke (\S+)$/m)

      if (adresse && marke) {
        clearTimeout(zeitgeber)
        const agent = new http.Agent({ keepAlive: true, maxSockets: verbindungen })
        gut({ kind, adresse: adresse[1], pfad: adresse[2], marke: marke[1], agent })
      }
    }

    kind.stdout.setEncoding('utf8').on('data', lesen)
    kind.stderr.setEncoding('utf8').on('data', (stueck) => { mitschrift += stueck })
    kind.on('exit', (code) => scheitern(new Error(`${programm} endete mit ${code}:\n${mitschrift}`)))
    kind.on('error', (fehler) => scheitern(new Error(`${programm} ließ sich nicht starten: ${fehler.message}`)))
  })
}

/**
 * starteNacheinander startet die Dienste der Reihe nach — jeder Auftrag ist
 * `[programm, argumente, umgebung, verbindungen]`. **Scheitert einer, werden die
 * schon gestarteten beendet** und der Fehler geht weiter; nacheinander und nicht
 * zugleich, weil beide denselben Bestand öffnen.
 */
export async function starteNacheinander(register, auftraege) {
  const dienste = []
  try {
    for (const [programm, argumente, umgebung, verbindungen] of auftraege) {
      dienste.push(await starteDienst(register, programm, argumente, umgebung, verbindungen))
    }
  } catch (fehler) {
    await Promise.all(dienste.map((dienst) => beende(dienst)))
    throw fehler
  }
  return dienste
}

/** beende schließt die Verbindungen eines Dienstes und beendet seinen Prozess. */
export async function beende(dienst, frist = 5000) {
  dienst.agent?.destroy()
  await beendeKind(dienst.kind, frist)
}

/**
 * mcpUeberStdio startet einen Werkzeugserver über die Standardeingabe und gibt
 * eine Funktion, die ein Werkzeug ruft und auf **seine** Antwort wartet.
 *
 * **Jeder offene Ruf endet**: mit der Antwort, mit dem Ende des Prozesses
 * (dann sind alle abgewiesen) oder nach `frist` Millisekunden. Ein Ruf nach dem
 * Ende wird sofort abgewiesen. Gelesen wird das Ende erst mit `close`, wenn die
 * Standardausgabe leer gelesen ist — eine Antwort, die noch unterwegs war, geht
 * nicht verloren.
 */
export async function mcpUeberStdio(register, programm, argumente, umgebung, { frist = 120_000 } = {}) {
  const kind = register.starte(programm, argumente, { env: umgebung, stdio: ['pipe', 'pipe', 'pipe'] })
  const offen = new Map()
  let rest = ''
  let kennung = 0
  let beendet = null

  const abweisen = (grund) => {
    for (const eintrag of offen.values()) {
      clearTimeout(eintrag.zeitgeber)
      eintrag.schlecht(new Error(grund))
    }
    offen.clear()
  }
  const endeSagen = (text) => {
    beendet ??= text
    abweisen(`${beendet}, bevor es antwortete`)
  }

  kind.stdout.setEncoding('utf8').on('data', (stueck) => {
    rest += stueck
    let ende = rest.indexOf('\n')

    while (ende >= 0) {
      const zeile = rest.slice(0, ende)
      rest = rest.slice(ende + 1)
      ende = rest.indexOf('\n')

      try {
        const id = JSON.parse(zeile).id
        const eintrag = offen.get(id)
        if (!eintrag) continue
        clearTimeout(eintrag.zeitgeber)
        offen.delete(id)
        eintrag.gut(zeile)
      } catch { /* eine Zeile, die keine Antwort ist, gehört keinem Ruf */ }
    }
  })

  kind.stderr.resume()
  kind.stdin.on('error', () => {})
  kind.on('close', (code, signal) => endeSagen(`${programm} endete (Code ${code}, Signal ${signal})`))
  kind.on('error', (fehler) => endeSagen(`${programm} ließ sich nicht starten (${fehler.message})`))

  const senden = (methode, parameter) => new Promise((gut, schlecht) => {
    if (beendet) {
      schlecht(new Error(`${beendet}, bevor der Ruf ${methode} ging`))
      return
    }

    kennung += 1
    const id = kennung
    const beginn = performance.now()
    const zeitgeber = setTimeout(() => {
      offen.delete(id)
      schlecht(new Error(`${programm}: keine Antwort auf ${methode} nach ${frist} ms`))
    }, frist)
    offen.set(id, { zeitgeber, schlecht, gut: (zeile) => gut({ ms: performance.now() - beginn, text: zeile }) })
    kind.stdin.write(`${JSON.stringify({ jsonrpc: '2.0', id, method: methode, params: parameter })}\n`)
  })

  await senden('initialize', { protocolVersion: '2025-06-18', capabilities: {}, clientInfo: { name: 'lastlauf', version: '1' } })

  return {
    kind,
    async ruf(werkzeug, argumente) {
      const antwort = await senden('tools/call', { name: werkzeug, arguments: argumente })

      return { ...antwort, status: 200, ...mcpErgebnis(antwort.text) }
    },
    /** ende schließt die Eingabe und wartet auf das Ende des Prozesses — danach SIGKILL. */
    async ende(wartezeit = 5000) {
      kind.stdin.end()
      if (!lebt(kind)) return
      await new Promise((fertig) => {
        const zwang = setTimeout(() => kind.kill('SIGKILL'), wartezeit)
        kind.once('exit', () => { clearTimeout(zwang); fertig() })
      })
    },
  }
}
