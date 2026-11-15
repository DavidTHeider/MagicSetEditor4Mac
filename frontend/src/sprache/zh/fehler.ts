export const fehler = {
  unbekannt:
    '发生了意外错误。具体原因记录在日志中——在“设置 › 诊断”下可以查看日志末尾并将其存储。',
  unbekanntMitVorgang:
    '发生了意外错误。具体原因记录在日志中——在“设置 › 诊断”下可以查看日志末尾并将其存储。请在其中搜索 {vorgang}：此操作的所有行都带有这个标识。',
  nichtGefunden: '未找到要查找的条目。也许被别人删除了——请重新载入列表。',
  ohnePeeridentitaet:
    '此档案还没有身份。请在“此档案如何证明身份”框中创建一个——没有身份，就无法向他人证明自己是谁。',
  peerNichtErreichbar:
    '目前无法连接到对方。对方是否在监听，地址是否正确？否则请稍后再试。',
  peerOhneAdresse:
    '尚未为此对方登记访问地址。请在其联系人的“访问地址和备注”中填写——没有它，本程序只能在同一网络中找到对方。',
  ohneNetzerlaubnis:
    '没有任何内容发送到本地网络。如果 macOS 刚刚请求了权限，请允许后再试一次。从这里无法确定具体原因：要么是缺少权限（请在“系统设置 → 隐私与安全性 → 本地网络”中授予），要么是网络不允许组播通过，例如处于 VPN 之后。',
  ohneNetzroute:
    '当前没有通往本地网络的路径。是否有网络接口已连接？可在“设置 → 连接”中选择指定的接口。',
  schluesselbundAblage:
    '无法将此个人资料的密钥存入钥匙串。登录钥匙串是否已解锁？请在解锁后再试一次。',
  einladungZuNeu:
    '此邀请来自本程序的较新版本。请向对方索取其较旧版本生成的邀请——或者你自己获取较新的版本。',
  fernhorchenLaeuft:
    '此档案已可通过其他网络访问。如果想要新的邀请，请先将其停止——旧的邀请随之失效。',
  anmeldungAbgelaufen: '登录未在十分钟内返回。在浏览器中准备好后，请重新开始。',
  umschlagFassung:
    '此信封来自本程序的另一个版本。请向对方索取其较旧版本生成的信封——或者你自己获取较新的版本。',
  umschlagUnterschrift:
    '此信封的签名不成立。它在传输途中被更改，或者并非来自其声称的发送者——请不要接受，并向对方核实。',
  umschlagInhalt:
    '此信封中有一部分与其清单不符。请对方重新发送——传输途中有内容损坏或被更改。',
  fremderNutzer:
    '只有该用户本人才能这样做。请先切换到该用户——此时应用程序会要求第二重验证。',
  nichtFreigegeben:
    '对方没有向你公开任何可交易的内容。请对方向你公开视图，或通过文件交易。',
  zeileNichtVerbuchbar:
    '第 {zeile} 行在接收方无法入账：缺少状态或语言，或者某项取值不存在。请删除该行，并带上状态、语言和工艺重新添加。',
  zusageUnpassend:
    '此实物不符合你给出的任何一行：状态、语言或工艺不同。请选择相符的一件，或将其添加到收藏。',
  erlaubnis: '你没有执行此操作所需的角色。管理此档案的人可以授予你。',
  unbekannteHerkunftsart: '程序不认识这种来源。请选择“取自档案”或“已导入”。',
  // The command centre (#1423).
  befehlBestaetigen: '此命令会更改一些难以撤销的内容——请阅读预览，然后确认。',
  befehlArgumente:
    '这些参数不是 JSON 对象。请输入用花括号括起的对象——每个参数一个字段。',
  herkunftOhneProfil:
    '取自档案的来源要写明该档案，导入则不写——请填写一个，或者留空。',
  nurLesend:
    '在此档案中你只能读取。它属于别人，你在其中的角色是“阅读者”。请切换到你自己的档案再进行工作，或请管理此档案的人授予你“协作者”角色。',
  bestandUnbrauchbar:
    '此文件已损坏或不是 MagicSetEditor4Mac 的资料库——无法打开。请恢复一份备份；最近的一份就在你当时存放它的位置。',
  bestandNeuereFassung:
    '此资料库来自 MagicSetEditor4Mac 的较新版本。文件本身没有问题——请不要替换它。请获取较新的版本，或恢复更早的备份。',
  bestandAeltereFassung:
    '此资料库来自 0.7.0 之前的版本，当前版本已无法再升级它。文件本身没有问题——请不要替换它。请先用 MagicSetEditor4Mac 0.7.0 打开一次；之后当前版本也能打开它。',
  bestandNurLesbar:
    '资料库可以读取，但无法写入——文件或其所在文件夹缺少写入权限。文件本身没有问题，请不要替换它。请在访达的“显示简介”中为它授予写入权限，然后重试。',
  bestandEA:
    '读取资料库时磁盘报告了错误。文件是否完好尚不清楚，请不要立即替换它。请先检查磁盘（例如使用磁盘工具），然后重试。',
  kaumPlatz:
    '资料库无法打开或写入，而且磁盘几乎没有可用空间。文件很可能完好——请不要替换它。请腾出空间后再试一次；如果仍然不行，请用“磁盘工具”等检查磁盘和写入权限。',
  startUngeklaert:
    '无法打开资料库，原因不明——原因显示在下方。请重试。只有在确定此文件已无法再打开时，才恢复备份。',
  sicherungVerworfen:
    '备份未通过检查，已被丢弃——没有保存任何内容。请重试，或备份到其他位置。',
  sicherungUnvermerkt:
    '备份已写入并通过检查——只是无法记录其时间，因此“上次备份”的时长仍显示之前那次。请保留该文件，并在数据空闲后再备份一次。',
  schluesselFehlt:
    '此资料库已加密，而它的密钥不在这台 Mac 的钥匙串中。文件本身没有问题——请不要替换它。密钥是在开启加密时生成的，无法重新计算；没有它，就再也无法访问这些数据。',
  belegt: '资料库当前正忙——可能是另一个窗口或某条命令正在写入。请稍后再试。',
  bestandWirdGetauscht:
    '资料库正在被替换——恢复备份、切换加密或重新开始。请稍候片刻，然后再试一次。',
  anwendungEndet:
    'MagicSetEditor4Mac 正在退出——此步骤未被执行。请重新启动应用以继续。',
  vorgangAbgebrochen: '操作在完成前已被中止。如仍需要，请重新开始。',
  konflikt: '该条目在此期间已被更改。请重新载入。',
  konfliktgrund: '操作未成功：{grund}。再次尝试之前，请先检查当前状态。',
  doppelt: '已经存在包含这些信息的条目。请更改其中一项，或编辑已有的条目。',
  doppelterName: '此处已有这个名称——请另选一个。',
  doppeltesExemplar:
    '收藏中已有这个印刷版本——品相、语言和表面处理都相同。请在那里增加数量，或更改其中一项信息。',
  gesperrt:
    '本程序无权访问此位置——它已被锁定。请另选一个位置，或在文件权限中，或在“系统设置 › 隐私与安全性 › 文件和文件夹”下授予访问权限，然后重试。',
  voll: '磁盘上已没有空间。请腾出空间后重试。',
  netzName:
    '无法解析名称“{wirt}”。请检查互联网连接，如已设置代理，也请检查“设置 › 连接”中的代理。连接报告的内容在“设置 › 诊断”中的 {vorgang} 下。',
  netzNichtErreichbar:
    '“{wirt}”上没有接受连接的服务。该服务是否在运行，地址是否正确？否则请稍后重试。连接报告的内容在“设置 › 诊断”中的 {vorgang} 下。',
  netzKeineAntwort:
    '“{wirt}”在 {sekunden} 秒内没有响应。请稍后重试。连接报告的内容在“设置 › 诊断”中的 {vorgang} 下。',
  netzAntwortStockt:
    '来自“{wirt}”的响应停滞了 {sekunden} 秒。请检查连接后重试。连接报告的内容在“设置 › 诊断”中的 {vorgang} 下。',
  netzAbgerissen:
    '与“{wirt}”的连接在响应完整到达之前中断了。请重试；如果反复出现，问题在于该服务或中间的代理。连接报告的内容在“设置 › 诊断”中的 {vorgang} 下。',
  netzProxyAnmeldung:
    '代理“{wirt}”要求登录。请通过环境变量 HTTPS_PROXY 提供其凭据——“设置 › 连接”中的代理有意不接受密码。连接报告的内容在“设置 › 诊断”中的 {vorgang} 下。',
  gedrosselt:
    '对方报告请求过多。请稍候再试。连接报告的内容在“设置 › 诊断”中的 {vorgang} 下。',
  gedrosseltBis:
    '“{wirt}”目前不接受请求。请在 {uhrzeit} 之后重试——在此之前本程序不会再次请求。连接报告的内容在“设置 › 诊断”中的 {vorgang} 下。',
  kontingent: '短时间内调用过多。请等待 {sekunden} 秒后重试。',
  prozessVeraltet:
    '此进程来自比资料库更旧的版本，正在退出——未写入任何内容。已安装的服务会以其安装时指定的程序自动重新启动；其他进程请用新版本重新启动，然后重试。',
  bestandVomDienstGehalten:
    '脚本服务正在占用资料库。请在 设置 › 诊断 ›「设置命令行和工具服务器」中撤销它，然后再次恢复。',
  falschesSpiel: '该套牌所属的游戏与此赛事不同。请选择同一游戏的套牌。',
  turnierformatSpiel:
    '该赛制所属的游戏与此赛事不同。请选择同一游戏的赛制——或不选。',
  paketGeaendert:
    '此软件包与安装时的状态不同。请确认替换——所做的更改会作为上一版本保留。',
  symbolZuVieleTeile:
    '此符号有 {anzahl} 个部分，最多允许 {grenze} 个。请合并部分，或选择更简单的图形。',
  symbolZuVielePunkte:
    '此符号有 {anzahl} 个点，最多允许 {grenze} 个。请简化形状，或选择更简单的图形。',
  symbolZuTief:
    '此符号的分组嵌套了 {anzahl} 层，最多允许 {grenze} 层。请选择层级更少的图形，或取消内层分组。',
  symbolZuGross:
    '此文件大于 {grenze}，因此不是符号文件。请选择 .mse-symbol 文件。',
  symbolGespeichertZuGross:
    '此符号保存后将超过 {grenze}，超出符号文件能容纳的大小。请简化形状或减少点数。',
  bildZuGross: '此图像太大，无法解压。请选择较小的图像，或先将其缩小。',
  setOhneSymbol: '此系列还没有符号。请先绘制一个并存储——之后才能添加。',
  // **Die eigenen Übersetzungen** — LOC-02 (#1688).
  sprachkennungUngueltig:
    '“{kennung}”不是符合 BCP 47 的语言标识符。请使用类似 pt-BR 或 zh-Hant 的标识符。',
  sprachkennungEingebaut:
    '“{kennung}”是内置语言。请选择自己的标识符，例如 {kennung}-x-meine。',
  platzhalterAbweichend:
    '此翻译使用的占位符与德文原句不同。请准确填入这些：{erwartet}。',
  auszeichnungUnsicher:
    '此翻译包含非文本内容，或包含未闭合的标记。请移除标签，并闭合 ** 和 `。',
  sprachpaketNeuer: '此翻译来自更新版本的程序（{format}）。请获取新版本。',
  sprachpaketOhneLizenz:
    '没有许可证，翻译无法对外发布。请填写一个，例如 CC0-1.0。',
  bereichAbgeschaltet:
    '此区域已关闭（{schalter}）。可在 设置 › 区域与功能 中重新开启——没有丢失任何内容。',
  dienstAus: '服务 {dienst} 已关闭。如需查询，可在 设置 › 服务 中开启。',
  fremdeEigeneKarten:
    '发送者的自有卡牌没有一并到达——它们的标识只在对方的资料库中有效：{karten}。如果需要，请打开对方的系列，把它们添加到这里。',
  nichtGelegtePlaetze:
    '这些条目无法在此添加：{plaetze}。牌组的其余部分已接收；请检查并手动添加。',
  fremdeKartenUndNichtGelegte:
    '发送者的自有卡牌没有一并到达——它们的标识只在对方的资料库中有效：{karten}。另外，这些条目无法在此添加：{plaetze}。请检查两者并手动添加。',
  symbolVerwendet:
    '这些系列带有该符号：{sets}。请确认删除——它们会保留各自的副本。',
  symbolOhneVorige:
    '此符号没有上一版本。请先存储一个新的——这样当前这个才能恢复。',
  turnierAbgeschlossen: '此赛事已结束。若要更改，请选择“重新开放”。',
  sitzGegangen: '此座位已退出对局，无法再行动。请选择仍在对局中的座位。',
  fremderWert:
    '在此期间，其他人修改了这个字段。请查看新的值，并决定是否保留你的草稿。',
  vorschlagNurGegenseite:
    '只有对方可以修改未决报价。请等待对方答复，或取消交易后重新报价。',
  turnierNichtFertig:
    '请先填写所有结果：赛事至少有一轮，且每组配对都有结果时，才能结束。',
  turniernameLeer: '赛事需要一个名称——请输入。',
  turnierspielLeer:
    '赛事必须属于某个游戏——否则没有可用于检查的赛制。请选择一个。',
  validierungsgrund: '操作未成功：{grund}。请更正输入后重试。',

  validierung: '输入无效。字段处的提示说明了原因——请在那里更正。',

  // **The rule itself is the remedy** (#1350).
  // **The password of a backup** (OPS-06, #1723).
  sicherungskennwortKurz:
    '备份密码至少要有十二个字符。请选择更长的——由四个词组成的一句话既好又易记。',
  sicherungskennwortFehlt: '此备份已加密。请输入其密码以查看或恢复它。',
  sicherungskennwortFalsch: '密码与此备份不符。请检查后重新输入。',
  // **Eine MTG-JSON-Datei, die sich nicht lesen lässt** (G01, #2248).
  mtgjsonKeinJSON:
    '{datei} 不是可读取的 JSON——文件在第 {stelle} 字节处中断或格式错误。请选择其他文件。',
  mtgjsonKeinSet:
    '{datei} 是 JSON，但不包含 MTG JSON 格式的系列（在第 {stelle} 字节处不符）。请选择包含系列的文件，或整个文件夹。',
  sicherungImKlartext:
    '资料库已加密——没有密码的话，备份将是明文。请在“设置 › 数据与备份”中设置密码，或在那里明确选择“不设密码备份”。',

  planVerschluesselt:
    '资料库已加密，而计划在没有密码的情况下备份——因此计划不会备份它。请使用密码手动备份。',

  paketnameUngueltig:
    '软件包名称由字母、数字、连字符和下划线组成——它会成为目录名。请从这些字符中选择。',
  ordnerangabe:
    '“{wert}”不是完整的位置。请输入以 / 或 ~/ 开头、中间不含 .. 的位置。',
  ordnerFehlt: '无法访问文件夹“{ordner}”。请连接该驱动器，或选择其他文件夹。',
  ordnerIstDatei:
    '“{ordner}”是文件，不是文件夹。请选择它所在的文件夹，或另选一个。',
  ordnerGesperrt:
    '本程序无权写入“{ordner}”。请选择其他文件夹，或授予对此文件夹的访问权限。',
  ordnerVoll:
    '“{ordner}”所在的驱动器只剩 {frei} 可用。请腾出空间，或选择其他文件夹。',
  karteInDeck:
    '这张卡牌用在 {anzahl} 副套牌中——{decks}。删除时也会将其从那些套牌中移除；如果你就是这个意思，请确认。',
  pflichtfeld: '此字段不能为空——请输入内容。',
  nichtEditierbar: '此字段无法编辑——模板只是显示它。请更改生成它的那个值。',
  berechnetesFeld:
    '此字段由其他字段生成，无法对多张卡牌统一设置。其组成部分单独列出——请选择其中之一。',
  keinSetMitStil:
    '没有已打开的系列使用此样式表。请打开使用它的系列，或更改已打开系列的样式表。',
  vorlageFehlt:
    '没有为这张卡牌安装模板。请在“软件包”页面查看该系列的游戏和样式表是否存在——没有它们，卡牌就没有字段和区域。',
  sicherungNeuereFassung:
    '此备份来自 MagicSetEditor4Mac 的较新版本。它包含 {unbekannt}，而当前版本最高支持到 {stand}。请获取较新的版本，或选择较旧的备份。',
  sicherungAeltereFassung:
    '此备份来自 0.7.0 之前的版本——它缺少直到 {grundlinie} 的架构步骤，当前版本已无法再升级它。请用 MagicSetEditor4Mac 0.7.0 恢复它；之后当前版本即可打开该资料库。',
  sicherungUnbrauchbar:
    '此文件不是 MagicSetEditor4Mac 的备份。请选择由本应用程序写入的备份。',
  // The parser's own sentence is German for now — see issue #1178.
  skriptFehler: '无法对该表达式求值：{meldung} 请在所指位置更正。',

  // **The missing second factor** — piece 7 of phase 4.12 (#1351).
  // **The local session** — ID-02 (#1701).
  kennwortFalsch: '密码不正确。请重新输入，或通过恢复页设置新密码。',
  kennwortLaenge:
    '密码至少 {mindestens} 个、最多 {hoechstens} 个字符。请选择符合该长度的密码。',
  sperreOhneRueckweg:
    '没有密码，也没有本机解锁，就无法解除锁定。请先设置密码。',
  lokaleEntsperrungFehlt:
    '这台 Mac 上尚未为该用户设置本机解锁。请改为输入密码。',
  codeFalsch:
    '此验证码不正确——或者已经用过。请核对恢复页上的验证码——每个只能使用一次。',
  zweiterFaktor:
    '此用户要求第二重验证。请输入验证器应用中的验证码，或使用 Touch ID 确认。',
  faktorGesperrt: '错误的验证码过多。请稍候再试。',
  sicherungFremdeFassung:
    '此备份与当前版本的 MagicSetEditor4Mac 不匹配。它标明的是 {dort}，而本应用程序是 {hier}。请用写入它的版本来恢复。',

  // **Die Sätze des Zerteilers und des Auswerters** —
  // #1178.
  //
  // **`durchgereicht` ist die Ausnahme, und sie hat einen Grund**: die 48
  // Sätze der Standardbibliothek nennen Funktions- und Argumentnamen. Ein
  // Eintrag je Funktion wäre eine zweite Abschrift der Bibliothek in zwei
  // Sprachen; dort steht der deutsche Satz, wie er kommt.
  skript: {
    fremdesZeichen: '字符 {0} 不属于该语言',
    einzelnesGleich: '单独的 `=`；赋值用 `:=`，比较用 `==`',
    offeneZeichenkette: '字符串缺少结束引号',
    zuTiefGeschachtelt: '表达式嵌套超过 {0} 层',
    keinSkript: '不存在名为 {0} 的脚本',
    skriptkreis: '这些字段循环互相读取：{0}',
    keinVerweis: '`+` 无法连接指向另一个脚本的引用',
    keineZahl: '{0} 不是数字',
    formUnbekannt: '无法对形式 {0} 求值',
    nameUnbekannt: '此处不认识名称 {0}',
    keineFunktion: '此处调用了 {0}，只有函数才能被调用',
    ohneRumpf: '此函数没有函数体',
    nichtWahrheitswert: '`not` 需要真值，此处是 {0}',
    vorzeichenZahl: '符号需要作用于数字，此处是 {0}',
    zaehlschleifeZahl: '计数循环的 {0} 需要是数字，此处是 {1}',
    rechnenLinks: '`{0}` 对数字进行运算，左侧是 {1}',
    rechnenRechts: '`{0}` 对数字进行运算，右侧是 {1}',
    vergleichLinks: '`{0}` 比较数字，左侧是 {1}',
    vergleichRechts: '`{0}` 比较数字，右侧是 {1}',
    keineListe: 'sort_list 用于排序列表，此处是 {0}',
    zuTief: '求值深入了超过 {0} 层——这里是不是有东西在调用自身？',
    schleifeSteht: '计数循环在 {0} 处无法推进——数字太大',
    funktionVerbinden: '函数只能与函数组合，此处是 {0} 和 {1}',
    fehltNach: '{0} 之后缺少 {1}',
    fehltHier: '此处缺少 {0}',
    klammerOffen: '这个 {0} 没有闭合——缺少 {1}',
    ausdruckGehtWeiter: '表达式到此结束，但文本还在继续',
    zuweisungBrauchtNamen: '赋值的对象必须是名称',
    durchgereicht: '{0}',
    bedingungWahrheit: '条件需要真值，此处是 {0}',
    fuerJedenListe: '`for each` 遍历列表，此处是 {0}',
    wahrheitLinks: '`{0}` 需要真值，左侧是 {1}',
    wahrheitRechts: '`{0}` 需要真值，右侧是 {1}',
    bindenFunktion: '`@` 将参数绑定到函数，此处是 {0}',
  },
} as const
