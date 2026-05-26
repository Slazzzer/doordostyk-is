import { useToastStore } from '../stores/toast.js'

export function useFlash() {
  const toast = useToastStore()

  function showFlash(message, type = 'success', duration) {
    toast.push(message, type, duration)
  }

  return { showFlash }
}
