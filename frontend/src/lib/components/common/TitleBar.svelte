<script lang="ts">
  import Icon from '@iconify/svelte'
</script>

<!-- Drag handle for the frameless window. No window buttons: minimise,
     maximise and close are left to the window manager. -->
<header class="h-10 flex items-center bg-muted/50 border-b border-border select-none shrink-0">
  <div class="flex-1 flex items-center gap-2 px-3 h-full" style="--wails-draggable: drag">
    <Icon icon="mdi:email-fast-outline" class="w-5 h-5 text-primary" />
    <span class="text-sm font-medium text-foreground">Aerion</span>
  </div>
</header>
