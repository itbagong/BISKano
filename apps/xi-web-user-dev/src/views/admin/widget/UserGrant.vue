<template>
    <div class>
        <data-list ref="grid" hide-title no-gap grid-hide-search grid-hide-sort grid-hide-select
            grid-config="admin/rolemember/gridconfig" grid-mode="grid" :grid-read="'admin/rolemember/gets?UserID='+user._id"
            grid-delete="admin/rolemember/delete" form-config="admin/rolemember/formconfig"
            form-read="admin/rolemember/get" form-insert="admin/rolemember/insert" form-update="admin/rolemember/update"
            :grid-fields="['Dimension']" :form-fields="['Dimension']" 
            @form-new-data="newRecord"
            @form-edit-data="openForm" 
            @alter-grid-config="alterGridConfig"
            @form-field-change="handleFieldChange">
            <template #grid_Dimension="{ item }">
                {{ item.Dimension ? item.Dimension.map(el => `${el.Kind}=${el.Value}`).join(", ") : "" }}
            </template>

            <template #form_input_Dimension="{ item }">
                <div>
                    <label clas="input_label">Dimension</label>
                    <key-value-editor v-model="item.Dimension" v-if="item.DimensionScope=='Custom'"></key-value-editor>
                </div>
            </template>
        </data-list>
    </div>
</template>

<script setup>
import { inject, onMounted, reactive, ref, nextTick } from 'vue';
import { DataList, util } from 'suimjs';
import KeyValueEditor from "@/components/common/KeyValueEditor.vue";

const props = defineProps({
    user: {type: Object, default: ()=>{}}
})

const axios = inject("axios")

const data = reactive({
    filter: {
        UserID: "",
        RoleID: "",
        FeatureID: ""
    },
    dimNames: [],
    gridConfig: {},
    records: [],
})

const grid = ref(null)

function newRecord(r) {
    r.UserID = props.user._id
    r.Dimension = []
    openForm(r)
}

function openForm(r) {
    if (r.Dimension==null) r.Dimension = [];

    util.nextTickN(2,()=>{
        grid.value.setFormFieldAttr("TenantID","readOnly",r.Scope=="GLOBAL");
        grid.value.setFormFieldAttr("UserID","hide",true)
        grid.value.setFormFieldAttr("UserID", "hide", true)
        grid.value.setFormFieldAttr("Hash", "hide", true)
    })
}

function alterGridConfig(cfg) {
    cfg.fields = cfg.fields.filter(el => el.field != 'UserID' && el.field != 'Hash');
}

function handleFieldChange(name, v1, v2, old, record) {
    switch (name) {
        case "Scope":
            grid.value.setFormFieldAttr("TenantID","readOnly",v1=="GLOBAL");
            if (v1=="GLOBAL") record.TenantID="";
    }
}

onMounted(() => {
})

</script>