<template>
    <div>
        Tenant {{  route.params.fid }}
    </div>
</template>

<script setup>
import { layoutStore } from '@/stores/layout';
import { useRoute } from 'vue-router';

layoutStore().name = 'tenant';

const route = useRoute();

</script>