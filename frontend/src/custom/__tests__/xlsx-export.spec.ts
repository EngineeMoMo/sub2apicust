import { describe, expect, it } from 'vitest'
import * as XLSX from 'xlsx'

describe('真实SheetJS导出兼容性', () => {
  it('分页追加的中文、金额、包号计费和公式样式文本在XLSX读回后保持原值', () => {
    const rows = [
      ['用户', '计费类型', '实扣', '备注'],
      ['测试用户', '包号', 0, '=1+1'],
      ['user@example.test', '订阅', 12.345678, '+SUM(A1:A2)'],
    ]
    const sheet = XLSX.utils.aoa_to_sheet([rows[0]])
    XLSX.utils.sheet_add_aoa(sheet, [rows[1]], { origin: -1 })
    XLSX.utils.sheet_add_aoa(sheet, [rows[2]], { origin: -1 })
    const workbook = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(workbook, sheet, '使用记录')
    const bytes = XLSX.write(workbook, { bookType: 'xlsx', type: 'array' })
    const restored = XLSX.read(bytes, { type: 'array' }).Sheets['使用记录']
    expect(XLSX.utils.sheet_to_json(restored, { header: 1 })).toEqual(rows)
    expect(restored.D2.t).toBe('s')
    expect(restored.D2.f).toBeUndefined()
    expect(restored.C2.v).toBe(0)
  })
})
