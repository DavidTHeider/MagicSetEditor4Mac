#!/usr/bin/env node
// **Die Hüllenprobe** — die Abnahme am nativen Fenster: ENG-02.AK-04
// (#1727),
// die Menüleiste aus #255 und die FIX-Abnahmen #1746–#1748.
//
//     make huellenprobe                         # baut und fährt jedes Drehbuch
//     node scripts/huellenprobe.mjs <drehbuch> … # nach ./scripts/bauen.sh abnahme
//
// # Wie
//
// Das Bündel wird **über LaunchServices** gestartet (`open -n -W --env …`),
// wie ein Doppelklick es täte, mit eigenem Bestand und Bund unter einem
// frischen Ort. Dann spricht die Probe über zwei Wege: die
// **Abnahmeleitung** (`internal/fenster/abnahme.go`) für alles, was AppKit
// gehört — Leiste, Kürzel, Dateidialoge, Fenster —, und **Wails' MCP**
// (`js_eval`) für die Seite. Bilder nimmt `screencapture` vom Fenster und von
// der Leiste; sie liegen unter `$TMPDIR/huellenprobe/` und nicht im Baum.
// Der Beleg — was geprüft wurde, mit den gemessenen Werten — geht nach
// `docs/belege/huellenprobe/<drehbuch>.md`.
//
// # Was sie nicht belegt
//
// Die Maus des Fensterservers: eine Maus, die in die Leiste fährt, ein Ziehen
// aus dem Finder. Tastendrücke gehen **ab AppKit** denselben Weg wie echte —
// Kürzel der Leiste, Schlüsselfenster, Webansicht —, aber sie kommen aus der
// Ereignisschlange der Anwendung und nicht von der Tastatur. Und ob ein
// Mensch findet, was er sucht, sieht nur ein Mensch (`docs/HANDPRUEFUNGEN.md`).
import { execFileSync, spawn } from 'node:child_process'
import { mkdirSync, mkdtempSync, openSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { fensterbild, schreibeBeleg } from './huellenbeleg.mjs'
import { fahreDrehbuecher } from './huellenlauf.mjs'
import { klappeAuf, klickeAuf } from './huellenklick.mjs'
import { kindumgebung } from './huellenorte.mjs'
import { Leitung, mcpRuf, schlafe } from './huellenleitung.mjs'
import { rufbuchHolen } from './huellenrufbuch.mjs'

const wurzel = resolve(dirname(fileURLToPath(import.meta.url)), '..')

// oberflaecheSteht ist die Frage von `bereit` — ausgeführt, damit der Gastlauf
// von B02 (`budgetsgast.mjs`) dieselbe stellt.
export const oberflaecheSteht = "return document.readyState === 'complete' && (document.querySelectorAll('nav button').length > 3 || !!document.querySelector('[data-wiederherstellung]') || !!document.querySelector('[data-sitzung]'))"

// textziel findet den sichtbaren Knopf, Reiter oder Verweis mit diesem Text —
// zuerst genau, sonst am Anfang.
// **Zuletzt ohne das Zeichen davor** — ein Knopf aus dem Aktionskatalog
// trägt eines („✓ Wahl übernehmen"), und g08 fand ihn nach der Umstellung der
// Lauschwahl nicht mehr (nativer Lauf am 2026-10-10).
export const textziel = (text) => `(() => { const glatt = (e) => (e.innerText ?? e.textContent ?? '').replace(/\\s+/g, ' ').trim(); const alle = [...document.querySelectorAll('button, a, [role=button], [role=tab], [role=menuitem], [role=option], summary, label')].filter((e) => e.checkVisibility()); return alle.find((e) => glatt(e) === ${JSON.stringify(text)}) || alle.find((e) => glatt(e).startsWith(${JSON.stringify(text)})) || alle.find((e) => glatt(e).replace(/^[^\\p{L}\\p{N}]+/u, '') === ${JSON.stringify(text)}) })()`
const buendel = join(wurzel, 'build/bin/MagicSetEditor4Mac.app')
// abnahmekennung ist die Kennung des Abnahmebaus (`scripts/bauen.sh`).
const abnahmekennung = 'io.github.davidtheider.magicseteditor4mac.abnahme'
const bilderort = join(tmpdir(), 'huellenprobe')


export class Probe {
  // orte sind die Orte des laufenden Drehbuchs — `huellenlauf.mjs` leert sie
  // davor und liest danach ihre Protokolle (SEC-17).
  static orte = []

  constructor(name) {
    this.name = name
    this.zeilen = []
    this.fehler = 0
    this.ohneBild = 0
    this.bilder = join(Probe.laufort ? join(Probe.laufort, 'bilder') : bilderort, name)
    mkdirSync(this.bilder, { recursive: true })
    this.port = 9190 + Math.floor(Math.random() * 400)
    this.aufruf = 0
    this.pids = []
  }

  // aufraeumen beendet jede Anwendung dieses Laufs, die noch steht — ein
  // Drehbuch, das mittendrin abbricht, ließe sonst ein Fenster zurück, und
  // das nächste stritte mit ihm um den Vordergrund.
  aufraeumen() {
    for (const pid of this.pids) {
      try {
        process.kill(pid)
      } catch {
        // schon fort
      }
    }
  }

  // ort legt einen frischen Bestand samt Bund an — eine zweite Anwendung
  // desselben Laufs bekommt ihren eigenen. **`zuhause` gibt ein eigenes
  // `HOME`**: Pakete liegen am Vorgabeort und nicht neben dem Bestand
  // (QA-10.NAT-01); wer sie anfasst, täte es sonst im echten `~/Library`.
  // **Mit eigenem Zuhause als Vorgabe** — T01 (#2095). Bis zum 2026-10-06
  // musste ein Drehbuch es verlangen; 63 von 78 taten es nicht, und ihre
  // Anwendung und ihr `mse` lasen und schrieben unter dem echten `HOME`.
  ort({ zuhause = true } = {}) {
    const ort = mkdtempSync(join(tmpdir(), 'huelle-'))
    const heim = zuhause ? join(ort, 'zuhause') : undefined
    if (heim) mkdirSync(heim)
    const bund = join(ort, 'bund.keychain-db')
    execFileSync('security', ['create-keychain', '-p', 'leer', bund])
    execFileSync('security', ['unlock-keychain', '-p', 'leer', bund])
    // **Und er sperrt sich nicht von selbst**: nach fünf Minuten hätte er
    // es getan, und `security` fragte dann nach dem Kennwort — in einem
    // Fenster, vor dem niemand sitzt. Ohne Angaben heißt: nie.
    execFileSync('security', ['set-keychain-settings', bund])

    const angelegt = { ort, bestand: join(ort, 'bestand.db'), bund, heim }
    // **Jeder Ort des Drehbuchs wird nach dem Lauf gelesen** — auch der
    // eines zweiten Prozesses wie des Gastes in `nat04` (SEC-17, #2164).
    Probe.orte.push(angelegt)

    return angelegt
  }

  // mse legt im Bestand dieser Probe an — derselbe Weg wie in
  // `scripts/webkitabnahme.sh`, und ein Fehler bricht ab.
  mse(ort, ...argumente) {
    return execFileSync(join(wurzel, '.go/bin/mse'), argumente, { cwd: wurzel, env: kindumgebung(ort) }).toString()
  }

  // starte öffnet das Bündel über LaunchServices; `datei` geht wie beim
  // Doppelklick mit, `sprachen` setzt `AppleLanguages` nur für diesen Prozess (#255),
  // `umgebung` legt Variablen nur für diesen Start dazu (SEC-17, #2164).
  // `takt`: wie oft `bereit` nachfragt — kürzer, wo `budgets` die Startzeit misst.
  // `einfuehrung`: die Einführung eines frischen Bestands **bleibt stehen** —
  // für Wege, die beim ersten Start beginnen (G01, #2248). `buendel`: ein anderes, etwa aus einem Tag (M01, #2345).
  async starte({ ort = this.ort(), datei, direkt = false, sprachen, umgebung = {}, takt = 150, einfuehrung = false, buendel: ziel = buendel } = {}) {
    this.jetzt = ort
    this.takt = takt
    this.einfuehrungLassen = einfuehrung
    // **Jeder Start bekommt seinen eigenen Hafen**: der MCP-Server der
    // vorigen Anwendung kann ihn noch halten, wenn die nächste schon lauscht.
    this.port = 9190 + Math.floor(Math.random() * 400)
    const sock = join(ort.ort, 'abnahme.sock')
    if (direkt) {
      // **Direkt gestartet** — zur Fehlersuche: die Ausgabe, die LaunchServices
      // verschluckt, kommt in `ort/ausgabe.log`.
      const ausgabe = openSync(join(ort.ort, 'ausgabe.log'), 'w')
      this.prozess = spawn(join(ziel, 'Contents/MacOS/MagicSetEditor4Mac'), datei ? [datei] : [], {
        env: { ...kindumgebung(ort), ...umgebung, MSE_ABNAHME: sock, WAILS_MCP_PORT: String(this.port), WAILS_MCP_HIDE_CURSOR: '1' },
        stdio: ['ignore', ausgabe, ausgabe],
      })
      this.beendet = new Promise((los) => this.prozess.once('exit', los))

      return this.verbinden(sock, ort)
    }
    const argumente = ['-n', '-W', '-a', ziel,
      ...(ort.heim ? ['--env', `HOME=${ort.heim}`] : []),
      '--env', `MSE_BESTAND=${ort.bestand}`, '--env', `MSE_SCHLUESSELBUND=${ort.bund}`,
      '--env', 'MSE_DIENST_GESPERRT=an', // kein Ruf an launchd (#2323, Nachlauf, Befund 6)
      '--env', `MSE_ABNAHME=${sock}`,
      '--env', `WAILS_MCP_PORT=${String(this.port)}`,
      '--env', 'WAILS_MCP_HIDE_CURSOR=1',
      ...Object.entries(umgebung).flatMap(([name, wert]) => ['--env', `${name}=${wert}`])]
    if (datei) argumente.push(datei)
    if (sprachen) argumente.push('--args', '-AppleLanguages', `(${sprachen})`)
    this.prozess = spawn('open', argumente, { stdio: 'ignore' })
    this.beendet = new Promise((los) => this.prozess.once('exit', los))

    return this.verbinden(sock, ort)
  }

  // verbinden öffnet die Leitung und holt über sie den **Schlüssel** für
  // Wails' MCP-Server — ohne ihn weist der Server jede Anfrage ab.
  async verbinden(sock, ort) {
    this.leitung = new Leitung(sock)
    await this.leitung.verbinde()
    this.schluessel = await this.frage({ was: 'schluessel' })
    this.pids.push((await this.fenster()).prozess)
    await this.bereit()

    return ort
  }

  // bereit wartet, bis die Oberfläche steht — die Leiste links ist gezeichnet
  // und hört zu. Ein Menüeintrag, der vorher ankommt, geht ins Leere: die
  // Seite meldet sich erst an, wenn sie gezeichnet ist.
  async bereit() {
    // **Die Leiste links, oder die Wiederherstellung**: ein unbrauchbarer
    // Bestand öffnet ohne Leiste (#962). **Oder die Sitzung** (ID-02, #1701):
    // gesperrt oder abgemeldet steht die Anwendung gar nicht da.
    await this.warte('die Oberfläche steht', () => this.seite(oberflaecheSteht), 30000, this.takt)
    this.bereitUm = Date.now() // **vor** der Pause — `budgets` maß sie bis zum vierten Lauf mit (B02)
    await schlafe(500)
    // **„Erste Schritte" geht mit Escape zu** — ein frischer Bestand zeigt
    // sie über allem. Erkannt an `data-einfuehrung`, nicht am Titel (LOC-01).
    const einfuehrung = "return !!document.querySelector('[data-einfuehrung]')?.checkVisibility()"
    if (!this.einfuehrungLassen && await this.seite(einfuehrung)) {
      await this.taste('Escape')
      await this.warte('die Einführung ist zu', async () => !(await this.seite(einfuehrung)))
    }
  }

  async frage(befehl) {
    const antwort = await this.leitung.frage(befehl)
    if (!antwort.ok) throw new Error(`${befehl.was}: ${antwort.fehler}`)

    return antwort.daten
  }

  menue() { return this.frage({ was: 'menue' }) }
  klick(...pfad) { return this.frage({ was: 'klick', pfad }) }
  // imFenster löst einen Eintrag aus, als wäre das Fenster vorn — eine
  // Handlung ohne eigenes Ziel geht an seinen Ersten Antwortenden
  // (`internal/fenster/abnahme_darwin.go`, `klicken`).
  imFenster(...pfad) { return this.frage({ was: 'imfenster', pfad }) }
  taste(taste, ...mit) { return this.frage({ was: 'taste', taste, mit }) }
  async tippe(text) {
    for (const zeichen of text) {
      const gross = zeichen !== zeichen.toLowerCase()
      await this.taste(zeichen.toLowerCase(), ...(gross ? ['umschalt'] : []))
    }
  }
  // tippeInsFeld tippt in das Feld mit dem Fokus und wartet, **bis der Wert
  // dasteht** — erst dann ist auch die Oberfläche so weit. Ein Klick direkt
  // danach nahm sonst den halben Namen: von „Nachtprobe" kam „Na" an.
  async tippeInsFeld(text) {
    const vorher = await this.seite('return document.activeElement?.value ?? ""')
    await this.tippe(text)
    await this.warte(`im Feld steht »${vorher}${text}«`, async () =>
      (await this.seite('return document.activeElement?.value ?? ""')) === `${vorher}${text}`)
    await schlafe(300)
  }

  // dialogantwort legt die Antwort für den nächsten Dateidialog bereit —
  // bestätigen lässt sich der echte von hier aus nicht
  // (`internal/fenster/abnahme_darwin.go`, `vorbereitet`).
  dialogantwort(datei) { return this.frage({ was: 'dialogantwort', datei }) }

  // dialogZeigtSich wartet auf den echten Dialog, liest ihn und bricht ihn
  // ab — so ist belegt, dass er aufgeht und dass ein Abbruch nichts tut.
  async dialogZeigtSich(art, was) {
    const blatt = await this.warte(`${was} öffnet den Dialog (${art})`, async () => {
      const b = await this.blatt()
      return b.art === art ? b : false
    })
    this.pruefe(`${was} öffnet den Dialog von macOS`, true,
      `${blatt.nachricht ?? ''}${blatt.name ? `, Vorschlag »${blatt.name}«` : ''}`)
    await this.blattabbruch()
    await this.warte('der Dialog ist wieder zu', async () => (await this.blatt()).art === '')
    // **Und der Vorgang dahinter ist zu Ende** — die Seite erfährt vom
    // Abbruch eine Nachricht später, und ein Handgriff davor träfe einen
    // Vorgang, der noch läuft (§7, Regel 4: er wird dann nicht angenommen).
    await schlafe(1500)

    return blatt
  }

  blatt() { return this.frage({ was: 'blatt' }) }
  blattantwort(datei) { return this.frage({ was: 'blattantwort', datei }) }
  blattabbruch() { return this.frage({ was: 'blattantwort', abbrechen: true }) }
  fenster() { return this.frage({ was: 'fenster' }) }
  aktiviere() { return this.frage({ was: 'aktivieren' }) }

  // klickeAuf klickt **nativ** in die Mitte eines Elements — Trefferprüfung
  // und Fokus wie bei einer Maus; warum ohne Bewegung gerollt und bis zum
  // `mouseup` gewartet wird, steht bei der Funktion (`huellenklick.mjs`).
  klickeAuf(ausdruck, beschreibung = ausdruck, art = {}) { return klickeAuf(this, ausdruck, beschreibung, art) }

  // klappeAuf öffnet eine Falte nativ und verlangt, dass sie danach offen ist.
  klappeAuf(ausdruck, beschreibung = ausdruck) { return klappeAuf(this, ausdruck, beschreibung) }

  // klickeText klickt, was `textziel` findet.
  klickeText(text) { return this.klickeAuf(textziel(text), `»${text}«`) }

  // ruhe wartet, bis die Seite nicht mehr wächst — drei Abfragen, dieselbe
  // Höhe. **Gemessen am 2026-10-05**: nach dem Wechsel der Gruppe in den
  // Einstellungen fügt die Seite 100 bis 130 ms später 296 Punkte über den
  // tieferen Abschnitten ein; wer dort sofort zielte, traf daneben (`ops02`,
  // `ops06`). Schneller, als ein Mensch zielt, deshalb ein Warten und kein
  // Befund.
  async ruhe(frist = 10000) {
    const hoehen = []
    await this.warte('die Seite steht', async () => {
      hoehen.push(await this.seite('return document.documentElement.scrollHeight'))

      return hoehen.length >= 3 && hoehen.slice(-3).every((h) => h === hoehen.at(-1))
    }, frist)
  }

  // wartetext wartet, bis der Text irgendwo sichtbar steht.
  wartetext(text, frist) {
    return this.warte(`»${text}« steht da`, () => this.seite(`return document.body.innerText.includes(${JSON.stringify(text)})`), frist)
  }

  // mcp ruft ein Werkzeug von Wails' MCP-Server und gibt seinen Text.
  async mcp(name, argumente, frist) {
    this.aufruf += 1

    return mcpRuf({ port: this.port, schluessel: this.schluessel, id: this.aufruf }, name, argumente, frist)
  }

  // seite wertet JavaScript in der Webansicht aus — über Wails' MCP, damit
  // die Antwort zurückkommt. Das Ergebnis reist als JSON.
  async seite(js) {
    const inhalt = await this.mcp('js_eval', { js: `return JSON.stringify(await (async () => { ${js} })())` })

    return inhalt === 'undefined' || inhalt === 'ok' ? undefined : JSON.parse(inhalt)
  }

  // groesse setzt die Fenstergröße über Wails — so, wie ein Ziehen am Rand
  // sie am Ende setzt.
  groesse(breite, hoehe) {
    return this.mcp('window_control', { action: 'set_size', width: breite, height: hoehe })
  }

  // warte fragt, bis `bedingung` wahr ist — **auf die Wirkung**, nicht auf
  // eine Frist (§3).
  //
  // **Die Frist gilt auch für einen Ruf, der nie zurückkehrt** (T02, #2096):
  // er läuft gegen den Rest der Frist, und was er offen hält, schließt seine
  // eigene Frist (`mcp`, `Leitung.frage`).
  async warte(beschreibung, bedingung, frist = 15000, takt = 150) {
    const ende = Date.now() + frist
    let zuletzt
    for (;;) {
      let uhr
      try {
        zuletzt = await Promise.race([
          bedingung(),
          new Promise((_, weg) => {
            uhr = setTimeout(() => weg(new Error('keine Antwort')), Math.max(0, ende - Date.now()) + 50)
          }),
        ])
        if (zuletzt) return zuletzt
      } catch (grund) {
        zuletzt = grund.message
      } finally {
        clearTimeout(uhr)
      }
      if (Date.now() > ende) {
        this.pruefe(`${beschreibung} (nach ${String(frist)} ms)`, false, String(zuletzt))
        throw new Error(`abgebrochen: ${beschreibung}`)
      }
      await schlafe(takt)
    }
  }

  pruefe(beschreibung, erfuellt, zeige = '') {
    if (!erfuellt) this.fehler += 1
    this.zeilen.push(`- ${erfuellt ? '✓' : '✗'} ${beschreibung}${zeige === '' ? '' : ` — ${zeige}`}`)

    return erfuellt
  }

  notiz(text) {
    this.zeilen.push('', `**${text}**`, '')
  }

  zeige(text) {
    this.zeilen.push(`- ${text}`)
  }

  // bild nimmt das Fenster auf — **oder vermerkt eine Lücke**, wenn der
  // Bildschirm schläft (`huellenbeleg.mjs`).
  async bild(name) {
    const { nummer } = await this.fenster()
    const pfad = join(this.bilder, `${name}.png`)
    if (!fensterbild(nummer, pfad)) {
      this.ohneBild++
      this.zeige(`Bild »${name}«: **fehlt** — der Bildschirm schlief, und \`screencapture\` bekam kein Fenster`)

      return undefined
    }
    this.zeige(`Bild: \`${pfad}\``)

    return pfad
  }

  async leistenbild(name) {
    const pfad = join(this.bilder, `${name}.png`)
    execFileSync('screencapture', ['-x', '-R', '0,0,1512,34', pfad])
    this.zeige(`Bild der Leiste: \`${pfad}\``)

    return pfad
  }

  // beende nimmt ⌘Q, wie ein Mensch — oder den Eintrag der Leiste — und
  // wartet, bis der Prozess weg ist. **Die Antwort darf ausbleiben**: wer die
  // Anwendung beendet, beendet auch die Leitung.
  async beende({ ueber = 'taste' } = {}) {
    const befehl = ueber === 'taste'
      ? { was: 'taste', taste: 'q', mit: ['befehl'] }
      : { was: 'klick', pfad: ['MagicSetEditor4Mac', 'MagicSetEditor4Mac beenden'] }
    await rufbuchHolen(this)
    const antwort = await this.leitung.frage(befehl)
    if (!antwort.ok && !antwort.zu) throw new Error(`beenden: ${antwort.fehler}`)
    const weg = await Promise.race([this.beendet.then(() => true), schlafe(15000).then(() => false)])
    this.leitung.schliesse()

    return weg
  }

  beleg() { return schreibeBeleg(this, wurzel) }
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  await fahreDrehbuecher(Probe, { wurzel, buendel, abnahmekennung, namen: process.argv.slice(2) })
}
