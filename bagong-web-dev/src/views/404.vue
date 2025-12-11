<template> 
    <div
      class=" p-4 flex flex-col gap-4 min-h-[400px] justify-center items-center"
    >
      <h1 class="text-6xl font-extrabold">404</h1>
      <div class="text-3xl font-bold">Page not found</div>
    </div> 
</template>
<script setup>
import { layoutStore } from "@/stores/layout";
import { useRoute } from "vue-router";

layoutStore().name = "tenant";

const route = useRoute();
</script>