// **Bereiche und Funktionen** — OPS-10
// (#1908).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge; die Namen und Werte werden dort hineingestreut.
// Ein Bereich heißt hier wie sein Eintrag in der Leiste.
export const seitenschalter = {
  schalter: {
    erklaerung:
      '关闭后，该区域会从侧栏和命令面板中消失，并且它的路径在每个界面上都会被拒绝。数据保持不变——重新开启即可恢复一切。',
    hinweis: '开关只负责隐藏，不负责保护：如需保护，请使用角色和共享设置。',
    reife: {
      stabil: '稳定',
      vorschau: '预览',
      auslaufend: '即将停用',
    },
    saetze: {
      sammeln: '收藏、持有物和愿望清单',
      spielen: '套牌、锦标赛、对局和牌桌',
      handeln: '交易、共享和信封',
      ressourcen: '模板包和模板',
      ki: '建议、图片、搜索索引、账户和费用',
      netz: '身份、联系人、邀请和可达性',
      markt: '交易市场的报价、价格和账户',
      turniere: '带轮次、配对和结果的锦标赛',
      dienst: '供脚本使用的 REST——关闭时正在运行的服务会停止',
      werkzeugserver: 'HTTP 上的 MCP——关闭时正在运行的服务器会停止',
    },
    abgeschaltet: {
      titel: '{seite} 已关闭',
      erklaerung:
        '可在 设置 › 区域与功能 中重新开启该区域。它的数据仍在，未作任何更改。',
      oeffnen: '打开区域与功能',
    },
  },
} as const

// schalternamen sind die Beschriftungen der zehn Schalter.
export const schalternamen = {
  'schalter.sammeln': '收藏',
  'schalter.spielen': '对战',
  'schalter.handeln': '交易',
  'schalter.ressourcen': '资源',
  'schalter.ki': 'AI',
  'schalter.netz': '网络与对方',
  'schalter.markt': '交易市场与价格',
  'schalter.turniere': '锦标赛',
  'schalter.dienst': '脚本服务',
  'schalter.werkzeugserver': 'HTTP 工具服务器',
}

// schalterwerte sind an und aus, je Schalter.
export const schalterwerte = {
  'schalter.sammeln.an': '开',
  'schalter.sammeln.aus': '关',
  'schalter.spielen.an': '开',
  'schalter.spielen.aus': '关',
  'schalter.handeln.an': '开',
  'schalter.handeln.aus': '关',
  'schalter.ressourcen.an': '开',
  'schalter.ressourcen.aus': '关',
  'schalter.ki.an': '开',
  'schalter.ki.aus': '关',
  'schalter.netz.an': '开',
  'schalter.netz.aus': '关',
  'schalter.markt.an': '开',
  'schalter.markt.aus': '关',
  'schalter.turniere.an': '开',
  'schalter.turniere.aus': '关',
  'schalter.dienst.an': '开',
  'schalter.dienst.aus': '关',
  'schalter.werkzeugserver.an': '开',
  'schalter.werkzeugserver.aus': '关',
}
