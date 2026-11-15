// Die Texte der **Werkzeugseiten** — Konsole und Schlüsselwörter.
//
// Aus `seiten.ts` herausgelöst, als die Datei über 400 Zeilen wuchs (§2).
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — an der Zahl der Kataloge hängt der Vertragstest gegen
// die Go-Seite, nicht an der Zahl der Dateien.

export const werkzeugseiten = {
  pack: {
    titel: '随机补充包',
    erklaerung: '按游戏包的规则，用此系列的卡牌组成一包补充包。',
  },
  symbol: {
    titel: '符号',
    erklaerung: '此系列的符号——绘制而成，而非加载。',
  },
  konsole: {
    titel: '控制台',
    erklaerung: '在当前打开的卡牌上下文中运行的 MSE 脚本表达式。',
  },
  schluesselwoerter: {
    titel: '关键字',
    erklaerung: '此系列的关键字——以及其游戏自带的关键字。',
  },
}
