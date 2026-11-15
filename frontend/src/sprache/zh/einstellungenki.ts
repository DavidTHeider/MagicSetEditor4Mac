// Die Einstellungen der KI — die lokale Laufzeit, Apples Gerätemodell und
// die Adresse (A7, #1446; AI-05, #2042). **Aus `seiteneinstellungen.ts`
// herausgelöst**, als der Katalog über die Zeilengrenze aus §2 ging.

export const kinamen = {
  'ki.laufzeit': '本地 AI 运行环境',
  'ki.geraetemodell': 'Apple 设备端模型',
  'ki.vorrangmodell': '优先使用的 Ollama 模型',
  'ki.adresse': 'AI 运行环境的地址',
} as const

export const kihinweise = {
  'ki.laufzeit':
    '是否会询问某个 AI 运行环境。**关闭**表示：不发出任何调用——连对这台电脑也不发。这里不会安装或启动任何内容。',
  'ki.geraetemodell':
    '是否使用 macOS 设备端模型处理小型文本任务——卡牌名称、风味文字。**自动**：若在此 Mac 上已启用（Apple 智能已开启、已接受 `fm` 许可），即会接入并优先使用；数据不会离开本机。**关闭**在你改回之前一直有效；仅对本次会话关闭可在 **AI** 中设置。',
  'ki.vorrangmodell':
    '带**优先**的建议在设备端模型之后询问哪个 Ollama 模型。**留空**表示第一个能以固定格式回答的模型，而 Ollama 把最近拉取的排在最前。若所选模型已不在 Ollama 的列表中，会给出提示，而不会悄悄换成另一个。',
  'ki.adresse':
    '运行环境监听的位置。留空表示这台电脑上的默认地址。**不在**这台电脑上的地址，会把其后的每个模型归为非本地——在“仅本地”下，则根本不会询问。',
} as const

export const kiwerte = {
  'ki.laufzeit.aus': '关闭',
  'ki.laufzeit.an': '开启',
  'ki.geraetemodell.automatisch': '自动',
  'ki.geraetemodell.aus': '关闭',
} as const
