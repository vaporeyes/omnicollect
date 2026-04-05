<script lang="ts" setup>
import {reactive} from 'vue'
import {GetActiveModules} from '../../wailsjs/go/main/App'

const data = reactive({
  resultText: "Loading modules...",
})

GetActiveModules().then((modules: any[]) => {
  if (modules.length === 0) {
    data.resultText = "No modules found. Add .json schemas to ~/.omnicollect/modules/"
  } else {
    data.resultText = `Loaded ${modules.length} module(s): ${modules.map((m: any) => m.displayName).join(', ')}`
  }
}).catch((err: any) => {
  data.resultText = `Error: ${err}`
})

</script>

<template>
  <main>
    <h2>OmniCollect</h2>
    <div id="result" class="result">{{ data.resultText }}</div>
  </main>
</template>

<style scoped>
.result {
  height: 20px;
  line-height: 20px;
  margin: 1.5rem auto;
}
</style>
