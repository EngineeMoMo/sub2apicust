# 工坊图片解码测试素材

`studio-valid.webp`是本仓`product-samples/magic-studio/assets/red-panda-moss-thumb.webp`的逐字节副本（SHA256及实际生图提示词见该目录PROVENANCE-CURATED-20261001.json），520×347，39690字节。这张无参考图的AI生成素材用于确认升级后的WebP解码与投稿校验仍接受正常图片，不是新的作品或第三方授权素材。

截断WebP以及最小／超限PNG尺寸案例由custom_studio_cleanup_test.go构造，不导入外部攻击文件，不分配巨幅位图。测试拒绝后检查没有记录、媒体或临时文件残留；另验证保存重命名失败保留原目标并清理临时记录。
