<script setup lang="ts">
import { SignedIn, SignedOut, SignInButton, UserButton, useAuth, useSession } from '@clerk/vue'

const { session } = useSession()
const { getToken, signOut } = useAuth()
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

      <div v-if="session">
        <p>Session:</p>
        <pre>{{ session.lastActiveOrganizationId }}</pre>
      </div>

      <button @click="signOut">Sign Out</button>
    </SignedIn>
  </header>
</template>