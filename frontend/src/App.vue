<script setup lang="ts">
import { SignedIn, SignedOut, SignInButton, UserButton, useAuth } from '@clerk/vue'

const { getToken } = useAuth()
</script>

<template>
  <header>
    <SignedOut>
      <SignInButton />
    </SignedOut>
    <SignedIn>
      <UserButton />

      <button @click="async () => {
          const token = await getToken()
          console.log('Token:', token)
      }">Get Token</button>
    </SignedIn>
  </header>
</template>