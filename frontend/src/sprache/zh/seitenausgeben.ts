// The texts of the **export assistant** — where to, and what each
// destination needs.
//
// Split out of `seiten.ts` when the assistant brought the file close to the
// 400-line limit (§2, UXR-09, #1917).

export const ausgeben = {
  titel: '导出',
  erklaerung: '将系列从此应用中取出的五种方式——以及各自的用途。',
  wohin: '导出到哪里？',
  keinZiel: '尚未选择导出方式',
  nichtGemerkt: '目标有效，但未被记住——下次不会预先选中：',
  keinZielErklaerung:
    '在“导出到哪里？”中选择系列以何种形式导出——这里就会显示需要设置的内容。',
  ohneKarten:
    '图片和印刷版面没有可显示的内容。卡牌在“卡牌”中创建、粘贴或导入。',
  ziele: {
    setdatei: {
      name: '系列文件',
      kurz: '系列本身，可以再次打开——一个 Magic Set Editor 2 也能读取的 `.mse-set` 文件。',
      lang: '快捷键是 ⌘S。**不包含**套牌、藏品、价格、标签以及各个值的来源——它们留在资料库中，因为 MSE2 在保存时会丢弃它不认识的内容。',
    },
    bilder: {
      kurz: '每张卡牌一个图片文件，可用于别处——论坛、电子表格、游戏程序。',
    },
    druckbogen: {
      kurz: '卡牌并排排在纸上，带有裁切标记。',
    },
    vorlage: {
      kurz: '输出文本而非图像：列表、表格、供其他程序使用的文件。',
      lang: '有哪些模板取决于已安装的资源包。',
    },
    mtgjson: {
      kurz: 'mtgjson.com 的卡牌格式——适用于能读取它的工具。',
      lang: '不适合在此应用中继续编辑：MTG JSON 无法容纳的内容会在导出后列出。',
    },
  },
  vorlagenEingrenzen: '筛选模板',
  weitereOptionen: '更多选项（{zahl}）',
  // **„结果预览"** — der Schritt vor dem Schreiben (UXR-F7, #2015).
  probelauf: {
    ueberschrift: '结果预览',
    rechnet: '正在计算…',
    hinweis: '点击下方按钮之前，不会写入任何内容。',
    bogen: {
      jeSeite: '每页 {zahl} 张 · {seiten}',
      masse: '{karten} · 纸张 {breite} × {hoehe} 毫米 · {dpi} dpi',
      letzte: '最后一页：{karten}。',
    },
    vorlage: {
      datei: '文件：{name}（{groesse}）',
      karten: '已对 {karten} 求值。',
      beigaben: '旁边的文件：{zahl}（文件夹“{ordner}”，{groesse}）',
      mindestens: '至少 {groesse}',
      befunde: '求值时发现：',
    },
  },
}
