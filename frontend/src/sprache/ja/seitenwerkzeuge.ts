// Die Texte der **Werkzeugseiten** — Konsole und Schlüsselwörter.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — an der Zahl der Kataloge hängt der Vertragstest gegen
// die Go-Seite, nicht an der Zahl der Dateien.

export const werkzeugseiten = {
  pack: {
    titel: 'ランダムパック',
    erklaerung:
      'このセットのカードから作るブースターです。ゲームパッケージのルールに従います。',
  },
  symbol: {
    titel: 'シンボル',
    erklaerung: 'このセットのシンボルです。読み込むのではなく、描画されます。',
  },
  konsole: {
    titel: 'コンソール',
    erklaerung: '開いているカードの文脈で評価される MSE スクリプトの式です。',
  },
  schluesselwoerter: {
    titel: 'キーワード',
    erklaerung: 'このセットのキーワードと、そのゲームが持つキーワードです。',
  },
}
