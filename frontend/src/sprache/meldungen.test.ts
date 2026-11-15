// **Jede Meldung sagt, was zu tun ist** — Werkstück 2 von Phase 4.13
// (#1352).
//
// # Die Zusage
//
// `AGENTS.md` §6: *„Leere Zustände erklären den nächsten Schritt, statt »Keine
// Daten« zu zeigen"*, und die Abnahme dieses Issues sagt es für Meldungen:
// **jede sagt, was zu tun ist, und nicht nur, was schiefging.**
//
// # Wie eine Maschine das prüfen kann — und wie nicht
//
// **Sie kann es nicht beurteilen.** „Sagt, was zu tun ist" ist eine Frage an
// den Leser, nicht an einen Regelausdruck. Was sie kann, ist **eine
// notwendige Bedingung** halten: ein Satz, der einen nächsten Schritt nennt,
// enthält ein **Handlungswort** — einen Imperativ oder ein „du kannst".
//
// **Am 2026-09-06 gezählt**: von 26 Meldungen im deutschen Katalog trugen
// **16 keines**. „Die Eingabe ist nicht gültig." sagt, was schiefging, und
// lässt jemanden stehen.
//
// # Was er nicht prüft
//
// **Ob der genannte Schritt der richtige ist.** „Lade die Liste neu" ist ein
// Handlungswort und wäre bei einem Schreibfehler der falsche Rat. Das ist eine
// Frage ans Review (§10).
//
// **Und ob der Satz gut ist.** Er sieht ein Wort, nicht einen Sinn.
import { describe, expect, it } from 'vitest'

import { kataloge, sprachen } from './kataloge'

// handlungswoerter sind die Wörter, an denen ein nächster Schritt zu erkennen
// ist — je Sprache.
//
// **Eine Liste und kein Wörterbuch**: sie führt, was in diesem Katalog
// vorkommt, und wächst mit ihm. Ein Wort, das fehlt, macht den Wächter rot und
// nicht blind — das ist die richtige Richtung.
const handlungswoerter: Record<string, RegExp> = {
  de: /\b(wechsle|wähle|bitte|versuche|ersetze|trage|lege|nimm|gib|prüfe|schließe|öffne|lösche|lade|melde|starte|warte|benenne|entferne|setze|speichere|ergänze|berichtige|bestätige|hole|spiele|beende|bearbeite|erhöhe|sichere|nachsehen|kannst|kann sie dir geben|lässt sich)\b/i,
  en: /\b(switch|pick|choose|please|try|replace|enter|put|take|give|check|close|open|delete|reload|report|start|wait|rename|remove|set|save|add|correct|confirm|get|restore|change|create|stop|back it up|you can)\b/i,
  // **Die sechs Übersetzungen aus LOC-01** (#1687) — Stämme statt Wörter,
  // denn Italienisch und Spanisch hängen das Fürwort an (`scegline`,
  // `corrígela`), und `\b` kennt kein Kyrillisch und keine CJK-Grenze. Sie
  // sind maschinell übersetzt; die Liste hält, dass der Schritt dabei nicht
  // verloren ging, nicht, dass er gut übersetzt ist.
  es: /\b(elig|selecci|intent|prueb|vuelv|reemplaz|introduc|escrib|comprueb|revis|cierr|abr|borr|elimin|recarg|inform|inici|esper|renombr|quit|establec|guard|añad|agreg|corrig|confirm|obt|restaur|cambi|cre|det|pued|por favor|pide|pregunt|busc|usa|utiliz|activ|desactiv|instal|actualiz|conect|vincul|descarg|import|export|sal|dibuj|dale|haz|ve a|consult|marc|asign|deja|aplic|consúlt|corríg|introdúc|defin)/i,
  fr: /\b(choisi|sélectionn|essay|réessay|essai|réessai|remplac|saisi|entr|vérifi|ferm|ouvr|supprim|recharg|signal|démarr|lanc|attend|renomm|retir|défin|enregistr|ajout|corrig|confirm|obten|restaur|chang|modifi|cré|arrêt|pouv|peux|veuill|demand|cherch|utilis|activ|désactiv|install|mett|connect|associ|télécharg|import|export|dessin|donn|consult|coch|attribu|laiss|appliqu|reprend|relanc|recommenc|sauvegard|indiqu|rends|va |passe|fais|peut te|procur)/i,
  it: /\b(scegli|selezion|prova|riprov|sostitu|inserisc|controll|verific|chiud|apri|elimin|ricaric|segnal|avvi|attend|aspett|rinomin|rimuov|impost|salv|aggiung|corregg|conferm|ottien|ripristin|cambi|modific|cre|ferm|puoi|può|per favore|chied|cerc|usa|utilizz|attiv|disattiv|install|mett|collega|associ|scaric|import|esport|disegn|dai|consult|segna|assegn|lascia|applic|riprend|indica|vai|fai|esegu|interromp|procur)/i,
  ja: /(ください|できます|してから|選んで|選び|試して|確認|再読み込み|入力|閉じ|開い|削除|保存|追加|修正|変更|作成|待って|しましょう|して下さい|お試し|戻して|直して|使って|設定して)/,
  ru: /(попробуй|выбери|проверь|введи|укажи|открой|закрой|удали|перезагрузи|сохрани|добавь|исправь|подтверди|можешь|можно|замени|запусти|подожди|переименуй|убери|установи|измени|создай|останови|пожалуйста|дай|возьми|сообщи|верни|заверши|загрузи|переключи|повтори|обнови|попроси|восстанови|ищи|уточни|не принимай|не заменяй|нарисуй|используй|включи|отключи|задай|назначь|оставь|примени|перейди|сделай|нажми|сними|отметь|может выдать|получи)/i,
  zh: /(请|可以|可在|重试|检查|选择|输入|关闭|打开|删除|重新|保存|添加|更正|确认|更改|创建|等待|替换|改正|使用|设置|先)/,
}

// ohneHandlung sind die Meldungen, die keinen nächsten Schritt nennen —
// **jede mit ihrem Grund**.
const ohneHandlung: Record<string, string> = {
  // **Der Schritt steht in der Meldung selbst**, die eingesetzt wird: die
  // Skriptfehler der Standardbibliothek nennen Zeile und Spalte, und der
  // Rahmen darum sagt „berichtige ihn an der genannten Stelle".
  skript: 'die Sätze der Standardbibliothek — der Rahmen nennt den Schritt',
}

describe('Die Meldungen', () => {
  it('nennen einen nächsten Schritt', () => {
    for (const sprache of sprachen) {
      const muster = handlungswoerter[sprache]
      expect(muster, `${sprache} führt keine Handlungswörter`).toBeDefined()

      for (const [schluessel, wert] of Object.entries(
        kataloge[sprache].fehler,
      )) {
        // **Die Ausnahme kommt aus dem Register, nicht aus der Datenform** —
        // Befund 8 des Zweitblicks über Phase 4.13 (2026-09-07).
        //
        // Hier stand `typeof wert !== 'string'` als **eigener** Grund zum
        // Überspringen. Damit wäre jeder neue verschachtelte Verbund still
        // durchgegangen — ohne Handlungsprüfung **und** ohne Begründung. Der
        // Wächter hätte die Klasse behauptet, die er nicht hält (§10).
        //
        // Jetzt gilt: übersprungen wird nur, was im Register steht. Ein
        // Verbund, der dort fehlt, macht den Lauf rot und verlangt seinen
        // Grund.
        if (schluessel in ohneHandlung) {
          continue
        }

        // **Ein `throw` und kein `expect`**: der Typprüfer engt `wert` damit
        // auf `string` ein, und die Zusicherung darunter darf ihn in ihre
        // Meldung setzen. Ein `expect` ließe ihn breit, und `${wert}` gäbe
        // dann `[object Object]`.
        if (typeof wert !== 'string') {
          throw new Error(
            `${sprache}.${schluessel} ist kein Satz und steht nicht in ` +
              `ohneHandlung — trage ihn mit seinem Grund ein`,
          )
        }

        expect(
          muster?.test(wert) ?? false,
          `${sprache}.${schluessel} nennt keinen nächsten Schritt: ${wert}`,
        ).toBe(true)
      }
    }
  })

  // **Und jede Ausnahme trägt ihren Grund** — sonst wäre die Liste die
  // Stelle, an der eine vergessene Meldung sich versteckt.
  it('begründen jede Ausnahme', () => {
    for (const [schluessel, grund] of Object.entries(ohneHandlung)) {
      expect(grund.trim(), `${schluessel} ohne Grund`).not.toBe('')
    }
  })
})
