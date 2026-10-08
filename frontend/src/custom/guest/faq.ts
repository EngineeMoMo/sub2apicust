export const faqGroups = ['入门', '充值与订阅', 'API 接入', '常见排错'] as const

export const faqItems = [
  { id: 'image-generation', group: 'API 接入', question: '本站支持生图吗？如何开通？', answer: '本站支持生图（图像生成），但需要联系管理员开通权限。使用前请先联系管理员，确认权限已开启，再按管理员提供的模型和接入方式配置。' },
  { id: 'choose-protocol', group: 'API 接入', question: 'Claude Code、Codex 和聊天客户端的地址能通用吗？', answer: '不能直接照搬完整路径。Claude Code 使用 Anthropic Messages，Codex 使用 OpenAI Responses，其他工具需核对其兼容协议。接入教程已按命令行、桌面应用和编辑器扩展分别列出步骤，并展示本站动态服务地址。' },
  { id: 'desktop-environment', group: 'API 接入', question: '终端能用，桌面应用为什么还报密钥错误？', answer: '终端中的临时环境变量通常只影响该终端启动的程序。请按桌面端教程配置本机环境或用户级配置，保存后完全退出应用、重新打开并新建会话，再核对提供商和模型选择。' },
  { id: 'verify-connection', group: '常见排错', question: '收到模型回复就代表已接入本站了吗？', answer: '还需在本站用量页核对测试时间和模型。客户端可能仍在使用原来的提供商；若没有对应记录，请先核对服务地址、密钥所属账户和当前选择的提供商，再进行批量任务。' },
  { id: 'leaked-key', group: 'API 接入', question: 'API Key 泄露后应该怎么处理？', answer: '立即在控制台停用或删除泄露密钥，检查近期用量记录，再创建新密钥并更新自己使用的客户端。仅从聊天或截图中删除文字不能让已经泄露的密钥失效。' },
  { id: 'service', group: '入门', question: '这里是什么服务？是聊天软件会员吗？', answer: '这里提供多模型 API 接入和账户管理。API 服务与模型厂商的聊天软件会员不是同一种产品，购买本站套餐不等于获得其他网站的会员权益。具体可用范围以套餐与控制台为准。' },
  { id: 'guest', group: '入门', question: '不注册可以先看看吗？', answer: '可以。产品介绍、在售订阅套餐和本页均可免登录查看。购买、充值、创建 API Key 或查看个人用量时才需要登录。浏览套餐不会创建订单。' },
  { id: 'start', group: '入门', question: '第一次使用应该从哪里开始？', answer: '先了解套餐权益与接入方式；需要使用时注册或登录，按实际需要充值或购买订阅，再到控制台创建 API Key，并在兼容的工具中配置服务地址、密钥和模型。' },
  { id: 'balance', group: '充值与订阅', question: '余额充值和订阅套餐有什么区别？', answer: '余额充值向账户增加可用余额；订阅购买指定套餐的有效期与权益。两者不是同一商品，不能把套餐价格当作到账余额。请以购买页面展示的额度、适用范围和有效期为准。' },
  { id: 'price', group: '充值与订阅', question: '套餐标价就是最终付款金额吗？', answer: '不一定。套餐页展示标价与币种，结账可能涉及币种换算或手续费。请在确认支付前检查结账页的实付金额。套餐或支付方式不可用时，不要重复提交付款。' },
  { id: 'expiry', group: '充值与订阅', question: '在哪里看套餐有效期和用量？', answer: '购买后登录控制台，在“我的订阅”查看订阅信息，在用量页面查看使用记录。不同套餐的有效期、额度和模型范围可能不同，请以你购买的具体套餐为准。' },
  { id: 'paid', group: '充值与订阅', question: '付款后没有看到余额或订阅怎么办？', answer: '先到订单页面核对订单状态，并确认当前登录的是购买时使用的账户。若状态仍异常，请保存订单号和付款凭证，通过站点已提供的客服渠道联系处理；不要提供完整 API Key、密码或验证码。' },
  { id: 'key', group: 'API 接入', question: 'API Key 是什么？可以发给别人吗？', answer: 'API Key 是调用服务时使用的凭证。登录后在控制台创建和管理，只填写到你信任的兼容工具中。不要放进公开仓库、截图或群聊；怀疑泄露时及时撤销相关密钥。' },
  { id: 'endpoint', group: 'API 接入', question: '服务地址和模型名称应该填什么？', answer: '请使用控制台或站点接入文档提供的地址及模型名称，并按照客户端要求填写。不同工具对地址路径的要求可能不同，不要随意重复添加接口路径。可用模型以账户实际权限和套餐范围为准。' },
  { id: 'unauthorized', group: '常见排错', question: '出现认证失败或权限不足怎么办？', answer: '先检查密钥是否完整、是否已被撤销、服务地址是否正确，再核对账户和套餐是否有对应模型权限。分享报错时可以提供时间、错误码和脱敏信息，但不要发送密钥。' },
  { id: 'limit', group: '常见排错', question: '提示额度不足或请求过于频繁怎么办？', answer: '先检查余额、订阅有效期与相关额度，并降低并发或调用频率。不同限制的原因不同，请结合具体错误信息判断；不要仅凭一次报错重复购买套餐。' },
  { id: 'support', group: '常见排错', question: '如何反馈问题？', answer: '准备问题发生时间、使用的客户端、模型名称、错误码及已脱敏的截图，通过站点提供的客服渠道反馈。涉及支付时补充订单号，不要附上密码、验证码或完整密钥。' }
]

export function searchFaq(query: string) {
  const keyword = query.trim().toLocaleLowerCase()
  return faqItems.filter(item => `${item.group} ${item.question} ${item.answer}`.toLocaleLowerCase().includes(keyword))
}
