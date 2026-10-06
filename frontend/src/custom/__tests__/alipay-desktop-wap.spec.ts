import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import AlipayDesktopWapSetting from '../components/AlipayDesktopWapSetting.vue'
import zhSettings from '@/i18n/locales/zh/admin/settings'
import enSettings from '@/i18n/locales/en/admin/settings'
import { decidePaymentLaunch, readPaymentRecoverySnapshot } from '@/components/payment/paymentFlow'

describe('电脑端支付宝 WAP 扫码', () => {
  it.each(['zh', 'en'])('设置有说明、可访问名称和显式开关事件（%s）', async (locale) => {
    const wrapper = mount(AlipayDesktopWapSetting, {
      props: { modelValue: false },
      global: { plugins: [createI18n({ legacy: false, locale, messages: {
        zh: { admin: { settings: { payment: {
          alipayDesktopWapQRCode: () => zhSettings.settings.payment.alipayDesktopWapQRCode,
          alipayDesktopWapQRCodeHint: () => zhSettings.settings.payment.alipayDesktopWapQRCodeHint,
        } } } },
        en: { admin: { settings: { payment: {
          alipayDesktopWapQRCode: () => enSettings.settings.payment.alipayDesktopWapQRCode,
          alipayDesktopWapQRCodeHint: () => enSettings.settings.payment.alipayDesktopWapQRCodeHint,
        } } } },
      } })] },
    })
    const button = wrapper.get('[role="switch"]')
    expect(button.attributes('aria-checked')).toBe('false')
    expect(button.attributes('aria-labelledby')).toBe('alipay-desktop-wap-label')
    expect(button.attributes('aria-describedby')).toBe('alipay-desktop-wap-hint')
    expect(wrapper.text()).not.toContain('admin.settings.payment.')
    await button.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[true]])
    await wrapper.setProps({ modelValue: true })
    expect(button.attributes('aria-checked')).toBe('true')
    wrapper.unmount()
  })

  it.each(['balance', 'subscription'] as const)('电脑端使用短二维码，刷新可恢复原订单（%s）', (orderType) => {
    const qr = 'https://merchant.example/api/v1/payment/public/alipay/wap/101?token=test'
    const result = { order_id: 101, amount: 10, pay_amount: 10, fee_rate: 0, expires_at: '2099-01-01T00:30:00Z',
      qr_code: qr, pay_url: 'https://openapi.alipay.com/gateway.do?method=alipay.trade.wap.pay', payment_mode: 'qrcode' }
    const launch = decidePaymentLaunch(result, { visibleMethod: 'alipay', orderType, isMobile: false })
    expect(launch.kind).toBe('qr_waiting')
    expect(launch.paymentState.qrCode).toBe(qr)
    const restored = readPaymentRecoverySnapshot(JSON.stringify(launch.recovery))!
    expect(restored.orderId).toBe(101)
    expect(restored.qrCode).toBe(qr)
    expect(restored.paymentMode).toBe('qrcode')
  })

  it('手机原 WAP 响应继续直接跳转，不要求扫码', () => {
    const launch = decidePaymentLaunch({ order_id: 101, amount: 10, pay_amount: 10, fee_rate: 0,
      expires_at: '2099-01-01T00:30:00Z', pay_url: 'https://openapi.alipay.com/gateway.do?method=alipay.trade.wap.pay', payment_mode: 'redirect' },
    { visibleMethod: 'alipay', orderType: 'balance', isMobile: true })
    expect(launch.kind).toBe('redirect_waiting')
    expect(launch.paymentState.qrCode).toBe('')
  })
})
