// **The “Users and profiles” page** — phase 4.12
// (#1351).
import { faktor, konto } from './seitenkonto'
import { sitzung } from './seitensitzung'

export const nutzer = {
  titel: '用户和配置文件',

  laschen: {
    allgemein: '通用',
    konto: '账户',
    profil: '配置文件',
    persoenlich: '个人',
  },

  allgemeinErklaerung:
    '用户就是这个人。配置文件是这个人的一个工作环境——有自己的系列、自己的收藏、自己的设置。当前谁在哪里工作，显示在侧边栏的应用名称下方；切换在这里进行。',

  kontoErklaerung:
    '此用户用什么登录。只要恰好只有一个用户且只有一个配置文件，就完全不会询问。',

  profilErklaerung:
    '想为比赛单独收藏的人，需要的是第二个配置文件，而不是第二个账户。',

  niemand: '还没有其他用户。第一个用户是在首次启动时创建的。',
  keineProfile: '此用户还没有配置文件。',

  neuerNutzer: '创建用户',
  neuerNutzerName: '新用户的显示名称',
  neuerNutzerNameHinweis: '必须唯一，且首尾不含空格——用来区分两个用户',
  neuerNutzerKnopf: '创建用户',
  nutzerAngelegt: '已创建：{name}',

  neuesProfil: '再创建一个配置文件',
  neuesProfilNutzer: '用于哪个用户',
  neuesProfilName: '配置文件名称',
  neuesProfilNameHinweis: '每个用户内唯一——两个人都可以有一个“比赛”',
  neuesProfilKnopf: '创建配置文件',
  profilAnlegenOeffnen: '创建配置文件…',
  profilAngelegt: '已创建：{name}',

  wahl: '在哪个配置文件中工作',
  wahlErklaerung:
    '系列、套牌、收藏、比赛和设置都属于所选的配置文件。切换不会删除任何内容——切换回来后，这里的一切仍然都在。',
  hier: '你当前在此工作',
  hierhin: '切换到这里',
  hierhinFuer: '切换到 {nutzer} · {profil}',

  nutzerwahl: '当前谁在工作',
  nutzerwahlErklaerung:
    '查看他人的配置文件不会改变你是谁——你自己的共享决定了你能去哪里。这与上方的配置文件是两回事。要成为另一个人，你需要证明其第二重验证；之后会在其第一个配置文件中工作。',
  dasBinIch: '这是你当前的身份',
  werden: '成为此人',
  werdenFuer: '成为 {nutzer}',

  anmeldung: '通过提供商登录',
  anmeldungErklaerung:
    '想通过提供商登录的人，请在**设置 › 用户和配置文件 › 通过提供商登录**下填入其 `client_id`。MagicSetEditor4Mac 不附带任何凭据；你需要在提供商处创建自己的 OAuth 应用，如果它要求密钥，密钥会存入钥匙串，而不是资料库。',
  anmelden: '登录',
  anbieterstand: '{eingerichtet}，{geprueft}',
  eingerichtet: '已设置',
  nichtEingerichtet: '未设置',
  geprueft: {
    simuliert: '仅针对模拟环境测试',
    realGeprueft: '已针对真实提供商测试',
  },
  anmeldenBei: '使用 {anbieter} 登录',
  anmeldungAbbrechen: '取消',
  anmeldungAbbrechenBei: '取消使用 {anbieter} 登录',
  verbunden: '已连接',
  verbundenAls: '已作为 {name} 连接',
  loesen: '断开连接',
  loesenVon: '断开与 {anbieter} 的连接',

  anbieternamen: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  } as Record<string, string>,

  pflege: '编辑用户和配置文件',
  pflegeErklaerung:
    '名称可以更改。删除时，名称必须保持原样——用它来确认指的是谁。',
  bearbeiten: '编辑 {name}',
  bearbeitenKurz: '编辑',
  neuerNameFuer: '名称',
  umbenennen: '重命名',
  loeschen: '删除',
  loeschenBestaetigt: '永久删除 {name}',
  loeschenFrage: '要永久删除 {name} 吗？',
  loeschenJa: '永久删除',
  abbrechen: '取消',
  nutzerLoeschenHinweis:
    '删除用户时，其配置文件也会一并删除。最后一个用户无法删除。',
  profilLoeschenHinweis:
    '用户的最后一个配置文件无法删除。配置文件中的内容仍保留在资料库中，但无法再访问。',

  angaben: '个人信息',
  angabenErklaerung:
    '除显示名称外，其余都是自愿填写的。**称谓**是自由填写的字段：任何列表都会漏掉某些人，而强迫人在错误选项中做选择的应用，就是称呼错了对方。地址、电话号码和密码在**账户**下。',
  angabenSichern: '保存信息',
  felder: {
    name: '显示名称',
    anrede: '称谓',
    namenszusatz: '姓名后缀',
    vorname: '名',
    weiterenamen: '其他名字',
    nachname: '姓',
    slogan: '个性签名',
    kurzebio: '简短简介',
    langebio: '详细简介',
  },

  // **Die Profil-Einstellungen** — Werkstück 9 von Phase 4.12 (#1351).
  badge: '配置文件徽章',
  badgeVon: '{name} 的配置文件徽章',
  badgeWaehlen: '选择徽章……',
  badgeEntfernen: '移除徽章',
  profilangaben: '配置文件信息',
  profilangabenErklaerung:
    '这些属于**此配置文件**，而不是你本人——拥有多个配置文件的人，可以在此加以区分。**副签名**显示在你自己的签名下方，不会取代它。**徽章**位于你的头像右下角：它显示你当前工作所用的角色。',
  profilangabenSichern: '保存配置文件',
  profilfelder: {
    name: '配置文件名称',
    subslogan: '副签名',
    kurzebeschreibung: '简短描述',
    langebeschreibung: '详细描述',
  },

  // **Technische Profile und ihre Zugriffstoken** — Werkstück 10 von
  // Phase 4.12 (#1351), Punkte 12 und 13.
  technisch: '技术配置文件——背后没有真人',
  technischErklaerung:
    '技术配置文件把服务自动写入的内容，与你亲手所做的内容区分开。让服务代为操作的人，不希望它写进**自己的**配置文件：那里出现的内容并非出自自己之手。它仍然属于你，因为总得有人能把它关掉。',
  ablauf: '到期日（留空：永不）',
  ablaufErklaerung:
    '此配置文件在该时间点之后即视为已过期——**并非删除指令**。不会有任何内容自行消失。',

  token: '访问令牌',
  tokenErklaerung:
    '自动化调用方借此在 REST 接口上表明身份——在**此**配置文件中，并具有此处设定的权限。窗口、命令行和 MCP 都不要求令牌：在那里已经确定是谁在工作。',
  tokenLeer: '此配置文件还没有访问令牌。',
  tokenBezeichnung: '用途',
  tokenScope: '权限',
  tokenLesen: '仅读取',
  tokenSchreiben: '读取和写入',
  tokenAblauf: '到期日（留空：永不）',
  tokenAnlegen: '创建令牌',
  tokenAendern: '应用',
  tokenLoeschen: '撤销',
  tokenLoeschenFrage: '要撤销令牌“{name}”吗？',
  tokenLoeschenErklaerung:
    '使用它的一方此后将无法再进入。新令牌必须在那边重新填入。',
  tokenLoeschenJa: '撤销',
  tokenAbgelaufen: '已过期',
  tokenUnbefristet: '永不过期',

  tokenEinmalig:
    '明文**只会出现一次**，就在创建之后。此后它不会保存在任何地方；丢失的人，需要创建新令牌。',
  tokenKlartext: '请立即复制——此值不会再次出现：',

  // **Rotation, bulk revocation, expiry presets and MCP client set-up** —
  // piece O2 of phase 5.2 (#1426).
  tokenPrincipal: '为谁工作',
  tokenAblaufwahl: '有效期',
  tokenAblaufwahlen: {
    tag: '一天',
    monat: '三十天',
    jahr: '一年',
    nie: '永不',
    eigen: '自定时间',
  },
  tokenRotieren: '新密钥',
  tokenRotierenFrage: '要生成新密钥吗？',
  tokenRotierenErklaerung:
    '旧密钥还会再有效一个小时，以便正在运行的任务能够完成。新的明文只会出现一次。',
  tokenAlleWiderrufen: '全部撤销',
  tokenAlleWiderrufenFrage: '要撤销此配置文件的所有访问令牌吗？',
  tokenAlleWiderrufenErklaerung:
    '使用其中任何一个的自动化调用方此后都将无法进入——即使在会话进行中也一样。用它们创建的内容会保留。',
  tokenGezogen: '已撤销 {zahl} 个令牌。',
  tokenLetzteNutzung: '上次使用 {zeit}',
  tokenNieBenutzt: '从未使用',
  tokenVorgaengerBis: '旧密钥有效至 {zeit}',
  tokenOnboarding: '如何设置 MCP 客户端',
  tokenOnboardingSchritt1: '创建一个令牌并复制明文——它只会出现一次。',
  tokenOnboardingSchritt2:
    '把它写入一个只有你能读取的文件（权限 0600）。其他人可读的文件会被“mse-mcp”拒绝。',
  tokenOnboardingSchritt3:
    '启动“mse-mcp”，并把 MSE_TOKEN_DATEI 设为该文件——作为环境变量，绝不要作为参数：那会出现在进程列表中。',
  tokenOnboardingSchritt4:
    '每次调用工具时都会检查令牌：撤销、到期或权限变小，会在下一次调用时生效。没有令牌时，“mse-mcp”与之前一样运行，在所显示的配置文件中不带身份。',
  tokenOnboardingBeispiel:
    '{ "mcpServers": { "mse": { "command": "/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp", "env": { "MSE_TOKEN_DATEI": "/Users/ich/.mse/mcp-token" } } } }',
  tokenOnboardingBeispielTitel: '设置示例',

  // **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
  // (#1351), Punkte 4 bis 7.
  vorgaben: '此配置文件的默认值',
  vorgabenErklaerung:
    '在**此配置文件**中，新记录以什么开头——一种赛制、一款游戏、资料库中的一个模板。与设置不同，默认值指向**资料库中的行**；能选什么，取决于那里有什么。',
  vorgabenamen: {
    format: '新套牌的默认赛制',
    spiel: '新系列的默认游戏',
    vorlage: '新系列的默认模板',
    deck: '偏好套牌',
    lagerort: '分配的存放位置',
  } as Record<string, string>,
  vorgabenOhne: '无默认值',
  vorgabenLeer: '资料库中还没有相关内容。',
  vorgabenVerwaist: '指向空处，因为它已不存在：',

  // **Rechte, Rollen und Sichtbarkeit** — Werkstück 11 von Phase 4.12
  // (#1351).
  freigaben: '谁可以在此配置文件中工作',
  freigabenErklaerung:
    '**读者**只能看到有这个配置文件——看不到其中的内容。**协作者**可以切换进去并在其中工作。**管理者**可以自行授予访问权限。如果这个配置文件是你创建的，它就属于你；那么无论如何都由你管理。',
  freigabenNurLesen: '只有管理此配置文件的人才能授予访问权限。',
  freigabenLeer: '目前除所有者外，没有人可以进入此配置文件。',
  freigabeNutzer: '授予谁',
  freigabeNiemand: '无人',
  freigabeRolle: '作为什么角色',
  freigabeVergeben: '授予',
  freigabeZurueckziehen: '撤销',
  rollennamen: {
    leser: '读者 — 能看到有这个配置文件',
    mitarbeiter: '协作者 — 切换进去并在其中工作',
    verwalter: '管理者 — 自行授予访问权限',
  } as Record<string, string>,

  bild: '头像',
  bildVon: '{name} 的头像',
  bildWaehlen: '选择图片……',
  bildEntfernen: '移除图片',
  ohneBild: '{name} 没有头像',

  nochNicht:
    '各配置文件之间的数据是分开的——前提是没有人用外部工具打开资料库文件。对此，可以在“设置 › 数据与备份”下加密资料库。',

  // **Die vier Kontoarten** — Werkstück 6 (#1351), in einer eigenen Datei.
  konto,
  sitzung,

  // **Der zweite Faktor** — Werkstück 7 (#1351).
  faktor,

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351),
  // Punkte 9, 10 und 11.
  //
  // **Am 2026-09-06 aus `seiteneinstellungen` hierher gezogen**, nach dem
  // Zweitblick über die Phase: die Einstellungen einer Person lagen an zwei
  // Orten, und wer sie suchte, suchte an beiden.

  persoenlich: '我的账户',

  persoenlichErklaerung:
    '这些设置属于**你**，而不是此配置文件——你在自己的每个配置文件中都会带着它们。上面两项会覆盖操作系统的设置，而且**只有一个方向**：只能开启，不能关闭。',

  auskunft: '我的数据',

  auskunftErklaerung:
    '本程序保存的关于你的内容——个人信息、登录方式、第二重验证的**种类**、设置和配置文件名称。**不含密码，也不含密钥**：这两样都不是关于你的数据，而是用来成为你的手段。系列、套牌和收藏不在其中——它们属于配置文件，通过各自的途径导出。',

  auskunftSichern: '保存我的数据……',

  datenschutz: '本应用对你了解多少',

  datenschutzHinweis:
    '记录了哪些信息、存放在哪里、哪些会离开这台电脑——以及明确不会的有哪些。见**帮助 → 本应用对你了解多少**。',
} as const
