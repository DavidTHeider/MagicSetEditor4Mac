// Die Texte der Technikseite über **Dienste und Wege** — welches Zeugnis
// ein Dienst vorzeigt (NET-03, #1707) und wie ein Gegenüber über fremde
// Netze angebunden ist (NET-02, #1706). Aus `seitentechnik.ts` herausgelöst,
// weil die Datei an der Zeilengrenze aus §2 stand.

export const dienstzeugnis = {
  eigenes: '由 {aussteller} 签发的自有证书，有效期至 {bis}',
  identitaet: '基于此个人资料身份的证书——连接的人会核对指纹。',
}

export const fernwege = {
  titel: '对方的连接方式',
  keine: '还没有人使用过这份邀请。',
  direkt: '直连，经由 {adresse}',
  relais: '经由中继 {adresse}——尚无直连路径',
  aufbau: '已知，但目前还无法通过任何路径访问',
  unbestimmt: '已连接：{anzahl} —— 路径尚不确定。',
}

// **Command line and tool server with one click** (OPS-09, #1907).
export const schnittstellen = {
  titel: '设置命令行和工具服务器',
  satz: '一个按钮会写入这台 Mac 上的程序找到命令行或工具服务器所需的内容，另一个按钮则恰好把它撤回。事先会显示涉及哪个文件、哪几行；他人已有的条目保持原样。',
  namen: {
    befehle: '~/.local/bin 中指向 mse 和 mse-mcp 的链接',
    befehlszeile: '登录 shell 的 PATH 中的命令行',
    codex: '供 Codex 使用的工具服务器',
    claudedesktop: 'Claude Desktop',
    claudecode: 'Claude Code',
    dienst: '作为 LaunchAgent 的脚本服务',
  },
  staende: {
    offen: '未设置',
    eingerichtet: '已设置',
    veraltet: '已过时',
    fremd: '他人的条目',
    nichtGefunden: '这台 Mac 上没有',
    unnoetig: '无需设置',
  },
  staendeSaetze: {
    veraltet: '应用程序现在在别处；“设置”可修复它。',
    fremd: '已有他人的条目——保持原样。',
    unnoetig: '已在 PATH 中。',
  },
  zeilen: '将添加这些行',
  vorschau: {
    json: '这条条目会写入文件',
    befehl: '会运行这条命令',
    agent: '会创建这个文件，并由这条命令加载',
  },
  nurEintrag:
    '{name} 此后已被修改——只移除了本应用自己的条目，其余内容保持你留下的样子。',
  einrichten: '设置',
  widerrufen: '撤销',
  alle: {
    titel: '一次全部',
    satz: '设置此 Mac 上存在且尚未设置的每个目标，或撤销所有已设置的内容。他人的条目保持不变。',
    vorschau: '将写入的内容（{anzahl} 个目标）',
    vorschauLeer: '没有缺少的内容：此处已有的均已设置。',
    einrichten: '全部设置',
    widerrufen: '全部撤销',
    ergebnis: '结果',
    ausgaenge: {
      eingerichtet: '已设置',
      widerrufen: '已撤销',
      fremd: '他人的条目，未更改',
      unveraendert: '未更改',
      gescheitert: '失败',
      abgebrochen: '已中止',
    },
  },
  lauschwahl: {
    titel: '服务的监听位置',
    satz: '适用于脚本服务（在窗口中启动以及作为 LaunchAgent）和通过 HTTP 的工具服务器，自其下次启动起生效。',
    wahl: '可通过以下方式访问',
    schleife: '仅此 Mac（127.0.0.1）',
    auswahl: '所选地址',
    alle: '所有接口（0.0.0.0）',
    alleSatz: '每个接口，包括以后新增的接口。',
    aussen:
      '所选地址会离开此 Mac：此时服务要求 TLS 和访问令牌，macOS 可能会请求本地网络权限。',
    inaktiv: '已关闭',
    ohneAdresse: '无地址',
    liste: '此 Mac 的接口',
    arten: {
      schleife: '环回',
      netz: '网络',
      tunnel: '隧道（VPN）',
      bruecke: '网桥',
      direkt: '直连无线',
    },
    speichern: '应用选择',
    gespeichert: '已应用：在下次启动服务时生效。',
    keineGewaehlt: '请至少选择一个地址。',
    platzhalter: '同“服务的监听位置”中的设置',
    weitere: '还监听于',
    ohne: '缺少 {adresse}：{grund}',
  },
}
