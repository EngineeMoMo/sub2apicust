"""只读核对每个交付位图的精确生成提示词或既有来源。"""
from pathlib import Path
from PIL import Image
from xml.etree import ElementTree
import json

assets = Path(__file__).parent / 'assets'
manifest = json.loads((assets / 'PROVENANCE.json').read_text(encoding='utf-8'))
checked = []
for item in manifest['assets']:
    with Image.open(assets / item['file']) as original:
        assert original.info['Description'] == item['prompt'], item['file']
    checked.append(item['file'])
    for suffix in ['', '-thumb']:
        name = item['id'] + suffix + '.webp'
        with Image.open(assets / name) as preview:
            tree = ElementTree.fromstring(preview.info['xmp'])
            description = tree.find('.//{http://www.w3.org/1999/02/22-rdf-syntax-ns#}Description')
            assert description.attrib['{http://purl.org/dc/elements/1.1/}description'] == item['prompt'], name
        checked.append(name)
with Image.open(assets / manifest['brand']['file']) as brand:
    assert brand.info['Description'] == manifest['brand']['origin']
checked.append(manifest['brand']['file'])
actual = {p.name for p in assets.iterdir() if p.suffix in {'.png', '.webp'}}
assert actual == set(checked), f'Unregistered rasters: {actual - set(checked)}'
print(json.dumps({'checked': len(checked), 'unregistered': 0, 'status': 'passed'}, ensure_ascii=False))
