import test from 'node:test'
import assert from 'node:assert/strict'
import { createPinia, setActivePinia } from 'pinia'
import { useTransactionStore } from '../src/stores/transaction.js'

function createMemoryStorage() {
  const store = new Map()
  return {
    getItem: (key) => (store.has(key) ? store.get(key) : null),
    setItem: (key, value) => store.set(key, String(value)),
    removeItem: (key) => store.delete(key),
    clear: () => store.clear()
  }
}

globalThis.localStorage = createMemoryStorage()

test('useTransactionStore: wizard step validation, cart totals, and localStorage persistence', () => {
  globalThis.localStorage.clear()
  setActivePinia(createPinia())
  const store = useTransactionStore()

  assert.equal(Boolean(store.isStep1Complete), false)
  assert.equal(Boolean(store.isStep2Complete), false)
  assert.equal(Boolean(store.canSubmitTransaction), false)

  store.setCustomerInfo({
    customerName: 'Budi Santoso',
    phoneNumber: '08123456789',
    merchantId: 3,
    merchant: { id: 3, name: 'Toko Sembako Jaya' }
  })

  assert.equal(Boolean(store.isStep1Complete), true)
  assert.equal(Boolean(store.canProceedToStep2), true)

  store.addProduct({ id: 1, name: 'Beras 5kg', price: 'Rp 75.000', quantity: 2 })
  store.addProduct({ id: 1, name: 'Beras 5kg', price: 'Rp 75.000', quantity: 1 })
  store.addProduct({ id: 2, name: 'Gula 1kg', price: 'Rp 17.500', quantity: 2 })

  assert.equal(store.assignedProducts.length, 2)
  assert.equal(store.totalItems, 5)
  assert.equal(store.totalAmount, 260000)
  assert.equal(Boolean(store.canSubmitTransaction), true)

  store.updateProductQuantity(2, 1)
  assert.equal(store.totalItems, 4)
  assert.equal(store.totalAmount, 242500)

  store.removeProduct(2)
  assert.equal(store.assignedProducts.length, 1)
  assert.equal(store.totalAmount, 225000)

  const summary = store.getTransactionSummary()
  assert.equal(summary.totalItems, 3)
  assert.equal(summary.totalAmount, 225000)
  assert.equal(summary.merchantName, 'Toko Sembako Jaya')

  store.nextStep()
  assert.equal(store.currentStep, 2)
  store.previousStep()
  assert.equal(store.currentStep, 1)

  // Verify persistence across fresh Pinia instance
  setActivePinia(createPinia())
  const reloaded = useTransactionStore()
  reloaded.loadFromLocalStorage()
  assert.equal(reloaded.customerInfo.customerName, 'Budi Santoso')
  assert.equal(reloaded.totalAmount, 225000)

  reloaded.clearTransaction()
  assert.equal(reloaded.assignedProducts.length, 0)
  assert.equal(reloaded.currentStep, 1)
  assert.equal(globalThis.localStorage.getItem('transactionWizard'), null)
})
