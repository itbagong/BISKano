<template>
    <div class="w-full">
        <!-- Tenant {{  route.params.fid }} -->
        <h2>Tenant Information</h2>
        <div class="w-full card mt-2 mb-4">
            <!-- <s-form v-if="data.mode=='register'" ref="TenantForm"
                v-model="data.record" :config="data.config" keep-label
                :buttons-on-top="false"
            ></s-form> -->

            <!-- {{ data.record }} -->
            <div class="section grow mb-2">
                <div class="flex flex-col gap-2">
                    <div class="w-full items-start grid gridCol1 bottom-4">
                        <h3>Nama Tenant</h3>
                        <Transition name="fade" mode="out-in">
                            <label v-if="!isLoading" class="px-2 py-1">{{ data.record.Name }}</label>
                            <label v-else
                                class="bg-gray-500 blur-sm w-80 px-2 py-1 text-white v-leave-active bg-gradient-to-r from-gray-600 to-gray-300">Loading
                                data</label>
                        </Transition>
                    </div>

                    <div class="w-full items-start grid gridCol1">
                        <h3>Kode Tenant</h3>
                        <Transition name="fade" mode="out-in">
                            <label v-if="!isLoading" class="px-2 py-1">{{ data.record.FID }}</label>
                            <label v-else
                                class="bg-gray-500 blur-sm w-80 px-2 py-1 text-white v-leave-active bg-gradient-to-r from-gray-600 to-gray-300">Loading
                                data</label>
                        </Transition>
                    </div>

                    <div class="w-full items-start">
                        <h3>Owner Name</h3>
                        <Transition name="fade" mode="out-in">
                            <label v-if="!isLoading" class="px-2 py-1">{{ data.record.OwnerName }}</label>
                            <label v-else
                                class="bg-gray-500 blur-sm w-40 px-2 py-1 text-white v-leave-active bg-gradient-to-r from-gray-600 to-gray-300 mr-2 inline-block">Loading
                                data</label>
                        </Transition>
                        <Transition name="fade" mode="out-in">
                            <label v-if="!isLoading" class="py-1">[{{ data.record.OwnerEmail }}]</label>
                            <label v-else
                                class="bg-gray-500 blur-sm w-40 px-2 py-1 text-white v-leave-active bg-gradient-to-r from-gray-600 to-gray-300 inline-block">Loading
                                data</label>
                        </Transition>
                    </div>
                </div>
            </div>
        </div>

        <h2>Tenant Information</h2>
        <div class=" w-full card mt-2 mb-4" v-if="route.params.fid != ''">
            <div class="flex flex-col">
                <div class="flex flex-wrap gap-5">
                    <div v-if="isLoading" ref="content">Loading data....</div>
                    <a v-for="app in data.apps"
                        class="flex gap-2 items-center p-3 w-[150px] cursor-pointer rounded-md bg-slate-400 text-black hover:bg-primary hover:text-white"
                        :href="app.Address" v-else>
                        <div v-if="app.IconType == 'MDI'">
                            <mdicon :name="app.IconValue" size="28" />
                        </div>
                        <div>{{ app.Name }}</div>
                    </a>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { layoutStore } from "@/stores/layout";
import { useRoute } from "vue-router";
import {
    DataList,
    util,
    SForm,
    createFormConfig,
    formInput,
    loadFormConfig,
} from "suimjs";
import { onMounted, reactive, ref, inject } from "vue";

layoutStore().name = "tenant";

// const TenantForm = ref(null);
const isLoading = ref(true);
const axios = inject("axios");

const route = useRoute();
const data = reactive({
    appMode: "grid",
    formMode: "edit",
    mode: "register",
    config: {},
    record: {},
    apps: [],
});

onMounted(() => {
    // loadFormConfig(axios, "/admin/tenant/formconfig").
    //     then(r => {
    //         data.config = r
    //     }, e => util.showError(e)

    // );
    showInfo();
    showApps();
});

function showInfo() {
    axios.post("/admin/tenant/get", [route.params.fid]).then(
        (d) => {
            data.record = d.data;
            showOwner();
        },
        (x) => util.showError(x)
    );
}

function showOwner() {
    axios.post("/admin/user/get", [data.record.OwnerID]).then(
        (o) => {
            data.record.OwnerName = o.data.DisplayName || '-';
            data.record.OwnerEmail = o.data.Email || '-';
            isLoading.value = false;
        },
        (x) => util.showError(x)
    );
}

function showApps() {
    axios.post("/iam/user/apps").then((r) => {
        data.apps = r.data.sort((a, b) => {
            if (a.Name < b.Name) return -1;
            if (a.Name > b.Name) return 1;
            return 0;
        });
    });
}
</script>

<style>
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.2s ease-in;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>