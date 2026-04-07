<!-- ABOUTME: Safe Markdown-to-HTML renderer using marked + DOMPurify. -->
<!-- ABOUTME: Wraps output in a .prose container for consistent typography. -->
<script lang="ts" setup>
import {computed} from 'vue'
import {marked} from 'marked'
import DOMPurify from 'dompurify'

const props = defineProps<{
  content: string
}>()

// Configure marked to add target="_blank" on links
const renderer = new marked.Renderer()
renderer.link = function ({href, text}) {
  return `<a href="${href}" target="_blank" rel="noopener noreferrer">${text}</a>`
}

marked.setOptions({renderer})

const renderedHtml = computed(() => {
  if (!props.content) return ''
  const raw = marked.parse(props.content) as string
  return DOMPurify.sanitize(raw, {
    ADD_ATTR: ['target', 'rel'],
  })
})
</script>

<template>
  <div v-if="content" class="prose" v-html="renderedHtml"></div>
</template>
