<template>
  <div class="home">
    <UserAccess/>
  </div>
</template>

<script>
// @ is an alias to /src
import UserAccess from '@/components/UserAccess.vue'

export default {
  name: 'Home',
  components: {
    UserAccess
  }
}
</script>
