"""导出压缩预览并记录来源；不改变生成图的画面。"""
from pathlib import Path
from PIL import Image, PngImagePlugin
from xml.sax.saxutils import escape
from io import BytesIO
import json

assets=Path(__file__).parent/'assets'
manifest=json.loads((assets/'PROVENANCE.json').read_text(encoding='utf-8'))
for item in manifest['assets']:
    source=assets/item['file']
    with Image.open(BytesIO(source.read_bytes())) as image:
        meta=PngImagePlugin.PngInfo()
        for key,value in image.info.items():
            if isinstance(value,str):
                meta.add_text(key,value)
        meta.add_itxt('Description',item['prompt'])
        image.save(source,pnginfo=meta)
        xmp=('<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/" dc:description="'+escape(item['prompt'],{'"':'&quot;'})+'"/></rdf:RDF></x:xmpmeta>').encode('utf-8')
        for suffix,size in [('',1440),('-thumb',640)]:
            preview=image.convert('RGB')
            preview.thumbnail((size,size),Image.Resampling.LANCZOS)
            preview.save(assets/f"{item['id']}{suffix}.webp",quality=88,method=6,xmp=xmp)
with Image.open(assets/manifest['brand']['file']) as logo:
    meta=PngImagePlugin.PngInfo()
    meta.add_itxt('Description',manifest['brand']['origin'])
    logo.save(assets/manifest['brand']['file'],pnginfo=meta)
print(f"已导出{len(manifest['assets'])}组压缩预览并保留提示词／来源。")
