<template>
    <div class=" w-full">
        <data-list class="card" ref="grid" title="Access Grant"
            grid-config="admin/rolemember/gridconfig" grid-mode="grid" grid-read="admin/rolemember/gets"
            grid-delete="admin/rolemember/delete" form-config="admin/rolemember/formconfig"
            form-read="admin/rolemember/get" form-insert="admin/rolemember/insert" form-update="admin/rolemember/update"
            :grid-fields="['Dimension']" :form-fields="['Dimension']" 
            @form-new-data="newRecord"
            @form-edit-data="openForm" 
            @alter-grid-config="alterGridConfig"
            @form-field-change="handleFieldChange">
            <template #grid_Dimension="{ item }">
                <dimension-text :dimension="item.Dimension"></dimension-text>
            </template>

            <template #form_input_Dimension="{ item }">
                <dimension-editor v-model="item.Dimension.Items" :dim-names="data.dimNames"></dimension-editor>
            </template>
        </data-list>
    </div>
</template>

<script setup>
import { inject, onMounted, reactive, ref, nextTick } from 'vue';
import { DataList, util } from 'suimjs';
import DimensionText from '@/components/common/DimensionText.vue';
import DimensionEditor from '@/components/common/DimensionEditor.vue';
import { layoutStore } from "@/stores/layout.js";
import { authStore } from '@/stores/auth';

authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})

layoutStore().name = "tenant"

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
    r.Dimension = {
        Items: [],
        Hash: ""
    }
    openForm(r)
}

function openForm(r) {
    util.nextTickN(2,()=>{
        grid.value.setFormFieldAttr("TenantID","lookupUrl",`/admin/tenantuser/find?UserID=${r.UserID}`)
        grid.value.setFormFieldAttr("TenantID","lookupKey","TenantID")
        grid.value.setFormFieldAttr("TenantID","lookupLabels",["TenantID","TenantName"])
        getDimensions(r)
       //grid.value.setFormFieldAttr("Dimension", "hide", true)
    })
}

function alterGridConfig(cfg) {
    //cfg.fields = cfg.fields.filter(el => el.field != 'Dimension')
    //console.log(cfg)
}

function handleFieldChange (name, v1, v2, old, record) {
    if (name=="TenantID") {
        getDimensions(record)
        return
    }

    if (name=="UserID") {
        grid.value.setFormFieldAttr("TenantID","lookupUrl",`/admin/tenantuser/find?UserID=${record.UserID}`)
        grid.value.setFormFieldAttr("TenantID","lookupKey","TenantID")
        grid.value.setFormFieldAttr("TenantID","lookupLabels",["TenantID", "TenantName"])
        return
        //getDimensions(record)
    }
}

function getDimensions (r) {
    if (r.TenantID==undefined || r.TenantID==null || r.TenantID=="") {
        data.dimNames = []
        return
    }

    axios.post("/admin/tenant/get",[r.TenantID]).then(co => {
        data.dimNames = co.data.Dimensions
    }, e => util.showError(e))
}

</script>