<template>
    <div class="w-full">
        <data-list class="card" ref="listControl" title="Site Journal Type"
            grid-config="/fico/siteentryjournaltype/gridconfig" form-config="/fico/siteentryjournaltype/formconfig"
            grid-read="/fico/siteentryjournaltype/gets" form-read="/fico/siteentryjournaltype/get" grid-mode="grid"
            grid-delete="/fico/siteentryjournaltype/delete" form-keep-label form-insert="/fico/siteentryjournaltype/save"
            form-update="/fico/siteentryjournaltype/save" :grid-fields="['Enable']"
            :form-tabs-edit="['General', 'Configuration']"
            :form-fields="['Actions', 'Previews', 'DefaultOffset', 'Dimension', 'Vendor']" :init-app-mode="data.appMode"
            :init-form-mode="data.formMode" @formNewData="newRecord" @formEditData="openForm" @postSave="saveConfig">
            <template #form_tab_Configuration="{ item }">
                <SiteJournalTypeConfiguration ref="journalConfig" :id="item._id"></SiteJournalTypeConfiguration>
            </template>
            <template #form_input_Actions="{ item }">
                <JournalTypeContext title="Action" v-model="item.Actions"></JournalTypeContext>
            </template>
            <template #form_input_Previews="{ item }">
                <JournalTypeContext title="Previews" v-model="item.Previews"></JournalTypeContext>
            </template>
            <template #form_input_DefaultOffset="{ item }">
                <AccountSelector v-model="item.DefaultOffset" row></AccountSelector>
            </template>
            <template #form_input_Dimension="{ item }">
                <dimension-editor v-model="item.Dimension"></dimension-editor>
            </template>
            <template #form_input_Vendor="{ item }">
                <s-input required :hide-label="false" label="Vendor" v-model="item.Vendor" class="w-full" use-list
                    :disabled="item.VendorGroup == ''" :lookup-url="`/tenant/vendor/find?GroupID=${item.VendorGroup}`"
                    lookup-key="_id" :lookup-labels="['Name']" :lookup-searchs="['_id', 'Name']"></s-input>
            </template>
        </data-list>
    </div>
</template>
    
<script setup>
import { reactive, ref, inject } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput } from "suimjs";

import JournalTypeContext from './widget/JournalTypeContext.vue';
import DimensionEditor from '@/components/common/DimensionEditorVertical.vue';
import AccountSelector from '@/components/common/AccountSelector.vue';
import SiteJournalTypeConfiguration from "./widget/SiteJournalTypeConfiguration.vue";

layoutStore().name = "tenant";

const listControl = ref(null);
const journalConfig = ref(null);
const axios = inject('axios');

const data = reactive({
    appMode: "grid",
    formMode: "edit",
});

function newRecord(record) {
    record._id = "";
    record.Name = "";
    record.Enable = true;
    record.Actions = [];
    record.Previews = [];

    openForm(record)
}

function openForm() {
    util.nextTickN(2, () => {
        listControl.value.setFormFieldAttr("_id", "rules", [
            (v) => {
                let vLen = 0;
                let consistsInvalidChar = false;

                v.split("").forEach((ch) => {
                    vLen++;
                    const validCar =
                        "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(ch) >= 0;
                    if (!validCar) consistsInvalidChar = true;
                    //console.log(ch,vLen,validCar)
                });

                if (vLen < 3 || consistsInvalidChar)
                    return "minimal length is 3 and alphabet only";
                return "";
            },
        ]);
    })
}

function saveConfig() {
    if (journalConfig.value && journalConfig.value.getDataValue()) {
        let dv = journalConfig.value.getDataValue()
        axios.post('/bagong/vendorjournaltypeconfiguration/save', dv).then(
            (r) => { },
            (e) => {
                data.loading = false;
            }
        );
    }
}
</script>