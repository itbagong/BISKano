<template>
    <div class="w-full">
        <data-list 
            class="card"
            ref="listControl" title="Job" grid-config="/admin/job/gridconfig" form-config="/admin/job/formconfig"
            grid-read="/admin/job/gets" form-read="/admin/job/get" grid-mode="list" grid-delete="/admin/job/delete"
            grid-hide-delete grid-hide-new :grid-fields="['Enable']" :init-app-mode="data.appMode"
            :form-default-mode="data.formMode">
            <template #list_item="{ item }">
                <div class="flex gap-2">
                    <div class="w-[70px] font-semibold bg-success text-white p-2" v-if="item.Status=='Done'">{{ item.Status }}</div>
                    <div class="w-[70px] font-semibold bg-primary text-white p-2" v-else-if="item.Status=='Running'">{{ item.Status }}</div>
                    <div class="w-[70px] font-semibold bg-error text-white p-2" v-else-if="item.Status=='Error'">{{ item.Status }}</div>
                    <div class="w-[70px] font-semibold bg-warning text-white p-2" v-else>{{ item.Status }}</div>
                    <div class="flex flex-col gap-1">
                        <div class="font-semibold">{{ item.Name }}</div>
                        <div class="text-[0.8em]">{{ item._id }}, 
                            {{ item.Status=="Running" ? "started " + moment(item.Start).fromNow() : item.Status=="Done" ? "completed " + moment(item.End).fromNow() : "" }}
                            {{ item.Txt }}
                        </div>
                        <div class="text-[0.8em] flex gap-1">
                            <div v-for="tag in item.Tags">{{ tag }}</div>
                        </div>
                    </div>
                </div>
            </template>

            <template #form_footer_2="{ item }">
                <div class="mt-5 flex flex-col gap-2">
                    <div class="bg-primary text-white p-2">Logs</div>

                    <s-list ref="listGrid" class="w-full" :config="data.gridConfig"
                        :read-url="'admin/joblog/gets?JobID=' + item._id" hide-delete-button hide-new-button>
                        <template #item="item">
                            <div class="flex gap-2">
                                <div class="w-[180px]">{{ moment(item.TimeStamp).format("DD-MMM-yyyy hh:mm:ss") }}</div>
                                <div class="w-[60px]">{{ item.LogType }}</div>
                                <div>{{ item.Txt }}</div>
                            </div>
                        </template>
                    </s-list>
                </div>
            </template>
        </data-list>
    </div>
</template>

<script setup>
import { inject, onMounted, reactive, ref } from 'vue';
import { DataList, SList, loadGridConfig } from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import moment from 'moment'

layoutStore().name = "tenant"

const listControl = ref(null)
const axios = inject("axios")

const data = reactive({
    appMode: "grid",
    formMode: "view",
    gridConfig: {},
})

onMounted(() => {
    loadGridConfig(axios, "admin/joblog/gridconfig").then(r => data.gridConfig = r, e => util.showError(e))
})

</script>