// Die Texte der **Werkzeugseiten** — Konsole und Schlüsselwörter.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — an der Zahl der Kataloge hängt der Vertragstest gegen
// die Go-Seite, nicht an der Zahl der Dateien.

export const werkzeugseiten = {
  pack: {
    titel: 'Случайный бустер',
    erklaerung: 'Бустер из карт этого набора по правилам игрового пакета.',
  },
  symbol: {
    titel: 'Символ',
    erklaerung: 'Символ этого набора — нарисованный, а не загруженный.',
  },
  konsole: {
    titel: 'Консоль',
    erklaerung: 'Выражение скрипта MSE в контексте открытой карты.',
  },
  schluesselwoerter: {
    titel: 'Ключевые слова',
    erklaerung:
      'Ключевые слова этого набора — и те, которые приносит с собой его игра.',
  },
}
