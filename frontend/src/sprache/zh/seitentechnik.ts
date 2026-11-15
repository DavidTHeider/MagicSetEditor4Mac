// The texts of the **technical page** — operating state, effect counter and
// the three tools without a window.
//
// Split out of `seiten.ts` when the file grew past 400 lines with the effect
// counter (work item O9, #1433).

import { betrieb, leistungsprofil, zwischenspeicher } from './seitenbetrieb'
import { dienstzeugnis, fernwege, schnittstellen } from './seitentechnikdienst'
import { einladung } from './einladung'
import { gegenueber } from './gegenueber'

export const technik = {
  betrieb,
  dienstzeugnis,
  einladung,
  gegenueber,
  fernwege,
  schnittstellen,
  leistungsprofil,
  zwischenspeicher,
  gesundheit: '运行状态',
  gesundheitErklaerung:
    '此进程是否存活，是否能够工作。外部服务不响应，并不会让它变成未就绪——没有网络时，编辑仍可继续。',
  lebt: '进程有响应。',
  bereit: '可以工作。',
  nichtBereit: '目前无法工作。',
  grundBestandZu: '没有打开任何资料库。',
  grundWanderungLaeuft: '资料库正在迁移到新的结构。',
  grundBestandUnlesbar: '资料库没有响应。',
  fremdeDienste: '外部服务',
  erreichbar: '可访问',
  nichtErreichbar: '无法访问',
  metriken: '启动以来的影响',
  metrikenErklaerung:
    '此进程对外部做过什么——按界面、影响类别和结果分类。只统计经过检查的影响；单纯的读取不会出现在这里。',
  metrikenGrenze: '最多 {grenze} 行——不会随卡牌、系列或个人资料的数量增长。',
  metrikenLeer: '尚未对外部产生任何影响。',
  spalteFlaeche: '界面',
  spalteWirkung: '影响',
  spalteAusgang: '结果',
  spalteAnzahl: '数量',
  flaechen: {
    fenster: '窗口',
    kommandozeile: '命令行',
    mcp: 'MCP',
    dienst: '启动供脚本使用的服务',
  },
  wirkungen: {
    lokalLesen: '本地读取',
    lokalSchreiben: '本地写入',
    externSenden: '向外发送',
    externSchreiben: '更改外部内容',
    verbindlich: '具有约束力',
  },
  ausgaenge: { gewaehrt: '已允许', abgewiesen: '已拒绝' },
  paket: {
    ueberschrift: '诊断包',
    erklaerung:
      '同样的信息按部分整理，并附带校验和，保存为文件——适用于仅靠文字不够的问题报告。它会创建在你指定的位置，不会自行发往任何地方。',
    inhalt:
      '其中包含版本、平台、资料库、资源包、进程计数、日志行和崩溃记录——不含卡牌、图片、收藏或凭据。',
    zusammenstellen: '整理',
    sichern: '存储诊断包',
    loeschen: '删除诊断包',
    groesse: '{bytes}，上限 {grenze}。',
    zeitraum: '日志行的时间范围为 {von} 至 {bis}。',
    ausgelassen: '已省略——{teil}：',
    teile: {
      fassung: '版本与平台',
      bestand: '资料库的位置与结构',
      pakete: '已安装的模板资源包',
      metriken: '此进程的计数',
      protokoll: '日志行',
    },
    spalteTeil: '部分',
    spalteDatei: '文件',
    spalteZeilen: '行数',
    spalteBytes: '大小',
    spalteMitnehmen: '包含',
    mitnehmen: '在诊断包中包含{teil}',
  },
  diagnose: {
    ueberschrift: '用于问题报告的诊断信息',
    erklaerung:
      '出现问题时，这段文字有助于查明原因。它只显示在这里——是否转交给别人，由你决定。',
    inhalt:
      '其中包含版本、电脑、数据库、日志和模板资源包——不含卡牌文字、字段值和系列内容。具体包括：版本、电脑类型、本程序所属的构建、窗口与程序之间的绑定约定、数据库的位置、结构状态、孤立行的数量、日志的位置、你的模板资源包，以及最近 200 行日志。你的用户文件夹显示为“~”。',
    anzeigen: '显示诊断信息',
    kopieren: '拷贝',
    kopiert: '已拷贝',
    kopierfehler: '拷贝失败——窗口没有剪贴板权限。请选中文字后手动拷贝。',
    sichern: '存储为文件',
    // Filter über die Zeilen (NET-02, S13.NAT-04, #1706).
    filter: '在行中搜索',
    zeilenGezeigt: '{alle} 行中的 {gezeigt} 行',
  },
  // **Die Aufschrift der Lasche — und nicht die des Abschnitts darin.**
  //
  // Sie hießen bis zum 2026-09-14 gleich, und die Lasche trug elf Kästen,
  // von denen einer „Werkzeuge ohne Fenster" war. Wer die
  // Netzeinstellungen suchte, suchte überall sonst.
  bereichUeberschrift: '技术与诊断',

  // **Das Gebiet, das dabei entstanden ist.**
  netzUeberschrift: '网络与可达性',

  ueberschrift: '无窗口工具',
  erklaerung:
    '本应用附带三个无需窗口即可运行的程序：用于命令行的 mse、用于语言模型的 mse-mcp、用于脚本服务的 mse-rest；它们位于应用程序包内，只有在下方设置后才会进入搜索路径，这里的各行给出完整位置。',
  keineShell: '这些内容用于拷贝，而不是运行——请在你自选的终端中使用。',
  ort: '它们位于 {ort}。',
  vorhanden: '已有：{namen}。',
  keineWerkzeuge:
    '三者都不存在。应用从源码运行时这是正常的——发布流程会将它们放入应用包。',
  kopieren: '拷贝命令',
  kopiert: '已拷贝。',
  kopierfehler: '拷贝失败——窗口没有剪贴板权限。请选中文字后手动拷贝。',
  befehle: {
    hilfe: '命令行能做什么',
    sets: '列出资料库中的系列',
    verknuepfen: '让 mse 出现在搜索路径中（会要求输入你的密码）',
    dienst: '在回环接口上启动服务',
    mcp: '启动 MCP 服务器——供语言模型使用，通过标准输入输出',
    karten: '列出已打开系列的卡牌',
    statistik: '统计已打开的系列',
  },

  // **The local search index** — work item A15 of phase 5.3
  // (#1514).
  suchindex: '本地搜索索引',
  suchindexErklaerung:
    '卡牌文字和规则段落，可在本地搜索——全文与向量检索合并为一份列表。每个命中都会注明出处，以及来源当时所处的修订版。',
  suchindexKeiner: '尚未建立搜索索引。在此之前，这里的搜索什么也找不到。',
  suchindexModell: '模型',
  suchindexLaufzeit: '运行环境',
  suchindexVerfahren: '方法',
  suchindexStaende: '各来源的修订版',
  suchindexSemantisch: '语义搜索',
  suchindexJa: '是',
  suchindexNein: '否',
  suchindexFrage: '要找什么？',
  suchindexSuchen: '搜索',
  suchindexLeer: '没有找到。',
  suchindexModus: '检索方式',
  suchindexModusHybrid: '全文与向量',
  suchindexModusLexikalisch: '仅全文',
  suchindexModusOhneIndex: '未检索——没有可用的索引',
  suchindexAufbau: '重建',
  suchindexAufbauErklaerung:
    '在运行中的索引旁建立新一代索引，到最后才切换。这需要几分钟，并且需要已配置的嵌入模型。',
  suchindexModellfeld: '嵌入模型',
  suchindexQuellfeld: '批量文件',
  suchindexRevisionfeld: '来源修订版',
  suchindexDelta: '仅读取有变化的内容',
  suchindexAufbauen: '建立',
  suchindexAbbrechen: '取消',
  suchindexLaeuft: '正在建立',
  suchindexGeschrieben: '个部分已写入',
  suchindexFertig: '一代索引已建立',
  // **Configuration for agents** — A14, #1453.
  agenten: '为智能体做好准备',
  agentenErklaerung:
    'Claude Code 和 Codex 使用 MCP。下面的片段带有 mse-mcp 在这台电脑上的完整路径——拷贝并粘贴即可。本应用不会写入其中任何文件，也不会启动任何智能体。',
  agentenOhneGeheimnis:
    '其中不含任何访问密钥。mse-mcp 会自行从钥匙串中读取——写在这里的任何内容，都会以明文存放在一个可能被分享的文件里。',
  agentenNamen: {
    claude: 'Claude Code',
    codex: 'Codex',
  } as Record<string, string>,
  befehlszeileTitel: '设置命令行',
  befehlszeileSatz:
    '在你自己的文件夹中放置指向 mse 和 mse-mcp 的链接——无需管理员权限，也不会改动你的 shell 配置文件。那里已有且并非由这里放置的内容会保持不变。',
  befehlszeileVerzeichnis: '文件夹：{verzeichnis}',
  befehlszeileNichtImPfad:
    '这个文件夹不在你的 PATH 中。请自行将下面这一行加入你的 shell 配置文件：',
  befehlszeileStaende: {
    frei: '尚未设置',
    eigen: '已设置',
    veraltet: '已过时——指向的程序已不存在',
    fremd: '那里有另一个命令，将保持不变',
  } as Record<string, string>,
  befehlszeileVerdeckt: '被 {pfad} 遮挡：输入该名称时运行的是那一个。',
  befehlszeilePruefen: '检查方式',
  befehlszeileEinrichten: '设置',
  befehlszeileHerausnehmen: '再次移除',
  befehlszeileKopierenPfad: '拷贝用于 shell 配置文件的那一行',
  befehlszeileKopierenPruefung: '拷贝检查命令',
  agentenDatei: '应放入 {datei}',
  agentenInhalt: '{datei} 的内容',
  agentenBefehl: '或通过智能体的命令行',
  agentenNachgelesen: '已查证：{quelle}',
  agentenKopierenDatei: '拷贝片段',
  agentenKopierenBefehl: '拷贝智能体的命令',

  peerTitel: '此个人资料如何表明身份',

  peerZeigen: '显示身份',
  peerErklaerung:
    '一对密钥，让别人确认自己确实在与你交谈。私钥保存在这台电脑的钥匙串中，不会发往任何地方；只有标识符会对外发出。',
  // Was diese Identität nicht ist (NET-02, S13.NAT-05, #1706).
  peerNichtKonto:
    '它属于此个人资料——既不是你的 Apple 账户，也不是这台电脑的名称。',
  peerStandOhne: '此个人资料还没有身份。',
  peerStandBereit: '就绪。',
  peerStandVerloren: '密钥缺失。',
  peerStandFremd: '钥匙串与数据库不一致。',
  peerKennung: '标识符',
  peerFingerabdruck: '指纹',
  peerFingerabdruckHinweis: '这是要向对方念出的内容——显示名称不能作为证明。',
  peerSeit: '自 {zeitpunkt}',
  peerVorher: '之前：{kennung}',
  peerAnlegen: '创建身份',
  peerRotieren: '更换身份',
  peerRotierenFrage: '要创建新身份吗？所有已确认当前指纹的人都必须重新核对。',
  peerRotierenJa: '更换',
  peerRotierenNein: '取消',

  // **The known counterparts** — work item N1 of the same issue.
  kontakteTitel: '此个人资料认识的人',
  kontakteErklaerung:
    '此个人资料已经遇到过的对方。仅仅遇到并不允许任何事：只有核对指纹才能使对方可被验证，只有按用途授予的权限才允许做事。',
  kontakteLeer:
    '尚未遇到任何人。之后有人联系时，会显示在这里——处于未确认状态。',
  kontakteZeigen: '显示联系人',
  kontakteAufnehmen: '添加联系人',
  kontakteAufnehmenErklaerung:
    '在没有连接时，标识符需要手动交换。在这里添加的人处于未确认状态，不能做任何事。',
  kontakteAufnehmenKennung: '对方的标识符',
  kontakteAufnehmenName: '你对他们的称呼',
  kontakteStandUnbestaetigt: '未确认',
  kontakteStandBestaetigt: '已确认',
  kontakteStandGesperrt: '已屏蔽',
  kontakteAbdruck: '指纹',
  kontakteNennt: '自称：{name}',
  kontakteNenntHinweis: '自选的名称不能证明身份。',
  kontakteGesehen: '上次见到：{zeitpunkt}',
  kontakteVergleichen: '输入你核对过的指纹',
  kontakteBestaetigen: '确认',
  kontakteSperren: '屏蔽',
  kontakteEntsperren: '取消屏蔽',
  kontakteRechte: '已授予',
  kontakteKeineRechte: '未授予任何权限。',
  zweckPaarung: '互相认识',
  zweckTeilen: '传输对象',
  zweckHandel: '交易卡牌',
  zweckPartie: '进行对局',
  kontakteAdresse: '访问地址',
  kontakteAdresseBeispiel: '192.168.1.7:41234',
  kontakteAdresseHinweis:
    '可以是同一网络中的地址——也可以是对方发给你的邀请。留空表示没有。',
  kontakteAdresseMerken: '记住',

  // **Die lokale Notiz und die freigegebenen Sichten** — Werkstück T3
  // von Phase 5.6 (#1468).
  kontakteNotiz: '备注',
  kontakteNotizBeispiel: '交易公道，回复较慢…',
  kontakteNotizHinweis:
    '仅供你自己查看。它绝不会传出——既不会进入视图，也不会进入共享。',
  kontakteNotizMerken: '存储备注',
  freigabenTitel: '此对等端能看到什么',
  freigabenErklaerung:
    '没有共享，它什么也看不到。多个范围会同时生效——范围较大的优先。',
  freigabenLeer: '尚未授予任何视图。',
  freigabenOhneZweck: '此对等端目前什么也不能做。请先为它授予一个用途。',
  freigabenZweck: '用途',
  freigabenUmfang: '内容',
  freigabenErteilen: '授予视图',
  freigabenWiderrufen: '撤销',
  freigabenVorschau: '查看它能看到什么',
  freigabenVorschauLeer: '它将什么也看不到。',
  fernTitel: '跨网络可达',
  fernErklaerung:
    '这样即使不在同一网络，别人也能联系到你——无需端口转发，无需账户。你会得到一份邀请，以你喜欢的任何方式发给对方即可。',
  fernStill: '此个人资料无法跨网络访问。',
  fernHinweis:
    '连接会先寻找直连路径，只有别无他法时才使用中继。经过中继的只有加密流量——没有卡牌，没有名称。',
  fernStarten: '创建邀请',
  fernBeenden: '停止可达',
  fernWeitergeben:
    '这是对方所需要的内容。截获它的人得到的只是一行文字，而不是身份——检查会在连接时通过指纹进行。',
  fernStirbt:
    '只要此个人资料保持可达，邀请就一直有效。停止后邀请即作废——下次会得到新的邀请。',
  fernKopieren: '拷贝邀请',
  fernKopiert: '已拷贝。',
  horchenTitel: '对他人可达',
  horchenErklaerung:
    '此个人资料在监听期间，已知的对等端可以向它发起连接。' +
    '这绝不会自行发生——重启后也不会——' +
    '切换个人资料会将其结束。',
  horchenAdresse: '监听的地址',
  horchenAdresseBeispiel: ':41234',
  horchenAdresseHinweis:
    '只写端口（:41234）会监听这台电脑的所有接口。' +
    '前面加上 127.0.0.1，则只监听这台电脑本身。',
  horchenStarten: '设为可达',
  horchenBeenden: '停止可达',
  horchenLaeuft: '正在监听 {adresse}',
  horchenNamen: '名称（更换地址后仍然有效）',
  horchenAdressen: '地址',
  horchenNurHier:
    '采用这种绑定时，外部没有人能联系到你：127.0.0.1 只监听这台电脑。请只填端口（:41234），这台 Mac 才能在其网络中被访问。',
  horchenNurDiesesNetz:
    '两者在这个网络中都有效。租用的地址可能会变，名称通常不受影响——而在任何地方都有效的地址，只有跨网络连接时才会出现。',
  horchenStill: '此个人资料目前没有在监听。',
  horchenWeitergeben: '这是对方所需要的内容——连同此个人资料的密钥。',
  sichtbarTitel: '被附近的设备发现',
  sichtbarErklaerung:
    '通告会让同一网络中的所有人知道此个人资料的存在——包括它的对等端 ID。' +
    '可达与被发现是两个不同的决定，这里是第二个。',
  sichtbarName: '通告所用的名称',
  sichtbarNameHinweis:
    '自选的名称不能证明任何事。这台 Mac 的主机名不会对外发出。',
  sichtbarWerden: '通告',
  sichtbarBeenden: '停止通告',
  sichtbarLaeuft: '此个人资料正在对外通告。',
  sichtbarStill: '此个人资料没有在对外通告。',
  sichtbarBrauchtHorchen:
    '请先设为可达——没有端口就通告，就像邀请人去敲一扇后面没有人的门。',
  nachbarnTitel: '附近有谁',
  nachbarnErklaerung:
    '此刻正在同一网络中通告自己的对等端。你看到的只是它们的自称——只有在连接时才会核实。',
  nachbarnSuchen: '看一看',
  nachbarnSucht: '正在查找…',
  nachbarnLeer: '此刻没有人在通告自己。',
  nachbarnUnbekannt: '未知',
  nachbarnAufnehmen: '添加为联系人',
  nachbarnGemerkt: '已记为',
  nachbarnSchonBekannt: '已在列表中',
  // Ein bekannter Peer an einer neuen Adresse (NET-02.NAT-02, #1706).
  nachbarnAdresseNeu: '该对等端位于新的地址。',
  nachbarnAdresseBisher: '此前记录的地址',
  nachbarnAdresseKeine: '无',
  nachbarnAdresseUebernehmen: '采用新地址',
  nachbarnAdresseHinweis:
    '这只会改变连接的目标地址。对它的信任不会比之前更多，其密钥在每次连接时都会核实。',
  nachbarnVerworfen: '有 {zahl} 条回复没有带对等端 ID，因此不在列表中。',

  // **How this program goes online** — N5
  // (#1571).
  netzwegTitel: '来自此进程的环境',
  netzwegErklaerung:
    '这些值来自程序启动时所处的环境。它们会覆盖下面的字段，且无法通过窗口更改。',
  netzwegProxyHTTP: 'http 代理',
  netzwegProxyHTTPS: 'https 代理',
  netzwegProxyAus: '不使用代理的地址',
  netzwegAusUmgebung: '{name}：{wert}',

  // **The two switchable services** — N6
  // (#1572).
  dienstRESTTitel: '脚本接口',
  dienstRESTErklaerung:
    '基于同一数据库的网络服务——用于从脚本读取和更改。它只在你让它运行时才运行，并随此窗口一起结束。',
  dienstMCPTitel: '基于 HTTP 的工具服务器',
  dienstMCPErklaerung:
    '与标准输入方式相同的工具，这里通过一个地址提供——其他设备也可以访问。',
  dienstRESTAdresse: '脚本接口的地址',
  dienstRESTPort: '脚本接口的端口',
  dienstMCPAdresse: '工具服务器的地址',
  dienstMCPPort: '工具服务器的端口',
  dienstMarke: '访问凭据',
  dienstMarkeKopieren: '拷贝访问凭据',
  dienstFingerabdruck: '指纹',
  dienstFingerabdruckKopieren: '拷贝指纹',
  dienstKopiert: '已拷贝',
  dienstStarten: '启动',
  dienstBeenden: '停止',
  dienstErreichbar: '访问地址：',
  dienstMitTLS:
    '已加密。证书为自行签发，并带有此个人资料的 ID——连接的人会核对其指纹。',
  dienstOhneTLS: '未加密，这里没有问题：该地址只在这台电脑上有效。',
  dienstHinweis:
    '127.0.0.1 表示：仅限这台电脑。0.0.0.0 表示：网络中的所有设备——此时必须加密并设置访问凭据，两者会自动设好。端口 0 表示：自动寻找空闲端口。',
}
