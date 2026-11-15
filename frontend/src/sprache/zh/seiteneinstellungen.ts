import { dienstkatalog } from './einstellungenkatalog'
import { einstellungsbelegung } from './einstellungsbelegung'
import { einstellungsgruppennamen } from './einstellungsgruppennamen'
import {
  schalternamen,
  schalterwerte,
  seitenschalter,
} from './einstellungenschalter'

// **Die Texte der Einstellungsseite** — ausgelagert für Phase 4.8
// (#1321).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als
// einer ankommt** (§9) — dieselbe Gestalt wie `seitenkarten.ts`. Mit den
// Dienstschaltern ging `seiten.ts` über die Zeilengrenze aus §2.
import { ki } from './seitenki'
import { netzbereich, netzhilfe, netznamen } from './einstellungennetz'
import { handelshilfe, handelsnamen, handelswerte } from './einstellungenhandel'
import {
  gestaltungsbeiwerk,
  gestaltungshinweise,
  gestaltungsnamen,
  gestaltungswerte,
} from './einstellungengestaltung'
import { zugaenge } from './einstellungenzugaenge'
import { ansichtsnamen, ansichtswerte } from './einstellungenansicht'
import { vorgabennamen } from './einstellungenvorgaben'
import {
  protokollhinweise,
  protokollnamen,
  protokollwerte,
} from './einstellungenprotokoll'
import { austausch } from './einstellungenaustausch'
import { eigenuebersetzung } from './eigenuebersetzung'
import { kihinweise, kinamen, kiwerte } from './einstellungenki'

export const einstellungen = {
  ...einstellungsgruppennamen,
  ...seitenschalter,

  titel: '设置',

  namen: {
    ...schalternamen,
    sprache: '语言',
    'sprache.eigene': '自定义翻译',
    darstellung: '地区格式',
    zeitzone: '时区',
    'filter.gespeichert': '已存储的筛选器',
    'kuerzel.ziffern': '⌘0 至 ⌘9 的对应关系',
    erscheinungsbild: '主题',
    schriftgroesse: '文字大小',
    aktualisierung: '检查新版本',
    ersteschritte: '启动时显示入门介绍',
    seitenleiste: '侧边栏',
    leistenlage: '侧边栏位置',
    fensterstart: '启动时的窗口',
    fensterbreite: '上次记住的宽度',
    fensterhoehe: '上次记住的高度',
    schriftart: '展示字体',
    feldtiefe: '卡牌表单',
    'tisch.autoanordnung': '自动整理牌桌上的卡牌',

    'barrierefreiheit.bewegung': '动态效果',
    'barrierefreiheit.kontrast': '对比度',
    'sicherheit.faktorbeimwechsel': '切换配置文件时的第二重验证',

    ...vorgabennamen,
    ...protokollnamen,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **A cloud provider's address** — piece 3 of A18 (#1533), labelled with
    // #1540.
    'dienst.anthropic.adresse': 'Anthropic 地址',
    'dienst.openai.adresse': 'OpenAI 地址',
    'dienst.xai.adresse': 'xAI 地址',
    ...netznamen,
    ...handelsnamen,
    ...gestaltungsnamen,
    ...ansichtsnamen,
    'konto.schau': '读取我自己的市场账户',
    ...kinamen,
    // **The client id per provider** — Werkstück 4 of Phase 4.12 (#1351).
    'anbieter.github.kennung': 'GitHub 的客户端 ID',
    'anbieter.google.kennung': 'Google 的客户端 ID',
    'anbieter.microsoft.kennung': 'Microsoft 的客户端 ID',
    'anbieter.x.kennung': 'X 的客户端 ID',

    // **The workspace and the two limits** — Werkstück O3, Phase 5.2 (#1427).
    'arbeitsplatz.navigator': '左栏',
    'arbeitsplatz.inspektor': '右栏',
    'arbeitsplatz.navigatorbreite': '左栏宽度',
    'arbeitsplatz.inspektorbreite': '右栏宽度',
    'richtlinie.nurlokal': '仅在本地工作',
    'richtlinie.kostengrenze': '费用上限，以最小单位计（欧元为分）',
  },

  // Origin, default and search — work item U11 of phase 5.1 (#1422).
  herkunftMarke: '由 {herkunft} 设定',
  wirksamKurz: '当前生效的是“{wert}”。',
  wirksamOffenKurz: '生效的是这台电脑的默认值。',
  wirksamGehalten: '当前生效的是“{wert}”——由一条策略锁定。',
  herkunftZeile: '由 {herkunft} 设定，层级为 {ebene}。',
  herkunftNamen: {
    profil: '此配置文件',
    nutzer: '你',
    geraet: '此设备',
    vorgabe: '无人（默认值）',
  },
  ebenenNamen: {
    profil: '配置文件',
    nutzer: '用户',
    geraet: '设备',
    arbeitsplatz: '工作区',
  },
  ebenenErklaerung:
    '每项设置都在一个层级上生效：设备、用户、配置文件或工作区。**设备**描述这台电脑——一个路径、一个窗口尺寸——交换时不会随之携带。**用户**对你的每个配置文件都生效。**配置文件**只在此处生效，并覆盖你在用户层级设定的内容。**工作区**是这个窗口中各栏的布局。生效的是配置文件中设定的内容，否则是你在用户层级设定的内容，否则是默认值——对于策略，则取最严格的值。没有任何设置需要管理员权限。',
  fachleute: '专业用户在这里还能设置什么',
  fachleuteSatz:
    '超时、通过其他网络的可达性、出站途径，以及供其他程序使用的服务。展开不会改变任何内容——只是显示更多。',
  zuruecksetzen: '重置为“{vorgabe}”',
  erbeMitWert: '{erbe}（{wert}）',
  zuruecksetzenVon: '将 {name} 重置为默认值',
  suche: '查找设置',
  suchePlatzhalter: '名称、键或值',
  sucheTreffer: '{zahl} 个匹配项',
  sucheKeine: '没有设置符合“{frage}”。',
  zumGebiet: '在“{gebiet}”区域中显示',

  // **Below the appearance settings** — UX-09 (#1681).
  gestaltung: gestaltungsbeiwerk,

  hinweise: {
    ...protokollhinweise,
    'sprache.eigene':
      '窗口使用的自定义翻译——留空表示不使用。若该翻译已不存在，则重新使用所选语言。',
    feldtiefe:
      '卡牌表单显示多少内容。**普通**是你要填写的字段；**高级**增加游戏包脚本计算出的字段；**专家**再增加游戏包完全没有说明的字段——`stylesheet` 及其同类。三者层层包含：越深，内容越多，而不是内容不同。按配置文件生效。',
    sprache:
      '适用于整个窗口和菜单栏。来自 MSE 包的内容——字段名称、游戏名称——保持文件中的原样。 德语是源语言；其他语言均为机器翻译，尚未经过人工审校。',
    darstellung:
      '数字、金额和日期的外观——例如 **de-CH** 或 **en-GB**。留空表示：与系统一致。文字的语言不受影响。',
    zeitzone:
      '时间以哪个时区显示——例如 **Europe/Berlin**。选择“与系统一致”即使用电脑自身的时区。按配置文件存储，交换时不会随之携带。',
    'filter.gespeichert':
      '所有视图已存储的筛选器，以 JSON 列表的形式呈现——名称、视图、标记，以及采用**且**还是**或**。它们在各自视图的导航器中创建；属于配置文件，并随设置导出一同带走。',
    'kuerzel.ziffern':
      '数字键加 ⌘ 会打开哪一页。留空表示：侧边栏中的各个区域。在“通用”下的快捷键表中设定。',
    erscheinungsbild: '浅色、深色，或跟随电脑设置。此选择只对本应用生效。',
    schriftgroesse:
      '缩放整个界面，而不只是文字——间距和边框也随之增大。卡牌预览保持不变：它显示的是将要印刷的样子。',
    aktualisierung:
      '默认为**关闭**：此时不会向网络发出任何请求，也不会询问任何内容。选择**通知**时，应用在启动时检查是否有较新版本并告知你——仅此而已。选择**下载**时，它还会提供一个按钮，把该版本获取到下载文件夹并比对其校验和；不会自行安装任何内容。',
    ersteschritte:
      '通往第一张卡牌的三个步骤，在启动时作为窗口显示。同一个开关也在该窗口内——在那里关闭后，可以在这里重新找到它。',
    seitenleiste:
      '左侧显示应用各个区域的栏。收起后，它的宽度归工作区所有；⌥⌘S、边缘上的按钮以及**显示**下的菜单项都可以收起和展开它。按配置文件存储，交换时不会随之携带。',
    leistenlage:
      '在左侧时，它竖直排列在内容旁边；在顶部时，它是内容上方的一行。条目、快捷键和方向键都相同——只是方向为横向。按配置文件存储，交换时不会随之携带。',
    fensterstart:
      '默认是**填满屏幕**——窗口占用它所能获得的空间，而**不是全屏**：全屏会把菜单栏一并占去。**正常大小**以内置尺寸打开，**与上次相同**则按你上次离开时的大小。按配置文件存储，交换时不会随之携带。',
    fensterbreite:
      '窗口上次关闭时测得的值——**仅用于“与上次相同”**。留空表示还没有上次，此时窗口以内置尺寸打开。',
    fensterhoehe:
      '高度同理。两者都在窗口关闭时写入；超出 700 至 8,192 点范围的测量值不会被记住——一个无法适应下一块屏幕的窗口，将无法再被抓取。',
    schriftart:
      '标题和系列名称所用的字体。正文仍是系统字体——装饰性字体用在整段文字上不易阅读。四种字体都随 macOS 提供；不会下载任何内容。',
    'barrierefreiheit.bewegung':
      '这项覆盖**只有一个方向**：它只能关闭动态效果，不能开启。在系统中设置了**减弱动态效果**的人，在这里也不会重新得到动态效果——推翻这一点的应用，等于在与他们作对。',
    'barrierefreiheit.kontrast':
      '方向相同：**高**会去掉纤维纹理并提高颜色，直到每段文字都达到 7 : 1。在系统中要求更高对比度的人，得到的结果相同。',
    'sicherheit.faktorbeimwechsel':
      '**唯一**会放宽的覆盖——而且只在**切换**配置文件时。它不改变登录：在那里，第二重验证发挥它应有的作用。在只有一个人使用的电脑上，每次切换都输入一次性验证码，只有麻烦，没有好处。',
    'vorgabe.zustand':
      '收藏中新行以什么开头。**无默认值**表示该字段保持为空，一如既往。',
    'vorgabe.waehrung':
      '新行的货币字段以什么开头，例如 **EUR**。不会做任何换算。',
    'vorgabe.einzelspiel':
      '新比赛的对阵是否只有**一局**。它决定的是显示方式，而不是计算——单局胜率和对手单局胜率无论如何都会保留。',
    'vorgabe.hochladepfad':
      '打开文件的对话框起始所在的文件夹——用于图像、MTG JSON 以及其他一切来自磁盘的内容。可以是完整路径，或以 **~/** 开头的路径；留空表示由操作系统决定打开的位置。',
    'vorgabe.ausfuhrpfad':
      '当操作没有自己的位置时，存储对话框起始所在的位置。如果存储时文件夹不存在——比如拔掉了驱动器——应用会告诉你，而不是存到别处。',
    'vorgabe.sicherungspfad':
      '备份你的资料库时，对话框起始所在的位置。文件夹缺失时，备份会带着一条提示中止。',
    'vorgabe.diagnosepfad':
      '诊断文本和诊断包的保存位置。文件夹缺失时会带着一条提示中止。',
    'bildausfuhr.ordner':
      '图片输出上次写入的位置。“图片”页会建议它，并把每个新选的文件夹记在这里——所有系列共用一个，与 MSE 相同。留空表示文件夹对话框从存储文件的文件夹打开。',
    'ausgabe.ziele':
      '“导出”在每个系列上次显示的目标。JSON 列表，最近的选择在末尾，最多 200 条；在“导出到哪里？”中选择时写入。留空表示不预先选择任何目标。',
    'dienst.anthropic.adresse':
      '请求发往哪里。留空表示 Anthropic 自己的接口。其他地址则是一个**网关**——一个转发请求的服务；它的密钥单独保存。',
    'dienst.openai.adresse':
      'OpenAI 同理。留空表示 `https://api.openai.com`。在此填入任何内容，意味着你不再与 OpenAI 对话，而是与该地址上应答的对象对话。',
    'dienst.xai.adresse': 'xAI 同理。留空表示 `https://api.x.ai`。',
    ...netzhilfe,
    ...handelshilfe,
    ...gestaltungshinweise,
    'konto.schau':
      '本程序是否可以读取你在交易市场上**自己的**账户——你的挂牌和订单。**关闭**表示：连看一眼都不行。仅开启交易市场并不够：那一项回答的是**其他**卖家的要价，这一项回答的是**你自己**这里有什么。在那里不会写入任何内容——没有挂牌，没有购物车，没有购买。',
    ...kihinweise,
    'anbieter.github.kennung':
      '你在 GitHub 创建的 OAuth 应用的 `client_id`。留空表示：不提供通过 GitHub 登录。',
    'anbieter.google.kennung':
      '你在 Google 创建的 OAuth 应用的 `client_id`。它通常以 `apps.googleusercontent.com` 结尾。',
    'anbieter.microsoft.kennung':
      'Microsoft Entra 门户中的应用 ID——一个 UUID。Microsoft 对这条途径**不**要求密钥。',
    'anbieter.x.kennung': '你在 X 开发者门户中创建的应用的 `client_id`。',
  },

  werte: {
    ...schalterwerte,
    ...protokollwerte,
    ...handelswerte,
    ...gestaltungswerte,
    ...ansichtswerte,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **Die Kontoschau** — T11, #1476. Dieselben Worte wie bei der
    // KI-Laufzeit darüber: zwei Schreibweisen für denselben Zustand
    // lernt niemand gern.
    'konto.schau.aus': '关闭',
    'konto.schau.an': '开启',
    ...kiwerte,
    'sprache.erben': '与用户设置相同',
    'erscheinungsbild.erben': '与用户设置相同',
    'schriftgroesse.erben': '与用户设置相同',
    'schriftart.erben': '与用户设置相同',
    'sprache.de': 'Deutsch',
    'sprache.en': 'English',
    'sprache.es': 'Español',
    'sprache.fr': 'Français',
    'sprache.it': 'Italiano',
    'sprache.ja': '日本語',
    'sprache.ru': 'Русский',
    'sprache.zh': '简体中文',
    'erscheinungsbild.system': '跟随系统',
    'erscheinungsbild.hell': '浅色',
    'erscheinungsbild.dunkel': '深色',
    'schriftgroesse.klein': '小',
    'schriftgroesse.normal': '正常',
    'schriftgroesse.gross': '大',
    'schriftgroesse.sehrgross': '特大',
    'aktualisierung.aus': '关闭',
    'aktualisierung.melden': '仅通知',
    'aktualisierung.holen': '通知并下载',
    'ersteschritte.an': '开启',
    'ersteschritte.aus': '关闭',
    'seitenleiste.sichtbar': '可见',
    'seitenleiste.verborgen': '隐藏',
    'leistenlage.senkrecht': '左侧',
    'leistenlage.waagerecht': '顶部',
    'fensterstart.maximiert': '填满屏幕',
    'fensterstart.normal': '正常大小',
    'fensterstart.zuletzt': '与上次相同',

    'feldtiefe.normal': '普通',
    'feldtiefe.erweitert': '高级',
    'feldtiefe.experte': '专家',
    'tisch.autoanordnung.an': '是——由牌桌来摆放',
    'tisch.autoanordnung.aus': '否——你放在哪里就留在哪里',
    'tisch.beschriftung': '牌桌上的标注',
    'tisch.beschriftung.aus': '关闭——只显示数量',
    'tisch.beschriftung.klein': '小',
    'tisch.beschriftung.normal': '中',
    'tisch.beschriftung.gross': '大',
    'schriftart.pergament': '羊皮纸 — 沉静易读',
    'schriftart.fantasy': '奇幻 — 活泼',
    'schriftart.antike': '古典 — 罗马大写字母',
    'schriftart.kalligrafie': '书法 — 手写',
    'barrierefreiheit.bewegung.system': '跟随系统',
    'barrierefreiheit.bewegung.reduziert': '始终减少',
    'barrierefreiheit.kontrast.system': '跟随系统',
    'barrierefreiheit.kontrast.hoch': '始终提高',
    'sicherheit.faktorbeimwechsel.verlangen': '切换时也要',
    'sicherheit.faktorbeimwechsel.nurAnmelden': '仅登录时',
    'vorgabe.zustand.ohne': '无默认值',
    'vorgabe.zustand.mint': '全新',
    'vorgabe.zustand.nearMint': '近全新',
    'vorgabe.zustand.excellent': '极佳',
    'vorgabe.zustand.good': '良好',
    'vorgabe.zustand.lightPlayed': '轻微磨损',
    'vorgabe.zustand.played': '有磨损',
    'vorgabe.zustand.poor': '较差',
    'vorgabe.einzelspiel.aus': '关闭',
    'vorgabe.einzelspiel.an': '开启',

    // **The workspace and the limit** — Werkstück O3 of Phase 5.2 (#1427).
    'arbeitsplatz.navigator.erben': '与用户设置相同',
    'arbeitsplatz.navigatorbreite.erben': '与用户设置相同',
    'arbeitsplatz.inspektorbreite.erben': '与用户设置相同',
    'arbeitsplatz.navigator.sichtbar': '可见',
    'arbeitsplatz.navigator.verborgen': '隐藏',
    'arbeitsplatz.inspektor.erben': '与用户设置相同',
    'arbeitsplatz.inspektor.sichtbar': '可见',
    'arbeitsplatz.inspektor.verborgen': '隐藏',
    'richtlinie.nurlokal.aus': '关闭',
    'richtlinie.nurlokal.an': '开启',
  },

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351).
  // **Where a login comes from in the first place** — follow-up to Werkstück 4
  // of Phase 4.12 (#1351).
  anmeldung: '通过提供商登录',
  anmeldungHinweis:
    '本应用**不附带任何凭据**。想通过 GitHub、Google、Microsoft 或 X 登录，需要自己在提供商处创建一个 **OAuth 应用**，并在此填入其 `client_id`。重定向 URI 请填写登录时浏览器中显示的地址——它指向这台电脑，且只在登录进行期间有效。',

  anmeldungGeheimnis:
    'GitHub 另外要求一个 `client_secret`——Google 可选填，Microsoft 则不允许这类程序使用。它存入**钥匙串**，而不是这里：资料库会被传给别人，钥匙串不会。请在**钥匙串访问**中将它创建为**应用程序密码**，名称为 `MagicSetEditor4Mac-Anmeldung`，账户为 `github`。在命令行中：`security add-generic-password -s MagicSetEditor4Mac-Anmeldung -a github -w`。',

  geheimnis: '密钥',
  geheimnisPflicht: '必填，但缺失',
  geheimnisVerboten: '不需要密钥，也会拒绝密钥——这里只需客户端 ID。',

  anbietername: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  },

  wirktNachNeustart: '重新启动后生效。',

  trefferInGruppe: '“{frage}”有 {zahl} 个匹配项',
  gesuchtMarke: '找到',
  sucheSprung: '跳到第一个匹配项',

  vorgaben: '默认值',

  // **The pointer to the help** — Werkstück 13 of Phase 4.12 (#1351).

  dienste: '第三方卡牌服务',
  diensteHinweis:
    '关闭表示：不会向此服务发出任何请求。没有任何服务时，应用照样运行——只是来自外部的内容不会出现。',

  kiAnbieter: '语言模型',
  kiAnbieterHinweis:
    '问题可以发往哪里。每个提供商都在自己的机器上计算，并能看到你发送的内容——所以每个提供商都要单独开启，没有任何一个会自动开启。',
  // See the German catalogue for why these two live here and not under
  // `werte`: they are the same for every service.
  dienstAus: '关闭',
  dienstAn: '开启',
  dienstUngeprueft: '从未对真实服务运行过',
  dienstUngeprueftHinweis:
    '这里没有此服务的账户。经过测试的是我们这一侧的边界——请求、请求头、映射。服务是否按我们假设的方式应答，则没有测量过。',
  fragen: '谁回答什么',
  fragenHinweis:
    '当多个服务都能回答同一个问题时，由这里决定询问哪一个。**不会静默回退**：如果所选服务没有响应，你会被告知——而不是悄悄看到来自另一个来源的答案。',
  'frage.suche': '卡牌搜索',

  // The service catalogue — work item A1 of phase 5.3 (#1440).
  ...dienstkatalog,

  dienstSchluesselFehlt: '未存储凭据',
  dienstSchluesselHinweis:
    '此服务需要凭据。它们存放在 macOS 钥匙串中，而不是收藏文件中——收藏文件会被传给别人。',
  zugaenge: '凭据',
  zugaengeHinweis:
    '它们存放在 macOS 钥匙串中，而**不是**收藏文件中——收藏文件会被传给别人。这里只显示是否存有内容；值本身绝不会再返回。',
  // See the German catalogue: these name the headers each service asks for
  // (documented in decision 0056).

  kuerzelUeberschrift: '键盘快捷键',
  ...einstellungsbelegung,
  kuerzelHinweis: '它们在整个窗口中生效。在文本字段内，⌘Z 和 ⌘F 归该字段所有。',
  kuerzelZweiteReihe:
    '带 ⌥ 的八个快捷键位于最上面一排按键上，从左数起。在布局不同的键盘上——例如 AZERTY——那里印着的是别的字符；按键则是同一个。',
  kuerzelTaste: '按键',
  kuerzelHandlung: '操作',

  austausch,
  eigenuebersetzung,
  // **An account's remaining quota at a provider** — A5, #1444.
  kontingentTitel: '提供商配额',
  kontingentErklaerung:
    '套餐包含多少积分，还剩多少。查询会通过网络发出，本身也要消耗一个积分。',
  kontingentHolen: '查询配额',
  kontingentHolenBei: '查询 {dienst} 的配额',
  kontingentLaeuft: '正在查询……',
  kontingentZahlen: '还剩 {rest} / {gesamt} 积分。',
  kontingentFrist:
    '提供商给出的这个数字可能滞后最多 {minuten} 分钟——它不是上限。这里真正起限制作用的，是此资料库的预算。',
  kontingentUeberzug: '超出套餐的一个积分费用为 {satz}。',
  // **What is known about AI models** — A6, #1445.
  ...ki,

  // **Die Zugangsdaten** (#1534) — eigene Datei, ein Katalog (§9).
  ...zugaenge,

  ...netzbereich,
}
