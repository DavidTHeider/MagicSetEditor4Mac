// The texts for **database maintenance** — back up and restore
// (#940).
//
// Split out of `seiten.ts` the same way the card page was: one catalogue may
// live in several files as long as it arrives as **one** (§9).

export const bestandspflege = {
  titel: '资料库',
  hinweis:
    '此应用保存的一切：系列、卡牌、字段值、历史记录，以及你放入的图像。一份备份把两者放在同一个文件里——资料库和图像。',
  sichern: '备份资料库……',
  waehlen: '恢复备份……',
  rueckfrage: '要恢复此备份吗？',
  datei: '文件',
  angelegtAm: '资料库创建于',
  sets: '系列',
  karten: '卡牌',
  unbekannt: '—',
  warnung:
    '当前资料库将被替换。它不会丢失：会作为 bestand-vor-rueckspielung.db 保留在旁边——直到下一次恢复。',
  bilderLage: '已放入的图像',
  bilderErsetzt:
    '此备份包含图像。它们将取代现有图像；原有图像会作为 beiwerk-vorher 保留在旁边。',
  bilderGeleert:
    '此备份不含图像——它来自一个尚未放入任何图像的状态。现有图像将被移除；它们会作为 beiwerk-vorher 保留在旁边。',
  bilderUnberuehrt:
    '此备份是 0.1.0 或 0.2.0 的单个数据库，不含图像。现有图像保持不变——在另一台电脑上它们就会完全缺失。',
  bestaetigen: '恢复',
  abbrechen: '取消',

  // **The password of a backup** (OPS-06, #1723).
  kennwort: '密码（可选）',
  kennwortWiederholen: '再次输入密码',
  kennwortHinweis:
    '设置密码后，备份会被加密，也可以在另一台电脑上打开。忘记的密码无法找回——备份也就此丢失。',
  kennwortVerschieden: '两次输入的密码不一致。',
  klartext: '不设密码备份——文件将是明文',
  klartextHinweis:
    '资料库已加密。不设密码的话，备份会不受保护地存放在它所写入的介质上。',
  gesperrt: '此备份已加密。',
  gesperrtKennwort: '备份的密码',
  entsperren: '打开',
  verschluesselt: '已加密',
  ja: '是',
  nein: '否',
  paketeEnthalten: '其中的自有包',
  paketeFehlen: '需重新获取',
  keine: '无',
  zurueckgespielt: '已恢复。原来的资料库现为 {beiseite}。',
  rueckkehrUnveraendert: '原来的资料库保持不变。',
  rueckkehrUnvollstaendig:
    '撤回也失败了：原资料库的部分内容以 bestand-vor-rueckspielung.db 和 beiwerk-vorher 的名称留在旁边。请恢复一个备份，或手动将它们放回原处。',
  wiederherstellungTitel: '无法打开资料库',
  wiederherstellungHinweis:
    '只有当一个可用的资料库文件就位后，应用才会重新工作。恢复备份是通往这一步的途径——它会替换当前出问题的文件。',
  wiederherstellungGrund: '收到的报告',
  wiederherstellungHinweisHeil:
    '资料库文件本身很可能没有问题——问题出在它周围。请按下方说明处理后重试；无需替换任何内容。',
  erneutVersuchen: '重试',
  weiterGescheitert: '已重试——资料库仍然无法打开。',

  wegZurueckTitel: '返回的途径',
  wegZurueck:
    '要返回，请使用你资料库所在文件夹中的“{datei}”——{sets} 个系列、{karten} 张卡牌、架构步骤 {stand}。请在下方选择它。',

  chiffreTitel: '资料库加密',
  chiffreAnHinweis:
    '资料库以加密形式存放在磁盘上。没有密钥就无法打开——连你自己也不行。',
  chiffreAusHinweis:
    '资料库以明文存放在磁盘上。拥有该文件的人可以读取其中的一切：姓名、出生日期、地址、照片和登录方式。',
  chiffreAnschalten: '开启加密……',
  chiffreAusschalten: '关闭加密……',
  chiffreFrageAn: '现在加密资料库吗？',
  chiffreFrageAus: '重新关闭加密吗？',
  chiffreWarnungAn:
    '密钥会在此时生成，并存入这台 Mac 的钥匙串。它无法重新计算，也不存在于其他任何地方：密钥一旦丢失，数据也就没了。请先做一份备份。',
  chiffreWarnungAus: '之后一切将重新以明文存放，密钥也会从钥匙串中移除。',
  chiffreGrenzen:
    '加密**不能**防止篡改。比较同一文件的多个副本的人，可以看出哪些部分发生了变化。**备份使用它自己的密码**——在上方备份时设置；不设密码则是明文，那就需要明确选择“不设密码备份”。',
  chiffreLaeuft: '正在重写资料库……',

  platzUeberschrift: '磁盘空间',
  platzErklaerung:
    '本应用存放了什么以及占多大空间——整个文件夹，因为不属于任何项目的内容都计入**其他**。**迁移前的备份**会在每次重建资料库的启动时生成；保留下来的是**返回的途径**——此版本能打开的最新一份备份。**日志**只保留正在使用的文件。',
  platzDateien: '{anzahl} 个文件',
  platzFreigeben: '清除较旧的',
  platzFreigebenFuer: '{posten}：清除较旧的',
  platzFreigebenFrage: '{posten}：确定要清除较旧的吗？',
  platzFreigebenErklaerung:
    '较旧的文件会被删除，且无法找回。作为返回途径所需的文件会保留。',
  platzLoeschenJa: '清除',
  platzNamen: {
    bestand: '资料库',
    bilder: '嵌入的图像',
    pakete: '已下载的包',
    vorsicherungen: '迁移前的备份',
    protokoll: '日志',
    sonstiges: '应用文件夹中的其他文件',
  },
  alter: {
    ueberschrift: '上次备份',
    unbekannt: '未知',
    unbekanntSatz:
      '此版本尚未备份过数据。未知就是未知——除非文件已被读回，这里绝不会显示“今天”。',
    heute: '今天',
    gestern: '1天前',
    vorTagen: '{tage}天前',
    alt: '{alter} · 超过14天',
    unstimmig: '时间晚于时钟',
    satz: '上次备份于{zeitpunkt}，通过{flaeche}写入{datei}，已读回并检查。',
    satzPlan: '上次备份于{zeitpunkt}，按计划写入{datei}，已读回并检查。',
    unstimmigSatz:
      '上次备份的时间为{zeitpunkt}，晚于这台 Mac 的时钟。无法判断其时长；为保险起见请备份。',
    flaechen: {
      fenster: '窗口',
      cli: '命令行',
      mcp: 'MCP',
      rest: '服务',
    },
    planSchalter: '每周备份到一个文件夹，保留七份',
    planHilfe:
      '仅在选择文件夹后启用。只删除自己的备份，绝不删除其他文件。窗口打开时进行备份。',
    planOrdner: '文件夹：{ordner}',
    planZuletzt: '按计划的上次备份：{zeitpunkt}。',
    planNie: '尚未按计划备份。',
    planFehler: '计划上次未能备份：{grund}',
  },
} as const
