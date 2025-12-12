<template>
  <div
    class="w-full font-semibold bg-error text-white p-2 mb-2"
    v-if="auth.appToken != '' && auth.appData.Status != 'Active'"
  >
    Please check your email to activate your account
  </div>

  <div class="w-full">
    <div class="flex flex-row gap-4 mb-4">
      <div
        class="p-4 basis-1/4 text-white h-36 bg-gradient-to-r from-blue-800 to-gray-400 rounded-md"
      >
        <div class="relative ...">
          <div class="opacity-60 absolute right-0 top-0 ...">
            <mdicon name="briefcase-clock" size="48" />
          </div>
        </div>
        <h3 class="mb-2 text-lg">Today Attendance</h3>
        <div class="mb-3 flex flex-wrap gap-2">
          <div class="p-1"><mdicon name="login" size="18" /></div>
          <div class="text-[16px]">Clock-In : 08:30</div>
        </div>
        <div class="flex flex-wrap gap-2">
          <div class="p-1"><mdicon name="logout" size="18" /></div>
          <div class="text-[16px]">Clock-Out : 16:30</div>
        </div>
      </div>
      <div
        class="p-4 basis-1/4 text-white h-36 bg-gradient-to-r from-green-800 to-gray-400 rounded-md"
      >
        <div class="relative ...">
          <div class="opacity-60 absolute right-0 top-0 ...">
            <mdicon name="bed-clock" size="48" />
          </div>
        </div>
        <h3 class="mb-2 text-lg">Leave Reminder</h3>
        <div class="mb-2 flex flex-wrap gap-2">
          <div class="text-[28px]">14</div>
        </div>
        <div class="flex flex-wrap gap-2">
          <div class="text-[14px]">You used 7 days of leave right</div>
        </div>
      </div>
      <!-- <div class="basis-1/2 h-36">
        <div class="relative ...">
          <div class="absolute right-0 top-0 h-16 w-16 ...">
            <mdicon name="table-clock" />
          </div>
        </div>
      </div> -->
    </div>

    <div class="flex w-full flex-col gap-4" v-if="auth.appToken != ''">
      <div class="card w-full flex flex-col">
        <h1>Recent Attendance</h1>
        <s-grid ref="grid" config="data.config" hideSearch hideSort hideControl>
        </s-grid>
      </div>
    </div>

    <div v-else>
      <div class="card">
        <h1>Welcome to Kara App</h1>
        Please login first !
      </div>
    </div>
  </div>
</template>

<script setup>
import { layoutStore } from "@/stores/layout";
import { authStore } from "@/stores/auth";
import { reactive, onMounted, inject, ref } from "vue";
import { initCustomFormatter } from "vue";
import { SGrid } from "suimjs";

const layout = layoutStore();
layout.name = "public";
layout.change("tenant");

const auth = authStore();
const grid = ref(null);
const data = reactive({
  apps: [],

  records: [],
  model: ["aa", "bb"],
  filter: {},
  config: {
    fields: [
      {
        field: "_id",
        kind: "text",
        label: "ID",
        halign: "start",
        valign: "start",
        labelField: "",
        length: 0,
        width: "",
        pos: 1000,
        readType: "show",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        filter: "show",
      },
      {
        field: "LoginID",
        kind: "text",
        label: "Login ID",
        halign: "start",
        valign: "start",
        labelField: "",
        length: 0,
        width: "",
        pos: 1000,
        readType: "show",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        filter: "show",
      },
      {
        field: "DisplayName",
        kind: "text",
        label: "Display name",
        halign: "start",
        valign: "start",
        labelField: "",
        length: 0,
        width: "",
        pos: 1000,
        readType: "show",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        filter: "show",
      },
      {
        field: "Email",
        kind: "text",
        label: "Email",
        halign: "start",
        valign: "start",
        labelField: "",
        length: 0,
        width: "",
        pos: 1000,
        readType: "show",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        filter: "show",
      },
      {
        field: "Enable",
        kind: "checkbox",
        label: "Enable",
        halign: "start",
        valign: "start",
        labelField: "",
        length: 0,
        width: "",
        pos: 1000,
        readType: "show",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        filter: "show",
      },
      {
        field: "Status",
        kind: "text",
        label: "Status",
        halign: "start",
        valign: "start",
        labelField: "",
        length: 0,
        width: "",
        pos: 1000,
        readType: "show",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        filter: "show",
      },
      {
        field: "WalletAddress",
        kind: "text",
        label: "Wallet address",
        halign: "start",
        valign: "start",
        labelField: "",
        length: 0,
        width: "",
        pos: 1000,
        readType: "show",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        filter: "show",
      },
      {
        field: "Use2FA",
        kind: "checkbox",
        label: "Use2 f a",
        halign: "start",
        valign: "start",
        labelField: "",
        length: 0,
        width: "",
        pos: 1000,
        readType: "show",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        filter: "show",
      },
    ],
  },
});
const axios = inject("axios");

function refreshApps() {
  axios.post("/iam/user/apps").then((r) => {
    data.apps = r.data.sort((a, b) => {
      if (a.Name < b.Name) return -1;
      if (a.Name > b.Name) return 1;
      return 0;
    });
  });
}

onMounted(() => {
  // refreshApps();
  loadData();
});

function loadData() {
  grid.setRecords = [{ aa: "111", bb: "ss" }];
}

//const basePath = import.meta.env.VITE_BASE_PATH;
</script>
