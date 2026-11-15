// **Die Leihe wählt Ziel und Umschlag im Dateidialog, am nativen Fenster** —
// UXR-F15 (#2030):
//
// - *Spielen › Verleihen oder leihen …* hat kein Textfeld für einen Pfad mehr;
//   statt dessen steht an jeder der zwei Wahlen ein Knopf und der Satz „Noch
//   kein … gewählt",
// - der Knopf fürs Ziel öffnet den **Sichern-Dialog von macOS** — mit dem
//   Namen `leihe.msepackage` vorbelegt, Endung samt Filter — und ein Abbruch
//   lässt den vorigen Stand stehen,
// - der gewählte Pfad steht lesbar daneben, die Aufschrift wechselt zu
//   „Anderen … wählen", und eine zweite Wahl ändert ihn (Dialog abgebrochen:
//   der Pfad bleibt),
// - dasselbe für den Umschlag, der angenommen wird: der **Öffnen-Dialog**,
//   Abbruch, Wahl, Änderung.
//
// **Was die Probe nicht belegt**: das Bestätigen im echten Dialog (er läuft in
// einem Dienst von macOS; für die Wahl legt sie die Antwort vorher bereit,
// `internal/fenster/abnahme_blatt_darwin.go`, `vorbereitet`) und das
// **Annehmen**. Der Umschlag, den die Probe ablegt, ist an ein Gegenüber
// gerichtet, und derselbe Bestand ist nicht dieses Gegenüber; die Annahme
// prüfen `internal/app/deckleihe_intern_test.go` und der Flächenlauf. Hier
// zählt der Weg bis zum gewählten Pfad.
import { existsSync, mkdtempSync, statSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { zurSeite } from './uirevue.mjs'

const zielname = 'Wohin der Umschlag gelegt wird'
const umschlagname = 'Umschlag einer Leihe'

const zeilenwert = (ausgabe, name) => ausgabe.split('\n').find((z) => z.startsWith(`${name}\t`))?.split('\t')[1] ?? ''

// gruppe ist die Wahl mit diesem Namen — eine Gruppe, deren erstes Kind die
// Beschriftung trägt (`bausteine/Pfadwahl.tsx`).
const gruppe = (name) => `[...document.querySelectorAll('[role=group]')].find((g) => g.checkVisibility() && g.firstElementChild?.innerText.trim() === ${JSON.stringify(name)})`

// knopf ist der einzige Knopf der Gruppe.
const knopf = (name) => `(${gruppe(name)})?.querySelector('button')`

// stand liest, was die Gruppe zeigt: Aufschrift, Pfad (oder Satz), Meldung.
const stand = (p, name) => p.seite(`const g = ${gruppe(name)}; if (!g) return false; const k = g.querySelector('button'); return { knopf: k.innerText.trim(), text: document.getElementById(k.getAttribute('aria-describedby'))?.innerText ?? '', meldung: g.querySelector('[role=alert]')?.innerText ?? '', frei: !k.disabled }`)

export default async function uxrf15(p) {
  const anna = p.ort()
  p.mse(anna, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(anna, 'peer-anlegen')

  // Ein Deck zum Verleihen und ein bestätigtes Gegenüber, an der
  // Kommandozeile — die Probe prüft die Wahl, nicht das Anlegen.
  p.mse(anna, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(anna, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(anna, 'deck', 'legen', '1', `--karte=${karte}`, '8')

  const bert = p.ort()
  p.mse(bert, 'peer-anlegen')
  const ich = p.mse(bert, 'peer')
  const kennung = zeilenwert(ich, 'kennung')
  p.mse(anna, 'kontakt-aufnehmen', kennung, 'Bert')
  p.mse(anna, 'kontakt-bestaetigen', kennung, zeilenwert(ich, 'fingerabdruck'))

  const ordner = mkdtempSync(join(tmpdir(), 'uxrf15-'))
  const erstes = join(ordner, 'leihe-fuer-bert.msepackage')
  const zweites = join(ordner, 'zweite-leihe.msepackage')

  await p.starte({ ort: anna })
  await zurSeite(p, 'Spielen')
  await p.klickeText('Verleihen oder leihen …')
  await p.warte('die Schublade steht', () => p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())"))
  await p.warte('die Wahl des Ziels steht', () => p.seite(`return !!(${gruppe(zielname)})`))

  p.notiz('1 — kein Textfeld, sondern ein Knopf und ein Satz')
  const vorher = await stand(p, zielname)
  p.pruefe('der Knopf heißt „Ort wählen …"', vorher.knopf === 'Ort wählen …', vorher.knopf)
  p.pruefe('daneben steht „Noch kein Ort gewählt."', vorher.text === 'Noch kein Ort gewählt.', vorher.text)
  p.pruefe('es gibt kein Textfeld für einen Pfad',
    await p.seite(`return !(${gruppe(zielname)}).querySelector('input, textarea')`))
  await p.bild('ohne-wahl')

  p.notiz('2 — der Sichern-Dialog von macOS: Name und Endung, und ein Abbruch ändert nichts')
  await p.klickeAuf(knopf(zielname), 'Ort wählen …')
  const blatt = await p.dialogZeigtSich('sichern', '»Ort wählen …«')
  p.pruefe('der Dialog schlägt Namen und Endung vor', blatt.name === 'leihe.msepackage', String(blatt.name))
  p.pruefe('und sagt, was er fragt', blatt.nachricht === 'Umschlag der Leihe ablegen', String(blatt.nachricht))
  const abgebrochen = await stand(p, zielname)
  p.pruefe('nach dem Abbruch steht weiter „Noch kein Ort gewählt."', abgebrochen.text === 'Noch kein Ort gewählt.', abgebrochen.text)
  p.pruefe('und keine Meldung', abgebrochen.meldung === '', abgebrochen.meldung)
  p.pruefe('der Knopf ist wieder frei', abgebrochen.frei)

  p.notiz('3 — der gewählte Pfad steht lesbar daneben')
  await p.dialogantwort(erstes)
  await p.klickeAuf(knopf(zielname), 'Ort wählen …')
  const gewaehlt = await p.warte('der Pfad steht da', async () => {
    const jetzt = await stand(p, zielname)

    return jetzt.text === erstes ? jetzt : false
  })
  p.pruefe('der Pfad steht im Ganzen da', gewaehlt.text === erstes, gewaehlt.text)
  p.pruefe('der Knopf heißt jetzt „Anderen Ort wählen …"', gewaehlt.knopf === 'Anderen Ort wählen …', gewaehlt.knopf)
  await p.bild('ziel-gewaehlt')

  p.notiz('4 — ein Abbruch lässt den gewählten Pfad stehen')
  await p.klickeAuf(knopf(zielname), 'Anderen Ort wählen …')
  await p.dialogZeigtSich('sichern', '»Anderen Ort wählen …«')
  const geblieben = await stand(p, zielname)
  p.pruefe('der Pfad ist derselbe', geblieben.text === erstes, geblieben.text)
  p.pruefe('und der Knopf frei', geblieben.frei)

  p.notiz('5 — eine zweite Wahl ändert ihn')
  await p.dialogantwort(zweites)
  await p.klickeAuf(knopf(zielname), 'Anderen Ort wählen …')
  const geaendert = await p.warte('der neue Pfad steht da', async () => {
    const jetzt = await stand(p, zielname)

    return jetzt.text === zweites ? jetzt : false
  })
  p.pruefe('der neue Pfad ersetzt den alten', geaendert.text === zweites, geaendert.text)

  p.notiz('6 — der Umschlag geht dorthin, wo der Pfad steht')
  await p.klickeAuf("[...document.querySelectorAll('label')].find((l) => l.checkVisibility() && l.innerText.trim().startsWith('Bert'))", 'Bert')
  // **Das Deck über seine Beschriftung**, nicht als erstes Auswahlfeld: seit
  // UXR-F1 (#2007) kann auch die Wahl des Gegenübers ein Auswahlfeld sein.
  await p.seite(`const s = [...document.querySelectorAll('select')].find((e) => e.checkVisibility() && [...(e.labels ?? [])].some((l) => l.innerText.trim().startsWith('Welches Deck'))); if (!s) return false; const o = [...s.options].find((x) => x.text === 'Probe-Deck'); const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, o.value); s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  const anbieten = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Verleihen und ablegen')"
  p.zeige(`vor dem Verleihen: ${JSON.stringify(await p.seite("return { auswahl: [...document.querySelectorAll('select, input[type=radio]:checked, input[role=combobox]')].filter((e) => e.checkVisibility()).map((e) => ((e.labels?.[0]?.innerText ?? '').trim().slice(0, 20)) + '=' + e.value), knopf: [...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Verleihen und ablegen')?.disabled ?? null }"))}`)
  await p.warte('„Verleihen und ablegen" ist frei', () => p.seite(`return ${anbieten}?.disabled === false`))
  await p.klickeAuf(anbieten, 'Verleihen und ablegen')
  await p.warte('der Umschlag ist abgelegt', () => p.seite(`return [...document.querySelectorAll('[role=status]')].some((s) => s.innerText.includes(${JSON.stringify(zweites)}))`), 20000)
  p.pruefe('der Umschlag liegt am gewählten Ort', existsSync(zweites) && statSync(zweites).size > 0)
  p.pruefe('und nicht am ersten', !existsSync(erstes))
  await p.bild('abgelegt')

  p.notiz('7 — der Umschlag, der angenommen wird, kommt aus dem Öffnen-Dialog')
  await p.klickeText('Eine Leihe annehmen')
  await p.warte('die Wahl des Umschlags steht', () => p.seite(`return !!(${gruppe(umschlagname)})`))
  const leer = await stand(p, umschlagname)
  p.pruefe('der Knopf heißt „Umschlag wählen …"', leer.knopf === 'Umschlag wählen …', leer.knopf)
  p.pruefe('daneben steht „Noch kein Umschlag gewählt."', leer.text === 'Noch kein Umschlag gewählt.', leer.text)
  p.pruefe('„Leihe annehmen" ist angehalten', await p.seite("return [...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Leihe annehmen')?.disabled === true"))

  await p.klickeAuf(knopf(umschlagname), 'Umschlag wählen …')
  await p.dialogZeigtSich('oeffnen', '»Umschlag wählen …«')
  p.pruefe('nach dem Abbruch steht weiter „Noch kein Umschlag gewählt."', (await stand(p, umschlagname)).text === 'Noch kein Umschlag gewählt.')

  await p.dialogantwort(zweites)
  await p.klickeAuf(knopf(umschlagname), 'Umschlag wählen …')
  const umschlag = await p.warte('der Umschlag steht da', async () => {
    const jetzt = await stand(p, umschlagname)

    return jetzt.text === zweites ? jetzt : false
  }, 20000)
  p.pruefe('der Pfad steht im Ganzen da', umschlag.text === zweites, umschlag.text)
  p.pruefe('der Knopf heißt jetzt „Anderen Umschlag wählen …"', umschlag.knopf === 'Anderen Umschlag wählen …', umschlag.knopf)
  p.pruefe('„Leihe annehmen" ist frei', await p.seite("return [...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Leihe annehmen')?.disabled === false"))
  await p.bild('umschlag-gewaehlt')

  p.notiz('8 — ein Abbruch lässt auch den Umschlag stehen')
  await p.klickeAuf(knopf(umschlagname), 'Anderen Umschlag wählen …')
  await p.dialogZeigtSich('oeffnen', '»Anderen Umschlag wählen …«')
  const bleibt = await stand(p, umschlagname)
  p.pruefe('der Pfad ist derselbe', bleibt.text === zweites, bleibt.text)
  p.pruefe('und keine Meldung', bleibt.meldung === '', bleibt.meldung)

  p.pruefe('⌘Q beendet', await p.beende())
}
