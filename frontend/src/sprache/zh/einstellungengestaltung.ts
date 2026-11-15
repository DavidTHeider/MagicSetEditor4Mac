// **Appearance on the settings page** — UX-09 (#1681). Mirrors the German
// catalogue part of the same name.

export const gestaltungsnamen = {
  farbwelt: '配色',
  schriftbereich: '展示字体适用于',
  hintergrund: '窗口背景图',
  gestaltung: '自定颜色',
}

export const gestaltungswerte = {
  'farbwelt.pergament': '羊皮纸 — 温暖，带纤维纹理',
  'farbwelt.klar': '清爽 — 中性，无纹理',
  'farbwelt.schiefer': '石板 — 冷调蓝灰',
  'farbwelt.moos': '苔藓 — 柔和的绿色',
  'schriftbereich.ueberschriften': '仅标题',
  'schriftbereich.ganz': '整个界面',
  'hintergrund.aus': '关闭',
  'hintergrund.fuellen': '填充',
  'hintergrund.einpassen': '适合',
  'hintergrund.kacheln': '平铺',
  'schriftart.buch': '书籍体 — New York',
  'schriftart.rund': '圆体 — SF Pro Rounded',
  'schriftart.serifenlos': '无衬线 — 系统字体',
}

export const gestaltungshinweise = {
  farbwelt:
    '界面的颜色，浅色与深色各自经过核算。开启高对比度时，无论此处如何选择，均采用墨色。',
  schriftbereich:
    '整个界面只使用文本字体；装饰字体仅用于标题。卡牌保留其模板的字体。',
  hintergrund:
    '把你自己的图片铺在背景之下，并覆上一层薄纱，让文字保持清晰可读。图片在下方选择；选择**关闭**会立即将其移除。',
  gestaltung:
    '为界面的十项属性自定颜色，每行一项。只接受十六进制颜色——不接受地址，也不接受字体。会导致难以阅读的设置不会生效。',
}

export const gestaltungsbeiwerk = {
  fensterbild: '你的窗口背景图',
  vorschau: '窗口背景图预览',
  keinBild:
    '还没有自己的图片。请选择一张——它会显示在这里，一旦窗口背景图未关闭，也会显示在界面后面。',
  aus: '图片已就绪，但窗口背景图设为**关闭**。',
  waehlen: '选择图片……',
  entfernen: '移除图片',
  merkmale: '自定颜色',
  merkmaleHilfe:
    '每行一项，例如“--mse-marke: #4e2f16;”。允许的属性有 {merkmale}。',
  uebernehmen: '应用',
  zeilenfehler:
    '第 {zeile} 行：应为类似“--mse-grund: #1e1b16;”的写法——即允许的属性之一加一个十六进制颜色。',
  nichtLesbar: '未应用——这样会难以阅读：',
  paar: '{vorn} 在 {hinten} 上：{verhaeltnis} : 1，要求 {schwelle} : 1',
  paarnamen: {
    'status:warnung': '警告',
    'status:erfolg': '成功信息',
    'fensterbild:dunkel': '深色窗口背景图上的 --mse-grund',
    'fensterbild:hell': '浅色窗口背景图上的 --mse-grund',
  } as Record<string, string>,
  angewandt: '已应用。',
  ruhtBeiKontrast: '已保存——开启高对比度时，自定颜色不生效；关闭后会恢复。',
  schmuckNurOben: '所选展示字体是装饰字体——它仅用于标题。',
  zuruecksetzenGescheitert: '无法重置外观：',
  zurVorgabe: '将自定颜色重置为默认值',
}
