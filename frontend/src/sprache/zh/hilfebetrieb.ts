// **Operation and trust** — OPS-07
// (#1724).
export const betrieb = {
  ueberschrift: '运行与信任',
  einleitung:
    '安装了什么、存放在哪里、哪些已验证、哪些没有——均读取自本应用程序自身。',
  fassungName: '版本',
  fassungSatz: '{fassung}，构建自 {stand}，即 {gebaut} 的状态。',
  fassungVeraendert: '此次构建包含尚未提交的更改。',
  datenName: '数据',
  datenSatz: '已应用 {zahl} 次迁移，最近一次是 {letzte}。',
  datenOhne: '未打开任何资料库。',
  signaturName: '签名',
  signaturSatz:
    '采用临时（ad-hoc）签名，**未经公证**。首次启动时 macOS 会拒绝该应用程序；请在“系统设置 › 隐私与安全性”中点按**仍要打开**，允许一次。',
  aktualisierungName: '更新',
  aktualisierung: {
    aus: '关闭——本应用程序不会查询新版本。',
    melden: '它会报告来自 GitHub 的新版本，但不下载任何内容。',
    holen: '它会从 GitHub 获取新版本，并在替换前询问。',
  },
  orteName: '所有内容的位置',
  orte: {
    bestand: '资料库',
    pakete: '模板包',
    bilder: '插入的图像',
    zwischenspeicher: '缓存',
    protokoll: '日志',
    leistungsprofile: '性能剖析',
  },
  lizenzenName: '许可证',
  lizenzenSatz:
    '源代码采用 GNU 通用公共许可证（GNU General Public License）第 2.0 版或更高版本授权。第三方程序部分及其许可证列在源代码旁的 `DRITTANBIETER.md` 中；模板包自带其自己的条款，AI 模型则遵循其提供方的条款。',
  empfaengerName: '数据流向哪里',
  empfaengerSatz:
    '只会发往你启用的服务——有哪些服务、发出什么，请见**帮助 › 本应用程序了解你的哪些信息**。不会发送使用数据。',
  grenzenName: '安全边界',
  grenzenSatz:
    '在没有另行设置时，服务和工具服务器只监听 `127.0.0.1`，并要求令牌。若在**设置 › 诊断**下为其中一个指定了网络地址，它只会通过 TLS 运行，并且页面会说明它会离开这台电脑；工具服务器通过 stdio 时使用它自己的数据流通信。资料库可以加密，位于**设置 › 数据与备份**。',
  kontaktName: '出现问题时',
  kontaktSatz:
    '在**设置 › 诊断**下可以找到诊断信息、诊断包和本进程的各项数字。错误报告请提交到 github.com/DavidTHeider/MagicSetEditor4Mac——不会自行发出任何内容。',
  pruefstandName: '已验证的内容',
  pruefstandSatz:
    '“已验证”指：有一次运行证明了它——一次探测，其证据连同提交记录保存在源代码旁；或者每条流水线都会运行的一个测试；只有通过它们的内容才会发布。“已有”指：已经构建，但从未针对原版或该平台运行过。',
  spalteBereich: '领域',
  spalteGegenstand: '项目',
  spalteStand: '状态',
  spalteBeleg: '依据',
  belegIssue: '议题 #{nummer}',
  bereiche: {
    plattform: '平台',
    sprache: '语言',
    barrierefreiheit: '无障碍',
    regeln: '规则',
    integration: '外部服务',
  },
  staende: {
    geprueft: '已验证',
    vorhanden: '已有',
    beschrieben: '已描述',
    blockiert: '等待人工处理',
  },
  gegenstaende: {
    'macOS 15.3 bis 26': 'macOS 15.3 至 26',
    'macOS 15.2 auf Apple Silicon': 'Apple 芯片上的 macOS 15.2',
    'macOS 27 auf Apple Silicon': 'Apple 芯片上的 macOS 27',
    'Tastatur und Vorleser im WebKit': 'WebKit 中的键盘与屏幕阅读器',
  },
}
