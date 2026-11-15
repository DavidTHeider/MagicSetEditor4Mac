// The **data hygiene** texts — work item O8 of phase 5.2
// (#1432).

export const datenpflege = {
  titel: '整理',
  hinweis:
    '资料库用得久了，会积累一些不再指向任何内容的行。这里列出它们是什么，以及其中哪些可以放心清除。',

  nachsehen: '检查',
  nachsehenLaeuft: '正在检查……',
  sauber: '未发现任何问题。资料库状态良好。',
  gueltigbis: '此概览有效至 {zeit}。',

  arten: {
    kartenwerteOhneKarte: '没有对应卡牌的字段值',
    setangabenOhneSet: '没有对应系列的系列信息',
    suchzeilenOhneKarte: '没有对应卡牌的搜索行',
    deckplaetzeOhneDeck: '没有对应套牌的套牌位置',
    zuordnungenOhneExemplar: '没有对应副本的套牌分配',
    verlaufOhneSet: '指向已不存在的系列的历史记录条目',
    abgelaufeneNachweise: '已过期的批准记录',
    abgelaufeneToken: '已过期的访问令牌',
    moeglicheDubletten: '看起来是重复项',
    offeneRueckkehren: '未完成的恢复',
  } as Record<string, string>,

  klassen: {
    sicher: '可以放心清除',
    entscheidung: '由你决定',
    nurInformieren: '不会被触及',
  } as Record<string, string>,

  klassenhinweise: {
    sicher: '这些行不指向任何内容，也不保存任何内容。它们已被预先选中。',
    entscheidung:
      '这看起来像遗留物——但也可能是一条记录。选择之前，请先看看示例。',
    nurInformieren:
      '这些不会被清除。卡牌编号相同并不等于重复——品相、语言、表面处理、获得方式和存放位置会把它们区分开——而未完成的恢复是一条记录。',
  } as Record<string, string>,

  anzahl: '资料库中的 {anzahl} 处',
  beispieleTitel: '例如',

  aufraeumen: '清除所选项',
  aufraeumenLaeuft: '正在清除……',
  aufraeumenHinweis: '会先备份资料库。之后会再检查一次。',

  bilanzTitel: '发生了什么',
  gesichert: '已备份到 {ort}',

  ausgaenge: {
    erledigt: '已清除',
    konflikt: '自概览以来已发生变化——未触及',
    schonWeg: '早已不存在',
    uebersprungen: '已跳过',
    gescheitert: '未能完成',
    bereit: '就绪',
  } as Record<string, string>,

  bilanzZeile: '{art}：{ausgang}（{anzahl}）',
  bilanzStaende: '之前 {vorher} 项发现，之后 {nachher} 项。',
}
