<template>
  <div class="w-full">
    <div class="w-full font-semibold bg-error text-white p-2 mb-2"
      v-if="auth.appToken != '' && auth.appData.Status != 'Active'">Please check
      your email to activate your account</div>

    <div class="flex w-full flex-col gap-4" v-if="auth.appToken != ''">
      <div class="card w-full flex flex-col">
        <h1>Hi {{ auth.appData.DisplayName }}</h1>
        <div>Dashboard Bagong BIS Business App</div>
        <!-- <div>Select app you want to working on:</div>
        <div class="flex flex-wrap gap-5">
          <a v-for="app in data.apps"
            class="flex gap-2 items-center p-3 w-[150px] cursor-pointer bg-slate-300 text-black hover:bg-primary hover:text-white"
            :href="app.Address">
            <div v-if="app.IconType == 'MDI'">
              <mdicon :name="app.IconValue" size="28" />
            </div>
            <div>{{ app.Name }}</div>
          </a>
        </div> -->
      </div>
    </div>

    <div v-else>
      <div class="card p-5">
        <h1>Welcome to Bagong BIS Business App</h1>
        Please login first !
      </div>
    </div>
  </div>
</template>

<script setup>
import { layoutStore } from '@/stores/layout';
import { authStore } from '@/stores/auth';
import { reactive, onMounted, inject, watch } from 'vue';

const layout = layoutStore();
layout.change('tenant');

const auth = authStore();
const data = reactive({
  apps: []
})
const axios = inject('axios');

function refreshApps() {
  axios.post('/iam/user/apps').
    then(r => {
      data.apps = r.data.sort((a, b) => {
        if (a.Name < b.Name) return -1;
        if (a.Name > b.Name) return 1;
        return 0;
      });
    })
}

watch(auth, () => {
  if (auth.appToken == '') {
    location.reload();
  }
})

onMounted(() => {
  refreshApps()
})

//const basePath = import.meta.env.VITE_BASE_PATH;
</script>
