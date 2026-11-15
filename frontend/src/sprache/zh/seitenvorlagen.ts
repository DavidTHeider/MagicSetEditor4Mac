// Die Texte der **Paketwerkstatt** (#1164) — eine Vorlage öffnen, ändern,
// speichern.
//
// **Eine eigene Datei**, weil `seiten.ts` an der Zeilengrenze aus §2 steht.
// Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als **einer**
// ankommt (§9).
export const vorlagen = {
  titel: '模板',
  erklaerung:
    '卡牌能做什么、外观如何，写在模板包内的文本文件里——字段、脚本、尺寸。可以在这里更改。先前的状态会作为备份保留在软件包旁边。',
  paket: '软件包',
  keinPaket: '尚未选择软件包。',
  waehleEines:
    '在上方选择一个软件包，即可查看并更改它的文本文件——字段、脚本、尺寸。',
  keineDateien: '此软件包没有可编辑的文件。',
  dateien: '文件',
  datei: '文件',
  inhalt: '内容',
  speichern: '存储',
  gespeichert: '已存储。',
  ungespeichert: '未存储。',
  warnung: '此文件适用于所有使用该软件包的系列——不仅是已打开的那个。',
  konflikt:
    '文件在打开之后已被更改。请载入其现有版本——你的文本将会丢失——或保留你自己的版本：这样，“更改”会显示你覆盖了其中的哪些内容。',
  neuLaden: '载入现有版本',
  behalten: '保留我的',
  vorschau: '预览',
  vorschauErklaerung:
    '使用此模板的第一个系列的第一张卡牌——按此处显示的文本绘制。此过程不会存储任何内容。',
  vorschauLaeuft: '正在绘制…',
  vorschauKeineKarte:
    '没有已打开的系列使用此模板。请在“系列”下打开一个指定它的系列——这样就能在这里看到更改的效果。',

  arbeitsplatz: '模板工坊',
  baumBezeichnung: '软件包与文件',
  inspectorBezeichnung: '预览与诊断',
  baumSuche: '搜索软件包与文件',
  baumSuchePlatzhalter: '软件包或文件…',
  baumLeer: '没有与搜索匹配的软件包或文件。',
  baumOhnePakete:
    '还没有可编辑的软件包。请在“软件包”下创建一个或获取一个——之后它会出现在这里。',
  nurLesbar: '随附——只读。可以编辑你自己的副本。',
  baumNurMitgeliefert:
    '这里所有的软件包都是随附的。请为其中之一创建自己的副本，或在“软件包”下新建一个——两者都可以编辑，并会显示在顶部。',
  dateienLaden: '正在读取文件…',
  inspectorLeer: '已打开文件的预览、诊断和来源显示在这里。',
  waehleImBaum:
    '在左侧的树中选择一个文件——字段、脚本、尺寸以文本形式存放在那里，可以在这里更改。',
  verwerfenFrage: '已打开的文件有尚未存储的更改。要放弃它们并打开 {datei} 吗？',
  verwerfen: '放弃并打开',
  hierBleiben: '留在此处',
  testkarte: '测试卡牌',
  testkarteVorgabe: '第一个系列的第一张卡牌',
  testkarteOhneNamen: '卡牌 {nummer}',
  testkarteErklaerung:
    '模板很少在第一张卡牌上出问题。示例系列“Werkstatt-Testkarten”带有长文本、特殊版式、第二面、缺失的图像和特殊字符——请在“系列”下打开它，并将它的样式切换到这个软件包。',
  veraltet:
    '已过期——文件在第 {zeile} 行有一个发现项；当前显示的是最近一次有效的状态。',
  wirdNeuGezeichnet: '正在重新绘制…',
  diagnose: '诊断',
  diagnoseFrei: '此文件中没有发现项。',
  feldartFehlt:
    '此字段缺少“type:”。MSE2 在读取软件包时会在此处中止；本应用程序将其视为文本字段。',
  feldartUnbekannt:
    'MSE2 不认识字段类型“{0}”，并会在此处中止读取软件包。本应用程序将该字段视为文本。已知的类型有：{1}。',
  diagnoseStelle: '第 {zeile} 行，第 {spalte} 列',
  zurStelle: '转到该位置',
  herkunft: '名称的来源',
  herkunftKeine: '将光标放在某个名称上——这里就会显示它来自哪里、谁在使用它。',
  herkunftUnbekannt:
    '此软件包没有定义“{name}”——该名称来自函数库或某个引入的软件包。',
  herkunftStelle: '{art}，位于 {datei} 第 {zeile} 行',
  zurDefinition: '转到定义',
  verwendungen: '引用',
  verwendungenKeine: '没有其他位置提到这个名称。',
  verwendungLaden: '正在查找引用…',
  symbolarten: {
    'card field': '卡牌字段',
    'set field': '系列字段',
    'styling field': '样式选项',
    'extra card field': '额外字段',
    script: '脚本名称',
  } as Record<string, string>,
  symbolartSonst: '在此定义',
  aenderungen: '存储前的更改',
  aenderungenZeigen: '显示更改',
  aenderungenVerbergen: '隐藏更改',
  aenderungenKeine: '与读取时的状态相比没有更改。',
  aenderungenZusammenfassung: '新增 {neu} 行，移除 {weg} 行。',
  aenderungenAbschnitt: '更改 {nummer}',
  zeileWeg: '已移除',
  zeileNeu: '新增',
  kopieHinweis: '软件包的来源',
  kopieren: '创建自己的副本',
  kopieMitgeliefert:
    '此软件包随应用程序一起提供，无法更改。请创建自己的副本——它之后会出现在你的软件包中，原件保持不变。',
  kopieFremd:
    '此软件包来自外部。在这里所做的更改算作本地更改，下次获取时会与新版本冲突。如果希望更改保留，请创建自己的副本。',
  kopieOriginal: '原件保持不变。你正在处理的是副本——{original} 本身没有改变。',
  dateiaktionen: '软件包文件',
  dateiNeu: '新建文件',
  dateiName: '新文件的名称',
  dateiAnlegen: '创建',
  dateiAbbrechen: '取消',
  dateiHinzufuegen: '添加图像或字体…',
  dateiUmbenennen: '重命名',
  dateiNeuerName: '新名称',
  dateiUmbenennenJa: '重命名',
  dateiLoeschen: '删除…',
  dateiLoeschenFrage: '要删除吗？该文件之后会保存在软件包旁的备份中。',
  dateiLoeschenJa: '是，删除',
  dateiErstSichern: '请先存储或放弃更改。',
  beiwerk: '附属文件',
  beiwerkHinweis: '不是文本——软件包按原样使用此文件。可以在这里重命名和删除。',
  beiwerkZeile: '附属文件（{anzahl}）',
  weitereZeigen: '显示更多（{gesamt} 项中的 {gezeigt} 项）',
} as const
