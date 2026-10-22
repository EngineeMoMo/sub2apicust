import { adminPaymentAPI } from '@/api/admin/payment'
import { apiClient } from '@/api/client'
import { useStepUp, isStepUpBlocked, isStepUpCancelled, stepUpBlockReason, type StepUpController } from '@/composables/useStepUp'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { useI18n } from 'vue-i18n'

export function usePaymentSecurity(controller?: StepUpController) {
  const stepUp = controller ?? useStepUp()
  const appStore = useAppStore()
  const { t } = useI18n()

  function protect<Arguments extends unknown[], Result>(operation: (...args: Arguments) => Promise<Result>) {
    return (...args: Arguments) => stepUp.run(() => operation(...args))
  }

  const api = {
    ...adminPaymentAPI,
    updateConfig: protect(adminPaymentAPI.updateConfig),
    cancelOrder: protect(adminPaymentAPI.cancelOrder),
    retryRecharge: protect(adminPaymentAPI.retryRecharge),
    refundOrder: protect(adminPaymentAPI.refundOrder),
    queryRefund: protect(adminPaymentAPI.queryRefund),
    createPlan: protect(adminPaymentAPI.createPlan),
    updatePlan: protect(adminPaymentAPI.updatePlan),
    deletePlan: protect(adminPaymentAPI.deletePlan),
    createProvider: protect(adminPaymentAPI.createProvider),
    updateProvider: protect(adminPaymentAPI.updateProvider),
    deleteProvider: protect(adminPaymentAPI.deleteProvider),
    recordCollection: protect((data: { amount: number; reference: string }) => apiClient.post('/admin/payment/collection/manual', data)),
    closeCollection: protect((orderID: number) => apiClient.post(`/admin/payment/collection/${orderID}/close`)),
    recheckCollection: protect(() => apiClient.post<{ released: number }>('/admin/payment/collection/recheck')),
    releaseCollection: protect((orderID: number, data: { reason: string; confirmed: boolean }) => apiClient.post(`/admin/payment/collection/${orderID}/release`, data)),
  }

  function errorMessage(error: unknown, fallback = t('common.error')): string {
    if (isStepUpCancelled(error)) return ''
    if (isStepUpBlocked(error)) {
      return t(stepUpBlockReason(error) === 'STEP_UP_ADMIN_API_KEY_FORBIDDEN' ? 'stepUp.adminApiKeyForbidden' : 'stepUp.notEnabled')
    }
    return extractI18nErrorMessage(error, t, 'payment.errors', fallback)
  }

  function showError(error: unknown, fallback?: string) {
    const message = errorMessage(error, fallback)
    if (message) appStore.showError(message)
  }

  return { api, stepUp, errorMessage, showError }
}
