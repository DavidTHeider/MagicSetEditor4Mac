// **The network settings** — work item N5 of phase 5.4
// (#1571).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungenzugaenge.ts`
// daneben. **Ein Katalog bleibt es trotzdem**: die drei Teile werden dort
// hineingestreut, und der Vertragstest gegen die Go-Seite sieht einen.

// netznamen sind die Beschriftungen der sechs Einstellungen.
export const netznamen = {
  'netz.proxy.http': 'http 代理',
  'netz.proxy.https': 'https 代理',
  'netz.proxy.aus': '不经代理直接访问',
  'netz.frist.verbindung': '超时：建立连接（秒）',
  'netz.frist.kopf': '超时：响应头（秒）',
  'netz.frist.stillstand': '超时：无进展（秒）',
  'netz.schnittstelle': '本地网络搜索所用的网络接口',
  'netz.lauschen': '自有服务的监听位置',
  'dienst.tls.zertifikat': '服务自有证书',
  'dienst.tls.schluessel': '自有证书的密钥',
  'dienst.tls.kette': '中间证书',
  'dienst.tls.anker': '证书的信任锚',
  'grenzen.kontingente': '服务的配额',
}

// netzhilfe ist der erklärende Satz je Einstellung.
export const netzhilfe = {
  'netz.proxy.http':
    '`http://` 地址所经过的代理。**留空表示采用环境中的 `HTTP_PROXY`**——此处填写的内容会覆盖它。需带协议前缀，例如 `http://proxy.company:3128`。**不含密码**——密码会以未加密的形式存放在磁盘上；需要登录的代理则通过环境中的 `HTTPS_PROXY` 提供。',
  'netz.proxy.https':
    '`https://` 地址的同类设置，对应 `HTTPS_PROXY`。**单独设置**，因为许多网络只为两者之一配置了代理。',
  'netz.proxy.aus':
    '**不**经过代理的地址——主机名和地址范围，以逗号分隔，与 `NO_PROXY` 相同。本机回环地址始终不经代理；位于自己网络中**另一台**电脑上的 Ollama 应填在此处。',
  'netz.frist.verbindung':
    '建立连接（含 TLS）最多可花费的时间。留空表示 15 秒。',
  'netz.frist.kopf': '服务器发回响应头的最长等待时间。留空表示 30 秒。',
  'netz.frist.stillstand':
    '正在进行的响应可在多长时间内没有进展。**需要调大的是这一项**——当经由慢速线路的大文件下载中断时，而不是上面两项。留空表示 60 秒。',
  'netz.schnittstelle':
    '本地网络搜索的查询和对外通告经由哪个网络接口发出，例如 `en0`。**留空表示自动**：使用所有适用且支持 IPv4 和组播的接口。当前不存在的名称按留空处理，并作为发现项记入日志。',
  'netz.lauschen':
    'REST 和 MCP 监听的地址，以逗号分隔。**留空表示：仅此 Mac**（`127.0.0.1`）。`0.0.0.0` 表示所有接口，须单独使用。**若某地址离开本机，则需要 TLS 和令牌**；在 “无窗口工具” 中选择更方便。在下次启动服务时生效。',
  'dienst.tls.zertifikat':
    'REST 与 MCP 服务在本机回环地址之外出示的证书，PEM 格式。**留空表示：基于此个人资料身份的证书。**若使用你自己的认证机构，客户端会按名称和证书链校验，而不是按指纹。下次启动服务时生效。',
  'dienst.tls.schluessel':
    '证书对应的私钥，PEM 格式。**仅你本人可读（0600）**，否则服务无法启动。同时替换两个文件后，无需重启，下一次握手时即生效。',
  'dienst.tls.kette': '若证书中未包含中间证书，可在此提供。留空表示：无。',
  'dienst.tls.anker':
    '你所用认证机构的根证书，启动前会据此校验证书链。留空表示：使用系统的信任锚。',
  'grenzen.kontingente':
    '机器访问每分钟可以通过 REST 和 MCP 做多少事——按每个访问计，也按全部合计。**关闭适用于受控环境**，即谁也不会挤占别人的份额，例如负载测试。访问控制、TLS、每类工作的并发名额、请求体和响应的上限、时限以及防猜测锁定都会保留。下次启动服务时生效。',
}

// netzbereich ist die Überschrift des Bereichs samt Hinweis.
export const netzbereich = {
  netz: '本程序如何联网',
  netzHinweis:
    '这些设置对**所有**服务同时生效——包括命令行、MCP 服务器和后台服务。留空表示采用环境中的设置。当前生效的设置及其来源，请见“网络与可达性”。',
}

// netzwerte sind die Namen der Werte — QA-08.AK-03 (#1739).
export const netzwerte = {
  'grenzen.kontingente.an': '开',
  'grenzen.kontingente.aus': '关',
}
