// The **staged reset** — work item O4 of phase 5.2
// (#1428).

export const ruecksetzung = {
  titel: '重置',
  hinweis:
    '共六种，从最轻到最重。每一种会影响什么，事先都会显示——' +
    '包括数量和名称。',

  artWaehlen: '要重置什么？',

  arten: {
    layout: '仅窗口的外观',
    profileinstellungen: '此个人资料的设置',
    nutzereinstellungen: '我自己的设置',
    profil: '此个人资料及其拥有的一切',
    nutzer: '我和我的所有个人资料',
    anwendung: '整个应用',
  } as Record<string, string>,

  nachsehen: '预览会影响哪些内容',
  nachsehenLaeuft: '正在查看…',

  nichtsBetroffen: '这里没有可以重置的内容。',
  gruppe: '{was}：{anzahl}',

  gruppen: {
    einstellungen: '设置',
    profile: '个人资料',
    sets: '系列',
    karten: '卡牌',
    decks: '套牌',
    exemplare: '收藏中的藏品',
    token: '访问令牌',
    freigaben: '授权记录',
    zugaenge: '配置文件访问权限',
    bilder: '图片',
  } as Record<string, string>,

  sicherung: '这里的数据将会丢失。请先备份数据库——在上方的区域中。',

  hindernisse: {
    offeneRueckkehr:
      '此个人资料中的数量在回滚后尚未结清。请先检查这些数量，' +
      '并确认回滚已处理，然后再删除。',
    fremdesZiel: '这不是你的。',
    letzterNutzer: '这是此数据库中的最后一位用户——删除后将无人留下。',
    laufenderAuftrag: '仍有任务在运行。请等它完成。',
    nachweisFehlt: '重置整个应用需要在这台电脑上给出明确的授权。',
  } as Record<string, string>,

  nameTippen: '输入名称以确认',
  freigabeSatz:
    '重置整个应用需要你在这台电脑上的授权。授权有效五分钟，仅限一次重新开始。',
  freigeben: '授予授权',
  freigebenLaeuft: '正在授权…',
  freigegeben: '授权有效至 {zeit}。',
  kennungFuerKommandozeile:
    '同一授权也可用于命令行（mse zuruecksetzen anwendung）：',

  tun: '重置',
  tunLaeuft: '正在重置…',

  bilanz: '已还原 {anzahl} 项设置。',
  beiseite: '之前的数据库位于 {ort}。',
}
