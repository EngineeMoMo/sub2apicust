# 本地验收证据 · 2026-09-30

只对magic-studio本地首版取证，不替代用户视觉确认、真实模型测试或生产验收。

已核实：

- 4179 HTTP200；21项Node／jsdom／HTTP定向通过，见tests.log（首轮18项，修复后新增3项）。
- 浏览器实际修改产品为“白色陶瓷香薰瓶”并复制到浏览器会话剪贴板，正文含该材料；分享仅`http://127.0.0.1:4179/#item=jade-product`。
- 下载工具事件等待超时，但文件实际落在本机Downloads；读取内容确认替换材料、官方来源及“未实测”，证据副本download-proof.txt。没有将超时假报为功能失败或仅凭toast称下载成功。
- 点击收藏后在“我的收藏”看到条目，刷新再进入仍存在；本轮测试收藏随后取消。空关键词结果→清除筛选→恢复12条、大图Esc关闭、二维运镜animationPlayState running后暂停均实操通过。
- 320／390／768／1280／1440／2560宽度未见外层横向溢出。320时innerWidth320、clientWidth305、scrollWidth305（15px原生滚动条），更多筛选right293、卡片right146.5／293，在内容边界内。
- 25个交付位图的PNG Description／WebP XMP与PROVENANCE精确一致，见asset-check.log。

截图口径：

- desktop.png：1440×900深色发现区全页；mobile.png：390×844浅色全页；user-1280.png：1280×720深色全页。
- width-320.png全页导出右边缘失真，以width-320-live.png实际首屏及上述DOM几何为准。width-768.png／width-2560.png是深色全页。
- skills.png全页导出漏掉下部，不用作有效证据；skills-live.png实际首屏＋skills-bottom.png实际底部替代。底部scrollY236、innerHeight900、pageHeight1136，页脚top786.95／bottom851.95，第二行卡片bottom756.95。
- video.png为有效暂停状态首屏；该状态全页导出出现错误布局，DOM实际宽度与普通截图正确，不把失真文件作为构建问题或合格证据。
- plays.png深色玩法全页；mobile-edit.png是手机编辑区实际截图。
- user.png是修正前历史截图，审查及交付使用user-1280.png。

限制：设计扫描器尝试exit127（引擎不可用），没有自动扫描通过结论；无真实目标模型／Skill执行、无付费服务接口请求、无生产发布。内置image_gen用于制作8张原创视觉，独立于产品运行时模型接入。

修复补证：默认发现区切Skill→Skills导航及两条Skill；工作流→两条指引；提示词→发现导航及12条。手机角色替换为“一只纸艺熊猫”后点返回，collection top88、search top184.39、焦点search、材料仍熊猫，截图mobile-edit.png／mobile-return.png。两个dialog通过对应非空getByRole名称定位，Enter打开／Esc关闭后焦点分别回预览按钮和about-button。新版已重采同尺寸截图；快速跨尺寸导出有帧滞后，320及2560随后以稳定实际状态重新覆盖。最后用户窗口恢复1280×720、scrollY0、clientWidth／scrollWidth均1265，图片均加载，preview.png为最终实际视口。
