<template>
  <RouterView v-slot="{Component, route}">
    <KeepAlive>
      <component v-if="route.meta?.keepAlive" :is="Component" />
    </KeepAlive>
    <component v-if="!route.meta?.keepAlive" :is="Component" />
  </RouterView>
</template>

