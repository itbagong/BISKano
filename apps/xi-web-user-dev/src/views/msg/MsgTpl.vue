<template>
    <div>
        <div class="card w-full">
            <div v-if="data.formCfg.setting" v-show="data.appMode == 'grid'" class="pt-3">
                    <s-list class="w-full" read-url="/msg/template/gets" delete-url="/msg/template/delete" 
                        v-if="data.gridCfg.setting" :config="data.gridCfg" form-keep-label
                        @select-data="selectData" @new-data="newData" ref="grid">
                        <template v-slot:item="{item}">
                            <div class="flex flex-col pb-2">
                                <div>{{ item.Name }} ({{ item.LanguageID }}) <span v-if="item.Group!='N/A' && item.Group!=''" class="font-semibold">|&nbsp;{{ item.Group }}</span></div>
                                <div class="text-[0.8em] text-slate-600">{{ item.Namex }}</div>
                                <div class="text-[0.8em] text-slate-600">{{ item.Message && item.Message.length > 130 ? item.Message.substring(0,120) + "..." : item.Message }}</div>
                            </div>
                        </template>
                    </s-list>
            </div>

            <s-form v-model="data.record" :config="data.formCfg" 
                class="pt-2" auto-focus keep-label
                @submitForm="save" @cancelForm="cancelForm"
                v-if="data.appMode == 'form' && data.formCfg.setting" />
        </div>
    </div>
</template>

<script setup>
import { layoutStore } from "@/stores/layout.js";
import { SCard, SForm, SList, loadFormConfig, loadGridConfig, util } from "suimjs";
import { reactive, inject, onMounted, ref, nextTick } from "vue";

const axios = inject("axios")

layoutStore().name = "tenant"

const data = reactive({
    appMode: "grid",
    formMode: "new",
    formCfg: {},
    gridCfg: {},
    content: '',
    record: {
        ChainCode: 0,
        Enable: true
    }
})

const grid = ref(null)

function selectData(dt, op) {
    axios.post("/msg/template/get", [dt._id]).then(r => {
        data.appMode = "form"
        data.formMode = "edit"
        data.record = r.data
    }, e => {

    })
}

function newData(dt) {
    data.appMode = "form"
    data.formMode = "new"
    data.record = {
        LanguageID: "en-us",
        Group: "N/A"
    }
}

function refreshData() {
    nextTick(() => {
        grid.value.refreshData()
    })
}

function cancelForm() {
    data.appMode = "grid"
    refreshData()
}

function save() {
    const saveUrl = data.formMode == "new" ? "/msg/template/insert" : "/msg/template/update"
    axios.post(saveUrl, data.record).then(r => {
        data.record = r
        data.appMode = "grid"
        refreshData()
    }, e => util.showError(e))
}

onMounted(() => {
    loadFormConfig(axios, "/msg/ui/template/formconfig").then(r => data.formCfg = r, e => util.showError(e))
    loadGridConfig(axios, "/msg/ui/template/gridconfig").then(r => data.gridCfg = r, e => util.showError(e))
})

</script>