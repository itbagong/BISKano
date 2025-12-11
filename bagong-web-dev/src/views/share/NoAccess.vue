<template>
    <div class="card flex flex-col gap-4">
        <h1 class="text-primary">Unauthorized</h1>
        <div>
            Ups !!! either accidentally someone removed your access or 
            you are not allowed to access this resource.
        </div>
        <div>
            You may need to contact system admin or tenant admin if you 
            need to access this resources.
        </div>
        <div class="flex gap-1">
            <a href="/bagong" class="btn_primary p-1 shadow cursor-pointer hover:opacity-[60%]">Back to home</a>
        </div>
    </div>
</template>

<script setup>
import { layoutStore } from "@/stores/layout";
import { ref } from "vue";
import { useRoute } from "vue-router";

// layoutStore().change('clear');

layoutStore().name = "tenant";
const route = useRoute();
const sourcePath = window.location.origin + "/bagong/" + route.query.route;
</script>