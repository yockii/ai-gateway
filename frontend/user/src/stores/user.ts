import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '@/types/models'

export const useUserStore = defineStore('user', () => {
  const profile = ref<User | null>(null)

  const updateProfile = (data: Partial<User>) => {
    if (profile.value) {
      profile.value = { ...profile.value, ...data }
    }
  }

  return { profile, updateProfile }
})
