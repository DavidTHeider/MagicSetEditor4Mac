// **Die Teilnehmer des Leistungsvergleichs** — QA-04.AK-01
// (#1735).
//
// Je Programm dieselben Aufgaben über **seine** kopflosen Wege. Welche das
// sind und warum gerade diese, steht in `docs/LEISTUNGSVERGLEICH.md` §1.
//
// **Gestartet wird im Läufer**: `messe` fährt einen ganzen Prozess unter
// `/usr/bin/time -l`, `starte` gibt einen laufenden für eine Sitzung. Hier
// steht nur, **welcher** Prozess mit welcher Eingabe.

import { mkdirSync, mkdtempSync } from 'node:fs'
import { basename, dirname, join } from 'node:path'
import { performance } from 'node:perf_hooks'

import { McpSitzung, Mse2Sitzung } from './leistungssitzung.mjs'

/**
 * mse2 baut den Teilnehmer MSE2.
 *
 * # Ein eigenes Zuhause — und `HOME` allein gibt keines
 *
 * **Gemessen am 2026-10-02**: MSE2 sucht sein Benutzerverzeichnis über
 * wxWidgets, und das fragt auf macOS **Core Foundation** und nicht `HOME`. Mit
 * einem leeren `HOME` lud es `magic.mse-game` trotzdem aus
 * `~/Library/Application Support/magicseteditor/data` — der Warnung nach. Erst
 * `CFFIXED_USER_HOME` lenkt es um; dann liest es nur `build/data`
 * (`PackageManager::openAny`, `src/util/io/package_manager.cpp`, Zeilen 47–54,
 * Fassung f9d9356 vom 2021-07-06).
 *
 * **Und das Verzeichnis darin muss es geben.** Ohne
 * `Library/Application Support/magicseteditor` hing MSE2 beim Beenden — es
 * schreibt dort `mse.config`, und ohne Ordner wartet es auf eine Antwort, die
 * kopflos nie kommt.
 */
export function mse2({ binaer, arbeit, messe, starte }) {
  const umgebung = () => {
    const zuhause = mkdtempSync(join(arbeit, 'mse2-zuhause-'))
    mkdirSync(join(zuhause, 'Library/Application Support/magicseteditor'), { recursive: true })
    return { ...process.env, HOME: zuhause, CFFIXED_USER_HOME: zuhause }
  }
  const cli = (set) => (set ? ['--cli', set.pfad, '--quiet', '--raw'] : ['--cli', '--quiet', '--raw'])
  const oeffnen = async (set) => {
    const m = await messe(binaer, cli(set), { umgebung: umgebung(), eingabe: ':info\n:quit\n' })
    // `:info` nennt `#cards:   N` (`src/cli/cli_main.cpp`, `handleCommand`,
    // Fassung f9d9356 vom 2021-07-06).
    const karten = m.aus.match(/#cards:\s*(\d+)/)
    return { ...m, karten: karten ? Number(karten[1]) : null }
  }
  return {
    name: 'mse2',
    start: () => messe(binaer, cli(null), { umgebung: umgebung(), eingabe: ':quit\n' }),
    oeffnen,
    sichern: (set, ziel) => messe(binaer, cli(set), {
      umgebung: umgebung(), ort: dirname(ziel),
      eingabe: `write_set_file(file:${JSON.stringify(basename(ziel))})\n:quit\n`,
    }),
    bilder: (set, verzeichnis) => messe(binaer,
      ['--export-images', set.pfad, join(verzeichnis, '{card.name}.png')], { umgebung: umgebung() }),
    ausfuhr: (set, vorlage, ziel) => messe(binaer, ['--export', vorlage, set.pfad, ziel], { umgebung: umgebung() }),
    async sitzung(set) {
      const anfang = performance.now()
      const lauf = starte(binaer, cli(set), { umgebung: umgebung() })
      const s = new Mse2Sitzung(lauf.kind)
      // **Bereit ist MSE2, wenn es den ersten Befehl beantwortet** — das Set
      // lädt es, bevor es die Eingabe liest (`src/main.cpp`, Zeile 225). Ohne
      // diesen Schritt trüge die erste Karte das Laden mit.
      await s.befehl(':info')
      const bereitMs = performance.now() - anfang
      const antwort = (satz) => ({ ok: satz.status < 2, dauerMs: satz.dauerMs, text: satz.zeilen.join('\n') })
      const scheitern = (fehler) => ({ ok: false, dauerMs: null, text: String(fehler.message ?? fehler) })
      return {
        lauf,
        bereitMs,
        // **Ein Dateiname wird nur einmal geschrieben** — ein zweiter Ruf mit
        // demselben Namen gibt ihn zurück, ohne zu zeichnen
        // (`src/script/functions/export.cpp`, Zeilen 404–408, Fassung f9d9356
        // vom 2021-07-06). Der Läufer gibt deshalb jedem Schritt einen eigenen.
        zeichne: (stelle, ziel) => s.befehl(`write_image_file(set.cards[${stelle}], file:${JSON.stringify(ziel)})`)
          .then(antwort, scheitern),
        name: (stelle) => s.befehl(`set.cards[${stelle}].name`).then(antwort, scheitern),
        skript: (ausdruck) => s.befehl(ausdruck).then(antwort, scheitern),
      }
    },
  }
}

/** feldwert liest ein Feld aus einer Karte von `lesen.karten`. */
function feldwert(karte, feld) {
  return (karte.Values ?? karte.werte ?? []).find?.((w) => w.Field === feld)?.Text ?? karte.name ?? null
}

/**
 * mse4mac baut einen Teilnehmer aus den gebauten `mse` und `mse-mcp`. Das
 * Zuhause trägt das Paketverzeichnis, der Bestand kommt je Aufruf, und der
 * Schlüsselbund ist einer, den es nicht gibt (`internal/flaechen`,
 * `eigenerBestand`).
 */
export function mse4mac({ name, verzeichnis, zuhause, messe, starte }) {
  const mse = join(verzeichnis, 'mse')
  const umgebung = (bestand) => ({
    ...process.env, HOME: zuhause, MSE_BESTAND: bestand,
    MSE_SCHLUESSELBUND: join(zuhause, 'keiner.keychain'),
  })
  const json = (m) => {
    try {
      return JSON.parse(m.aus)
    } catch {
      return null
    }
  }
  return {
    name,
    verzeichnis,
    umgebung,
    start: (bestand) => messe(mse, ['sets', '--json'], { umgebung: umgebung(bestand) }),
    async oeffnen(set, bestand) {
      const m = await messe(mse, ['oeffnen', set.pfad, '--json'], { umgebung: umgebung(bestand) })
      const antwort = json(m)
      return { ...m, karten: antwort?.karten ?? null, setID: antwort?.setID ?? antwort?.id ?? null }
    },
    async wiederoeffnen(set, bestand) {
      const m = await messe(mse, ['karten', String(set.id), '--json'], { umgebung: umgebung(bestand) })
      const antwort = json(m)
      const karten = Array.isArray(antwort) ? antwort : antwort?.karten
      return { ...m, karten: Array.isArray(karten) ? karten.length : null }
    },
    // kontingente schaltet das Guthaben im Bestand — `an` oder `aus`.
    //
    // **Der Vergleich misst die Maschine, die Lastgrenze die Drossel** (#2005):
    // seit Zeichnen schnell ist, griff das Guthaben von `mse-mcp` nach 71
    // Karten. MSE2 hat keines; `lastlauf.mjs` schaltet es für „was die
    // Maschine kann“ ebenso ab (QA-08). Der Läufer schaltet es einmal je Lauf
    // ab und stellt es vor der Lastgrenze und am Ende zurück
    // (`leistungskontingent.mjs`).
    kontingente: (bestand, stand) => messe(mse, ['einstellungen', 'setzen', 'grenzen.kontingente', stand],
      { umgebung: umgebung(bestand) }),
    // kontingenteStand fragt die Einstellungen desselben Bestands — der
    // Läufer stellt nach dem Vergleich zurück, was er vorgefunden hat.
    kontingenteStand: (bestand) => messe(mse, ['einstellungen', '--json'], { umgebung: umgebung(bestand) }),
    sichern: (set, ziel, bestand) => messe(mse, ['sichern', String(set.id), ziel], { umgebung: umgebung(bestand) }),
    ausfuhr: (set, vorlage, ziel, bestand) => messe(mse,
      ['vorlagen', String(set.id), 'ausfuehren', vorlage, ziel], { umgebung: umgebung(bestand) }),
    async sitzung(set, bestand) {
      const anfang = performance.now()
      const lauf = starte(join(verzeichnis, 'mse-mcp'), [], { umgebung: umgebung(bestand) })
      const s = new McpSitzung(lauf.kind)
      await s.beginne()
      const liste = await s.werkzeug('lesen.karten', { setID: Number(set.id) })
      const karten = Array.isArray(liste.wert) ? liste.wert : liste.wert.karten
      // **Bereit ist mse-mcp, wenn die Kartenliste da ist** — vorher kennt der
      // Aufrufer keine Kennung, die er zeichnen lassen könnte.
      const bereitMs = performance.now() - anfang
      const kennung = (stelle) => Number(karten[stelle].ID ?? karten[stelle].id)
      const scheitern = (fehler) => ({ ok: false, dauerMs: null, text: String(fehler.message ?? fehler) })
      return {
        lauf,
        bereitMs,
        karten,
        zeichne: (stelle, ziel) => s.werkzeug('schreiben.karte.zeichnen', { karteID: kennung(stelle), ziel })
          .then((a) => ({ ok: true, dauerMs: a.dauerMs, text: JSON.stringify(a.wert) }), scheitern),
        name: async (stelle) => ({ ok: true, dauerMs: 0, text: feldwert(karten[stelle], 'name') }),
        skript: (ausdruck, stelle = 0) => s.werkzeug('lesen.skript', { karteID: kennung(stelle), ausdruck })
          .then((a) => ({ ok: true, dauerMs: a.dauerMs, text: a.wert?.Quelltext ?? String(a.wert) }), scheitern),
      }
    },
  }
}
