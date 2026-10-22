import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises } from '@vue/test-utils'
import { usePaymentSecurity } from '@/custom/usePaymentSecurity'
import { StepUpCancelledError } from '@/composables/useStepUp'
import { useAppStore } from '@/stores/app'

const paymentMocks = vi.hoisted(() => ({ refundOrder: vi.fn(), getConfig: vi.fn(), post: vi.fn() }))
vi.mock('@/api/admin/payment', () => ({ adminPaymentAPI: paymentMocks }))
vi.mock('@/api/client', () => ({ apiClient: { post: paymentMocks.post } }))
vi.mock('vue-i18n', async (importOriginal) => ({ ...(await importOriginal<typeof import('vue-i18n')>()), useI18n: () => ({ t: (key: string) => key }) }))

describe('支付敏感操作二次验证', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.resetAllMocks()
  })

  it('已有验证授权时直接执行退款', async () => {
    paymentMocks.refundOrder.mockResolvedValue({ data: { success: true } })
    const security = usePaymentSecurity()
    await security.api.refundOrder(1, { amount: 10, reason: 'test', deduct_balance: true, force: false })
    expect(paymentMocks.refundOrder).toHaveBeenCalledTimes(1)
    expect(security.stepUp.visible.value).toBe(false)
  })

  it('缺少验证时弹窗，验证成功后仅重试原退款一次', async () => {
    paymentMocks.refundOrder.mockRejectedValueOnce({ status: 403, code: 'STEP_UP_REQUIRED' }).mockResolvedValueOnce({ data: { success: true } })
    const security = usePaymentSecurity()
    const payload = { amount: 10, reason: 'test', deduct_balance: true, force: false }
    const result = security.api.refundOrder(7, payload)
    await flushPromises()
    expect(security.stepUp.visible.value).toBe(true)
    expect(paymentMocks.refundOrder).toHaveBeenCalledTimes(1)
    security.stepUp.onVerified()
    await expect(result).resolves.toEqual({ data: { success: true } })
    expect(paymentMocks.refundOrder.mock.calls).toEqual([[7, payload], [7, payload]])
  })

  it('取消验证不退款也不弹业务错误', async () => {
    paymentMocks.refundOrder.mockRejectedValueOnce({ code: 'STEP_UP_REQUIRED' })
    const toast = vi.spyOn(useAppStore(), 'showError')
    const security = usePaymentSecurity()
    const result = security.api.refundOrder(7, { amount: 10, reason: 'test', deduct_balance: true, force: false })
    const rejected = expect(result).rejects.toBeInstanceOf(StepUpCancelledError)
    await flushPromises()
    security.stepUp.onCancel()
    await rejected
    security.showError(new StepUpCancelledError())
    expect(paymentMocks.refundOrder).toHaveBeenCalledTimes(1)
    expect(toast).not.toHaveBeenCalled()
  })

  it.each(['STEP_UP_TOTP_NOT_ENABLED', 'STEP_UP_ADMIN_API_KEY_FORBIDDEN'])('无法输入验证码解决的拒绝直接提示：%s', async (code) => {
    paymentMocks.refundOrder.mockRejectedValueOnce({ code })
    const security = usePaymentSecurity()
    await expect(security.api.refundOrder(7, { amount: 10, reason: 'test', deduct_balance: true, force: false })).rejects.toEqual({ code })
    expect(security.stepUp.visible.value).toBe(false)
    expect(security.errorMessage({ code })).toBe(code === 'STEP_UP_ADMIN_API_KEY_FORBIDDEN' ? 'stepUp.adminApiKeyForbidden' : 'stepUp.notEnabled')
    expect(paymentMocks.refundOrder).toHaveBeenCalledTimes(1)
  })

  it('验证后的业务错误不会无限重试', async () => {
    const failure = { code: 'REFUND_FAILED', message: 'provider unavailable' }
    paymentMocks.refundOrder.mockRejectedValueOnce({ code: 'STEP_UP_REQUIRED' }).mockRejectedValueOnce(failure)
    const security = usePaymentSecurity()
    const result = security.api.refundOrder(7, { amount: 10, reason: 'test', deduct_balance: true, force: false })
    const rejected = expect(result).rejects.toEqual(failure)
    await flushPromises()
    security.stepUp.onVerified()
    await rejected
    expect(paymentMocks.refundOrder).toHaveBeenCalledTimes(2)
  })

  it('人工释放同样弹窗，保持订单编号和确认参数', async () => {
    paymentMocks.post.mockRejectedValueOnce({ reason: 'STEP_UP_REQUIRED' }).mockResolvedValueOnce({ data: {} })
    const security = usePaymentSecurity()
    const payload = { reason: '商户账单已核实', confirmed: true }
    const result = security.api.releaseCollection(8, payload)
    await flushPromises()
    expect(security.stepUp.visible.value).toBe(true)
    security.stepUp.onVerified()
    await result
    expect(paymentMocks.post.mock.calls).toEqual([['/admin/payment/collection/8/release', payload], ['/admin/payment/collection/8/release', payload]])
  })

  it('普通配置读取保持原接口', async () => {
    paymentMocks.getConfig.mockResolvedValue({ data: {} })
    const security = usePaymentSecurity()
    await security.api.getConfig()
    expect(paymentMocks.getConfig).toHaveBeenCalledTimes(1)
    expect(security.stepUp.visible.value).toBe(false)
  })
})
